package vuln

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pkg/vuln/domain"
)

func nvdPageMinPath() string {
	return filepath.Join("..", "..", "..", "..", "pkg", "nvd", "parse", "testdata", "nvd_page_min.json")
}

func TestTransform_table(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	raw, err := os.ReadFile(nvdPageMinPath())
	if err != nil {
		t.Fatal(err)
	}
	nvdPayload, err := json.Marshal(harvest.VulnNVDPage{RawJSON: string(raw)})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		kind    string
		payload []byte
		wantErr string
		wantLen int
		check   func(t *testing.T, out []*commit.Envelope)
	}{
		{
			name: "cve_upsert",
			kind: harvest.KindVulnCVEUpsert,
			payload: mustJSON(t, domain.Vulnerability{
				CVE: "CVE-2024-1234", ID: "CVE-2024-1234",
				Summary: "test summary", CWE: []string{"CWE-79"},
			}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				e := out[0]
				if e.Source != commit.SourceVuln || e.Kind != commit.KindVulnUpsert {
					t.Fatalf("envelope: %+v", e)
				}
				wantKey := commit.VulnUpsertIdempotencyKey("CVE-2024-1234")
				if e.IdempotencyKey != wantKey {
					t.Fatalf("idempotency key: got %q want %q", e.IdempotencyKey, wantKey)
				}
				var got domain.Vulnerability
				decodePayload(t, e.Payload, &got)
				if got.CVE != "CVE-2024-1234" || len(got.CWE) != 1 {
					t.Fatalf("payload: %#v", got)
				}
			},
		},
		{
			name: "cve_upsert_whitespace_only",
			kind: harvest.KindVulnCVEUpsert,
			payload: mustJSON(t, domain.Vulnerability{CVE: "   ", Summary: "no id"}),
			wantErr: "vuln: empty CVE",
		},
		{
			name:    "cve_upsert_invalid_payload",
			kind:    harvest.KindVulnCVEUpsert,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name: "merge_exploit",
			kind: harvest.KindVulnMergeExploit,
			payload: mustJSON(t, harvest.VulnMergeExploit{
				CVE: "CVE-2024-9999", Source: "exploitdb", RefID: "EDB-1",
				URL: "https://example.com/edb/1",
			}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				e := out[0]
				if e.Kind != commit.KindVulnMergeExploit {
					t.Fatalf("envelope: %+v", e)
				}
				var pl commit.VulnMergeExploitPayload
				decodePayload(t, e.Payload, &pl)
				if pl.CVE != "CVE-2024-9999" || pl.RefID != "EDB-1" {
					t.Fatalf("payload: %#v", pl)
				}
			},
		},
		{
			name:    "merge_exploit_invalid_payload",
			kind:    harvest.KindVulnMergeExploit,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "nvd_page",
			kind:    harvest.KindVulnNVDPage,
			payload: nvdPayload,
			wantLen: 2,
			check: func(t *testing.T, out []*commit.Envelope) {
				var v domain.Vulnerability
				decodePayload(t, out[0].Payload, &v)
				if len(v.CWE) == 0 || len(v.CPEs) == 0 {
					t.Fatalf("first CVE: %#v", v)
				}
			},
		},
		{
			name:    "nvd_page_invalid_payload",
			kind:    harvest.KindVulnNVDPage,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "nvd_page_parse_error",
			kind:    harvest.KindVulnNVDPage,
			payload: mustJSON(t, harvest.VulnNVDPage{RawJSON: "{not-json"}),
			wantErr: "invalid",
		},
		{
			name:    "unknown_kind",
			kind:    "scrape_vuln_unknown",
			payload: []byte(`{}`),
			wantErr: "pipeline vuln: unknown kind",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := &harvest.Envelope{Kind: tt.kind, Payload: tt.payload}
			out, err := Transform(ctx, env)
			if tt.wantErr != "" {
				assertErrContains(t, err, tt.wantErr)
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

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func decodePayload(t *testing.T, raw json.RawMessage, dst any) {
	t.Helper()
	if err := json.Unmarshal(raw, dst); err != nil {
		t.Fatal(err)
	}
}

func assertErrContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("err=%v want substring %q", err, substr)
	}
}
