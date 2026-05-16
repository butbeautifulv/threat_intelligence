package audit

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ExportBatch is the webhook/SIEM payload shape.
type ExportBatch struct {
	Source string  `json:"source"`
	SentAt string  `json:"sent_at"`
	Events []Event `json:"events"`
}

// ExportWebhook POSTs audit events to url with optional HMAC signature.
func ExportWebhook(ctx context.Context, url, secret string, events []Event) error {
	if url == "" {
		return fmt.Errorf("webhook url required")
	}
	batch := ExportBatch{
		Source: "veil-engage",
		SentAt: time.Now().UTC().Format(time.RFC3339),
		Events: events,
	}
	body, err := json.Marshal(batch)
	if err != nil {
		return err
	}
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		if secret != "" {
			mac := hmac.New(sha256.New, []byte(secret))
			_, _ = mac.Write(body)
			req.Header.Set("X-Engage-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("webhook status %d", resp.StatusCode)
	}
	return lastErr
}

// WebhookConfigFromEnv reads ENGAGE_AUDIT_WEBHOOK_URL and ENGAGE_AUDIT_WEBHOOK_SECRET.
func WebhookConfigFromEnv() (url, secret string) {
	url = strings.TrimSpace(os.Getenv("ENGAGE_AUDIT_WEBHOOK_URL"))
	secret = strings.TrimSpace(os.Getenv("ENGAGE_AUDIT_WEBHOOK_SECRET"))
	return url, secret
}
