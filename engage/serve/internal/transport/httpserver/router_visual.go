package httpserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
	"github.com/butbeautifulv/veil/pkg/report"
)

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
