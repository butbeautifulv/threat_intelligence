package report

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	wkhtmlLookPath  = exec.LookPath
	wkhtmlMkdirTemp = os.MkdirTemp
	wkhtmlWriteFile = os.WriteFile
	wkhtmlReadFile  = os.ReadFile
	wkhtmlCommand   = exec.Command
)

// RenderPDF chooses gofpdf or wkhtmltopdf based on ENGAGE_PDF_ENGINE.
func RenderPDF(summary SummaryReport, branding ...Branding) ([]byte, error) {
	engine := strings.ToLower(strings.TrimSpace(os.Getenv("ENGAGE_PDF_ENGINE")))
	if engine == "wkhtml" || engine == "wkhtmltopdf" {
		return renderPDFWkhtml(summary, branding...)
	}
	return ToPDF(summary, branding...)
}

func renderPDFWkhtml(summary SummaryReport, branding ...Branding) ([]byte, error) {
	if _, err := wkhtmlLookPath("wkhtmltopdf"); err != nil {
		return nil, fmt.Errorf("wkhtmltopdf not found on PATH")
	}
	b := DefaultBranding()
	if len(branding) > 0 {
		b = branding[0]
	}
	htmlBytes := []byte(RenderAssessmentHTML(summary, b))
	dir, err := wkhtmlMkdirTemp("", "engage-pdf-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	htmlPath := filepath.Join(dir, "report.html")
	pdfPath := filepath.Join(dir, "report.pdf")
	if err := wkhtmlWriteFile(htmlPath, htmlBytes, 0o600); err != nil {
		return nil, err
	}
	cmd := wkhtmlCommand("wkhtmltopdf", "--quiet", htmlPath, pdfPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("wkhtmltopdf: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return wkhtmlReadFile(pdfPath)
}
