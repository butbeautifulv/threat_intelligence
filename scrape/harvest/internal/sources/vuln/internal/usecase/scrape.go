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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/butbeautifulv/threat_intelligence/scrape/harvest/internal/feeds"
	"github.com/butbeautifulv/threat_intelligence/scrape/harvest/internal/ledger"
	"github.com/butbeautifulv/threat_intelligence/scrape/pkg/proxypool"
	"github.com/butbeautifulv/threat_intelligence/scrape/harvest/internal/sources/vuln/internal/repository"
)

const nvdBaseURL = "https://services.nvd.nist.gov/rest/json/cves/2.0"

type ScraperUsecase struct {
	repo   repository.VulnerabilityRepository
	logger *slog.Logger
	apiKey string
	http   *http.Client
	cache  string
	delay  time.Duration
	feeds  *feeds.Client
	ledger *ledger.Store
}

func NewScraperUsecase(repo repository.VulnerabilityRepository, logger *slog.Logger, apiKey string, fc *feeds.Client, led *ledger.Store) *ScraperUsecase {
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.TLSHandshakeTimeout = 30 * time.Second

	var rt http.RoundTripper = base
	if env := strings.TrimSpace(os.Getenv("VULN_PROXY_URLS")); env != "" {
		p, err := proxypool.New(proxypool.SplitEnvList(env), 2*time.Minute)
		if err == nil {
			only := strings.EqualFold(strings.TrimSpace(os.Getenv("VULN_PROXY_MODE")), "only")
			rt = proxypool.NewTransport(base, p, only)
			logger.Info("vuln proxy pool enabled", slog.Int("count", len(proxypool.SplitEnvList(env))))
		} else {
			logger.Warn("vuln proxy pool invalid; running direct", slog.String("err", err.Error()))
		}
	}

	cache := firstNonEmpty(os.Getenv("VULN_CACHE_DIR"), filepath.Join(".", "data", "cache"))
	if fc == nil {
		fc = feeds.NewClient(cache, logger)
	}
	if fc.Cache == "" {
		fc.Cache = cache
	}
	hc := &http.Client{Timeout: 60 * time.Second, Transport: rt}
	fc.HTTP = hc
	return &ScraperUsecase{
		repo:   repo,
		logger: logger,
		apiKey: apiKey,
		http:   hc,
		cache:  fc.Cache,
		delay:  parseDelayEnv(os.Getenv("VULN_REQUEST_DELAY"), 1200*time.Millisecond),
		feeds:  fc,
		ledger: led,
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
	// Accept Go durations ("1500ms", "2s") or plain milliseconds ("1200").
	if d, err := time.ParseDuration(v); err == nil && d >= 0 {
		return d
	}
	if ms, err := strconv.Atoi(v); err == nil && ms >= 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return def
}

func (u *ScraperUsecase) downloadNVDPage(ctx context.Context, startIndex, resultsPerPage int) ([]byte, error) {
	cachePath := fmt.Sprintf("nvd/start_%d_size_%d.json", startIndex, resultsPerPage)
	uu, _ := url.Parse(nvdBaseURL)
	q := uu.Query()
	q.Set("startIndex", strconv.Itoa(startIndex))
	q.Set("resultsPerPage", strconv.Itoa(resultsPerPage))
	uu.RawQuery = q.Encode()
	urlStr := uu.String()

	if u.feeds != nil && u.ledger != nil {
		key := fmt.Sprintf("nvd:page:%d:%d", startIndex, resultsPerPage)
		res, err := feeds.FetchIfDue(ctx, u.feeds, u.ledger, key, "vuln", urlStr, ledger.PolicyPeriodic, cachePath, func() (*http.Request, error) {
			return u.nvdRequest(ctx, urlStr)
		})
		if err != nil {
			return nil, err
		}
		if res.Skipped {
			if len(res.Body) > 0 {
				return res.Body, nil
			}
			return nil, fmt.Errorf("nvd page %d skipped by ledger (no cache)", startIndex)
		}
		if len(res.Body) > 0 {
			return res.Body, nil
		}
	}

	req, err := u.nvdRequest(ctx, urlStr)
	if err != nil {
		return nil, err
	}

	backoff := 1 * time.Second
	for attempt := 0; attempt < 6; attempt++ {
		if u.delay > 0 {
			time.Sleep(u.delay)
		}
		resp, err := u.http.Do(req)
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
		if rerr == nil && u.feeds != nil {
			_ = u.feeds.WriteCache(cachePath, b)
		}
		return b, rerr
	}
	return nil, fmt.Errorf("nvd fetch failed after retries")
}

func (u *ScraperUsecase) nvdRequest(ctx context.Context, urlStr string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "veil-scrape/1.0")
	if u.apiKey != "" {
		req.Header.Set("apiKey", u.apiKey)
	}
	return req, nil
}

// nvdPageStats reads pagination metadata from an NVD CVE API 2.0 JSON page without full CVE parsing.
func nvdPageStats(data []byte) (totalResults, itemCount int, err error) {
	var meta struct {
		TotalResults    int               `json:"totalResults"`
		Vulnerabilities []json.RawMessage `json:"vulnerabilities"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return 0, 0, err
	}
	return meta.TotalResults, len(meta.Vulnerabilities), nil
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
		pub, ok := u.repo.(interface {
			PublishNVDPage(context.Context, int, []byte) error
		})
		if !ok {
			return fmt.Errorf("vuln: repository must support PublishNVDPage")
		}
		tr, itemCount, err := nvdPageStats(data)
		if err != nil {
			return err
		}
		if total < 0 {
			total = tr
		}
		if err := pub.PublishNVDPage(ctx, start, data); err != nil {
			return err
		}
		count += itemCount
		u.logger.Info("nvd page published", slog.Int("startIndex", start), slog.Int("itemCount", itemCount), slog.Int("totalResults", total))
		if itemCount == 0 {
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

