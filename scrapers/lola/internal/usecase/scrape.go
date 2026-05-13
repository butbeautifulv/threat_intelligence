package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"lola/internal/domain"
	"lola/internal/proxypool"
	"lola/internal/repository"
)

const (
	lolbasOwner = "LOLBAS-Project"
	lolbasRepo  = "LOLBAS"
	lolbasPath  = "yml"

	gtfobinsOwner = "GTFOBins"
	gtfobinsRepo  = "GTFOBins.github.io"
	gtfobinsPath  = "_gtfobins"
)

type ScraperUsecase struct {
	repo   repository.LolaRepository
	logger *slog.Logger
	http   *http.Client
	cache  string
}

func NewScraperUsecase(repo repository.LolaRepository, logger *slog.Logger, cacheDir string) *ScraperUsecase {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSHandshakeTimeout = 30 * time.Second
	var rt http.RoundTripper = tr
	if env := strings.TrimSpace(os.Getenv("LOLA_PROXY_URLS")); env != "" {
		p, err := proxypool.New(proxypool.SplitEnvList(env), 2*time.Minute)
		if err == nil {
			only := strings.EqualFold(strings.TrimSpace(os.Getenv("LOLA_PROXY_MODE")), "only")
			rt = proxypool.NewTransport(tr, p, only)
			logger.Info("lola proxy pool enabled", slog.Int("count", len(proxypool.SplitEnvList(env))))
		} else {
			logger.Warn("lola proxy pool invalid; running direct", slog.String("err", err.Error()))
		}
	}
	return &ScraperUsecase{
		repo:   repo,
		logger: logger,
		http:   &http.Client{Timeout: 120 * time.Second, Transport: rt},
		cache:  cacheDir,
	}
}

type ghContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

func (u *ScraperUsecase) Run(ctx context.Context) error {
	if err := u.repo.EnsureSchema(ctx); err != nil {
		return err
	}
	if err := u.IngestLOLBAS(ctx); err != nil {
		return err
	}
	if err := u.IngestGTFOBins(ctx); err != nil {
		return err
	}
	if err := u.IngestLOFTS(ctx); err != nil {
		return err
	}
	if err := u.IngestMITREEnterprise(ctx); err != nil {
		return err
	}
	return nil
}

func (u *ScraperUsecase) IngestLOLBAS(ctx context.Context) error {
	u.logger.Info("ingesting LOLBAS definitions")
	cats, err := u.githubListDir(ctx, lolbasOwner, lolbasRepo, lolbasPath)
	if err != nil {
		return err
	}
	n := 0
	for _, cat := range cats {
		if cat.Type != "dir" {
			continue
		}
		items, err := u.githubListDir(ctx, lolbasOwner, lolbasRepo, cat.Path)
		if err != nil {
			u.logger.Warn("lolbas category list failed", slog.String("path", cat.Path), slog.String("err", err.Error()))
			continue
		}
		for _, it := range items {
			if it.Type != "file" {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(it.Name), ".yml") && !strings.HasSuffix(strings.ToLower(it.Name), ".yaml") {
				continue
			}
			cacheName := strings.ReplaceAll(it.Path, "/", "__")
			raw, err := u.fetchBytes(ctx, it.DownloadURL, filepath.Join(u.cache, "lolbas", cacheName))
			if err != nil {
				u.logger.Warn("lolbas file fetch failed", slog.String("path", it.Path), slog.String("err", err.Error()))
				continue
			}
			a, err := parseLOLBASYAML(raw)
			if err != nil {
				u.logger.Warn("lolbas parse failed", slog.String("path", it.Path), slog.String("err", err.Error()))
				continue
			}
			// Capture category from folder name when not present.
			if a.Category == "" {
				a.Category = strings.TrimPrefix(cat.Path, lolbasPath+"/")
			}
			if err := u.repo.UpsertArtifact(ctx, "lolbas", a); err != nil {
				return err
			}
			n++
		}
	}
	u.logger.Info("LOLBAS ingest done", slog.Int("count", n))
	return nil
}

func (u *ScraperUsecase) IngestGTFOBins(ctx context.Context) error {
	u.logger.Info("ingesting GTFOBins pages")
	items, err := u.githubListDir(ctx, gtfobinsOwner, gtfobinsRepo, gtfobinsPath)
	if err != nil {
		return err
	}
	n := 0
	for _, it := range items {
		if it.Type != "file" || !strings.HasSuffix(it.Name, ".md") {
			continue
		}
		raw, err := u.fetchBytes(ctx, it.DownloadURL, filepath.Join(u.cache, "gtfobins", it.Name))
		if err != nil {
			u.logger.Warn("gtfobins fetch failed", slog.String("name", it.Name), slog.String("err", err.Error()))
			continue
		}
		a := parseGTFOBinsMarkdown(it.Name, string(raw))
		if err := u.repo.UpsertArtifact(ctx, "gtfobins", a); err != nil {
			return err
		}
		n++
	}
	u.logger.Info("GTFOBins ingest done", slog.Int("count", n))
	return nil
}

func (u *ScraperUsecase) githubListDir(ctx context.Context, owner, repo, path string) ([]ghContent, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "threat_intelligence-lola/1.0")
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := u.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("github list %s: %d %s", url, resp.StatusCode, string(b))
	}
	var out []ghContent
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (u *ScraperUsecase) fetchBytes(ctx context.Context, downloadURL, cacheFile string) ([]byte, error) {
	if u.cache != "" && cacheFile != "" {
		if b, err := os.ReadFile(cacheFile); err == nil && len(b) > 0 {
			return b, nil
		}
	}
	backoff := 1 * time.Second
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "threat_intelligence-lola/1.0")
		resp, err := u.http.Do(req)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}
			lastErr = err
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
			_ = resp.Body.Close()
			return nil, fmt.Errorf("download %s: %d %s", downloadURL, resp.StatusCode, string(b))
		}
		b, rerr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if rerr != nil {
			lastErr = rerr
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		if u.cache != "" && cacheFile != "" {
			_ = os.MkdirAll(filepath.Dir(cacheFile), 0o755)
			_ = os.WriteFile(cacheFile, b, 0o644)
		}
		return b, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("download failed: %s", downloadURL)
}

func parseLOLBASYAML(raw []byte) (*domain.Artifact, error) {
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, err
	}
	a := &domain.Artifact{}
	if v, ok := root["Name"].(string); ok {
		a.Name = strings.TrimSpace(v)
	}
	if v, ok := root["Description"].(string); ok {
		a.Description = strings.TrimSpace(v)
	}
	if v, ok := root["Author"].(string); ok && v != "" {
		a.Description = strings.TrimSpace(a.Description + "\n\nAuthor: " + v)
	}
	if v, ok := root["OS"].(string); ok && v != "" {
		a.OS = []string{v}
	}
	if v, ok := root["MITRE"].(map[string]any); ok {
		if id, ok := v["ID"].(string); ok {
			a.MitreID = id
		}
	}
	if v, ok := root["Categories"].([]any); ok && len(v) > 0 {
		if s, ok := v[0].(string); ok {
			a.Category = s
		}
	}
	if v, ok := root["Full Path"].([]any); ok {
		for _, p := range v {
			if s, ok := p.(string); ok {
				a.Paths = append(a.Paths, s)
			}
		}
	}
	if v, ok := root["Commands"].([]any); ok {
		for _, c := range v {
			cm, ok := c.(map[string]any)
			if !ok {
				continue
			}
			cmd := domain.Command{}
			if s, ok := cm["Command"].(string); ok {
				cmd.Command = s
			}
			if s, ok := cm["Description"].(string); ok {
				cmd.Description = s
			}
			if s, ok := cm["Usecase"].(string); ok {
				cmd.Usecase = s
			}
			a.Commands = append(a.Commands, cmd)
		}
	}
	if det, ok := root["Detection"].(map[string]any); ok {
		if sig, ok := det["Sigma"].([]any); ok {
			for _, s := range sig {
				if str, ok := s.(string); ok {
					a.Detection.Sigma = append(a.Detection.Sigma, str)
				}
			}
		}
		if yar, ok := det["YARA"].([]any); ok {
			for _, s := range yar {
				if str, ok := s.(string); ok {
					a.Detection.Yara = append(a.Detection.Yara, str)
				}
			}
		}
	}
	if a.Name == "" {
		return nil, fmt.Errorf("missing Name")
	}
	return a, nil
}

func parseGTFOBinsMarkdown(filename, body string) *domain.Artifact {
	name := strings.TrimSuffix(filename, ".md")
	name = strings.ReplaceAll(name, "_", "/")
	a := &domain.Artifact{
		Name:        name,
		Description: firstParagraph(body),
		OS:          []string{"linux"},
		Category:    "gtfobins",
	}
	a.Commands = []domain.Command{
		{Command: "(see GTFOBins page)", Description: truncate(body, 8000)},
	}
	return a
}

func firstParagraph(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "\n## "); idx > 0 {
		s = s[:idx]
	}
	return truncate(strings.TrimSpace(s), 4000)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n\n…"
}
