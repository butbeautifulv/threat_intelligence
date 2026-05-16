package intelligence

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// ComprehensiveAPIAuditRequest mirrors legacy comprehensive_api_audit inputs.
type ComprehensiveAPIAuditRequest struct {
	BaseURL          string
	SchemaURL        string
	JWTToken         string
	GraphQLEndpoint  string
}

// ComprehensiveAPIAudit runs phased API security checks (HexStrike aggregate parity).
func (s *Service) ComprehensiveAPIAudit(ctx context.Context, subject string, req ComprehensiveAPIAuditRequest) map[string]any {
	base := strings.TrimSpace(req.BaseURL)
	out := map[string]any{
		"base_url":              base,
		"audit_timestamp":       time.Now().UTC().Unix(),
		"tests_performed":       []string{},
		"total_vulnerabilities": 0,
		"recommendations": []string{
			"Implement proper authentication and authorization",
			"Use HTTPS for all API communications",
			"Validate and sanitize all input parameters",
			"Implement rate limiting and request throttling",
			"Add comprehensive logging and monitoring",
		},
	}
	if base == "" {
		out["success"] = false
		out["error"] = "base_url required"
		return out
	}

	performed := []string{}
	findings := 0

	// Phase 1: endpoint discovery / probe
	discovery := s.phaseAPIDiscovery(ctx, subject, base)
	out["api_discovery"] = discovery
	if discovery["success"] == true {
		performed = append(performed, "api_discovery")
		findings += toInt(discovery["findings_count"], 0)
	}

	// Phase 2: schema analysis
	if u := strings.TrimSpace(req.SchemaURL); u != "" {
		schema := s.phaseSchemaAnalysis(ctx, subject, u)
		out["schema_analysis"] = schema
		if schema["success"] == true {
			performed = append(performed, "schema_analysis")
			findings += toInt(schema["findings_count"], 0)
		}
	}

	// Phase 3: JWT analysis
	if tok := strings.TrimSpace(req.JWTToken); tok != "" {
		jwtRes := JWTAnalysis(tok)
		out["jwt_analysis"] = jwtRes
		if jwtRes["success"] == true {
			performed = append(performed, "jwt_analysis")
			findings += toInt(jwtRes["findings_count"], 0)
		}
	}

	// Phase 4: GraphQL
	if gq := strings.TrimSpace(req.GraphQLEndpoint); gq != "" {
		if toolRes, ok := s.runCatalogTool(ctx, subject, "graphql_scanner", gq, map[string]string{
			"graphql_endpoint": gq,
		}); ok {
			out["graphql_scanning"] = map[string]any{
				"success": toolRes.Success, "tool_run": toolRes, "findings_count": countToolFindings(toolRes),
			}
			if toolRes.Success {
				performed = append(performed, "graphql_scanning")
				findings += countToolFindings(toolRes)
			}
		} else {
			gql := s.phaseGraphQL(ctx, gq)
			out["graphql_scanning"] = gql
			if gql["success"] == true {
				performed = append(performed, "graphql_scanning")
				findings += toInt(gql["findings_count"], 0)
			}
		}
	}

	out["tests_performed"] = performed
	out["total_vulnerabilities"] = findings
	out["success"] = len(performed) > 0
	out["summary"] = map[string]any{
		"phases":   len(performed),
		"findings": findings,
	}
	return out
}

func (s *Service) phaseAPIDiscovery(ctx context.Context, subject, baseURL string) map[string]any {
	res := map[string]any{"success": false, "findings_count": 0}
	if toolRes, ok := s.runCatalogTool(ctx, subject, "api_fuzzer", baseURL, map[string]string{
		"base_url": baseURL,
	}); ok {
		res["success"] = toolRes.Success
		res["tool_run"] = toolRes
		res["findings_count"] = countToolFindings(toolRes)
		return res
	}
	if toolRes, ok := s.runCatalogTool(ctx, subject, "httpx_probe", baseURL, nil); ok {
		res["success"] = toolRes.Success
		res["tool_run"] = toolRes
		res["findings_count"] = countToolFindings(toolRes)
		return res
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normalizeURL(baseURL), nil)
	if err != nil {
		res["error"] = err.Error()
		return res
	}
	req.Header.Set("User-Agent", "veil-engage/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		res["error"] = err.Error()
		return res
	}
	defer resp.Body.Close()
	res["success"] = true
	res["status_code"] = resp.StatusCode
	findings := 0
	if resp.StatusCode >= 500 {
		findings++
		res["issues"] = []string{"server_error_status"}
	}
	if resp.Header.Get("Access-Control-Allow-Origin") == "*" {
		findings++
		issues, _ := res["issues"].([]string)
		res["issues"] = append(issues, "permissive_cors")
	}
	res["findings_count"] = findings
	return res
}

func (s *Service) phaseSchemaAnalysis(ctx context.Context, subject, schemaURL string) map[string]any {
	if toolRes, ok := s.runCatalogTool(ctx, subject, "api_schema_analyzer", schemaURL, map[string]string{
		"schema_url": schemaURL,
	}); ok {
		return map[string]any{
			"success":        toolRes.Success,
			"schema_url":     schemaURL,
			"tool_run":       toolRes,
			"findings_count": countToolFindings(toolRes),
		}
	}
	res := map[string]any{"success": false, "findings_count": 0, "schema_url": schemaURL}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, schemaURL, nil)
	if err != nil {
		res["error"] = err.Error()
		return res
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		res["error"] = err.Error()
		return res
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	low := strings.ToLower(string(body))
	res["success"] = resp.StatusCode == http.StatusOK
	findings := 0
	issues := []string{}
	if !strings.Contains(low, "openapi") && !strings.Contains(low, "swagger") {
		issues = append(issues, "missing_openapi_marker")
		findings++
	}
	if strings.Contains(low, `"security":[]`) || !strings.Contains(low, "security") {
		issues = append(issues, "weak_or_missing_security_scheme")
		findings++
	}
	res["issues"] = issues
	res["findings_count"] = findings
	return res
}

// JWTAnalysis decodes a JWT and returns heuristic security findings.
func JWTAnalysis(token string) map[string]any {
	res := map[string]any{"success": true, "findings_count": 0}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		res["success"] = false
		res["error"] = "invalid jwt format"
		return res
	}
	headerJSON, _ := base64.RawURLEncoding.DecodeString(parts[0])
	payloadJSON, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var header, payload map[string]any
	_ = json.Unmarshal(headerJSON, &header)
	_ = json.Unmarshal(payloadJSON, &payload)
	res["header"] = header
	res["payload"] = payload
	findings := 0
	issues := []string{}
	if alg, _ := header["alg"].(string); strings.EqualFold(alg, "none") {
		issues = append(issues, "alg_none")
		findings++
	}
	if alg, _ := header["alg"].(string); strings.EqualFold(alg, "HS256") {
		issues = append(issues, "symmetric_alg_review")
		findings++
	}
	res["issues"] = issues
	res["findings_count"] = findings
	return res
}

func (s *Service) phaseGraphQL(ctx context.Context, endpoint string) map[string]any {
	res := map[string]any{"success": false, "findings_count": 0, "endpoint": endpoint}
	query := `{"query":"{ __schema { queryType { name } } }"}`
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(query))
	if err != nil {
		res["error"] = err.Error()
		return res
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		res["error"] = err.Error()
		return res
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 65536))
	res["success"] = resp.StatusCode == http.StatusOK
	findings := 0
	issues := []string{}
	if strings.Contains(string(body), "__schema") {
		issues = append(issues, "introspection_enabled")
		findings++
	}
	res["issues"] = issues
	res["findings_count"] = findings
	return res
}

func toInt(v any, def int) int {
	switch t := v.(type) {
	case int:
		return t
	case float64:
		return int(t)
	default:
		return def
	}
}
