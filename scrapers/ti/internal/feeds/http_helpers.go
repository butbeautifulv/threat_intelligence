package feeds

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const tiUserAgent = "threat_intelligence-ti/1.0"

const httpMaxAttempts = 5

func retryableHTTPStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func truncateForLog(b []byte) string {
	s := string(b)
	if len(s) > 512 {
		return s[:512] + "..."
	}
	return s
}

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

func (r *Runner) doGETWithRetries(ctx context.Context, urlStr string) ([]byte, error) {
	var lastErr error
	for attempt := range httpMaxAttempts {
		if attempt > 0 {
			backoff := time.Duration(500*(1<<min(attempt-1, 6))) * time.Millisecond
			backoff += time.Duration(rand.Int64N(400)) * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", tiUserAgent)
		resp, err := r.HTTP.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		b, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return b, nil
		}
		if retryableHTTPStatus(resp.StatusCode) && attempt < httpMaxAttempts-1 {
			lastErr = fmt.Errorf("http %d: %s", resp.StatusCode, truncateForLog(b))
			continue
		}
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, truncateForLog(b))
	}
	if lastErr != nil {
		return nil, fmt.Errorf("after retries: %w", lastErr)
	}
	return nil, fmt.Errorf("after retries: unknown error")
}

func writeCacheFile(cacheFile string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(cacheFile), 0o755); err != nil {
		return err
	}
	return os.WriteFile(cacheFile, b, 0o644)
}

// postJSONAuthCached sends a JSON POST with optional Auth-Key header, using delay/retry; caches successful body.
func (r *Runner) postJSONAuthCached(ctx context.Context, urlStr, cacheFile, authKey string, jsonBody []byte) ([]byte, error) {
	if r.Cache != "" && cacheFile != "" {
		if b, err := os.ReadFile(cacheFile); err == nil && len(b) > 0 {
			return b, nil
		}
	}
	r.preNetworkDelay()
	var lastErr error
	for attempt := range httpMaxAttempts {
		if attempt > 0 {
			backoff := time.Duration(500*(1<<min(attempt-1, 6))) * time.Millisecond
			backoff += time.Duration(rand.Int64N(400)) * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(jsonBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", tiUserAgent)
		req.Header.Set("Content-Type", "application/json")
		if strings.TrimSpace(authKey) != "" {
			req.Header.Set("Auth-Key", authKey)
		}
		resp, err := r.HTTP.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		b, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if r.Cache != "" && cacheFile != "" && len(b) > 0 {
				_ = writeCacheFile(cacheFile, b)
			}
			return b, nil
		}
		if retryableHTTPStatus(resp.StatusCode) && attempt < httpMaxAttempts-1 {
			lastErr = fmt.Errorf("http %d: %s", resp.StatusCode, truncateForLog(b))
			continue
		}
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, truncateForLog(b))
	}
	if lastErr != nil {
		return nil, fmt.Errorf("after retries: %w", lastErr)
	}
	return nil, fmt.Errorf("after retries: unknown error")
}
