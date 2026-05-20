package ti

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	tidomain "github.com/butbeautifulv/veil/pkg/ti/domain"
	tinormalize "github.com/butbeautifulv/veil/pkg/ti/normalize"
)

func stubCommitEnvelope(t *testing.T) {
	t.Helper()
	orig := newCommitEnvelope
	newCommitEnvelope = func(_, _ string, _ string, _ any) (*commit.Envelope, error) {
		return nil, errors.New("envelope stub")
	}
	t.Cleanup(func() { newCommitEnvelope = orig })
}

func TestJsonlToIngest_envelopeErrors(t *testing.T) {
	stubCommitEnvelope(t)

	tests := []struct {
		name string
		line JSONLEnvelope
	}{
		{"ioc", JSONLEnvelope{IOC: &tidomain.IOC{Type: tidomain.IOCURL, Value: "https://a.com", Source: "s"}}},
		{"campaign", JSONLEnvelope{Campaign: &tidomain.Campaign{ID: "c1", Name: "C"}}},
		{"cluster", JSONLEnvelope{Cluster: &tidomain.Cluster{ID: "cl1", Name: "Cl"}}},
		{"actor", JSONLEnvelope{Actor: &tidomain.Actor{ID: "a1", Name: "A"}}},
		{"report", JSONLEnvelope{Report: &tidomain.Report{Title: "T", Provider: "P", Link: "https://x"}}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := jsonlToIngest(tt.line)
			if err == nil || err.Error() != "envelope stub" {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestJsonlToIngest_table(t *testing.T) {

	validURLIOC := tidomain.IOC{
		Type: tidomain.IOCURL, Value: "https://evil.com/x", Source: "feed",
	}

	tests := []struct {
		name    string
		line    JSONLEnvelope
		wantLen int
		wantErr string
		check   func(t *testing.T, out []*commit.Envelope)
	}{
		{
			name:    "ioc",
			line:    JSONLEnvelope{IOC: &validURLIOC},
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				assertCommit(t, out[0], commit.KindTIIoC, out[0].IdempotencyKey)
			},
		},
		{
			name:    "ioc_skip",
			line:    JSONLEnvelope{IOC: &tidomain.IOC{Type: tidomain.IOCIP, Value: "bad"}},
			wantLen: 0,
		},
		{
			name:    "campaign",
			line:    JSONLEnvelope{Campaign: &tidomain.Campaign{ID: "c1", Name: "Camp"}},
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				assertCommit(t, out[0], commit.KindTICampaign, commit.TICampaignIdempotencyKey("c1"))
			},
		},
		{
			name:    "cluster",
			line:    JSONLEnvelope{Cluster: &tidomain.Cluster{ID: "cl1", Name: "Cl"}},
			wantLen: 1,
		},
		{
			name:    "actor_id",
			line:    JSONLEnvelope{Actor: &tidomain.Actor{ID: "a1", Name: "A"}},
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				assertCommit(t, out[0], commit.KindTIActor, commit.TIActorIdempotencyKey("a1"))
			},
		},
		{
			name:    "actor_stable",
			line:    JSONLEnvelope{Actor: &tidomain.Actor{Name: "APT"}},
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				sid := tinormalize.ActorStableID("APT")
				assertCommit(t, out[0], commit.KindTIActor, commit.TIActorIdempotencyKey(sid))
			},
		},
		{
			name:    "report",
			line:    JSONLEnvelope{Report: &tidomain.Report{Title: "T", Provider: "P", Link: "https://x/r"}},
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				sid := tinormalize.ReportStableID("https://x/r")
				assertCommit(t, out[0], commit.KindTIReport, commit.TIReportIdempotencyKey(sid))
			},
		},
		{
			name:    "empty",
			line:    JSONLEnvelope{},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := jsonlToIngest(tt.line)
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

func assertErrContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("err=%v want %q", err, substr)
	}
}

func TestTransform(t *testing.T) {
	validURLIOC := tidomain.IOC{
		Type:   tidomain.IOCURL,
		Value:  "HTTPS://Evil.COM/path",
		Source: "urlhaus",
	}
	invalidIOC := tidomain.IOC{Type: tidomain.IOCIP, Value: "not-an-ip", Source: "feed"}

	tests := []struct {
		name    string
		kind    string
		payload []byte
		wantErr string
		wantLen int
		check   func(t *testing.T, out []*commit.Envelope)
	}{
		{
			name: "KindTIKEVRow",
			kind: harvest.KindTIKEVRow,
			payload: mustJSON(t, harvest.TIKEVRow{
				CVEID: "cve-2024-0001", VendorProject: "vendor", Product: "prod",
				ShortDesc: "desc", DateAdded: "2024-01-01",
			}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				assertCommit(t, out[0], commit.KindTIKEVVulnerability, commit.TIKEVIdempotencyKey("CVE-2024-0001"))
				var p commit.TIKEVVulnPayload
				decodePayload(t, out[0].Payload, &p)
				if p.CVEID != "cve-2024-0001" || p.VendorProject != "vendor" {
					t.Fatalf("payload %#v", p)
				}
			},
		},
		{
			name:    "KindTIKEVRow_invalidJSON",
			kind:    harvest.KindTIKEVRow,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "KindTIJSONLLine_actor",
			kind:    harvest.KindTIJSONLLine,
			payload: jsonlLinePayload(t, JSONLEnvelope{Actor: &tidomain.Actor{Name: "APT29", ID: "actor-1"}}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				assertCommit(t, out[0], commit.KindTIActor, commit.TIActorIdempotencyKey("actor-1"))
			},
		},
		{
			name:    "KindTIJSONLLine_actor_stableID",
			kind:    harvest.KindTIJSONLLine,
			payload: jsonlLinePayload(t, JSONLEnvelope{Actor: &tidomain.Actor{Name: "APT29"}}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				sid := tinormalize.ActorStableID("APT29")
				assertCommit(t, out[0], commit.KindTIActor, commit.TIActorIdempotencyKey(sid))
			},
		},
		{
			name:    "KindTIJSONLLine_campaign",
			kind:    harvest.KindTIJSONLLine,
			payload: jsonlLinePayload(t, JSONLEnvelope{Campaign: &tidomain.Campaign{ID: "c1", Name: "Camp"}}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				assertCommit(t, out[0], commit.KindTICampaign, commit.TICampaignIdempotencyKey("c1"))
				var c tidomain.Campaign
				decodePayload(t, out[0].Payload, &c)
				if c.Name != "Camp" {
					t.Fatalf("campaign %#v", c)
				}
			},
		},
		{
			name:    "KindTIJSONLLine_cluster",
			kind:    harvest.KindTIJSONLLine,
			payload: jsonlLinePayload(t, JSONLEnvelope{Cluster: &tidomain.Cluster{ID: "cl1", Name: "Cluster"}}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				assertCommit(t, out[0], commit.KindTICluster, commit.TIClusterIdempotencyKey("cl1"))
			},
		},
		{
			name:    "KindTIJSONLLine_ioc",
			kind:    harvest.KindTIJSONLLine,
			payload: jsonlLinePayload(t, JSONLEnvelope{IOC: &validURLIOC}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				var ni tidomain.IOC
				decodePayload(t, out[0].Payload, &ni)
				wantKey := commit.TIIoCIdempotencyKey(tinormalize.CanonicalID(ni))
				assertCommit(t, out[0], commit.KindTIIoC, wantKey)
				if ni.Value != "https://evil.com/path" {
					t.Fatalf("normalized value = %q", ni.Value)
				}
			},
		},
		{
			name:    "KindTIJSONLLine_ioc_normalize_skip",
			kind:    harvest.KindTIJSONLLine,
			payload: jsonlLinePayload(t, JSONLEnvelope{IOC: &invalidIOC}),
			wantLen: 0,
		},
		{
			name:    "KindTIJSONLLine_report",
			kind:    harvest.KindTIJSONLLine,
			payload: jsonlLinePayload(t, JSONLEnvelope{Report: &tidomain.Report{Title: "T", Provider: "P", Link: "https://example.com/r"}}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				sid := tinormalize.ReportStableID("https://example.com/r")
				assertCommit(t, out[0], commit.KindTIReport, commit.TIReportIdempotencyKey(sid))
			},
		},
		{
			name:    "KindTIJSONLLine_invalid_outerJSON",
			kind:    harvest.KindTIJSONLLine,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "KindTIJSONLLine_invalid_innerJSON",
			kind:    harvest.KindTIJSONLLine,
			payload: []byte(`{"line":1}`),
			wantErr: "ti jsonl:",
		},
		{
			name:    "KindTIJSONLLine_empty_record",
			kind:    harvest.KindTIJSONLLine,
			payload: jsonlLinePayload(t, JSONLEnvelope{}),
			wantLen: 0,
		},
		{
			name:    "KindTIIoCRaw",
			kind:    harvest.KindTIIoCRaw,
			payload: mustJSON(t, validURLIOC),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				var ni tidomain.IOC
				decodePayload(t, out[0].Payload, &ni)
				wantKey := commit.TIIoCIdempotencyKey(tinormalize.CanonicalID(ni))
				assertCommit(t, out[0], commit.KindTIIoC, wantKey)
				if ni.Value != "https://evil.com/path" {
					t.Fatalf("normalized value = %q", ni.Value)
				}
			},
		},
		{
			name:    "KindTIIoCRaw_invalidJSON",
			kind:    harvest.KindTIIoCRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "KindTIIoCRaw_normalize_skip",
			kind:    harvest.KindTIIoCRaw,
			payload: mustJSON(t, invalidIOC),
			wantLen: 0,
		},
		{
			name:    "KindTIReportRaw",
			kind:    harvest.KindTIReportRaw,
			payload: mustJSON(t, tidomain.Report{Title: "Report", Provider: "vendor", Link: "https://ti.example/report"}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				sid := tinormalize.ReportStableID("https://ti.example/report")
				assertCommit(t, out[0], commit.KindTIReport, commit.TIReportIdempotencyKey(sid))
			},
		},
		{
			name:    "KindTIReportRaw_invalidJSON",
			kind:    harvest.KindTIReportRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "KindTICampaignRaw",
			kind:    harvest.KindTICampaignRaw,
			payload: mustJSON(t, tidomain.Campaign{ID: " c1 ", Name: " Camp ", Source: "feed"}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				assertCommit(t, out[0], commit.KindTICampaign, commit.TICampaignIdempotencyKey("c1"))
				var c tidomain.Campaign
				decodePayload(t, out[0].Payload, &c)
				if c.ID != "c1" || c.Name != "Camp" {
					t.Fatalf("normalized campaign %#v", c)
				}
			},
		},
		{
			name:    "KindTICampaignRaw_missing_id_name",
			kind:    harvest.KindTICampaignRaw,
			payload: mustJSON(t, tidomain.Campaign{ID: "  ", Name: "  "}),
			wantErr: "ti campaign requires id and name",
		},
		{
			name:    "KindTICampaignRaw_invalidJSON",
			kind:    harvest.KindTICampaignRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "KindTIClusterRaw",
			kind:    harvest.KindTIClusterRaw,
			payload: mustJSON(t, tidomain.Cluster{ID: " cl1 ", Name: " Cluster ", Description: "d"}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				assertCommit(t, out[0], commit.KindTICluster, commit.TIClusterIdempotencyKey("cl1"))
				var cl tidomain.Cluster
				decodePayload(t, out[0].Payload, &cl)
				if cl.Name != "Cluster" {
					t.Fatalf("cluster %#v", cl)
				}
			},
		},
		{
			name:    "KindTIClusterRaw_invalidJSON",
			kind:    harvest.KindTIClusterRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "KindTIActorRaw_withID",
			kind:    harvest.KindTIActorRaw,
			payload: mustJSON(t, tidomain.Actor{ID: " apt-1 ", Name: "Actor"}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				assertCommit(t, out[0], commit.KindTIActor, commit.TIActorIdempotencyKey("apt-1"))
			},
		},
		{
			name:    "KindTIActorRaw_stableID",
			kind:    harvest.KindTIActorRaw,
			payload: mustJSON(t, tidomain.Actor{Name: "APT29"}),
			wantLen: 1,
			check: func(t *testing.T, out []*commit.Envelope) {
				sid := tinormalize.ActorStableID("APT29")
				assertCommit(t, out[0], commit.KindTIActor, commit.TIActorIdempotencyKey(sid))
			},
		},
		{
			name:    "KindTIActorRaw_invalidJSON",
			kind:    harvest.KindTIActorRaw,
			payload: []byte(`{`),
			wantErr: "unexpected",
		},
		{
			name:    "unknown_kind",
			kind:    "scrape_ti_unknown",
			payload: []byte(`{}`),
			wantErr: "pipeline ti: unknown kind",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := &harvest.Envelope{
				SchemaVersion: harvest.CurrentSchemaVersion,
				Source:        harvest.SourceTI,
				Kind:          tt.kind,
				Payload:       tt.payload,
			}
			out, err := Transform(context.Background(), env)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != tt.wantLen {
				t.Fatalf("len(out) = %d want %d", len(out), tt.wantLen)
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

func jsonlLinePayload(t *testing.T, inner JSONLEnvelope) []byte {
	t.Helper()
	line, err := json.Marshal(inner)
	if err != nil {
		t.Fatal(err)
	}
	return mustJSON(t, harvest.TIJSONLLine{Line: line})
}

func decodePayload(t *testing.T, raw json.RawMessage, dst any) {
	t.Helper()
	if err := json.Unmarshal(raw, dst); err != nil {
		t.Fatal(err)
	}
}

func assertCommit(t *testing.T, e *commit.Envelope, kind string, idempotency string) {
	t.Helper()
	if e.Source != commit.SourceTI {
		t.Fatalf("source = %q want %q", e.Source, commit.SourceTI)
	}
	if e.Kind != kind {
		t.Fatalf("kind = %s want %s", e.Kind, kind)
	}
	if e.IdempotencyKey != idempotency {
		t.Fatalf("idempotency %q want %q", e.IdempotencyKey, idempotency)
	}
}
