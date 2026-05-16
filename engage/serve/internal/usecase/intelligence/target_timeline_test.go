package intelligence

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
)

type mockAuditReader struct {
	events []audit.Event
}

func (m *mockAuditReader) Recent(limit int) ([]audit.Event, error) {
	if limit <= 0 || limit >= len(m.events) {
		return m.events, nil
	}
	return m.events[:limit], nil
}

func (m *mockAuditReader) ExportNDJSON(_ time.Time) ([]byte, error) {
	return nil, nil
}

type mockVeilTimeline struct {
	search map[string]json.RawMessage
	ctx    json.RawMessage
}

func (m *mockVeilTimeline) Enabled() bool { return true }

func (m *mockVeilTimeline) Categories(context.Context) (json.RawMessage, error) {
	return nil, nil
}

func (m *mockVeilTimeline) Search(_ context.Context, cat, _ string) (json.RawMessage, error) {
	return m.search[cat], nil
}

func (m *mockVeilTimeline) EngageContext(_ context.Context, _ string) (json.RawMessage, error) {
	return m.ctx, nil
}

func (m *mockVeilTimeline) GetNode(context.Context, string) (json.RawMessage, error) {
	return nil, nil
}

func (m *mockVeilTimeline) Neighbors(context.Context, string, int) (json.RawMessage, error) {
	return nil, nil
}

func TestTargetTimeline_mergesAuditAndGraph(t *testing.T) {
	now := time.Now().UTC()
	s := &Service{
		Audit: &mockAuditReader{events: []audit.Event{{
			Tool: "nmap_scan", Target: "https://example.com", At: now, Success: true,
		}}},
	}
	// Veil is *veilgraph.Client - use nil and only test audit path
	out := s.TargetTimeline(context.Background(), TargetTimelineRequest{Target: "https://example.com", Limit: 10})
	if len(out.AuditEvents) != 1 {
		t.Fatalf("audit events: %d", len(out.AuditEvents))
	}
	if len(out.Timeline) < 1 {
		t.Fatal("expected timeline entries")
	}
	if out.Host != "example.com" {
		t.Fatalf("host: %q", out.Host)
	}
}

func TestTargetTimeline_withGraphReader(t *testing.T) {
	ctxRaw := json.RawMessage(`{"found":true,"context":{"tool_runs":[{"props":{"at":"2026-01-01T00:00:00Z","tool":"nmap"}}]}}`)
	s := &Service{
		Veil: &mockVeilTimeline{
			search: map[string]json.RawMessage{"engage": json.RawMessage(`{}`)},
			ctx:    ctxRaw,
		},
	}
	out := s.TargetTimeline(context.Background(), TargetTimelineRequest{
		Target: "https://example.com", Limit: 10, IncludeGraph: true,
	})
	if out.Host != "example.com" {
		t.Fatalf("host %q", out.Host)
	}
	if len(out.EngageContext) == 0 {
		t.Fatal("expected engage context")
	}
}
