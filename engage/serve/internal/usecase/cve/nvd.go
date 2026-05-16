package cve

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const nvdBaseURL = "https://services.nvd.nist.gov/rest/json/cves/2.0"

// NVDClient fetches CVE data from NVD API 2.0.
type NVDClient interface {
	FetchCVE(ctx context.Context, cveID string) (*CVEEntry, error)
	FetchRecent(ctx context.Context, hours int, severityFilter string) ([]CVEEntry, error)
}

type httpNVDClient struct {
	baseURL string
	http    *http.Client
	apiKey  string
}

// DefaultNVDClient returns a production NVD HTTP client.
func DefaultNVDClient() NVDClient {
	return &httpNVDClient{
		baseURL: nvdBaseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
		apiKey:  strings.TrimSpace(os.Getenv("ENGAGE_NVD_API_KEY")),
	}
}

// NewNVDClient allows tests to override base URL.
func NewNVDClient(baseURL string, client *http.Client) NVDClient {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &httpNVDClient{baseURL: baseURL, http: client}
}

func (c *httpNVDClient) FetchCVE(ctx context.Context, cveID string) (*CVEEntry, error) {
	cveID = normalizeCVEID(cveID)
	if cveID == "" {
		return nil, fmt.Errorf("cve_id required")
	}
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("cveId", cveID)
	u.RawQuery = q.Encode()
	body, err := c.get(ctx, u.String())
	if err != nil {
		return nil, err
	}
	entries, err := parseNVDResponse(body)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("CVE %s not found in NVD database", cveID)
	}
	return &entries[0], nil
}

func (c *httpNVDClient) FetchRecent(ctx context.Context, hours int, severityFilter string) ([]CVEEntry, error) {
	if hours <= 0 {
		hours = 24
	}
	end := time.Now().UTC()
	start := end.Add(-time.Duration(hours) * time.Hour)
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("lastModStartDate", start.Format("2006-01-02T15:04:05.000"))
	q.Set("lastModEndDate", end.Format("2006-01-02T15:04:05.000"))
	q.Set("resultsPerPage", "100")
	u.RawQuery = q.Encode()
	body, err := c.get(ctx, u.String())
	if err != nil {
		return nil, err
	}
	entries, err := parseNVDResponse(body)
	if err != nil {
		return nil, err
	}
	allowed := parseSeverityFilter(severityFilter)
	if len(allowed) == 0 {
		return entries, nil
	}
	var out []CVEEntry
	for _, e := range entries {
		if severityAllowed(e.Severity, allowed) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (c *httpNVDClient) get(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("apiKey", c.apiKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NVD API %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func parseNVDResponse(body []byte) ([]CVEEntry, error) {
	var wrap struct {
		Vulnerabilities []struct {
			CVE nvdCVE `json:"cve"`
		} `json:"vulnerabilities"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		return nil, err
	}
	out := make([]CVEEntry, 0, len(wrap.Vulnerabilities))
	for _, v := range wrap.Vulnerabilities {
		if e := v.CVE.toEntry(); e.CVEID != "" {
			out = append(out, e)
		}
	}
	return out, nil
}

type nvdCVE struct {
	ID           string `json:"id"`
	Descriptions []struct {
		Lang  string `json:"lang"`
		Value string `json:"value"`
	} `json:"descriptions"`
	Metrics struct {
		CVSSMetricV31 []struct {
			CVSSData struct {
				BaseScore            float64 `json:"baseScore"`
				BaseSeverity         string  `json:"baseSeverity"`
				AttackVector         string  `json:"attackVector"`
				AttackComplexity     string  `json:"attackComplexity"`
				ExploitabilityScore  float64 `json:"exploitabilityScore"`
			} `json:"cvssData"`
		} `json:"cvssMetricV31"`
		CVSSMetricV30 []struct {
			CVSSData struct {
				BaseScore        float64 `json:"baseScore"`
				BaseSeverity     string  `json:"baseSeverity"`
				AttackVector     string  `json:"attackVector"`
				AttackComplexity string  `json:"attackComplexity"`
			} `json:"cvssData"`
		} `json:"cvssMetricV30"`
		CVSSMetricV2 []struct {
			CVSSData struct {
				BaseScore float64 `json:"baseScore"`
			} `json:"cvssData"`
		} `json:"cvssMetricV2"`
	} `json:"metrics"`
	References []struct {
		URL string `json:"url"`
	} `json:"references"`
}

func (c nvdCVE) toEntry() CVEEntry {
	e := CVEEntry{CVEID: c.ID}
	for _, d := range c.Descriptions {
		if d.Lang == "en" && d.Value != "" {
			e.Description = d.Value
			break
		}
	}
	if e.Description == "" && len(c.Descriptions) > 0 {
		e.Description = c.Descriptions[0].Value
	}
	if len(c.Metrics.CVSSMetricV31) > 0 {
		cv := c.Metrics.CVSSMetricV31[0].CVSSData
		e.CVSSScore = cv.BaseScore
		e.Severity = strings.ToUpper(cv.BaseSeverity)
		e.AttackVector = cv.AttackVector
		e.AttackComplexity = cv.AttackComplexity
	} else if len(c.Metrics.CVSSMetricV30) > 0 {
		cv := c.Metrics.CVSSMetricV30[0].CVSSData
		e.CVSSScore = cv.BaseScore
		e.Severity = strings.ToUpper(cv.BaseSeverity)
		e.AttackVector = cv.AttackVector
		e.AttackComplexity = cv.AttackComplexity
	} else if len(c.Metrics.CVSSMetricV2) > 0 {
		e.CVSSScore = c.Metrics.CVSSMetricV2[0].CVSSData.BaseScore
		e.Severity = cvssV2Severity(e.CVSSScore)
	}
	for _, r := range c.References {
		if r.URL != "" {
			e.References = append(e.References, r.URL)
		}
	}
	if len(e.References) > 5 {
		e.References = e.References[:5]
	}
	return e
}

func cvssV2Severity(score float64) string {
	switch {
	case score >= 9.0:
		return "CRITICAL"
	case score >= 7.0:
		return "HIGH"
	case score >= 4.0:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func normalizeCVEID(id string) string {
	id = strings.TrimSpace(strings.ToUpper(id))
	if strings.HasPrefix(id, "CVE-") {
		return id
	}
	return ""
}

func parseSeverityFilter(filter string) map[string]struct{} {
	filter = strings.TrimSpace(strings.ToUpper(filter))
	if filter == "" || filter == "ALL" {
		return nil
	}
	out := map[string]struct{}{}
	for _, p := range strings.Split(filter, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out[p] = struct{}{}
		}
	}
	return out
}

func severityAllowed(sev string, allowed map[string]struct{}) bool {
	if len(allowed) == 0 {
		return true
	}
	_, ok := allowed[strings.ToUpper(sev)]
	return ok
}
