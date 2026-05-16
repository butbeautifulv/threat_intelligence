package findings

import (
	"testing"

	domainreport "github.com/butbeautifulv/veil/engage/serve/internal/domain/report"
)

func TestParseGeneric_high(t *testing.T) {
	got := parseGeneric("https://x.com", "test", "found HIGH severity issue")
	if len(got) == 0 {
		t.Fatal("expected findings")
	}
	if got[0].Severity != domainreport.SeverityHigh {
		t.Fatalf("severity %s", got[0].Severity)
	}
}

func TestParseNmap_openPort(t *testing.T) {
	out := `22/tcp open ssh`
	got := parseNmap("10.0.0.1", "nmap_scan", out)
	if len(got) != 1 {
		t.Fatalf("got %d", len(got))
	}
}

func TestParseNuclei_jsonLine(t *testing.T) {
	line := `{"info":{"name":"test","severity":"critical"},"matcher-name":"x"}`
	got := parseNuclei("https://x.com", "nuclei_scan", line)
	if len(got) != 1 || got[0].Severity != domainreport.SeverityCritical {
		t.Fatalf("got %+v", got)
	}
}
