package nvdmap_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	nvdmap "github.com/butbeautifulv/veil/pipeline/pkg/nvd/map"
	"github.com/butbeautifulv/veil/pipeline/pkg/nvd/parse"
)

func TestFromNVD_mapsFixtureToCanonical(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "parse", "testdata", "nvd_page_min.json"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, total, err := parse.ParsePage(raw)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(parsed) != 2 {
		t.Fatalf("parsed: total=%d len=%d", total, len(parsed))
	}

	m := nvdmap.FromNVD(parsed[0])
	if m.CVE != "CVE-2024-0001" || m.ID != "CVE-2024-0001" {
		t.Fatalf("id/cve: %#v", m)
	}
	if m.Summary == "" {
		t.Fatal("expected summary")
	}
	if len(m.CWE) != 1 || m.CWE[0] != "CWE-79" {
		t.Fatalf("cwe: %#v", m.CWE)
	}
	if len(m.CPEs) != 1 || m.CPEs[0].URI == "" {
		t.Fatalf("cpes: %#v", m.CPEs)
	}
	if m.CVSS == nil || m.CVSS.Base != 9.8 || m.CVSS.Version != "3.1" {
		t.Fatalf("cvss: %#v", m.CVSS)
	}
}

func TestFromNVD_toCommitEnvelopes(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "parse", "testdata", "nvd_page_min.json"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _, err := parse.ParsePage(raw)
	if err != nil {
		t.Fatal(err)
	}

	var envelopes []*commit.Envelope
	for _, p := range parsed {
		m := nvdmap.FromNVD(p)
		e, err := commit.NewEnvelope(commit.SourceVuln, commit.KindVulnUpsert, commit.VulnUpsertIdempotencyKey(m.CVE), m)
		if err != nil {
			t.Fatal(err)
		}
		envelopes = append(envelopes, e)
	}

	if len(envelopes) != 2 {
		t.Fatalf("envelopes: got %d want 2", len(envelopes))
	}
	if envelopes[0].Kind != commit.KindVulnUpsert || envelopes[0].Source != commit.SourceVuln {
		t.Fatalf("first envelope: source=%s kind=%s", envelopes[0].Source, envelopes[0].Kind)
	}
	var v0 nvdmap.Vulnerability
	if err := json.Unmarshal(envelopes[0].Payload, &v0); err != nil {
		t.Fatal(err)
	}
	if v0.CVE != "CVE-2024-0001" || len(v0.CWE) == 0 || len(v0.CPEs) == 0 || v0.CVSS == nil {
		t.Fatalf("payload: %#v", v0)
	}
	if envelopes[0].IdempotencyKey != commit.VulnUpsertIdempotencyKey("CVE-2024-0001") {
		t.Fatalf("idempotency %q", envelopes[0].IdempotencyKey)
	}
}

func TestFromNVD_table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input parse.Vulnerability
		check func(t *testing.T, m nvdmap.Vulnerability)
	}{
		{
			name:  "empty CVE skipped by parse not mapped",
			input: parse.Vulnerability{},
			check: func(t *testing.T, m nvdmap.Vulnerability) {
				if m.CVE != "" || m.ID != "" {
					t.Fatalf("expected empty id/cve, got %#v", m)
				}
				if len(m.CWE) != 0 || len(m.CPEs) != 0 || m.CVSS != nil {
					t.Fatalf("expected zero optional fields, got %#v", m)
				}
			},
		},
		{
			name: "partial fields id and summary only",
			input: parse.Vulnerability{
				ID:      "CVE-2024-0003",
				CVE:     "CVE-2024-0003",
				Summary: "minimal record",
			},
			check: func(t *testing.T, m nvdmap.Vulnerability) {
				if m.CVE != "CVE-2024-0003" || m.Summary != "minimal record" {
					t.Fatalf("id/summary: %#v", m)
				}
				if len(m.CWE) != 0 || len(m.CPEs) != 0 || m.CVSS != nil {
					t.Fatalf("optional fields: %#v", m)
				}
			},
		},
		{
			name: "cwe without cpe or cvss",
			input: parse.Vulnerability{
				ID:  "CVE-2024-0005",
				CVE: "CVE-2024-0005",
				CWE: []string{"CWE-20"},
			},
			check: func(t *testing.T, m nvdmap.Vulnerability) {
				if len(m.CWE) != 1 || m.CWE[0] != "CWE-20" {
					t.Fatalf("cwe: %#v", m.CWE)
				}
				if len(m.CPEs) != 0 || m.CVSS != nil {
					t.Fatalf("cpes/cvss: %#v", m)
				}
			},
		},
		{
			name: "nil cvss not copied",
			input: parse.Vulnerability{
				ID:   "CVE-2024-0006",
				CVE:  "CVE-2024-0006",
				CPEs: []parse.CPE{{URI: "cpe:2.3:a:vendor:product:*:*:*:*:*:*:*:*"}},
			},
			check: func(t *testing.T, m nvdmap.Vulnerability) {
				if len(m.CPEs) != 1 || m.CPEs[0].URI == "" {
					t.Fatalf("cpes: %#v", m.CPEs)
				}
				if m.CVSS != nil {
					t.Fatalf("cvss: %#v want nil", m.CVSS)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := nvdmap.FromNVD(tt.input)
			tt.check(t, m)
		})
	}
}

func TestFromNVD_parseSkipsEmptyCVE(t *testing.T) {
	raw := []byte(`{
		"totalResults": 3,
		"vulnerabilities": [
			{"cve": {"id": ""}},
			{"cve": {}},
			{"cve": {"id": "CVE-2024-0010", "descriptions": [{"lang": "en", "value": "valid"}]}}
		]
	}`)
	parsed, total, err := parse.ParsePage(raw)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(parsed) != 1 {
		t.Fatalf("parse: total=%d len=%d want 3/1", total, len(parsed))
	}
	m := nvdmap.FromNVD(parsed[0])
	if m.CVE != "CVE-2024-0010" || m.Summary != "valid" {
		t.Fatalf("mapped: %#v", m)
	}
}
