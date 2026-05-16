package httpserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	domainjob "github.com/butbeautifulv/veil/engage/serve/internal/domain/job"
	domainreport "github.com/butbeautifulv/veil/engage/serve/internal/domain/report"
	"github.com/butbeautifulv/veil/engage/serve/internal/telemetry"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/browser"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/ctf"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/payloads"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/report"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

func Register(mux *http.ServeMux, c *components.APIComponents) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":         true,
			"service":    "veil-engage",
			"tool_count": c.Registry.Count(),
		})
	})

	mux.HandleFunc("GET /api/tools", func(w http.ResponseWriter, r *http.Request) {
		list := c.Tools.List()
		out := make([]map[string]any, 0, len(list))
		for _, s := range list {
			out = append(out, map[string]any{
				"name": s.Name, "category": string(s.Category), "description": s.Description,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"tools": out})
	})

	mux.HandleFunc("POST /api/tools/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		var req contract.ToolRunRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
			return
		}
		sub := subject(r)
		writeJSON(w, http.StatusOK, c.Tools.Run(r.Context(), sub, name, req))
	})

	registerJobs(mux, c)
	registerIntel(mux, c)
	registerWorkflows(mux, c)
	registerCTF(mux, c)
	registerVulnIntel(mux, c)
	registerErrorHandling(mux)
	registerProcessRoutes(mux, c)
	registerBrowser(mux, c)
	registerVisual(mux, c)
	registerFiles(mux, c)
	registerCommand(mux, c)
	registerPayloads(mux, c)
	registerAdmin(mux, c)
	registerPlaybooks(mux, c)
}

func registerPlaybooks(mux *http.ServeMux, c *components.APIComponents) {
	mux.HandleFunc("GET /api/playbooks", func(w http.ResponseWriter, r *http.Request) {
		list, err := workflow.LoadAllPlaybooks(c.CatalogPath)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"playbooks": list})
	})
	mux.HandleFunc("POST /api/playbooks/{name}/run", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		var body struct {
			Target string `json:"target"`
			Async  bool   `json:"async"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
			return
		}
		if strings.TrimSpace(body.Target) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "target required"})
			return
		}
		list, err := workflow.LoadAllPlaybooks(c.CatalogPath)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		pb, ok := workflow.FindPlaybook(list, name)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "playbook not found"})
			return
		}
		if strings.HasPrefix(pb.Workflow, "ctf-") && c.CTF != nil {
			writeJSON(w, http.StatusOK, c.CTF.RunPlaybook(r.Context(), subject(r), pb, body.Target, !body.Async))
			return
		}
		if isBugBountyPlaybook(pb.Workflow, pb.Name) && c.BugBounty != nil {
			writeJSON(w, http.StatusOK, c.BugBounty.RunPlaybook(r.Context(), subject(r), pb.Name, pb.Workflow, body.Target, body.Async, pb.MaxTools))
			return
		}
		if c.Workflows == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "workflows not configured"})
			return
		}
		out := c.Workflows.RunPlaybook(r.Context(), subject(r), pb, body.Target, body.Async)
		writeJSON(w, http.StatusOK, out)
	})
}

func isBugBountyPlaybook(workflow, name string) bool {
	switch workflow {
	case "reconnaissance", "vuln-hunt", "business-logic", "osint", "file-upload", "comprehensive":
		return true
	}
	switch name {
	case "reconnaissance", "vuln-hunt", "business-logic", "osint", "file-upload", "comprehensive":
		return true
	}
	return false
}

func registerJobs(mux *http.ServeMux, c *components.APIComponents) {
	mux.HandleFunc("POST /api/jobs", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Tool       string            `json:"tool"`
			Target     string            `json:"target"`
			Parameters map[string]string `json:"parameters"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
			return
		}
		j, err := c.Jobs.Enqueue(body.Tool, body.Target, subject(r), body.Parameters)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusAccepted, j)
	})
	mux.HandleFunc("GET /api/jobs", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 50
		}
		list, err := c.Jobs.List(domainjob.Status(status), limit)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"jobs": list})
	})
	mux.HandleFunc("GET /api/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		j, ok := c.Jobs.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "job not found"})
			return
		}
		writeJSON(w, http.StatusOK, j)
	})
	mux.HandleFunc("POST /api/jobs/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := c.Jobs.Cancel(id); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		j, _ := c.Jobs.Get(id)
		writeJSON(w, http.StatusOK, j)
	})
}

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

func registerVulnIntel(mux *http.ServeMux, c *components.APIComponents) {
	if c.CVE == nil {
		return
	}
	postJSON(mux, "POST /api/vuln-intel/cve-monitor", func(r *http.Request, body map[string]any) (any, int) {
		return c.CVE.MonitorFromBody(r.Context(), body), http.StatusOK
	})
	postJSON(mux, "POST /api/vuln-intel/exploit-generate", func(r *http.Request, body map[string]any) (any, int) {
		return c.CVE.GenerateExploitFromCVE(r.Context(), body), http.StatusOK
	})
	postJSON(mux, "POST /api/vuln-intel/cve-lookup", func(r *http.Request, body map[string]any) (any, int) {
		id := toString(body["cve_id"])
		if id == "" {
			return map[string]any{"success": false, "error": "cve_id is required"}, http.StatusBadRequest
		}
		return c.CVE.Lookup(r.Context(), id), http.StatusOK
	})
	postJSON(mux, "POST /api/vuln-intel/attack-chains", func(r *http.Request, body map[string]any) (any, int) {
		if c.Intel == nil {
			return map[string]any{"success": false, "error": "intelligence not configured"}, http.StatusServiceUnavailable
		}
		out := c.Intel.DiscoverAttackChains(r.Context(), toString(body["target"]), toString(body["objective"]))
		out["alias_of"] = "/api/intelligence/discover-attack-chains"
		return out, http.StatusOK
	})
	postJSON(mux, "POST /api/vuln-intel/threat-feeds", func(r *http.Request, body map[string]any) (any, int) {
		return map[string]any{
			"alias_of": "/api/vuln-intel/cve-monitor",
			"note":     "threat-feeds merged into cve-monitor (NVD recent CVEs)",
			"result":   c.CVE.MonitorFromBody(r.Context(), body),
		}, http.StatusOK
	})
	postJSON(mux, "POST /api/vuln-intel/zero-day-research", func(r *http.Request, body map[string]any) (any, int) {
		target := toString(body["target"])
		if target == "" {
			target = toString(body["cve_id"])
		}
		if target == "" {
			return map[string]any{
				"success": false,
				"error":   "target or cve_id required",
				"note":    "heuristic stub — no LLM zero-day research in engage",
			}, http.StatusBadRequest
		}
		if strings.HasPrefix(strings.ToUpper(target), "CVE-") {
			return c.CVE.Lookup(r.Context(), target), http.StatusOK
		}
		if c.Intel != nil {
			return c.Intel.DiscoverAttackChains(r.Context(), target, "comprehensive"), http.StatusOK
		}
		return map[string]any{
			"success": true,
			"target":  target,
			"note":    "heuristic stub — use cve-lookup for CVE IDs or discover-attack-chains for targets",
		}, http.StatusOK
	})
}

func registerCTF(mux *http.ServeMux, c *components.APIComponents) {
	if c.CTF == nil {
		return
	}
	postJSON(mux, "POST /api/ctf/create-challenge-workflow", func(r *http.Request, body map[string]any) (any, int) {
		ch := ctf.ChallengeFromBody(body)
		out, err := c.CTF.CreateChallengeWorkflow(ch)
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}, http.StatusBadRequest
		}
		return out, http.StatusOK
	})
	postJSON(mux, "POST /api/ctf/auto-solve-challenge", func(r *http.Request, body map[string]any) (any, int) {
		ch := ctf.ChallengeFromBody(body)
		exec := body["execute_tools"] == nil || body["execute_tools"] == true
		maxSteps := toInt(body["max_steps"], 8)
		out, err := c.CTF.AutoSolve(r.Context(), subject(r), ch, exec, maxSteps)
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}, http.StatusBadRequest
		}
		return out, http.StatusOK
	})
	postJSON(mux, "POST /api/ctf/suggest-tools", func(r *http.Request, body map[string]any) (any, int) {
		desc := toString(body["description"])
		if desc == "" {
			return map[string]any{"success": false, "error": "description is required"}, http.StatusBadRequest
		}
		return c.CTF.SuggestTools(desc, toString(body["category"]), toString(body["target"])), http.StatusOK
	})
	postJSON(mux, "POST /api/ctf/team-strategy", func(r *http.Request, body map[string]any) (any, int) {
		raw, _ := body["challenges"].([]any)
		if len(raw) == 0 {
			return map[string]any{"success": false, "error": "challenges data is required"}, http.StatusBadRequest
		}
		var challenges []ctf.Challenge
		for _, item := range raw {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			challenges = append(challenges, ctf.ChallengeFromBody(m))
		}
		skills := map[string][]string{}
		if ts, ok := body["team_skills"].(map[string]any); ok {
			for member, v := range ts {
				if arr, ok := v.([]any); ok {
					for _, s := range arr {
						if str, ok := s.(string); ok {
							skills[member] = append(skills[member], str)
						}
					}
				}
			}
		}
		return c.CTF.TeamStrategy(challenges, skills), http.StatusOK
	})
	postJSON(mux, "POST /api/ctf/cryptography-solver", func(r *http.Request, body map[string]any) (any, int) {
		text := toString(body["cipher_text"])
		if text == "" {
			return map[string]any{"success": false, "error": "cipher text is required"}, http.StatusBadRequest
		}
		return c.CTF.AnalyzeCrypto(text, toString(body["cipher_type"]), toString(body["key_hint"]),
			toString(body["known_plaintext"]), toString(body["additional_info"])), http.StatusOK
	})
	postJSON(mux, "POST /api/ctf/forensics-analyzer", func(r *http.Request, body map[string]any) (any, int) {
		path := toString(body["file_path"])
		if path == "" {
			return map[string]any{"success": false, "error": "file path is required"}, http.StatusBadRequest
		}
		return c.CTF.AnalyzeForensics(r.Context(), subject(r), path, ctf.ForensicsOptions{
			AnalysisType:       toString(body["analysis_type"]),
			ExtractHidden:      body["extract_hidden"] != false,
			CheckSteganography: body["check_steganography"] != false,
		}), http.StatusOK
	})
	postJSON(mux, "POST /api/ctf/binary-analyzer", func(r *http.Request, body map[string]any) (any, int) {
		path := toString(body["binary_path"])
		if path == "" {
			return map[string]any{"success": false, "error": "binary path is required"}, http.StatusBadRequest
		}
		return c.CTF.AnalyzeBinary(r.Context(), subject(r), path, ctf.BinaryOptions{
			AnalysisDepth:    toString(body["analysis_depth"]),
			CheckProtections: body["check_protections"] != false,
			FindGadgets:      body["find_gadgets"] != false,
		}), http.StatusOK
	})
}

func registerWorkflows(mux *http.ServeMux, c *components.APIComponents) {
	wf := func(path, wfName string) {
		postJSON(mux, "POST "+path, func(r *http.Request, body map[string]any) (any, int) {
			if c.Workflows == nil {
				return map[string]any{"success": false, "error": "workflows not configured"}, http.StatusServiceUnavailable
			}
			return c.Workflows.RunWorkflowWithBody(r.Context(), subject(r), wfName, body), http.StatusOK
		})
	}
	wf("/api/bugbounty/reconnaissance-workflow", "reconnaissance")
	wf("/api/bugbounty/vulnerability-hunting-workflow", "vuln-hunt")
	wf("/api/bugbounty/business-logic-workflow", "business-logic")
	wf("/api/bugbounty/osint-workflow", "osint")
	wf("/api/bugbounty/file-upload-testing", "file-upload")
	wf("/api/bugbounty/comprehensive-assessment", "comprehensive")
}

func registerBrowser(mux *http.ServeMux, c *components.APIComponents) {
	if c.Browser == nil {
		return
	}
	postJSON(mux, "POST /api/browser/inspect", func(r *http.Request, body map[string]any) (any, int) {
		url := toString(body["url"])
		if url == "" {
			url = toString(body["target"])
		}
		if url == "" {
			return map[string]any{"success": false, "error": "url or target required"}, http.StatusBadRequest
		}
		params := map[string]string{}
		for k, v := range body {
			if s, ok := v.(string); ok {
				params[k] = s
			}
		}
		out := c.Browser.Inspect(r.Context(), browserInspectReqFromBody(url, body, params))
		if !out.Success {
			return out, http.StatusOK
		}
		return out, http.StatusOK
	})
}

func browserInspectReqFromBody(url string, body map[string]any, params map[string]string) browser.InspectRequest {
	req := browser.InspectFromParams(url, params)
	if body != nil {
		if v, ok := body["wait_time"]; ok {
			switch n := v.(type) {
			case float64:
				req.WaitTime = int(n)
			case int:
				req.WaitTime = n
			}
		}
		if body["active_tests"] == true {
			req.ActiveTests = true
		}
	}
	return req
}

func registerVisual(mux *http.ServeMux, c *components.APIComponents) {
	mux.HandleFunc("GET /api/visual/scan-progress/{id}", func(w http.ResponseWriter, r *http.Request) {
		if c.Progress == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "progress store not configured"})
			return
		}
		id := r.PathValue("id")
		if sp, ok := c.Progress.Get(id); ok {
			writeJSON(w, http.StatusOK, sp)
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "scan not found"})
	})
	postJSON(mux, "POST /api/visual/summary-report", func(r *http.Request, body map[string]any) (any, int) {
		target, _ := body["target"].(string)
		sections, _ := body["sections"].(map[string]any)
		if sections == nil {
			sections = body
		}
		if raw, ok := body["executive_summary"]; ok {
			sections["executive_summary"] = raw
		}
		findings := parseFindings(body["findings"])
		return report.NewSummary(target, sections, findings), http.StatusOK
	})
	postJSON(mux, "POST /api/visual/vulnerability-card", func(r *http.Request, body map[string]any) (any, int) {
		f := domainreport.Finding{
			Title:       toString(body["title"]),
			Severity:    domainreport.Severity(toString(body["severity"])),
			Description: toString(body["description"]),
			Target:      toString(body["target"]),
			Tool:        toString(body["tool"]),
			Evidence:    toString(body["evidence"]),
		}
		if f.Severity == "" {
			f.Severity = domainreport.SeverityMedium
		}
		return report.NewVulnerabilityCard(f), http.StatusOK
	})
	postJSON(mux, "POST /api/visual/tool-output", func(r *http.Request, body map[string]any) (any, int) {
		return report.ToolOutput{
			Tool:   toString(body["tool"]),
			Target: toString(body["target"]),
			Output: toString(body["output"]),
			OK:     body["success"] == true,
		}, http.StatusOK
	})
	postJSON(mux, "POST /api/visual/export-report", func(r *http.Request, body map[string]any) (any, int) {
		var summary report.SummaryReport
		if raw, ok := body["summary_report"]; ok {
			b, _ := json.Marshal(raw)
			_ = json.Unmarshal(b, &summary)
		}
		if summary.Target == "" {
			target := toString(body["target"])
			sections, _ := body["sections"].(map[string]any)
			if sections == nil {
				sections = map[string]any{}
			}
			findings := parseFindings(body["findings"])
			summary = report.NewSummary(target, sections, findings)
		}
		if summary.Target == "" {
			return map[string]any{"error": "target or summary_report required"}, http.StatusBadRequest
		}
		branding := report.DefaultBranding()
		if raw, ok := body["branding"].(map[string]any); ok {
			branding.Organization = toString(raw["organization"])
			branding.Classification = toString(raw["classification"])
			branding.Footer = toString(raw["footer"])
			branding.LogoURL = toString(raw["logo_url"])
		}
		format := strings.ToLower(toString(body["format"]))
		if format == "" {
			format = "pdf"
		}
		out := map[string]any{"target": summary.Target, "format": format}
		if format == "html" {
			html := report.RenderAssessmentHTML(summary, branding)
			out["size_bytes"] = len(html)
			out["html"] = html
		} else {
			pdfBytes, err := report.RenderPDF(summary, branding)
			if err != nil {
				return map[string]any{"error": err.Error()}, http.StatusInternalServerError
			}
			out["size_bytes"] = len(pdfBytes)
			out["pdf_base64"] = base64.StdEncoding.EncodeToString(pdfBytes)
		}
		if c.Files != nil && body["save_file"] != false {
			fname := toString(body["filename"])
			if fname == "" {
				fname = fmt.Sprintf("assessment-%d.%s", time.Now().Unix(), format)
			}
			var data []byte
			if format == "html" {
				data = []byte(out["html"].(string))
			} else {
				data, _ = base64.StdEncoding.DecodeString(out["pdf_base64"].(string))
			}
			if res, err := c.Files.CreateBytes(fname, data); err == nil {
				out["file"] = res
			}
		}
		return out, http.StatusOK
	})
}

func registerFiles(mux *http.ServeMux, c *components.APIComponents) {
	if c.Files == nil {
		return
	}
	postJSON(mux, "POST /api/files/create", func(r *http.Request, body map[string]any) (any, int) {
		res, err := c.Files.Create(toString(body["filename"]), toString(body["content"]), body["binary"] == true)
		if err != nil {
			return map[string]any{"error": err.Error()}, http.StatusBadRequest
		}
		return res, http.StatusOK
	})
	postJSON(mux, "POST /api/files/modify", func(r *http.Request, body map[string]any) (any, int) {
		res, err := c.Files.Modify(toString(body["filename"]), toString(body["content"]), body["append"] == true)
		if err != nil {
			return map[string]any{"error": err.Error()}, http.StatusBadRequest
		}
		return res, http.StatusOK
	})
	postJSON(mux, "POST /api/files/delete", func(r *http.Request, body map[string]any) (any, int) {
		res, err := c.Files.Delete(toString(body["filename"]))
		if err != nil {
			return map[string]any{"error": err.Error()}, http.StatusBadRequest
		}
		return res, http.StatusOK
	})
	mux.HandleFunc("GET /api/files/list", func(w http.ResponseWriter, r *http.Request) {
		dir := r.URL.Query().Get("directory")
		res, err := c.Files.List(dir)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
}

func registerPayloads(mux *http.ServeMux, c *components.APIComponents) {
	postJSON(mux, "POST /api/payloads/generate", func(r *http.Request, body map[string]any) (any, int) {
		req := payloads.Request{
			Type:     toString(body["type"]),
			Size:     toInt(body["size"], 1024),
			Pattern:  toString(body["pattern"]),
			Filename: toString(body["filename"]),
		}
		if req.Pattern == "" {
			req.Pattern = "A"
		}
		res, err := payloads.Generate(c.Files, req)
		if err != nil {
			return map[string]any{"error": err.Error()}, http.StatusBadRequest
		}
		return res, http.StatusOK
	})
}

func registerCommand(mux *http.ServeMux, c *components.APIComponents) {
	postJSON(mux, "POST /api/command", func(r *http.Request, body map[string]any) (any, int) {
		cmd, _ := body["command"].(string)
		useCache := true
		if body["use_cache"] == false {
			useCache = false
		}
		return c.Command.Run(r.Context(), cmd, useCache, c.Cache), http.StatusOK
	})
}

func registerAdmin(mux *http.ServeMux, c *components.APIComponents) {
	mux.HandleFunc("GET /api/cache/stats", func(w http.ResponseWriter, r *http.Request) {
		if c.Cache != nil {
			writeJSON(w, http.StatusOK, c.Cache.Stats())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"entries": 0})
	})
	mux.HandleFunc("POST /api/cache/clear", func(w http.ResponseWriter, r *http.Request) {
		cleared := 0
		if c.Cache != nil {
			cleared = c.Cache.Clear()
		}
		writeJSON(w, http.StatusOK, map[string]any{"cleared": cleared})
	})
	mux.HandleFunc("GET /api/telemetry", func(w http.ResponseWriter, r *http.Request) {
		out := map[string]any{
			"uptime_sec":      int(time.Since(c.StartedAt).Seconds()),
			"tools_enabled":   len(c.Tools.List()),
			"processes_total": len(c.Processes.List()),
		}
		running := 0
		for _, p := range c.Processes.List() {
			if p.Status == "running" {
				running++
			}
		}
		out["processes_running"] = running
		if c.Cache != nil {
			stats := c.Cache.Stats()
			out["cache_entries"] = stats["entries"]
			if entries, ok := stats["entries"].(int); ok {
				telemetry.SetCacheEntries(entries)
			}
		}
		if c.Jobs != nil {
			if n, err := c.Jobs.CountByStatus(domainjob.StatusPending); err == nil {
				out["jobs_pending"] = n
				telemetry.SetJobsPending(n)
			}
			if n, err := c.Jobs.CountByStatus(domainjob.StatusRunning); err == nil {
				out["jobs_running"] = n
			}
		}
		writeJSON(w, http.StatusOK, out)
	})
	mux.HandleFunc("GET /api/processes/list", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"processes": c.Processes.List()})
	})
	mux.HandleFunc("GET /api/processes/status/{pid}", func(w http.ResponseWriter, r *http.Request) {
		pid, _ := strconv.Atoi(r.PathValue("pid"))
		rec, ok := c.Processes.Get(pid)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, rec)
	})
	mux.HandleFunc("POST /api/processes/terminate/{pid}", func(w http.ResponseWriter, r *http.Request) {
		pid, _ := strconv.Atoi(r.PathValue("pid"))
		if err := c.Processes.Terminate(r.Context(), pid); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"terminated": pid})
	})
	mux.HandleFunc("POST /api/processes/pause/{pid}", func(w http.ResponseWriter, r *http.Request) {
		pid, _ := strconv.Atoi(r.PathValue("pid"))
		if err := c.Processes.Pause(pid); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"paused": pid})
	})
	mux.HandleFunc("POST /api/processes/resume/{pid}", func(w http.ResponseWriter, r *http.Request) {
		pid, _ := strconv.Atoi(r.PathValue("pid"))
		if err := c.Processes.Resume(pid); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"resumed": pid})
	})
	mux.HandleFunc("GET /api/processes/dashboard", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, c.Processes.Dashboard())
	})
	mux.HandleFunc("GET /api/audit/recent", func(w http.ResponseWriter, r *http.Request) {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if c.AuditReader == nil {
			writeJSON(w, http.StatusOK, map[string]any{"events": []any{}})
			return
		}
		events, err := c.AuditReader.Recent(limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"events": events})
	})
	mux.HandleFunc("GET /api/audit/export", func(w http.ResponseWriter, r *http.Request) {
		if c.AuditReader == nil {
			http.Error(w, "audit store not configured", http.StatusServiceUnavailable)
			return
		}
		var since time.Time
		if raw := strings.TrimSpace(r.URL.Query().Get("since")); raw != "" {
			if t, err := time.Parse(time.RFC3339, raw); err == nil {
				since = t
			}
		}
		data, err := c.AuditReader.ExportNDJSON(since)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
	postJSON(mux, "POST /api/audit/export-webhook", func(r *http.Request, body map[string]any) (any, int) {
		url := toString(body["url"])
		if url == "" {
			url = c.AuditWebhookURL
		}
		if url == "" {
			return map[string]any{"error": "webhook url required"}, http.StatusBadRequest
		}
		if c.AuditReader == nil {
			return map[string]any{"error": "audit store not configured"}, http.StatusServiceUnavailable
		}
		events, err := c.AuditReader.Recent(500)
		if err != nil {
			return map[string]any{"error": err.Error()}, http.StatusInternalServerError
		}
		secret := c.AuditWebhookSecret
		if s := toString(body["secret"]); s != "" {
			secret = s
		}
		if err := audit.ExportWebhook(r.Context(), url, secret, events); err != nil {
			return map[string]any{"error": err.Error()}, http.StatusBadGateway
		}
		return map[string]any{"exported": len(events), "url": url}, http.StatusOK
	})
	if c.MetricsEnabled {
		mux.Handle("GET /metrics", telemetry.Handler())
	}
}

func parseFindings(raw any) []domainreport.Finding {
	if raw == nil {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var findings []domainreport.Finding
	if err := json.Unmarshal(b, &findings); err != nil {
		return nil
	}
	return findings
}

func postJSON(mux *http.ServeMux, pattern string, fn func(*http.Request, map[string]any) (any, int)) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		if body == nil {
			body = map[string]any{}
		}
		res, code := fn(r, body)
		writeJSON(w, code, res)
	})
}

func subject(r *http.Request) string {
	if sub, ok := auth.SubjectFromContext(r.Context()); ok {
		return sub.Sub
	}
	return ""
}

func toInt(v any, def int) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return def
	}
}

func toBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		switch strings.ToLower(strings.TrimSpace(t)) {
		case "1", "true", "yes", "on":
			return true
		}
	case float64:
		return t != 0
	case int:
		return t != 0
	}
	return false
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
