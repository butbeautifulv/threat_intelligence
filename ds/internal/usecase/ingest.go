package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"ds/internal/proxypool"
	neo4jstore "ds/internal/storage/neo4j"
)

type Ingestor struct {
	store  *neo4jstore.Store
	logger *slog.Logger
	http   *http.Client
	cache  string
}

func NewIngestor(store *neo4jstore.Store, logger *slog.Logger, cacheDir string) *Ingestor {
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.TLSHandshakeTimeout = 30 * time.Second
	var rt http.RoundTripper = base
	if env := strings.TrimSpace(os.Getenv("DS_PROXY_URLS")); env != "" {
		p, err := proxypool.New(proxypool.SplitEnvList(env), 2*time.Minute)
		if err == nil {
			only := strings.EqualFold(strings.TrimSpace(os.Getenv("DS_PROXY_MODE")), "only")
			rt = proxypool.NewTransport(base, p, only)
			logger.Info("ds proxy pool enabled", slog.Int("count", len(proxypool.SplitEnvList(env))))
		} else {
			logger.Warn("ds proxy pool invalid; running direct", slog.String("err", err.Error()))
		}
	}
	return &Ingestor{store: store, logger: logger, http: &http.Client{Timeout: 120 * time.Second, Transport: rt}, cache: cacheDir}
}

type ghContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

func (u *Ingestor) Run(ctx context.Context) error {
	if err := u.store.EnsureSchema(ctx); err != nil {
		return err
	}
	maxSigma := envInt("DS_MAX_SIGMA", 250)
	maxYara := envInt("DS_MAX_YARA", 120)
	maxAtomic := envInt("DS_MAX_ATOMIC", 200)

	if err := u.ingestSigmaDir(ctx, "SigmaHQ", "sigma", "rules/windows/process_creation", maxSigma); err != nil {
		return err
	}
	if err := u.ingestYaraDir(ctx, "Neo23x0", "signature-base", firstNonEmpty(os.Getenv("DS_YARA_PATH"), "yara"), maxYara); err != nil {
		u.logger.Warn("yara ingest primary path failed", slog.String("err", err.Error()))
		if err2 := u.ingestYaraDir(ctx, "Neo23x0", "signature-base", "iocs/yara", maxYara); err2 != nil {
			return err2
		}
	}
	if err := u.ingestAtomicTests(ctx, maxAtomic); err != nil {
		return err
	}
	return nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func (u *Ingestor) ingestSigmaDir(ctx context.Context, owner, repo, path string, limit int) error {
	u.logger.Info("ingesting Sigma rules", slog.String("path", path), slog.Int("limit", limit))
	items, err := u.githubListDir(ctx, owner, repo, path)
	if err != nil {
		return err
	}
	n := 0
	for _, it := range items {
		if n >= limit {
			break
		}
		if it.Type != "file" || (!strings.HasSuffix(it.Name, ".yml") && !strings.HasSuffix(it.Name, ".yaml")) {
			continue
		}
		raw, err := u.fetchBytes(ctx, it.DownloadURL, filepath.Join(u.cache, "sigma", it.Name))
		if err != nil {
			continue
		}
		var root map[string]any
		if err := yaml.Unmarshal(raw, &root); err != nil {
			continue
		}
		id, _ := root["id"].(string)
		title, _ := root["title"].(string)
		level, _ := root["level"].(string)
		var logProduct, logService string
		if ls, ok := root["logsource"].(map[string]any); ok {
			logProduct, _ = ls["product"].(string)
			logService, _ = ls["service"].(string)
		}
		tags := tagsToJSON(root["tags"])
		if id == "" {
			id = neo4jstore.StableID("sigma", it.Path)
		}
		md := fmt.Sprintf("# %s\n\n**id:** `%s`  \n**level:** %s  \n\n```yaml\n%s\n```\n", title, id, level, string(raw))
		if err := u.store.UpsertSigmaRule(ctx, id, title, level, logProduct, logService, tags, md, "sigmahq"); err != nil {
			return err
		}
		n++
	}
	u.logger.Info("sigma ingest done", slog.Int("count", n))
	return nil
}

func tagsToJSON(v any) string {
	arr, ok := v.([]any)
	if !ok {
		return "[]"
	}
	ss := make([]string, 0, len(arr))
	for _, x := range arr {
		if s, ok := x.(string); ok {
			ss = append(ss, s)
		}
	}
	b, _ := json.Marshal(ss)
	return string(b)
}

func (u *Ingestor) ingestYaraDir(ctx context.Context, owner, repo, path string, limit int) error {
	u.logger.Info("ingesting YARA rules", slog.Int("limit", limit))
	items, err := u.githubListDir(ctx, owner, repo, path)
	if err != nil {
		return err
	}
	n := 0
	for _, it := range items {
		if n >= limit {
			break
		}
		if it.Type != "file" || !strings.HasSuffix(it.Name, ".yar") {
			continue
		}
		raw, err := u.fetchBytes(ctx, it.DownloadURL, filepath.Join(u.cache, "yara", it.Name))
		if err != nil {
			continue
		}
		body := string(raw)
		name := parseYaraRuleName(body, it.Name)
		id := neo4jstore.StableID("yara", it.Path)
		md := fmt.Sprintf("# %s\n\n```yara\n%s\n```\n", name, body)
		if err := u.store.UpsertYaraRule(ctx, id, name, "", "[]", md, "neo23x0-signature-base"); err != nil {
			return err
		}
		n++
	}
	u.logger.Info("yara ingest done", slog.Int("count", n))
	return nil
}

func parseYaraRuleName(body, fallback string) string {
	lines := strings.Split(body, "\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "rule ") {
			p := strings.TrimPrefix(ln, "rule ")
			if idx := strings.IndexAny(p, " \t{"); idx > 0 {
				return strings.TrimSpace(p[:idx])
			}
			return strings.TrimSpace(p)
		}
	}
	return strings.TrimSuffix(fallback, ".yar")
}

func (u *Ingestor) ingestAtomicTests(ctx context.Context, limit int) error {
	u.logger.Info("ingesting Atomic Red Team YAML", slog.Int("limit", limit))
	items, err := u.githubListDir(ctx, "redcanaryco", "atomic-red-team", "atomics")
	if err != nil {
		return err
	}
	n := 0
	for _, dir := range items {
		if n >= limit {
			break
		}
		if dir.Type != "dir" {
			continue
		}
		// technique folder contains Txxxx.yaml
		sub, err := u.githubListDir(ctx, "redcanaryco", "atomic-red-team", dir.Path)
		if err != nil {
			continue
		}
		for _, f := range sub {
			if n >= limit {
				break
			}
			if f.Type != "file" || (!strings.HasSuffix(f.Name, ".yaml") && !strings.HasSuffix(f.Name, ".yml")) {
				continue
			}
			raw, err := u.fetchBytes(ctx, f.DownloadURL, filepath.Join(u.cache, "atomic", f.Name))
			if err != nil {
				continue
			}
			added, err := u.parseAtomicYAML(ctx, raw, f.Path)
			if err != nil {
				u.logger.Debug("atomic parse skip", slog.String("path", f.Path), slog.String("err", err.Error()))
				continue
			}
			n += added
			if n >= limit {
				break
			}
		}
		if n >= limit {
			break
		}
	}
	u.logger.Info("atomic ingest done", slog.Int("count", n))
	return nil
}

func (u *Ingestor) parseAtomicYAML(ctx context.Context, raw []byte, path string) (int, error) {
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return 0, err
	}
	attackID, _ := root["attack_technique"].(string)
	name, _ := root["display_name"].(string)
	_ = name
	atomicTests, _ := root["atomic_tests"].([]any)
	if len(atomicTests) == 0 {
		return 0, fmt.Errorf("no atomic_tests")
	}
	count := 0
	for i, t := range atomicTests {
		tm, ok := t.(map[string]any)
		if !ok {
			continue
		}
		tid, _ := tm["auto_generated_guid"].(string)
		if tid == "" {
			tid = fmt.Sprintf("%s-%d", attackID, i)
		}
		tname, _ := tm["name"].(string)
		tactic := ""
		if ta, ok := tm["tactics"].([]any); ok && len(ta) > 0 {
			if s, ok := ta[0].(string); ok {
				tactic = s
			}
		}
		execName, execCmd := "", ""
		if ex, ok := tm["executor"].(map[string]any); ok {
			execName, _ = ex["name"].(string)
			execCmd, _ = ex["command"].(string)
		}
		md := fmt.Sprintf("# %s\n\n**Technique:** %s  \n**Test:** %s  \n\n```yaml\n%s\n```\n", tname, attackID, tid, string(raw))
		if err := u.store.UpsertAtomicTest(ctx, tid, tname, tactic, attackID, execName, execCmd, md, "atomic-red-team"); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (u *Ingestor) githubListDir(ctx context.Context, owner, repo, path string) ([]ghContent, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "threat_intelligence-ds/1.0")
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
		return nil, fmt.Errorf("github %s: %d %s", url, resp.StatusCode, string(b))
	}
	var out []ghContent
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (u *Ingestor) fetchBytes(ctx context.Context, downloadURL, cacheFile string) ([]byte, error) {
	if u.cache != "" && cacheFile != "" {
		if b, err := os.ReadFile(cacheFile); err == nil && len(b) > 0 {
			return b, nil
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "threat_intelligence-ds/1.0")
	resp, err := u.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("download %s: %d %s", downloadURL, resp.StatusCode, string(b))
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if u.cache != "" && cacheFile != "" {
		_ = os.MkdirAll(filepath.Dir(cacheFile), 0o755)
		_ = os.WriteFile(cacheFile, b, 0o644)
	}
	return b, nil
}
