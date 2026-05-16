package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	domainreport "github.com/butbeautifulv/veil/engage/serve/internal/domain/report"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/report"
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
	registerVisual(mux, c)
	registerAdmin(mux, c)
}

func registerJobs(mux *http.ServeMux, c *components.APIComponents) {
	mux.HandleFunc("POST /api/jobs", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Tool   string `json:"tool"`
			Target string `json:"target"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
			return
		}
		j, err := c.Jobs.Enqueue(body.Tool, body.Target, subject(r))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusAccepted, j)
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
		return c.Workflows.Comprehensive(r.Context(), subject(r), target), http.StatusOK
	})
	postJSON(mux, "POST /api/intelligence/technology-detection", func(r *http.Request, body map[string]any) (any, int) {
		target, _ := body["target"].(string)
		return c.Intel.AnalyzeTarget(r.Context(), contract.AnalyzeTargetRequest{Target: target}), http.StatusOK
	})
}

func registerWorkflows(mux *http.ServeMux, c *components.APIComponents) {
	wf := func(path, wfName string) {
		postJSON(mux, "POST "+path, func(r *http.Request, body map[string]any) (any, int) {
			target, _ := body["target"].(string)
			return c.Workflows.RunWorkflow(r.Context(), subject(r), wfName, target), http.StatusOK
		})
	}
	wf("/api/bugbounty/reconnaissance-workflow", "reconnaissance")
	wf("/api/bugbounty/vulnerability-hunting-workflow", "vuln-hunt")
	wf("/api/bugbounty/business-logic-workflow", "business-logic")
	wf("/api/bugbounty/osint-workflow", "osint")
	wf("/api/bugbounty/file-upload-testing", "file-upload")
	wf("/api/bugbounty/comprehensive-assessment", "comprehensive")
}

func registerVisual(mux *http.ServeMux, c *components.APIComponents) {
	postJSON(mux, "POST /api/visual/summary-report", func(r *http.Request, body map[string]any) (any, int) {
		target, _ := body["target"].(string)
		sections, _ := body["sections"].(map[string]any)
		if sections == nil {
			sections = body
		}
		return report.NewSummary(target, sections, nil), http.StatusOK
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
	_ = c
}

func registerAdmin(mux *http.ServeMux, c *components.APIComponents) {
	mux.HandleFunc("GET /api/cache/stats", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"entries": 0})
	})
	mux.HandleFunc("POST /api/cache/clear", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"cleared": true})
	})
	mux.HandleFunc("GET /api/telemetry", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"uptime": "ok", "tools_enabled": len(c.Tools.List())})
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
	mux.HandleFunc("GET /api/processes/dashboard", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, c.Processes.Dashboard())
	})
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
