package ds

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
)

func stubCommitEnvelope(t *testing.T) {
	t.Helper()
	orig := newCommitEnvelope
	newCommitEnvelope = func(_, _ string, _ string, _ any) (*commit.Envelope, error) {
		return nil, errors.New("envelope stub")
	}
	t.Cleanup(func() { newCommitEnvelope = orig })
}

func TestTransform_table(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		kind      string
		payload   any
		wantErr   string
		wantLen   int
		wantKind  string
		checkFirst func(t *testing.T, env *commit.Envelope)
	}{
		{
			name: "sigma",
			kind: harvest.KindDSSigmaRaw,
			payload: harvest.DSSigmaRaw{
				Path: "rules/test.yml",
				RawYAML: `id: test-sigma-1
title: Test Sigma
level: high
logsource:
  product: windows
  service: security
tags:
  - attack.execution
  - 42
`,
			},
			wantLen:  1,
			wantKind: commit.KindDSUpsertSigma,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertSigmaPayload
				decodePayload(t, env.Payload, &pl)
				if pl.ID != "test-sigma-1" || pl.LogProduct != "windows" || pl.TagsJSON != `["attack.execution"]` {
					t.Fatalf("payload: %#v", pl)
				}
			},
		},
		{
			name: "sigma_stable_id",
			kind: harvest.KindDSSigmaRaw,
			payload: harvest.DSSigmaRaw{
				Path:    "rules/no-id.yml",
				RawYAML: "title: No ID\nlevel: medium\n",
			},
			wantLen:  1,
			wantKind: commit.KindDSUpsertSigma,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertSigmaPayload
				decodePayload(t, env.Payload, &pl)
				wantID := stableID("sigma", "rules/no-id.yml")
				if pl.ID != wantID {
					t.Fatalf("id %q want %q", pl.ID, wantID)
				}
			},
		},
		{
			name: "yara_named_rule",
			kind: harvest.KindDSYaraRaw,
			payload: harvest.DSYaraRaw{
				Path: "rules/test.yar",
				RawBody: `rule TestRule {
  strings:
    $a = "test"
  condition:
    $a
}`,
			},
			wantLen:  1,
			wantKind: commit.KindDSUpsertYara,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertYaraPayload
				decodePayload(t, env.Payload, &pl)
				if pl.Name != "TestRule" {
					t.Fatalf("name: %q", pl.Name)
				}
			},
		},
		{
			name: "yara_rule_no_delimiter",
			kind: harvest.KindDSYaraRaw,
			payload: harvest.DSYaraRaw{
				Path:    "rules/solo.yar",
				RawBody: "rule SoloName\n",
			},
			wantLen:  1,
			wantKind: commit.KindDSUpsertYara,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertYaraPayload
				decodePayload(t, env.Payload, &pl)
				if pl.Name != "SoloName" {
					t.Fatalf("name: %q", pl.Name)
				}
			},
		},
		{
			name: "yara_parse_name",
			kind: harvest.KindDSYaraRaw,
			payload: harvest.DSYaraRaw{
				Path:    "rules/fallback.yar",
				RawBody: "rule TabRule {\ncondition:\n  true\n}",
			},
			wantLen:  1,
			wantKind: commit.KindDSUpsertYara,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertYaraPayload
				decodePayload(t, env.Payload, &pl)
				if pl.Name != "TabRule" {
					t.Fatalf("name: %q", pl.Name)
				}
			},
		},
		{
			name: "yara_fallback_path",
			kind: harvest.KindDSYaraRaw,
			payload: harvest.DSYaraRaw{
				Path:    "rules/no-rule.yar",
				RawBody: "// no rule keyword\n",
			},
			wantLen:  1,
			wantKind: commit.KindDSUpsertYara,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertYaraPayload
				decodePayload(t, env.Payload, &pl)
				if pl.Name != "rules/no-rule" {
					t.Fatalf("name: %q", pl.Name)
				}
			},
		},
		{
			name: "atomic",
			kind: harvest.KindDSAtomicRaw,
			payload: harvest.DSAtomicRaw{
				Path: "atomics/T1003.yml",
				RawYAML: `attack_technique: T1003
atomic_tests:
  - name: Dump LSASS
    auto_generated_guid: guid-abc
    tactics:
      - credential-access
    executor:
      name: powershell
      command: Get-Process lsass
  - not-a-map
  - name: Second
    tactics:
      - 99
    executor:
      name: sh
`,
			},
			wantLen:  2,
			wantKind: commit.KindDSUpsertAtomic,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertAtomicPayload
				decodePayload(t, env.Payload, &pl)
				if pl.ID != "guid-abc" || pl.Technique != "T1003" {
					t.Fatalf("payload: %#v", pl)
				}
			},
		},
		{
			name: "atomic_generated_guid",
			kind: harvest.KindDSAtomicRaw,
			payload: harvest.DSAtomicRaw{
				Path: "atomics/T1003.yml",
				RawYAML: `attack_technique: T1003
atomic_tests:
  - name: No GUID
`,
			},
			wantLen:  1,
			wantKind: commit.KindDSUpsertAtomic,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertAtomicPayload
				decodePayload(t, env.Payload, &pl)
				if pl.ID != "T1003-0" {
					t.Fatalf("id %q", pl.ID)
				}
			},
		},
		{
			name: "caldera_single",
			kind: harvest.KindDSCalderaRaw,
			payload: harvest.DSCalderaRaw{
				Path:     "abilities/test.yml",
				FileName: "test.yml",
				RawBody: `id: ability-1
name: List files
description: Enumerate files
tactic: collection
technique:
  attack_id: T1005
`,
			},
			wantLen:  1,
			wantKind: commit.KindDSUpsertCaldera,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertCalderaPayload
				decodePayload(t, env.Payload, &pl)
				if pl.ID != "ability-1" || pl.TechniqueID != "T1005" {
					t.Fatalf("payload: %#v", pl)
				}
			},
		},
		{
			name: "caldera_sequence",
			kind: harvest.KindDSCalderaRaw,
			payload: harvest.DSCalderaRaw{
				Path: "abilities/seq.yml",
				RawBody: `- id: ability-a
  name: A
  description: first
  tactic: discovery
  technique:
    attack_id: T1082
- id: ""
  name: skip
- id: ability-b
  name: B
  description: second
  tactic: execution
`,
			},
			wantLen:  2,
			wantKind: commit.KindDSUpsertCaldera,
		},
		{
			name: "caldera_empty_id_single",
			kind: harvest.KindDSCalderaRaw,
			payload: harvest.DSCalderaRaw{
				Path:    "abilities/empty.yml",
				RawBody: "name: no id\n",
			},
			wantLen: 0,
		},
		{
			name: "sigma_invalid_payload",
			kind: harvest.KindDSSigmaRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name: "sigma_invalid_yaml",
			kind: harvest.KindDSSigmaRaw,
			payload: harvest.DSSigmaRaw{Path: "x.yml", RawYAML: ":\n\tbad"},
			wantErr: "yaml",
		},
		{
			name: "yara_invalid_payload",
			kind: harvest.KindDSYaraRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name: "atomic_invalid_payload",
			kind: harvest.KindDSAtomicRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name: "atomic_invalid_yaml",
			kind: harvest.KindDSAtomicRaw,
			payload: harvest.DSAtomicRaw{Path: "a.yml", RawYAML: ":\n\tbad"},
			wantErr: "yaml",
		},
		{
			name: "atomic_no_tests",
			kind: harvest.KindDSAtomicRaw,
			payload: harvest.DSAtomicRaw{
				Path: "atomics/T1.yml", RawYAML: "attack_technique: T1\n",
			},
			wantErr: "no atomic_tests",
		},
		{
			name: "caldera_invalid_payload",
			kind: harvest.KindDSCalderaRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name: "caldera_invalid_yaml",
			kind: harvest.KindDSCalderaRaw,
			payload: harvest.DSCalderaRaw{Path: "c.yml", RawBody: ":\n\tbad"},
			wantErr: "yaml",
		},
		{
			name:    "unknown_kind",
			kind:    "scrape_ds_unknown",
			payload: struct{}{},
			wantErr: "pipeline ds: unknown kind",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var payload []byte
			var err error
			switch p := tt.payload.(type) {
			case []byte:
				payload = p
			default:
				payload, err = json.Marshal(p)
				if err != nil {
					t.Fatal(err)
				}
			}
			env := &harvest.Envelope{Kind: tt.kind, Payload: payload}
			out, err := Transform(ctx, env)
			if tt.wantErr != "" {
				assertErrContains(t, err, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != tt.wantLen {
				t.Fatalf("len(out) = %d want %d", len(out), tt.wantLen)
			}
			if tt.wantLen == 0 {
				return
			}
			if out[0].Kind != tt.wantKind {
				t.Fatalf("kind = %s want %s", out[0].Kind, tt.wantKind)
			}
			if out[0].Source != commit.SourceDS {
				t.Fatalf("source = %s", out[0].Source)
			}
			if tt.checkFirst != nil {
				tt.checkFirst(t, out[0])
			}
		})
	}
}

func TestAtomicFromYAML_envelopeError(t *testing.T) {
	stubCommitEnvelope(t)
	_, err := atomicFromYAML("a.yml", `attack_technique: T1
atomic_tests:
  - name: t
`)
	if err == nil || err.Error() != "envelope stub" {
		t.Fatalf("err=%v", err)
	}
}

func TestCalderaRootToEnvelope_envelopeError(t *testing.T) {
	stubCommitEnvelope(t)
	root := map[string]any{
		"id": "a1", "name": "n", "description": "d", "tactic": "t",
	}
	if env := calderaRootToEnvelope(root, "p.yml"); env != nil {
		t.Fatalf("expected nil envelope, got %+v", env)
	}
}

func TestTagsToJSON_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "not_array", in: "tags", want: "[]"},
		{name: "strings", in: []any{"a", "b"}, want: `["a","b"]`},
		{name: "mixed", in: []any{"a", 1, "b"}, want: `["a","b"]`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tagsToJSON(tt.in); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
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
