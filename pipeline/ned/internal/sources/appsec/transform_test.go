package appsec

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
)

func TestTransformSBOM_table(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name    string
		kind    string
		payload []byte
		wantErr string
		wantLen int
		check   func(t *testing.T, out []*commit.Envelope)
	}{
		{
			name: "osv_json",
			kind: harvest.KindSBOMOSVJSON,
			payload: mustJSON(t, harvest.SBOMOSVRaw{
				OSVID:   "OSV-1",
				CVE:     "CVE-2024-1000",
				RawJSON: `{"id":"OSV-1","affected":[{"package":{"ecosystem":"npm","name":"pkg"}},"not-a-map"]}`,
			}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				if out[0].Kind != commit.KindSBOMOSVRecord {
					t.Fatalf("kind %s", out[0].Kind)
				}
				var pl commit.SBOMOSVPayload
				decodePayload(t, out[0].Payload, &pl)
				if pl.CVE != "CVE-2024-1000" || len(pl.Affected) != 1 {
					t.Fatalf("payload %+v", pl)
				}
			},
		},
		{
			name: "ghsa_path",
			kind: harvest.KindSBOMGHSAPath,
			payload: mustJSON(t, harvest.SBOMGHSARaw{
				Path: "/advisories/GHSA-xxxx",
				Doc:  map[string]any{"summary": "advisory"},
			}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				if out[0].Kind != commit.KindSBOMGHSADocument {
					t.Fatalf("kind %s", out[0].Kind)
				}
				wantKey := commit.SBOMGHSAIdempotencyKey("/advisories/GHSA-xxxx")
				if out[0].IdempotencyKey != wantKey {
					t.Fatalf("idempotency %q want %q", out[0].IdempotencyKey, wantKey)
				}
				var pl commit.SBOMGHSAPathPayload
				decodePayload(t, out[0].Payload, &pl)
				if pl.Path != "/advisories/GHSA-xxxx" {
					t.Fatalf("path %q", pl.Path)
				}
			},
		},
		{
			name:    "invalid_payload",
			kind:    harvest.KindSBOMOSVJSON,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "ghsa_invalid_payload",
			kind:    harvest.KindSBOMGHSAPath,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name: "invalid_inner_raw_json",
			kind: harvest.KindSBOMOSVJSON,
			payload: mustJSON(t, harvest.SBOMOSVRaw{
				OSVID: "OSV-1", CVE: "CVE-1", RawJSON: `{`,
			}),
			wantErr: "unexpected",
		},
		{
			name:    "unknown_kind",
			kind:    "scrape_sbom_unknown",
			payload: []byte(`{}`),
			wantErr: "pipeline sbom: unknown kind",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := &harvest.Envelope{Kind: tt.kind, Payload: tt.payload}
			out, err := TransformSBOM(ctx, env)
			if tt.wantErr != "" {
				assertErrContains(t, err, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != tt.wantLen {
				t.Fatalf("len(out)=%d want %d", len(out), tt.wantLen)
			}
			if tt.check != nil {
				tt.check(t, out)
			}
		})
	}
}

func TestTransformCoderules_table(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name    string
		kind    string
		payload []byte
		wantErr string
		wantLen int
		check   func(t *testing.T, out []*commit.Envelope)
	}{
		{
			name: "cwe_row",
			kind: harvest.KindCoderulesCWERaw,
			payload: mustJSON(t, harvest.CoderulesCWERaw{
				ID: "CWE-79", Name: "XSS", Description: "d", Status: "Stable",
			}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				if out[0].Source != commit.SourceCoderules || out[0].Kind != commit.KindCoderulesCWERow {
					t.Fatalf("envelope %+v", out[0])
				}
			},
		},
		{
			name: "semgrep_raw",
			kind: harvest.KindCoderulesSemgrepRaw,
			payload: mustJSON(t, harvest.CoderulesSemgrepRaw{
				Path: "python/rule.yml",
				RawYAML: `rules:
  - id: semgrep-rule
    message: SQL injection
    metadata:
      cwe: CWE-89
`,
			}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				if out[0].Kind != commit.KindCoderulesSemgrep {
					t.Fatalf("kind %s", out[0].Kind)
				}
				var pl commit.CoderulesSemgrepPayload
				decodePayload(t, out[0].Payload, &pl)
				if pl.Language != "python" || pl.RuleID != "semgrep-rule" || len(pl.CWEs) == 0 {
					t.Fatalf("payload %#v", pl)
				}
			},
		},
		{
			name: "codeql_raw",
			kind: harvest.KindCoderulesCodeQLRaw,
			payload: mustJSON(t, harvest.CoderulesCodeQLRaw{
				Path: "javascript/query.ql",
				Body: "/**\n * @name Test\n * @kind path-problem\n * @id js/test\n * @description d\n * @problem.severity warning\n * @precision medium\n * @tags security\n *       external/cwe/cwe-079\n */\n",
			}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				if out[0].Kind != commit.KindCoderulesCodeQL {
					t.Fatalf("kind %s", out[0].Kind)
				}
				var pl commit.CoderulesCodeQLPayload
				decodePayload(t, out[0].Payload, &pl)
				if pl.Lang != "javascript" || len(pl.CWEs) == 0 {
					t.Fatalf("payload %#v", pl)
				}
			},
		},
		{
			name:    "cwe_invalid_payload",
			kind:    harvest.KindCoderulesCWERaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "semgrep_invalid_payload",
			kind:    harvest.KindCoderulesSemgrepRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name: "semgrep_invalid_yaml",
			kind: harvest.KindCoderulesSemgrepRaw,
			payload: mustJSON(t, harvest.CoderulesSemgrepRaw{
				Path: "x.yml", RawYAML: ":\n\tbad",
			}),
			wantErr: "yaml",
		},
		{
			name:    "codeql_invalid_payload",
			kind:    harvest.KindCoderulesCodeQLRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "unknown_kind",
			kind:    "scrape_coderules_unknown",
			payload: []byte(`{}`),
			wantErr: "pipeline coderules: unknown kind",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := &harvest.Envelope{Kind: tt.kind, Payload: tt.payload}
			out, err := TransformCoderules(ctx, env)
			if tt.wantErr != "" {
				assertErrContains(t, err, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != tt.wantLen {
				t.Fatalf("len(out)=%d want %d", len(out), tt.wantLen)
			}
			if tt.check != nil {
				tt.check(t, out)
			}
		})
	}
}

func TestTransformNuclei_table(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name    string
		kind    string
		payload []byte
		wantErr string
		wantLen int
		check   func(t *testing.T, out []*commit.Envelope)
	}{
		{
			name: harvest.KindNucleiTemplateRaw,
			kind: harvest.KindNucleiTemplateRaw,
			payload: mustJSON(t, harvest.NucleiTemplateRaw{
				Path:    "/tmp/http-missing.yaml",
				RawYAML: "id: http-missing\ninfo:\n  name: Missing\n  severity: info\n",
			}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				if out[0].Kind != commit.KindNucleiTemplate {
					t.Fatalf("kind %s", out[0].Kind)
				}
			},
		},
		{
			name:    "unknown_kind",
			kind:    "scrape_nuclei_unknown",
			payload: []byte(`{}`),
			wantErr: "pipeline nuclei: unknown kind",
		},
		{
			name:    "invalid_payload",
			kind:    harvest.KindNucleiTemplateRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name: "invalid_yaml",
			kind: harvest.KindNucleiTemplateRaw,
			payload: mustJSON(t, harvest.NucleiTemplateRaw{
				Path: "/tmp/bad.yaml", RawYAML: ":\n\tbad",
			}),
			wantErr: "yaml",
		},
		{
			name: "empty_template_id",
			kind: harvest.KindNucleiTemplateRaw,
			payload: mustJSON(t, harvest.NucleiTemplateRaw{
				Path: "/tmp/no-id.yaml",
				RawYAML: "info:\n  name: No ID\n  severity: info\n",
			}),
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := &harvest.Envelope{Kind: tt.kind, Payload: tt.payload}
			out, err := TransformNuclei(ctx, env)
			if tt.wantErr != "" {
				assertErrContains(t, err, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != tt.wantLen {
				t.Fatalf("len(out)=%d want %d", len(out), tt.wantLen)
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
