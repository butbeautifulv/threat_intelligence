package intelligence

import (
	"context"
	"encoding/json"
	"testing"
)

type mockGraphReader struct {
	search map[string]json.RawMessage
	ctx    json.RawMessage
}

func (m *mockGraphReader) Enabled() bool { return true }

func (m *mockGraphReader) Categories(context.Context) (json.RawMessage, error) {
	return nil, nil
}

func (m *mockGraphReader) Search(_ context.Context, cat, _ string) (json.RawMessage, error) {
	return m.search[cat], nil
}

func (m *mockGraphReader) EngageContext(context.Context, string) (json.RawMessage, error) {
	return m.ctx, nil
}

func (m *mockGraphReader) GetNode(context.Context, string) (json.RawMessage, error) {
	return nil, nil
}

func (m *mockGraphReader) Neighbors(context.Context, string, int) (json.RawMessage, error) {
	return nil, nil
}

func (m *mockGraphReader) PlaybooksByTechnique(context.Context, string) (json.RawMessage, error) {
	return nil, nil
}

func (m *mockGraphReader) PlaybookRecommendTools(context.Context, string, string) (json.RawMessage, error) {
	return nil, nil
}

func TestLoadTargetGraph_mergesHitsAndContext(t *testing.T) {
	ctxRaw := json.RawMessage(`{"found":true,"context":{"vulnerabilities":[{"props":{"cve":"CVE-2024-1"}}]}}`)
	veil := &mockGraphReader{
		search: map[string]json.RawMessage{
			"vuln": json.RawMessage(`{"nodes":[]}`),
		},
		ctx: ctxRaw,
	}
	state := LoadTargetGraph(context.Background(), veil, "https://Example.COM", TargetGraphLoadOpts{
		IncludeEngageContext: true,
	})
	if state.Host != "example.com" {
		t.Fatalf("host %q", state.Host)
	}
	if len(state.Hits) != 1 {
		t.Fatalf("hits %d", len(state.Hits))
	}
	if !state.EngageFound || len(state.RelatedCVEs) != 1 {
		t.Fatalf("engage found=%v cves=%v", state.EngageFound, state.RelatedCVEs)
	}
}

func TestParseEngageContextCVEs_dedupes(t *testing.T) {
	raw := json.RawMessage(`{"context":{"vulnerabilities":[{"props":{"cve":"CVE-1"}},{"props":{"id":"CVE-1"}}]}}`)
	ids := ParseEngageContextCVEs(raw)
	if len(ids) != 1 || ids[0] != "CVE-1" {
		t.Fatalf("ids %v", ids)
	}
}
