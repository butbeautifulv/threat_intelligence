package feeds

import (
	"bytes"
	"context"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	scrapefeeds "github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
)

const tiUserAgent = "veil-ti/1.0"

// preNetworkDelay sleeps the configured base delay plus a small random jitter (reduces thundering herd).
func (r *Runner) preNetworkDelay() {
	if r.Delay <= 0 {
		return
	}
	extra := time.Duration(0)
	if r.Delay > 4 {
		extra = time.Duration(rand.Int64N(int64(r.Delay / 4)))
	}
	time.Sleep(r.Delay + extra)
}

// fetchLedgerPOST wraps FetchIfDue for JSON POST feeds (ThreatFox API, MalwareBazaar).
func (r *Runner) fetchLedgerPOST(ctx context.Context, key, url, cacheRel string, policy ledger.FetchPolicy, authKey string, jsonBody []byte) (scrapefeeds.FetchResult, error) {
	r.preNetworkDelay()
	body := append([]byte(nil), jsonBody...)
	return scrapefeeds.FetchIfDue(ctx, r.Feeds, r.Ledger, key, "ti", url, policy, cacheRel, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", tiUserAgent)
		req.Header.Set("Content-Type", "application/json")
		if strings.TrimSpace(authKey) != "" {
			req.Header.Set("Auth-Key", authKey)
		}
		return req, nil
	})
}

