package veilgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Config for veil-api access (service account).
type Config struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	TokenURL     string
	HTTPClient   *http.Client
}

// Client calls veil graph read API with OAuth2 client credentials.
type Client struct {
	cfg   Config
	mu    sync.Mutex
	token string
	exp   time.Time
}

func New(cfg Config) *Client {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	cfg.BaseURL = strings.TrimSuffix(strings.TrimSpace(cfg.BaseURL), "/")
	return &Client{cfg: cfg}
}

func (c *Client) Enabled() bool {
	return c.cfg.BaseURL != ""
}

func (c *Client) oauthConfigured() bool {
	return strings.TrimSpace(c.cfg.ClientID) != "" &&
		strings.TrimSpace(c.cfg.ClientSecret) != "" &&
		strings.TrimSpace(c.cfg.TokenURL) != ""
}

func (c *Client) GetJSON(ctx context.Context, path string) (json.RawMessage, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("veil api client not configured")
	}
	url := c.cfg.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.oauthConfigured() {
		tok, err := c.bearer(ctx)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("veil api %s: %d %s", path, resp.StatusCode, string(body))
	}
	return json.RawMessage(body), nil
}

func (c *Client) Categories(ctx context.Context) (json.RawMessage, error) {
	return c.GetJSON(ctx, "/v1/categories")
}

// Search queries veil-api category search endpoint.
func (c *Client) Search(ctx context.Context, category, query string) (json.RawMessage, error) {
	category = strings.TrimSpace(category)
	query = strings.TrimSpace(query)
	if category == "" || query == "" {
		return nil, fmt.Errorf("category and query required")
	}
	path := fmt.Sprintf("/v1/categories/%s/search?q=%s", category, url.QueryEscape(query))
	return c.GetJSON(ctx, path)
}

// EngageContext loads structured engage subgraph for a target host.
func (c *Client) EngageContext(ctx context.Context, host string) (json.RawMessage, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, fmt.Errorf("host required")
	}
	path := fmt.Sprintf("/v1/categories/engage/context?q=%s", url.QueryEscape(host))
	return c.GetJSON(ctx, path)
}

// GetNode fetches a node by elementId or business key (including EngageTarget.name).
func (c *Client) GetNode(ctx context.Context, id string) (json.RawMessage, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id required")
	}
	path := fmt.Sprintf("/v1/nodes/%s", url.PathEscape(id))
	return c.GetJSON(ctx, path)
}

// PlaybookProcedure returns structured procedure for a skill id.
func (c *Client) PlaybookProcedure(ctx context.Context, skillID string) (json.RawMessage, error) {
	skillID = strings.TrimSpace(skillID)
	if skillID == "" {
		return nil, fmt.Errorf("skill_id required")
	}
	path := fmt.Sprintf("/v1/playbooks/%s/procedure", url.PathEscape(skillID))
	return c.GetJSON(ctx, path)
}

// PlaybookRecommendTools returns engage catalog tools linked to a skill or technique.
func (c *Client) PlaybookRecommendTools(ctx context.Context, skillID, techniqueID string) (json.RawMessage, error) {
	if strings.TrimSpace(techniqueID) != "" {
		path := fmt.Sprintf("/v1/playbooks/ontology/technique/%s/skills", url.PathEscape(techniqueID))
		return c.GetJSON(ctx, path)
	}
	skillID = strings.TrimSpace(skillID)
	if skillID == "" {
		return nil, fmt.Errorf("skill_id or technique_id required")
	}
	path := fmt.Sprintf("/v1/playbooks/%s/recommend-tools", url.PathEscape(skillID))
	return c.GetJSON(ctx, path)
}

// PlaybooksByTechnique lists cybersecurity playbooks for a MITRE ATT&CK technique id.
func (c *Client) PlaybooksByTechnique(ctx context.Context, techniqueID string) (json.RawMessage, error) {
	techniqueID = strings.TrimSpace(techniqueID)
	if techniqueID == "" {
		return nil, fmt.Errorf("technique_id required")
	}
	path := fmt.Sprintf("/v1/playbooks/by-technique/%s", url.PathEscape(techniqueID))
	return c.GetJSON(ctx, path)
}

// Neighbors returns a k-hop subgraph around a node id.
func (c *Client) Neighbors(ctx context.Context, id string, depth int) (json.RawMessage, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id required")
	}
	if depth <= 0 {
		depth = 1
	}
	path := fmt.Sprintf("/v1/nodes/%s/neighbors?depth=%d", url.PathEscape(id), depth)
	return c.GetJSON(ctx, path)
}

func (c *Client) bearer(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.exp.Add(-30*time.Second)) {
		return c.token, nil
	}
	if c.cfg.ClientID == "" || c.cfg.ClientSecret == "" || c.cfg.TokenURL == "" {
		return "", fmt.Errorf("veil api oauth not configured")
	}
	form := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
		c.cfg.ClientID, c.cfg.ClientSecret)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenURL,
		strings.NewReader(form))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("empty access_token")
	}
	c.token = out.AccessToken
	c.exp = time.Now().Add(time.Duration(out.ExpiresIn) * time.Second)
	return c.token, nil
}
