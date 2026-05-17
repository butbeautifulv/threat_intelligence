package report

import (
	"testing"
	"time"

	domain "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

func TestToPDF_nonEmpty(t *testing.T) {
	summary := NewSummary("https://example.com", map[string]any{
		"severity_breakdown": map[string]int{"high": 1, "info": 2},
	}, []domain.Finding{
		{Title: "Test finding", Severity: domain.SeverityHigh, Target: "https://example.com"},
	})
	b, err := ToPDF(summary)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 100 {
		t.Fatalf("pdf too small: %d bytes", len(b))
	}
	if b[0] != '%' || b[1] != 'P' {
		t.Fatalf("not PDF header: %q", b[:4])
	}
	_ = time.Now()
}
