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

// BrowserProxy runs browser catalog tools via ENGAGE_BROWSER_URL sidecar.
type BrowserProxy struct {
	BaseURL string
	Client  *http.Client
}

func NewBrowserProxyFromEnv() *BrowserProxy {
	base := strings.TrimSpace(os.Getenv("ENGAGE_BROWSER_URL"))
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

func (b *BrowserProxy) Exec(ctx context.Context, target string, args []string) Result {
	payload := map[string]any{
		"target": target,
		"args":   args,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.BaseURL+"/exec", bytes.NewReader(body))
	if err != nil {
		return Result{ExitCode: -1, Err: err}
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.Client.Do(req)
	if err != nil {
		return Result{ExitCode: -1, Err: err}
	}
	defer resp.Body.Close()
	var out struct {
		Stdout   string `json:"stdout"`
		Stderr   string `json:"stderr"`
		ExitCode int    `json:"exit_code"`
		Error    string `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Result{ExitCode: -1, Err: fmt.Errorf("browser sidecar: %w", err)}
	}
	res := Result{Stdout: out.Stdout, Stderr: out.Stderr, ExitCode: out.ExitCode}
	if out.Error != "" {
		res.Err = fmt.Errorf("%s", out.Error)
	}
	if resp.StatusCode >= 400 && res.Err == nil {
		res.Err = fmt.Errorf("browser sidecar http %d", resp.StatusCode)
		res.ExitCode = resp.StatusCode
	}
	return res
}
