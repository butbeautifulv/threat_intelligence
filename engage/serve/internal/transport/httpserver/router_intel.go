package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

func registerIntel(mux *http.ServeMux, c *components.APIComponents) {
	postJSON(mux, "POST /api/intelligence/analyze-target", func(r *http.Request, body map[string]any) (any, int) {
		var req contract.AnalyzeTargetRequest
		b, _ := json.Marshal(body)
		_ = json.Unmarshal(b, &req)
		return c.Intel.AnalyzeTarget(r.Context(), req), http.StatusOK
	})
	postJSON(mux, "POST /api/intelligence/select-tools", func(r *http.Request, body map[string]any) (any, int) {
		tt, _ := body["target_type"].(string)
		obj, _ := body["objective"].(string)
		return map[string]any{"tools": c.Intel.SelectTools(r.Context(), tt, obj)}, http.StatusOK
	})
	postJSON(mux, "POST /api/intelligence/optimize-parameters", func(r *http.Request, body map[string]any) (any, int) {
		tt, _ := body["target_type"].(string)
		toolName, _ := body["tool"].(string)
		params, _ := body["parameters"].(map[string]any)
		pm := map[string]string{}
		for k, v := range params {
			pm[k] = toString(v)
		}
		return map[string]any{"parameters": c.Intel.OptimizeParameters(tt, toolName, pm)}, http.StatusOK
	})
	postJSON(mux, "POST /api/intelligence/create-attack-chain", func(r *http.Request, body map[string]any) (any, int) {
		target, _ := body["target"].(string)
		obj, _ := body["objective"].(string)
		return c.Intel.CreateAttackChain(r.Context(), target, obj), http.StatusOK
	})
	postJSON(mux, "POST /api/intelligence/smart-scan", func(r *http.Request, body map[string]any) (any, int) {
		target, _ := body["target"].(string)
		obj, _ := body["objective"].(string)
		maxTools := toInt(body["max_tools"], 5)
		async, _ := body["async"].(bool)
		return c.Workflows.SmartScan(r.Context(), subject(r), workflow.SmartScanRequest{
			Target: target, Objective: obj, MaxTools: maxTools, Async: async,
			RateLimitCheck: toBool(body["rate_limit_check"]),
		}), http.StatusOK
	})
	postJSON(mux, "POST /api/intelligence/assessment-report", func(r *http.Request, body map[string]any) (any, int) {
		target, _ := body["target"].(string)
		obj, _ := body["objective"].(string)
		maxTools := toInt(body["max_tools"], 5)
		return c.Workflows.AssessmentReport(r.Context(), subject(r), workflow.SmartScanRequest{
			Target: target, Objective: obj, MaxTools: maxTools, Async: false,
		}), http.StatusOK
	})
	postJSON(mux, "POST /api/intelligence/technology-detection", func(r *http.Request, body map[string]any) (any, int) {
		target, _ := body["target"].(string)
		return c.Intel.TechnologyDetection(r.Context(), target), http.StatusOK
	})
	postJSON(mux, "POST /api/intelligence/comprehensive-api-audit", func(r *http.Request, body map[string]any) (any, int) {
		return c.Intel.ComprehensiveAPIAudit(r.Context(), subject(r), intelligence.ComprehensiveAPIAuditRequest{
			BaseURL:         toString(body["base_url"]),
			SchemaURL:       toString(body["schema_url"]),
			JWTToken:        toString(body["jwt_token"]),
			GraphQLEndpoint: toString(body["graphql_endpoint"]),
		}), http.StatusOK
	})
	postJSON(mux, "POST /api/intelligence/correlate-threat", func(r *http.Request, body map[string]any) (any, int) {
		return c.Intel.CorrelateThreatIntelligence(r.Context(), toString(body["target"]), toString(body["indicators"])), http.StatusOK
	})
	postJSON(mux, "POST /api/intelligence/target-graph", func(r *http.Request, body map[string]any) (any, int) {
		return c.Intel.TargetGraph(r.Context(), toString(body["target"]), toString(body["indicators"])), http.StatusOK
	})
	mux.HandleFunc("GET /api/intelligence/target-graph", func(w http.ResponseWriter, r *http.Request) {
		out := c.Intel.TargetGraph(r.Context(), r.URL.Query().Get("target"), r.URL.Query().Get("indicators"))
		writeJSON(w, http.StatusOK, out)
	})
	postJSON(mux, "POST /api/intelligence/target-timeline", func(r *http.Request, body map[string]any) (any, int) {
		return c.Intel.TargetTimeline(r.Context(), intelligence.TargetTimelineRequest{
			Target:       toString(body["target"]),
			Limit:        toInt(body["limit"], 50),
			IncludeGraph: body["include_graph"] == nil || body["include_graph"] == true,
		}), http.StatusOK
	})
	mux.HandleFunc("GET /api/intelligence/target-timeline", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		includeGraph := r.URL.Query().Get("include_graph") != "false"
		out := c.Intel.TargetTimeline(r.Context(), intelligence.TargetTimelineRequest{
			Target: target, Limit: limit, IncludeGraph: includeGraph,
		})
		writeJSON(w, http.StatusOK, out)
	})
	postJSON(mux, "POST /api/intelligence/discover-attack-chains", func(r *http.Request, body map[string]any) (any, int) {
		return c.Intel.DiscoverAttackChains(r.Context(), toString(body["target"]), toString(body["objective"])), http.StatusOK
	})
	postJSON(mux, "POST /api/intelligence/execute-attack-chain", func(r *http.Request, body map[string]any) (any, int) {
		return c.Intel.ExecuteAttackChain(r.Context(), subject(r), toString(body["target"]), toString(body["objective"]), toBool(body["parallel"])), http.StatusOK
	})
}
