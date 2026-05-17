package report

import (
	"bytes"
	"fmt"
	"html"
	"strings"
	"time"
)

// RenderAssessmentHTML produces a branded HTML assessment report.
func RenderAssessmentHTML(summary SummaryReport, branding Branding) string {
	if branding.Organization == "" {
		branding = DefaultBranding()
	}
	var buf bytes.Buffer
	buf.WriteString("<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>")
	buf.WriteString(html.EscapeString(summary.Target))
	buf.WriteString(" — Assessment</title><style>")
	buf.WriteString(`body{font-family:system-ui,sans-serif;margin:2rem;color:#1a1a1a}
header{border-bottom:2px solid #2563eb;padding-bottom:1rem;margin-bottom:1.5rem}
h1{margin:0;font-size:1.5rem} .meta{color:#555;font-size:0.9rem}
table{border-collapse:collapse;width:100%;margin-top:1rem}
th,td{border:1px solid #ddd;padding:0.5rem;text-align:left}
th{background:#f3f4f6} .banner{background:#fef3c7;padding:0.25rem 0.5rem;display:inline-block;font-size:0.75rem}
footer{margin-top:2rem;font-size:0.8rem;color:#666}`)
	buf.WriteString("</style></head><body><header>")
	if branding.LogoURL != "" {
		buf.WriteString(`<img src="`)
		buf.WriteString(html.EscapeString(branding.LogoURL))
		buf.WriteString(`" alt="logo" style="max-height:48px;margin-bottom:0.5rem"/><br/>`)
	}
	if branding.Classification != "" {
		buf.WriteString(`<span class="banner">`)
		buf.WriteString(html.EscapeString(branding.Classification))
		buf.WriteString(`</span> `)
	}
	buf.WriteString("<h1>")
	buf.WriteString(html.EscapeString(branding.Organization))
	buf.WriteString(" — Security Assessment</h1>")
	buf.WriteString(`<p class="meta">Target: <strong>`)
	buf.WriteString(html.EscapeString(summary.Target))
	buf.WriteString("</strong> · Generated: ")
	buf.WriteString(html.EscapeString(summary.Generated.Format(time.RFC3339)))
	buf.WriteString("</p></header>")

	buf.WriteString("<h2>Executive summary</h2><p>")
	findings := len(summary.Findings)
	buf.WriteString(fmt.Sprintf("Assessment completed with <strong>%d</strong> recorded finding(s).</p>", findings))

	if br := severityMap(summary.Sections["severity_breakdown"]); len(br) > 0 {
		buf.WriteString("<h2>Severity breakdown</h2><table><tr><th>Severity</th><th>Count</th></tr>")
		for _, sev := range []string{"critical", "high", "medium", "low", "info"} {
			if br[sev] > 0 {
				buf.WriteString("<tr><td>")
				buf.WriteString(html.EscapeString(sev))
				buf.WriteString("</td><td>")
				buf.WriteString(fmt.Sprintf("%d", br[sev]))
				buf.WriteString("</td></tr>")
			}
		}
		buf.WriteString("</table>")
	}

	if len(summary.Findings) > 0 {
		buf.WriteString("<h2>Findings</h2><table><tr><th>Severity</th><th>Title</th><th>Tool</th></tr>")
		for i, f := range summary.Findings {
			if i >= 100 {
				break
			}
			buf.WriteString("<tr><td>")
			buf.WriteString(html.EscapeString(string(f.Severity)))
			buf.WriteString("</td><td>")
			buf.WriteString(html.EscapeString(f.Title))
			buf.WriteString("</td><td>")
			buf.WriteString(html.EscapeString(f.Tool))
			buf.WriteString("</td></tr>")
		}
		buf.WriteString("</table>")
	}

	if rec, ok := summary.Sections["recommendations"].([]any); ok && len(rec) > 0 {
		buf.WriteString("<h2>Recommendations</h2><ul>")
		for _, r := range rec {
			if s, ok := r.(string); ok && strings.TrimSpace(s) != "" {
				buf.WriteString("<li>")
				buf.WriteString(html.EscapeString(s))
				buf.WriteString("</li>")
			}
		}
		buf.WriteString("</ul>")
	}

	buf.WriteString("<footer>")
	if branding.Footer != "" {
		buf.WriteString(html.EscapeString(branding.Footer))
	}
	buf.WriteString("</footer></body></html>")
	return buf.String()
}
