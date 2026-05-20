package report

import (
	"bytes"
	"fmt"
	"io"

	"github.com/jung-kurt/gofpdf"
)

var pdfWriteOutput = func(pdf *gofpdf.Fpdf, w io.Writer) error { return pdf.Output(w) }

// ToPDF renders a summary report as a PDF document with optional branding.
func ToPDF(summary SummaryReport, branding ...Branding) ([]byte, error) {
	b := DefaultBranding()
	if len(branding) > 0 {
		b = branding[0]
	}
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	title := b.Organization + " — Assessment Report"
	pdf.Cell(40, 10, title)
	pdf.Ln(8)
	if b.Classification != "" {
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(180, 100, 0)
		pdf.Cell(40, 6, b.Classification)
		pdf.SetTextColor(0, 0, 0)
		pdf.Ln(8)
	}
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(40, 8, fmt.Sprintf("Target: %s", summary.Target))
	pdf.Ln(8)
	pdf.Cell(40, 8, fmt.Sprintf("Generated: %s", summary.Generated.Format("2006-01-02 15:04 UTC")))
	pdf.Ln(10)
	br := severityMap(summary.Sections["severity_breakdown"])
	if len(br) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(40, 8, "Severity breakdown")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 10)
		for _, sev := range []string{"critical", "high", "medium", "low", "info"} {
			pdf.Cell(40, 6, fmt.Sprintf("  %s: %d", sev, br[sev]))
			pdf.Ln(6)
		}
	}
	if len(summary.Findings) > 0 {
		pdf.Ln(4)
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(40, 8, fmt.Sprintf("Findings (%d)", len(summary.Findings)))
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 9)
		for i, f := range summary.Findings {
			if i >= 50 {
				pdf.Cell(40, 6, fmt.Sprintf("  ... and %d more", len(summary.Findings)-50))
				break
			}
			line := fmt.Sprintf("[%s] %s", f.Severity, f.Title)
			pdf.MultiCell(0, 5, line, "", "", false)
		}
	}
	if b.Footer != "" {
		pdf.Ln(6)
		pdf.SetFont("Arial", "I", 8)
		pdf.MultiCell(0, 4, b.Footer, "", "", false)
	}
	var buf bytes.Buffer
	if err := pdfWriteOutput(pdf, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func severityMap(v any) map[string]int {
	if m, ok := v.(map[string]int); ok {
		return m
	}
	if m, ok := v.(map[string]any); ok {
		out := map[string]int{}
		for k, val := range m {
			switch t := val.(type) {
			case float64:
				out[k] = int(t)
			case int:
				out[k] = t
			}
		}
		return out
	}
	return nil
}
