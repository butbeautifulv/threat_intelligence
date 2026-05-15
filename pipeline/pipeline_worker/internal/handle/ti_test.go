package handle

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/butbeautifulv/threat_intelligence/pipeline/contract/ingestv1"
	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"
	"github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/tidomain"
	"github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/ti"
)

func TestHandleTI_IoCRaw_normalizesInPipeline(t *testing.T) {
	raw := tidomain.IOC{
		Type:   tidomain.IOCURL,
		Value:  "HTTPS://Evil.COM/path",
		Source: "urlhaus",
	}
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	env := &scrapev1.Envelope{
		SchemaVersion: scrapev1.CurrentSchemaVersion,
		Source:        scrapev1.SourceTI,
		Kind:          scrapev1.KindTIIoCRaw,
		ContentKey:    "ti:ioc:url:https://evil.com/path",
		ScrapedAt:     "2026-05-15T12:00:00Z",
		Payload:       payload,
	}
	out, err := HandleTI(context.Background(), env)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("len(out) = %d", len(out))
	}
	if out[0].Kind != ingestv1.KindTIIoC {
		t.Fatalf("kind = %s", out[0].Kind)
	}
	var ni tidomain.IOC
	if err := json.Unmarshal(out[0].Payload, &ni); err != nil {
		t.Fatal(err)
	}
	wantID := tinormalize.CanonicalID(ni)
	if out[0].IdempotencyKey != ingestv1.TIIoCIdempotencyKey(wantID) {
		t.Fatalf("idempotency %q want %q", out[0].IdempotencyKey, ingestv1.TIIoCIdempotencyKey(wantID))
	}
	if ni.Value != "https://evil.com/path" {
		t.Fatalf("normalized value = %q", ni.Value)
	}
}
