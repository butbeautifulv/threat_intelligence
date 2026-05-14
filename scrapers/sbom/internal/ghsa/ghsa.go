package ghsa

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const rawBase = "https://raw.githubusercontent.com/github/advisory-database/main"

type Client struct {
	http  *http.Client
	token string
}

func NewClient(token string) *Client {
	return &Client{
		http:  &http.Client{Timeout: 120 * time.Second},
		token: strings.TrimSpace(token),
	}
}

type contentItem struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func (c *Client) getJSON(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ghsa: %s: %s", resp.Status, string(body[:min(300, len(body))]))
	}
	return body, nil
}

// CollectAdvisoryPaths walks advisories/github-reviewed from newest years downward until maxPaths or minYear.
func (c *Client) CollectAdvisoryPaths(ctx context.Context, maxPaths int, minYear int) ([]string, error) {
	if maxPaths <= 0 {
		maxPaths = 50
	}
	if minYear <= 0 {
		minYear = 2017
	}
	base := "https://api.github.com/repos/github/advisory-database/contents/advisories/github-reviewed"
	var out []string
	nowY := time.Now().UTC().Year()
	for year := nowY; year >= minYear && len(out) < maxPaths; year-- {
		yearURL := fmt.Sprintf("%s/%d", base, year)
		body, err := c.getJSON(ctx, yearURL)
		if err != nil {
			continue
		}
		var months []contentItem
		if err := json.Unmarshal(body, &months); err != nil {
			continue
		}
		for mi := len(months) - 1; mi >= 0 && len(out) < maxPaths; mi-- {
			m := months[mi]
			if m.Type != "dir" {
				continue
			}
			mURL := fmt.Sprintf("%s/%d/%s", base, year, m.Name)
			mb, err := c.getJSON(ctx, mURL)
			if err != nil {
				continue
			}
			var advisories []contentItem
			if err := json.Unmarshal(mb, &advisories); err != nil {
				continue
			}
			for ai := len(advisories) - 1; ai >= 0 && len(out) < maxPaths; ai-- {
				a := advisories[ai]
				if a.Type != "dir" || !strings.HasPrefix(strings.ToUpper(a.Name), "GHSA-") {
					continue
				}
				rel := fmt.Sprintf("advisories/github-reviewed/%d/%s/%s/%s.json", year, m.Name, a.Name, a.Name)
				out = append(out, rel)
			}
		}
	}
	return out, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FetchAdvisoryJSON downloads raw JSON (OSV schema) for a relative path from CollectAdvisoryPaths.
func (c *Client) FetchAdvisoryJSON(ctx context.Context, relPath string) (map[string]any, error) {
	relPath = strings.TrimPrefix(relPath, "/")
	u := rawBase + "/" + relPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ghsa raw: %s: %s", resp.Status, string(body[:min(200, len(body))]))
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}
