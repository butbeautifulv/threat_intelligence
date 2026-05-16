package intelligence

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// RateLimitProbe is the result of an optional pre-flight rate-limit check.
type RateLimitProbe struct {
	Detected bool   `json:"rate_limit_detected"`
	Source   string `json:"source"`
	Detail   string `json:"detail,omitempty"`
}

// ProbeRateLimit checks whether the target appears rate-limited (advisory only).
func (s *Service) ProbeRateLimit(ctx context.Context, subject, target string) RateLimitProbe {
	if target == "" {
		return RateLimitProbe{Source: "none"}
	}
	if s.Tools != nil && s.Registry != nil {
		if spec, ok := s.Registry.Get("httpx_probe"); ok && spec.Enabled {
			res := s.Tools.Run(ctx, subject, "httpx_probe", contract.ToolRunRequest{
				Target: target,
				Parameters: map[string]string{
					"target":          target,
					"additional_args": "-status-code -silent -timeout 5",
				},
			})
			if detected, detail := parseRateLimitOutput(res.Output); detected {
				return RateLimitProbe{Detected: true, Source: "httpx", Detail: detail}
			}
			if res.Success && res.Output != "" {
				return RateLimitProbe{Source: "httpx", Detail: "no rate limit indicators"}
			}
		}
	}
	if probe := probeHTTPRateLimit(ctx, target); probe.Detected {
		return probe
	}
	return RateLimitProbe{Source: "none"}
}

func probeHTTPRateLimit(ctx context.Context, target string) RateLimitProbe {
	url := target
	if !strings.HasPrefix(strings.ToLower(url), "http") {
		url = "https://" + url
	}
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodHead, url, nil)
	if err != nil {
		return RateLimitProbe{Source: "http", Detail: err.Error()}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// GET fallback for servers that reject HEAD
		req2, err2 := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
		if err2 != nil {
			return RateLimitProbe{Source: "http", Detail: err.Error()}
		}
		resp, err = http.DefaultClient.Do(req2)
		if err != nil {
			return RateLimitProbe{Source: "http", Detail: err.Error()}
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return RateLimitProbe{Detected: true, Source: "http", Detail: "HTTP 429"}
	}
	if resp.Header.Get("Retry-After") != "" {
		return RateLimitProbe{Detected: true, Source: "http", Detail: "Retry-After header"}
	}
	return RateLimitProbe{Source: "http", Detail: "no rate limit indicators"}
}

func parseRateLimitOutput(output string) (bool, string) {
	low := strings.ToLower(output)
	for _, sig := range []string{"429", "rate limit", "too many requests", "throttled"} {
		if strings.Contains(low, sig) {
			return true, sig
		}
	}
	return false, ""
}
