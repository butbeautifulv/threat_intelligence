package enrich

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/vuln/domain"
	"github.com/butbeautifulv/veil/pipeline/pkg/nvd/parse"
)

func stubCommitEnvelope(t *testing.T) {
	t.Helper()
	orig := newCommitEnvelope
	newCommitEnvelope = func(_, _ string, _ string, _ any) (*commit.Envelope, error) {
		return nil, errors.New("envelope stub")
	}
	t.Cleanup(func() { newCommitEnvelope = orig })
}

func nvdPageMinPath() string {
	return filepath.Join("..", "..", "..", "..", "..", "pkg", "nvd", "parse", "testdata", "nvd_page_min.json")
}

func TestFromNVDPage_table(t *testing.T) {
	t.Parallel()

	rawMin, err := os.ReadFile(nvdPageMinPath())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		raw     string
		wantErr string
		wantLen int
		check   func(t *testing.T, out []*commit.Envelope)
	}{
		{
			name:    "parse_error",
			raw:     "{not-json",
			wantErr: "invalid",
		},
		{
			name: "empty_cve_skip",
			raw: `{
  "totalResults": 2,
  "vulnerabilities": [
    {"cve": {"id": "   ", "descriptions": [{"lang": "en", "value": "whitespace id"}]}},
    {"cve": {"id": "CVE-2024-0001", "descriptions": [{"lang": "en", "value": "valid"}]}}
  ]
}`,
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				var v domain.Vulnerability
				decodePayload(t, out[0].Payload, &v)
				if v.CVE != "CVE-2024-0001" {
					t.Fatalf("CVE: got %q", v.CVE)
				}
			},
		},
		{
			name:    "min_fixture",
			raw:     string(rawMin),
			wantLen: 2,
			check: func(t *testing.T, out []*commit.Envelope) {
				var first domain.Vulnerability
				decodePayload(t, out[0].Payload, &first)
				if len(first.CWE) == 0 || len(first.CPEs) == 0 {
					t.Fatalf("first CVE: %#v", first)
				}
				if first.CVSS == nil || first.CVSS.Base != 9.8 {
					t.Fatalf("expected CVSS on first CVE, got %#v", first.CVSS)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := FromNVDPage(tt.raw)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != tt.wantLen {
				t.Fatalf("envelopes: got %d want %d", len(out), tt.wantLen)
			}
			if tt.check != nil {
				tt.check(t, out)
			}
		})
	}
}

func TestFromNVDPage_envelopeError(t *testing.T) {
	stubCommitEnvelope(t)
	raw := `{"totalResults":1,"vulnerabilities":[{"cve":{"id":"CVE-2024-1","descriptions":[{"lang":"en","value":"x"}]}}]}`
	_, err := FromNVDPage(raw)
	if err == nil || err.Error() != "envelope stub" {
		t.Fatalf("err=%v", err)
	}
}

func TestNVDToDomain_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   parse.Vulnerability
		check func(t *testing.T, v domain.Vulnerability)
	}{
		{
			name: "cvss_and_cpe",
			in: parse.Vulnerability{
				ID: "CVE-2024-1", CVE: "CVE-2024-1", Summary: "s",
				CWE: []string{"CWE-1"},
				CPEs: []parse.CPE{{URI: "cpe:2.3:a:v:p:1:*:*:*:*:*:*:*"}},
				CVSS: &parse.CVSS{Version: "3.1", Base: 7.5, Vector: "V"},
			},
			check: func(t *testing.T, v domain.Vulnerability) {
				if len(v.CPEs) != 1 || v.CVSS == nil || v.CVSS.Base != 7.5 {
					t.Fatalf("got %#v", v)
				}
			},
		},
		{
			name: "no_cpes_no_cvss",
			in:   parse.Vulnerability{ID: "CVE-1", CVE: "CVE-1", Summary: "s"},
			check: func(t *testing.T, v domain.Vulnerability) {
				if len(v.CPEs) != 0 || v.CVSS != nil {
					t.Fatalf("got %#v", v)
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := nvdToDomain(tt.in)
			tt.check(t, v)
		})
	}
}

func decodePayload(t *testing.T, raw json.RawMessage, dst any) {
	t.Helper()
	if err := json.Unmarshal(raw, dst); err != nil {
		t.Fatal(err)
	}
}
