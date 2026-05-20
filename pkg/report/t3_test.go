package report

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	domain "github.com/butbeautifulv/veil/pkg/engage/domain/report"
	"github.com/jung-kurt/gofpdf"
)

func TestNewSummary_and_VulnerabilityCard(t *testing.T) {
	f := domain.Finding{Title: "XSS", Severity: domain.SeverityHigh, Target: "https://x", Tool: "dalfox"}
	sum := NewSummary("https://x", map[string]any{"k": 1}, []domain.Finding{f})
	if sum.ReportType != "summary" || sum.Target != "https://x" || len(sum.Findings) != 1 {
		t.Fatalf("summary: %+v", sum)
	}
	if sum.Generated.IsZero() {
		t.Fatal("expected generated time")
	}
	card := NewVulnerabilityCard(f)
	if card.Title != "XSS" || card.Severity != "high" {
		t.Fatalf("card: %+v", card)
	}
}

func TestBuildExecutiveSummary_riskBranches(t *testing.T) {
	if got := riskPostureFromFindings(map[string]int{"critical": 1}, ""); got != "critical" {
		t.Fatalf("critical: %s", got)
	}
	if got := riskPostureFromFindings(map[string]int{"medium": 1}, ""); got != "medium" {
		t.Fatalf("medium: %s", got)
	}
	if got := riskPostureFromFindings(map[string]int{}, "HIGH"); got != "high" {
		t.Fatalf("profile: %s", got)
	}
	if got := riskPostureFromFindings(map[string]int{"low": 1}, ""); got != "low" {
		t.Fatalf("low: %s", got)
	}
	if got := riskPostureFromFindings(map[string]int{}, ""); got != "low" {
		t.Fatalf("default: %s", got)
	}
}

func TestBuildExecutiveSummary_toolsExecutedAny(t *testing.T) {
	es := BuildExecutiveSummary("t", map[string]any{
		"tools_executed": []any{map[string]any{"tool": "nmap"}},
	}, nil, "", nil)
	if es.ToolsExecuted != 1 {
		t.Fatalf("tools %d", es.ToolsExecuted)
	}
}

func TestTopRiskTitles_skipsEmpty(t *testing.T) {
	out := topRiskTitles([]domain.Finding{{Title: ""}, {Title: "A", Severity: domain.SeverityLow}}, 5)
	if len(out) != 1 || out[0] != "A" {
		t.Fatalf("got %v", out)
	}
}

func TestTopRiskTitles_sortsBySeverityThenTitle(t *testing.T) {
	findings := []domain.Finding{
		{Title: "Bravo", Severity: domain.SeverityHigh},
		{Title: "Alpha", Severity: domain.SeverityHigh},
		{Title: "Critical issue", Severity: domain.SeverityCritical},
	}
	out := topRiskTitles(findings, 5)
	if len(out) != 3 || out[0] != "Critical issue" || out[1] != "Alpha" || out[2] != "Bravo" {
		t.Fatalf("got %v", out)
	}
}

func TestTopRiskTitles_limitsCount(t *testing.T) {
	findings := make([]domain.Finding, 8)
	for i := range findings {
		findings[i] = domain.Finding{Title: fmt.Sprintf("Finding %d", i), Severity: domain.SeverityLow}
	}
	out := topRiskTitles(findings, 5)
	if len(out) != 5 {
		t.Fatalf("got %d titles", len(out))
	}
}

func TestRecommendations_noFindings(t *testing.T) {
	rec := recommendations(map[string]int{}, nil)
	if len(rec) != 1 {
		t.Fatalf("got %v", rec)
	}
}

func TestRecommendations_mediumOnly(t *testing.T) {
	rec := recommendations(map[string]int{"medium": 2}, nil)
	if len(rec) != 1 || !strings.Contains(rec[0], "medium-severity") {
		t.Fatalf("got %v", rec)
	}
}

func TestRenderPDF_wkhtmlMissingBinary(t *testing.T) {
	t.Setenv("ENGAGE_PDF_ENGINE", "wkhtml")
	summary := SummaryReport{
		ReportType: "summary",
		Target:     "https://example.com",
		Generated:  time.Now().UTC(),
	}
	_, err := RenderPDF(summary)
	if err == nil {
		t.Fatal("expected error when wkhtmltopdf missing")
	}
	if !strings.Contains(err.Error(), "wkhtmltopdf not found") {
		t.Fatalf("err: %v", err)
	}
}

func TestRenderPDF_wkhtmlSuccess(t *testing.T) {
	restore := mockWkhtmlHooks(t, wkhtmlHooks{
		lookPath: func(string) (string, error) { return "/usr/bin/wkhtmltopdf", nil },
		command: func(name string, arg ...string) *exec.Cmd {
			pdfPath := arg[len(arg)-1]
			return exec.Command("sh", "-c", fmt.Sprintf("echo -n '%%PDF-1.4' > %q", pdfPath))
		},
	})
	defer restore()
	t.Setenv("ENGAGE_PDF_ENGINE", "wkhtmltopdf")
	summary := SummaryReport{Target: "https://example.com", Generated: time.Now().UTC()}
	b, err := RenderPDF(summary, Branding{Organization: "Acme"})
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 4 || string(b[:4]) != "%PDF" {
		t.Fatalf("pdf bytes: %q", b)
	}
}

func TestRenderPDF_wkhtmlMkdirTempError(t *testing.T) {
	restore := mockWkhtmlHooks(t, wkhtmlHooks{
		lookPath:  func(string) (string, error) { return "/usr/bin/wkhtmltopdf", nil },
		mkdirTemp: func(string, string) (string, error) { return "", errors.New("temp dir failed") },
	})
	defer restore()
	t.Setenv("ENGAGE_PDF_ENGINE", "wkhtml")
	_, err := RenderPDF(SummaryReport{Target: "x", Generated: time.Now().UTC()})
	if err == nil || !strings.Contains(err.Error(), "temp dir failed") {
		t.Fatalf("err: %v", err)
	}
}

func TestRenderPDF_wkhtmlWriteFileError(t *testing.T) {
	restore := mockWkhtmlHooks(t, wkhtmlHooks{
		lookPath: func(string) (string, error) { return "/usr/bin/wkhtmltopdf", nil },
		writeFile: func(string, []byte, os.FileMode) error { return errors.New("write failed") },
	})
	defer restore()
	t.Setenv("ENGAGE_PDF_ENGINE", "wkhtml")
	_, err := RenderPDF(SummaryReport{Target: "x", Generated: time.Now().UTC()})
	if err == nil || !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("err: %v", err)
	}
}

func TestRenderPDF_wkhtmlCommandError(t *testing.T) {
	restore := mockWkhtmlHooks(t, wkhtmlHooks{
		lookPath: func(string) (string, error) { return "/usr/bin/wkhtmltopdf", nil },
		command:  func(string, ...string) *exec.Cmd { return exec.Command("false") },
	})
	defer restore()
	t.Setenv("ENGAGE_PDF_ENGINE", "wkhtml")
	_, err := RenderPDF(SummaryReport{Target: "x", Generated: time.Now().UTC()})
	if err == nil || !strings.Contains(err.Error(), "wkhtmltopdf:") {
		t.Fatalf("err: %v", err)
	}
}

type wkhtmlHooks struct {
	lookPath  func(string) (string, error)
	mkdirTemp func(string, string) (string, error)
	writeFile func(string, []byte, os.FileMode) error
	readFile  func(string) ([]byte, error)
	command   func(string, ...string) *exec.Cmd
}

func mockWkhtmlHooks(t *testing.T, h wkhtmlHooks) func() {
	t.Helper()
	oldLook, oldMk, oldWrite, oldRead, oldCmd := wkhtmlLookPath, wkhtmlMkdirTemp, wkhtmlWriteFile, wkhtmlReadFile, wkhtmlCommand
	if h.lookPath != nil {
		wkhtmlLookPath = h.lookPath
	}
	if h.mkdirTemp != nil {
		wkhtmlMkdirTemp = h.mkdirTemp
	} else {
		wkhtmlMkdirTemp = func(prefix, pattern string) (string, error) { return t.TempDir(), nil }
	}
	if h.writeFile != nil {
		wkhtmlWriteFile = h.writeFile
	}
	if h.readFile != nil {
		wkhtmlReadFile = h.readFile
	}
	if h.command != nil {
		wkhtmlCommand = h.command
	}
	return func() {
		wkhtmlLookPath = oldLook
		wkhtmlMkdirTemp = oldMk
		wkhtmlWriteFile = oldWrite
		wkhtmlReadFile = oldRead
		wkhtmlCommand = oldCmd
	}
}

func TestRenderPDF_defaultGofpdf(t *testing.T) {
	t.Setenv("ENGAGE_PDF_ENGINE", "")
	summary := SummaryReport{Target: "x", Generated: time.Now().UTC()}
	b, err := RenderPDF(summary)
	if err != nil || len(b) < 10 {
		t.Fatalf("pdf: len=%d err=%v", len(b), err)
	}
}

func TestToPDF_severityMapAny(t *testing.T) {
	summary := SummaryReport{
		Target: "x",
		Generated: time.Now().UTC(),
		Sections: map[string]any{
			"severity_breakdown": map[string]any{"high": float64(2), "low": 1},
		},
	}
	b, err := ToPDF(summary)
	if err != nil || len(b) < 50 {
		t.Fatalf("pdf err=%v len=%d", err, len(b))
	}
}

func TestToPDF_truncatesFindings(t *testing.T) {
	findings := make([]domain.Finding, 55)
	for i := range findings {
		findings[i] = domain.Finding{Title: fmt.Sprintf("Finding %d", i), Severity: domain.SeverityLow}
	}
	summary := SummaryReport{Target: "x", Generated: time.Now().UTC(), Findings: findings}
	b, err := ToPDF(summary)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 50 {
		t.Fatalf("pdf too small: %d", len(b))
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("pdf write failed") }

func TestToPDF_outputError(t *testing.T) {
	old := pdfWriteOutput
	defer func() { pdfWriteOutput = old }()
	pdfWriteOutput = func(pdf *gofpdf.Fpdf, w io.Writer) error { return pdf.Output(errWriter{}) }
	summary := SummaryReport{Target: "x", Generated: time.Now().UTC()}
	_, err := ToPDF(summary)
	if err == nil || !strings.Contains(err.Error(), "pdf write failed") {
		t.Fatalf("err: %v", err)
	}
}

func TestRenderAssessmentHTML_defaultBranding(t *testing.T) {
	s := SummaryReport{Target: "x", Generated: time.Now().UTC()}
	html := RenderAssessmentHTML(s, Branding{})
	if !strings.Contains(html, "Veil Engage") {
		t.Fatal("expected default branding")
	}
}

func TestRenderAssessmentHTML_recommendationsAndFindingCap(t *testing.T) {
	findings := make([]domain.Finding, 105)
	for i := range findings {
		findings[i] = domain.Finding{
			Title:    fmt.Sprintf("Finding %d", i),
			Severity: domain.SeverityLow,
			Tool:     "nmap",
		}
	}
	s := SummaryReport{
		Target:    "x",
		Generated: time.Now().UTC(),
		Sections: map[string]any{
			"severity_breakdown": map[string]any{"low": float64(105)},
			"recommendations":    []any{"Patch nginx", "  ", 42, "Rotate keys"},
		},
		Findings: findings,
	}
	html := RenderAssessmentHTML(s, DefaultBranding())
	if !strings.Contains(html, "Recommendations") || !strings.Contains(html, "Patch nginx") {
		t.Fatal("missing recommendations section")
	}
	if strings.Contains(html, "Finding 104") {
		t.Fatal("expected findings table capped at 100 rows")
	}
	if !strings.Contains(html, "Finding 99") {
		t.Fatal("expected first 100 findings rendered")
	}
}

func TestRenderAssessmentHTML_brandingLogo(t *testing.T) {
	s := SummaryReport{
		Target: "x",
		Findings: []domain.Finding{
			{Title: "f", Severity: domain.SeverityCritical},
		},
	}
	b := Branding{Organization: "Org", LogoURL: "https://logo.example/x.png", Classification: "TLP:GREEN"}
	html := RenderAssessmentHTML(s, b)
	if html == "" || !contains(html, "logo.example") {
		t.Fatal("expected logo in html")
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0) }
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
