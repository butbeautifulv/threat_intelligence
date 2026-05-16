package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
)

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
