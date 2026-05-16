package findings

import (
	"os"
	"path/filepath"
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

func TestParseFfuf_jsonFixture(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "ffuf_sample.json"))
	if err != nil {
		t.Fatal(err)
	}
	got := parseFfuf("https://example.com", "ffuf_scan", string(raw))
	if len(got) < 2 {
		t.Fatalf("got %d findings", len(got))
	}
	if got[0].Severity != domainreport.SeverityMedium {
		t.Fatalf("403 severity %s", got[0].Severity)
	}
}

func TestParseFfuf_statusLineFallback(t *testing.T) {
	got := parseFfuf("https://x.com", "ffuf", "https://x.com/admin [Status: 200]")
	if len(got) != 1 || got[0].Severity != domainreport.SeverityLow {
		t.Fatalf("got %+v", got)
	}
}

func TestParseSqlmap_injectionFixture(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "sqlmap_injection.txt"))
	if err != nil {
		t.Fatal(err)
	}
	got := parseSqlmap("https://example.com?id=1", "sqlmap_scan", string(raw))
	if len(got) < 2 {
		t.Fatalf("got %d findings", len(got))
	}
	foundHigh := false
	for _, f := range got {
		if f.Severity == domainreport.SeverityHigh {
			foundHigh = true
		}
	}
	if !foundHigh {
		t.Fatal("expected high severity finding")
	}
}
