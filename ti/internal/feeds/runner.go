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
	"strconv"
	"strings"
	"time"

	"ti/internal/domain"
	"ti/internal/normalize"
	neo4jstore "ti/internal/storage/neo4j"
)

const kevURL = "https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json"

type Runner struct {
	Store  *neo4jstore.Store
	Logger *slog.Logger
	HTTP   *http.Client
}

func NewRunner(store *neo4jstore.Store, logger *slog.Logger) *Runner {
	return &Runner{
		Store:  store,
		Logger: logger,
		HTTP:   &http.Client{Timeout: 120 * time.Second},
	}
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, kevURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "threat_intelligence-ti/1.0")
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("kev http %d: %s", resp.StatusCode, string(b))
	}
	var doc kevFile
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "threat_intelligence-ti/1.0")
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "threat_intelligence-ti/1.0")
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
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
