package httpserver

import (
	"net/http"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
)

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
