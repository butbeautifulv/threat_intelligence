package usecase

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"log/slog"

	"vuln/internal/domain"
	"vuln/internal/repository"
)

const nvdURL = "https://services.nvd.nist.gov/rest/json/cves/2.0?resultsPerPage=2000"

type ScraperUsecase struct {
	repo   repository.VulnerabilityRepository
	logger *slog.Logger
}

func NewScraperUsecase(repo repository.VulnerabilityRepository, logger *slog.Logger) *ScraperUsecase {
	return &ScraperUsecase{repo: repo, logger: logger}
}

func (u *ScraperUsecase) downloadNVDFeed(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nvdURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// parseNVD extracts a minimal set of fields into domain.Vulnerability.
func (u *ScraperUsecase) parseNVD(data []byte) ([]domain.Vulnerability, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	items, _ := raw["vulnerabilities"].([]any)
	var out []domain.Vulnerability

	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		cveBlock, _ := m["cve"].(map[string]any)
		id, _ := cveBlock["id"].(string)

		// description
		var desc string
		if descs, ok := cveBlock["descriptions"].([]any); ok {
			for _, d := range descs {
				if dm, ok := d.(map[string]any); ok {
					if lang, _ := dm["lang"].(string); lang == "en" {
						if v, _ := dm["value"].(string); v != "" {
							desc = v
							break
						}
					}
				}
			}
			if desc == "" && len(descs) > 0 {
				if dm, ok := descs[0].(map[string]any); ok {
					desc, _ = dm["value"].(string)
				}
			}
		}

		// weaknesses -> CWE
		var cwes []string
		if weaknesses, ok := cveBlock["weaknesses"].([]any); ok {
			for _, w := range weaknesses {
				if wm, ok := w.(map[string]any); ok {
					if descs, ok := wm["description"].([]any); ok {
						for _, dd := range descs {
							if dm, ok := dd.(map[string]any); ok {
								if v, _ := dm["value"].(string); v != "" {
									cwes = append(cwes, v)
								}
							}
						}
					}
				}
			}
		}

		// configurations -> cpes
		var cpes []domain.CPE
		if confs, ok := cveBlock["configurations"].([]any); ok {
			for _, c := range confs {
				if cm, ok := c.(map[string]any); ok {
					if nodes, ok := cm["nodes"].([]any); ok {
						for _, n := range nodes {
							if nm, ok := n.(map[string]any); ok {
								if matches, ok := nm["cpeMatch"].([]any); ok {
									for _, mm := range matches {
										if mmm, ok := mm.(map[string]any); ok {
											if uri, _ := mmm["criteria"].(string); uri != "" {
												cpes = append(cpes, domain.CPE{URI: uri})
											} else if uri2, _ := mmm["cpe23Uri"].(string); uri2 != "" {
												cpes = append(cpes, domain.CPE{URI: uri2})
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}

		v := domain.Vulnerability{
			ID:      id,
			CVE:     id,
			Summary: desc,
			CWE:     cwes,
			CPEs:    cpes,
		}

		out = append(out, v)
	}

	return out, nil
}

func (u *ScraperUsecase) ScrapeNVD(ctx context.Context) error {
	u.logger.Info("starting NVD scraping")

	data, err := u.downloadNVDFeed(ctx)
	if err != nil {
		return err
	}

	vulns, err := u.parseNVD(data)
	if err != nil {
		return err
	}

	for _, v := range vulns {
		if err := u.repo.Upsert(ctx, &v); err != nil {
			u.logger.Error("failed to upsert vuln", slog.String("cve", v.CVE), slog.String("error", err.Error()))
			return err
		}
	}

	u.logger.Info("finished NVD scraping", slog.Int("count", len(vulns)))
	return nil
}

func (u *ScraperUsecase) Run(ctx context.Context) error {
	return u.ScrapeNVD(ctx)
}
