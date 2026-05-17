package findings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
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

func TestDedupe_duplicateNucleiViaParseToolOutput(t *testing.T) {
	dup := "{\"info\":{\"name\":\"Same\",\"severity\":\"high\"}}\n" + "{\"info\":{\"name\":\"Same\",\"severity\":\"high\"}}\n"
	got := ParseToolOutput("nuclei", "https://example.com/", dup)
	if len(got) != 1 {
		t.Fatalf("expected 1 after dedupe, got %d: %+v", len(got), got)
	}
}

func TestDedupe_mixedTools(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "dedup_fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []domainreport.Finding
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures) != 7 {
		t.Fatalf("fixture count %d", len(fixtures))
	}
	got := DedupeFindings(fixtures)
	if len(got) != 4 {
		t.Fatalf("want 4 unique rows, got %d", len(got))
	}
}

func TestParseMasscan_grepableHosts(t *testing.T) {
	out := `Host: 10.0.0.2 () Ports: 22/open/tcp//ssh/, 80/open/tcp///`
	got := parseMasscan("10.0.0.2", "masscan_fast", out)
	if len(got) != 2 {
		t.Fatalf("got %d %+v", len(got), got)
	}
}

func TestParseMasscan_openLine(t *testing.T) {
	got := parseMasscan("", "masscan_fast", "open tcp 443 203.0.113.10")
	if len(got) != 1 || got[0].Title != "open 443/tcp" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseWpscan_jsonVulns(t *testing.T) {
	doc := `{"vulnerabilities":[{"title":"XSS issue","description":"stored"}],"interesting_findings":[{"url":"/wp-login.php","found_by":"Enumeration"}]}`
	got := ParseToolOutput("wpscan", "https://wp.example", doc)
	if len(got) < 2 {
		t.Fatalf("got %d %+v", len(got), got)
	}
}

func TestParseWpscan_textLine(t *testing.T) {
	out := `[!] WordPress security issue in theme`
	got := ParseToolOutput("wpscan", "https://wp.example", out)
	if len(got) != 1 || !strings.Contains(got[0].Title, "WordPress security") {
		t.Fatalf("%+v", got)
	}
}
