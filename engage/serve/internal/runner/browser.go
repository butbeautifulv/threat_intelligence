package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// BrowserInspectOpts configures discovery browser inspection via HTTP proxy.
type BrowserInspectOpts struct {
	Target      string
	WaitTime    int
	Headless    bool
	Screenshot  bool
	ActiveTests bool
}

// BrowserProxy runs catalog browser tools via DISCOVERY_BROWSER_URL (or legacy ENGAGE_BROWSER_URL).
type BrowserProxy struct {
	BaseURL string
	Client  *http.Client
}

func NewBrowserProxyFromEnv() *BrowserProxy {
	base := strings.TrimSpace(os.Getenv("DISCOVERY_BROWSER_URL"))
	if base == "" {
		base = strings.TrimSpace(os.Getenv("ENGAGE_BROWSER_URL"))
	}
	if base == "" {
		return nil
	}
	return &BrowserProxy{
		BaseURL: strings.TrimRight(base, "/"),
		Client:  &http.Client{Timeout: 5 * time.Minute},
	}
}

func (b *BrowserProxy) Enabled() bool {
	return b != nil && b.BaseURL != ""
}

func IsBrowserBinary(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), "browser")
}

func (b *BrowserProxy) Inspect(ctx context.Context, opts BrowserInspectOpts) Result {
	payload := map[string]any{
		"url":          opts.Target,
		"target":       opts.Target,
		"wait_time":    opts.WaitTime,
		"headless":     opts.Headless,
		"screenshot":    opts.Screenshot,
		"active_tests": opts.ActiveTests,
		"inspect":      true,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.BaseURL+"/inspect", bytes.NewReader(body))
	if err != nil {
		return Result{ExitCode: -1, Err: err}
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.Client.Do(req)
	if err != nil {
		return Result{ExitCode: -1, Err: err}
	}
	defer resp.Body.Close()
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return Result{ExitCode: -1, Err: fmt.Errorf("discovery browser: %w", err)}
	}
	if raw["success"] == false {
		errMsg, _ := raw["error"].(string)
		return Result{ExitCode: 1, Err: fmt.Errorf("%s", errMsg), Stderr: errMsg}
	}
	out, _ := json.Marshal(raw)
	return Result{Stdout: string(out) + "\n", ExitCode: 0}
}

func (b *BrowserProxy) Exec(ctx context.Context, target string, args []string) Result {
	opts := BrowserInspectOpts{Target: target, WaitTime: 5, Headless: true, Screenshot: true}
	return b.Inspect(ctx, opts)
}
