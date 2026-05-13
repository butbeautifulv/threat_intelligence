package feeds

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ti/internal/domain"
	"ti/internal/normalize"
	"ti/internal/proxypool"
	neo4jstore "ti/internal/storage/neo4j"
)

const kevURL = "https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json"

type Runner struct {
	Store  *neo4jstore.Store
	Logger *slog.Logger
	HTTP   *http.Client
	Cache  string
	Delay  time.Duration
}

func NewRunner(store *neo4jstore.Store, logger *slog.Logger) *Runner {
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.TLSHandshakeTimeout = 30 * time.Second
	var rt http.RoundTripper = base
	if env := strings.TrimSpace(os.Getenv("TI_PROXY_URLS")); env != "" {
		p, err := proxypool.New(proxypool.SplitEnvList(env), 2*time.Minute)
		if err == nil {
			only := strings.EqualFold(strings.TrimSpace(os.Getenv("TI_PROXY_MODE")), "only")
			rt = proxypool.NewTransport(base, p, only)
			logger.Info("ti proxy pool enabled", slog.Int("count", len(proxypool.SplitEnvList(env))))
		} else {
			logger.Warn("ti proxy pool invalid; running direct", slog.String("err", err.Error()))
		}
	}
	return &Runner{
		Store:  store,
		Logger: logger,
		HTTP:   &http.Client{Timeout: 120 * time.Second, Transport: rt},
		Cache:  firstNonEmpty(os.Getenv("TI_CACHE_DIR"), filepath.Join(".", "data", "cache")),
		Delay:  parseDelayEnv(os.Getenv("TI_REQUEST_DELAY"), 1200*time.Millisecond),
	}
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func parseDelayEnv(v string, def time.Duration) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return def
	}
	if d, err := time.ParseDuration(v); err == nil && d >= 0 {
		return d
	}
	if ms, err := strconv.Atoi(v); err == nil && ms >= 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return def
}

func (r *Runner) getBytesCached(ctx context.Context, urlStr, cacheFile string) ([]byte, error) {
	if r.Cache != "" && cacheFile != "" {
		if b, err := os.ReadFile(cacheFile); err == nil && len(b) > 0 {
			return b, nil
		}
	}
	if r.Delay > 0 {
		time.Sleep(r.Delay)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "threat_intelligence-ti/1.0")
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(b))
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if r.Cache != "" && cacheFile != "" && len(b) > 0 {
		_ = os.MkdirAll(filepath.Dir(cacheFile), 0o755)
		_ = os.WriteFile(cacheFile, b, 0o644)
	}
	return b, nil
}

func (r *Runner) Run(ctx context.Context, kinds []string) error {
	if err := r.Store.EnsureSchema(ctx); err != nil {
		return err
	}
	for _, k := range kinds {
		switch strings.TrimSpace(strings.ToLower(k)) {
		case "kev":
			if err := r.runKEV(ctx); err != nil {
				return err
			}
		case "pt":
			if err := r.runPTRSS(ctx); err != nil {
				return err
			}
		case "urlhaus":
			if err := r.runURLhaus(ctx); err != nil {
				return err
			}
		default:
			r.Logger.Info("unknown feed kind", slog.String("kind", k))
		}
	}
	return nil
}

type kevFile struct {
	Vulnerabilities []struct {
		CVEID         string `json:"cveID"`
		VendorProject string `json:"vendorProject"`
		Product       string `json:"product"`
		ShortDesc     string `json:"shortDescription"`
		DateAdded     string `json:"dateAdded"`
	} `json:"vulnerabilities"`
}

func (r *Runner) runKEV(ctx context.Context) error {
	r.Logger.Info("ingesting CISA KEV")
	b, err := r.getBytesCached(ctx, kevURL, filepath.Join(r.Cache, "ti", "kev.json"))
	if err != nil {
		return err
	}
	var doc kevFile
	if err := json.Unmarshal(b, &doc); err != nil {
		return err
	}
	limitN := len(doc.Vulnerabilities)
	if v := os.Getenv("TI_KEV_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n < limitN {
			limitN = n
		}
	}
	for i := 0; i < limitN; i++ {
		v := doc.Vulnerabilities[i]
		if err := r.Store.UpsertKEVVulnerability(ctx, v.CVEID, v.VendorProject, v.Product, v.ShortDesc, v.DateAdded); err != nil {
			return err
		}
	}
	r.Logger.Info("KEV ingest done", slog.Int("count", limitN))
	return nil
}

type rss struct {
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

func (r *Runner) runPTRSS(ctx context.Context) error {
	u := os.Getenv("PT_RSS_URL")
	if u == "" {
		u = "https://www.ptsecurity.com/rss/all.xml"
	}
	r.Logger.Info("ingesting PT RSS", slog.String("url", u))
	cacheFile := filepath.Join(r.Cache, "ti", "pt.xml")
	b, err := r.getBytesCached(ctx, u, cacheFile)
	if err != nil {
		return err
	}
	var doc rss
	if err := xml.Unmarshal(b, &doc); err != nil {
		return err
	}
	limitN := len(doc.Channel.Items)
	if v := os.Getenv("TI_PT_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n < limitN {
			limitN = n
		}
	}
	n := 0
	for i := 0; i < limitN; i++ {
		it := doc.Channel.Items[i]
		if !strings.Contains(strings.ToLower(it.Link), "ptsecurity") && !strings.Contains(strings.ToLower(it.Title), "pt") {
			continue
		}
		rep := domain.Report{
			Title:        it.Title,
			Provider:     "Positive Technologies",
			Link:         it.Link,
			PublishedAt:  it.PubDate,
			BodyMarkdown: stripHTML(it.Description),
			Source:       "pt-rss",
		}
		if err := r.Store.UpsertReport(ctx, rep); err != nil {
			return err
		}
		rid := normalize.ReportStableID(it.Link)
		for _, ioc := range extractIOCsFromText(it.Title + "\n" + it.Description + "\n" + rep.BodyMarkdown) {
			ni, ok := normalize.NormalizeIOC(ioc)
			if !ok {
				continue
			}
			if err := r.Store.UpsertIOC(ctx, ni); err != nil {
				return err
			}
			if err := r.Store.LinkReportMentionsIOC(ctx, rid, ni); err != nil {
				return err
			}
		}
		n++
	}
	r.Logger.Info("PT RSS ingest done", slog.Int("reports", n))
	return nil
}

func stripHTML(s string) string {
	// minimal: remove tags
	out := s
	for {
		i := strings.Index(out, "<")
		if i < 0 {
			break
		}
		j := strings.Index(out[i:], ">")
		if j < 0 {
			break
		}
		out = out[:i] + out[i+j+1:]
	}
	return strings.TrimSpace(out)
}

func (r *Runner) runURLhaus(ctx context.Context) error {
	u := "https://urlhaus.abuse.ch/downloads/csv_recent/"
	r.Logger.Info("ingesting URLhaus recent CSV", slog.String("url", u))
	b, err := r.getBytesCached(ctx, u, filepath.Join(r.Cache, "ti", "urlhaus_recent.csv"))
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	max := 500
	if v := os.Getenv("TI_URLHAUS_MAX"); v != "" {
		var n int
		_, _ = fmt.Sscanf(v, "%d", &n)
		if n > 0 {
			max = n
		}
	}
	count := 0
	for _, ln := range lines {
		if count >= max {
			break
		}
		if strings.HasPrefix(ln, "#") || strings.TrimSpace(ln) == "" {
			continue
		}
		// CSV: id,dateadded,url,url_status,last_online,threat,tags,urlhaus_link,reporter
		parts := strings.Split(ln, `","`)
		if len(parts) < 3 {
			parts = strings.Split(ln, ",")
		}
		var urlStr string
		for _, p := range parts {
			p = strings.Trim(p, `"`)
			if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
				urlStr = p
				break
			}
		}
		if urlStr == "" {
			continue
		}
		ioc := domain.IOC{Type: domain.IOCURL, Value: urlStr, Source: "urlhaus"}
		ni, ok := normalize.NormalizeIOC(ioc)
		if !ok {
			continue
		}
		if err := r.Store.UpsertIOC(ctx, ni); err != nil {
			return err
		}
		count++
	}
	r.Logger.Info("URLhaus ingest done", slog.Int("iocs", count))
	return nil
}
