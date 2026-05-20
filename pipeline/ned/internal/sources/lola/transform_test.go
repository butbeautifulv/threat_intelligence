package lola

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
)

func TestTransform_table(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name     string
		kind     string
		payload  any
		wantErr  string
		wantKind string
		check    func(t *testing.T, env *commit.Envelope)
	}{
		{
			name: harvest.KindLolaArtifactRaw,
			kind: harvest.KindLolaArtifactRaw,
			payload: harvest.LolaArtifactRaw{
				Source:  "misp",
				Path:    "obj/1.json",
				RawBody: `{"name":"sample-malware","type":"malware"}`,
			},
			wantKind: commit.KindLolaArtifact,
			check: func(t *testing.T, env *commit.Envelope) {
				var pl commit.LolaArtifactPayload
				decodePayload(t, env.Payload, &pl)
				if pl.Source != "misp" {
					t.Fatalf("source: %q", pl.Source)
				}
				wantKey := commit.LolaArtifactIdempotencyKey("misp", "sample-malware")
				if env.IdempotencyKey != wantKey {
					t.Fatalf("idempotency %q want %q", env.IdempotencyKey, wantKey)
				}
			},
		},
		{
			name: "artifact_empty_name",
			kind: harvest.KindLolaArtifactRaw,
			payload: harvest.LolaArtifactRaw{
				Source: "misp", Path: "obj/2.json", RawBody: `{"type":"malware"}`,
			},
			wantKind: commit.KindLolaArtifact,
			check: func(t *testing.T, env *commit.Envelope) {
				wantKey := commit.LolaArtifactIdempotencyKey("misp", "")
				if env.IdempotencyKey != wantKey {
					t.Fatalf("idempotency %q want %q", env.IdempotencyKey, wantKey)
				}
			},
		},
		{
			name: harvest.KindLolaLoftsRaw,
			kind: harvest.KindLolaLoftsRaw,
			payload: harvest.LolaLoftsRaw{
				Title: "LOFTS entry", Category: "technique",
				LinkURL: "https://example.com/lofts/1", Markdown: "# LOFTS",
			},
			wantKind: commit.KindLolaLofts,
			check: func(t *testing.T, env *commit.Envelope) {
				var pl commit.LolaLoftsPayload
				decodePayload(t, env.Payload, &pl)
				if pl.LinkURL != "https://example.com/lofts/1" {
					t.Fatalf("payload: %#v", pl)
				}
			},
		},
		{
			name: harvest.KindLolaAttackTechnique,
			kind: harvest.KindLolaAttackTechnique,
			payload: harvest.LolaAttackTechnique{
				ID: "T1059", Name: "Command and Scripting Interpreter", Description: "desc",
			},
			wantKind: commit.KindLolaAttackTechnique,
			check: func(t *testing.T, env *commit.Envelope) {
				var pl commit.LolaAttackTechniquePayload
				decodePayload(t, env.Payload, &pl)
				if pl.ID != "T1059" {
					t.Fatalf("id: %q", pl.ID)
				}
			},
		},
		{
			name: harvest.KindLolaAttackTactic,
			kind: harvest.KindLolaAttackTactic,
			payload: harvest.LolaAttackTactic{ID: "TA0002", Name: "Execution"},
			wantKind: commit.KindLolaAttackTactic,
			check: func(t *testing.T, env *commit.Envelope) {
				if env.IdempotencyKey != commit.LolaTacticIdempotencyKey("TA0002") {
					t.Fatalf("idempotency %q", env.IdempotencyKey)
				}
			},
		},
		{
			name: harvest.KindLolaMergeTacticTechnique,
			kind: harvest.KindLolaMergeTacticTechnique,
			payload: harvest.LolaMergeTacticTechnique{TacticID: "TA0002", TechniqueID: "T1059"},
			wantKind: commit.KindLolaMergeTacticTechnique,
			check: func(t *testing.T, env *commit.Envelope) {
				want := commit.LolaMergeTacticTechniqueIdempotencyKey("TA0002", "T1059")
				if env.IdempotencyKey != want {
					t.Fatalf("idempotency %q want %q", env.IdempotencyKey, want)
				}
			},
		},
		{
			name: harvest.KindLolaMergeSubtechnique,
			kind: harvest.KindLolaMergeSubtechnique,
			payload: harvest.LolaMergeSubtechnique{
				ParentTechniqueID: "T1059.001", ChildTechniqueID: "T1059.001.01",
			},
			wantKind: commit.KindLolaMergeSubtechnique,
			check: func(t *testing.T, env *commit.Envelope) {
				want := commit.LolaMergeSubtechniqueIdempotencyKey("T1059.001", "T1059.001.01")
				if env.IdempotencyKey != want {
					t.Fatalf("idempotency %q want %q", env.IdempotencyKey, want)
				}
				var pl commit.LolaMergeSubtechniquePayload
				decodePayload(t, env.Payload, &pl)
				if pl.ParentTechniqueID != "T1059.001" {
					t.Fatalf("payload %#v", pl)
				}
			},
		},
		{
			name:     harvest.KindLolaLinkArtifacts,
			kind:     harvest.KindLolaLinkArtifacts,
			payload:  struct{}{},
			wantKind: commit.KindLolaLinkArtifacts,
			check: func(t *testing.T, env *commit.Envelope) {
				if env.IdempotencyKey != commit.LolaLinkArtifactsIdempotencyKey() {
					t.Fatalf("idempotency %q", env.IdempotencyKey)
				}
			},
		},
		{
			name:    "artifact_invalid_payload",
			kind:    harvest.KindLolaArtifactRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "lofts_invalid_payload",
			kind:    harvest.KindLolaLoftsRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "technique_invalid_payload",
			kind:    harvest.KindLolaAttackTechnique,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "tactic_invalid_payload",
			kind:    harvest.KindLolaAttackTactic,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "merge_tactic_technique_invalid_payload",
			kind:    harvest.KindLolaMergeTacticTechnique,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "merge_subtechnique_invalid_payload",
			kind:    harvest.KindLolaMergeSubtechnique,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "unknown_kind",
			kind:    "scrape_lola_unknown",
			payload: struct{}{},
			wantErr: "pipeline lola: unknown kind",
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
			if len(out) != 1 {
				t.Fatalf("len(out) = %d", len(out))
			}
			if out[0].Kind != tt.wantKind || out[0].Source != commit.SourceLola {
				t.Fatalf("envelope source=%s kind=%s", out[0].Source, out[0].Kind)
			}
			if tt.check != nil {
				tt.check(t, out[0])
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
