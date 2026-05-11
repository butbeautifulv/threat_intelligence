package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"log/slog"

	"vuln/internal/domain"
	"vuln/internal/repository"
)

const nvdBaseURL = "https://services.nvd.nist.gov/rest/json/cves/2.0"

type ScraperUsecase struct {
	repo   repository.VulnerabilityRepository
	logger *slog.Logger
	apiKey string
}

func NewScraperUsecase(repo repository.VulnerabilityRepository, logger *slog.Logger, apiKey string) *ScraperUsecase {
	return &ScraperUsecase{repo: repo, logger: logger, apiKey: apiKey}
}

func (u *ScraperUsecase) downloadNVDPage(ctx context.Context, startIndex, resultsPerPage int) ([]byte, error) {
	uu, _ := url.Parse(nvdBaseURL)
	q := uu.Query()
	q.Set("startIndex", strconv.Itoa(startIndex))
	q.Set("resultsPerPage", strconv.Itoa(resultsPerPage))
	uu.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uu.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "threat_intelligence/1.0")
	if u.apiKey != "" {
		req.Header.Set("apiKey", u.apiKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	backoff := 1 * time.Second
	for attempt := 0; attempt < 6; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			// Read and discard to reuse connections.
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			_ = resp.Body.Close()
			return nil, fmt.Errorf("nvd http %d: %s", resp.StatusCode, string(b))
		}
		b, rerr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return b, rerr
	}
	return nil, fmt.Errorf("nvd fetch failed after retries")
}

// parseNVD extracts a minimal set of fields into domain.Vulnerability.
func (u *ScraperUsecase) parseNVD(data []byte) ([]domain.Vulnerability, int, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, 0, err
	}

	items, _ := raw["vulnerabilities"].([]any)
	total := 0
	if tr, ok := raw["totalResults"].(float64); ok {
		total = int(tr)
	}
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

		// metrics -> CVSS
		var cvss *domain.CVSS
		if metrics, ok := cveBlock["metrics"].(map[string]any); ok {
			// Prefer v3.1 then v3.0
			cvss = pickCVSS(metrics, "cvssMetricV31")
			if cvss == nil {
				cvss = pickCVSS(metrics, "cvssMetricV30")
			}
		}

		v := domain.Vulnerability{
			ID:      id,
			CVE:     id,
			Summary: desc,
			CWE:     cwes,
			CPEs:    cpes,
			CVSS:    cvss,
		}

		out = append(out, v)
	}

	return out, total, nil
}

func (u *ScraperUsecase) ScrapeNVD(ctx context.Context) error {
	u.logger.Info("starting NVD scraping")

	const pageSize = 2000
	start := 0
	total := -1
	count := 0
	maxPages := 0
	if v := os.Getenv("NVD_MAX_PAGES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxPages = n
		}
	}
	pages := 0

	for total < 0 || start < total {
		data, err := u.downloadNVDPage(ctx, start, pageSize)
		if err != nil {
			return err
		}
		vulns, tr, err := u.parseNVD(data)
		if err != nil {
			return err
		}
		if total < 0 {
			total = tr
		}

		for _, v := range vulns {
			if err := u.repo.Upsert(ctx, &v); err != nil {
				u.logger.Error("failed to upsert vuln", slog.String("cve", v.CVE), slog.String("error", err.Error()))
				return err
			}
			count++
		}

		u.logger.Info("nvd page ingested", slog.Int("startIndex", start), slog.Int("pageCount", len(vulns)), slog.Int("totalResults", total))
		if len(vulns) == 0 {
			break
		}
		pages++
		if maxPages > 0 && pages >= maxPages {
			u.logger.Info("nvd stopping early (NVD_MAX_PAGES)", slog.Int("maxPages", maxPages))
			break
		}
		start += pageSize
	}

	u.logger.Info("finished NVD scraping", slog.Int("count", count))
	return nil
}

func (u *ScraperUsecase) Run(ctx context.Context) error {
	return u.ScrapeNVD(ctx)
}

func pickCVSS(metrics map[string]any, key string) *domain.CVSS {
	ms, ok := metrics[key].([]any)
	if !ok || len(ms) == 0 {
		return nil
	}
	// Take first metric block
	m0, ok := ms[0].(map[string]any)
	if !ok {
		return nil
	}
	cv, _ := m0["cvssData"].(map[string]any)
	if cv == nil {
		return nil
	}
	ver, _ := cv["version"].(string)
	vec, _ := cv["vectorString"].(string)
	base, _ := cv["baseScore"].(float64)
	if ver == "" && vec == "" && base == 0 {
		return nil
	}
	return &domain.CVSS{Version: ver, Base: base, Vector: vec}
}
