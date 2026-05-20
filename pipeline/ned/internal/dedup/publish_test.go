package dedup

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	tidomain "github.com/butbeautifulv/veil/pkg/ti/domain"
)

type recordPublisher struct {
	subject string
	envs    []*commit.Envelope
}

func (r *recordPublisher) PublishJSON(_ context.Context, subject string, env *commit.Envelope) error {
	r.subject = subject
	r.envs = append(r.envs, env)
	return nil
}

func TestProcessScrapeMessage_tiIoCRaw_publishesIngest(t *testing.T) {
	raw := tidomain.IOC{Type: tidomain.IOCURL, Value: "https://evil.com/x", Source: "test"}
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	env := &harvest.Envelope{
		SchemaVersion: harvest.CurrentSchemaVersion,
		Source:        harvest.SourceTI,
		Kind:          harvest.KindTIIoCRaw,
		ContentKey:    "ti:ioc:url:https://evil.com/x",
		ScrapedAt:     "2026-05-15T12:00:00Z",
		Payload:       payload,
	}
	rec := &recordPublisher{}
	if err := ProcessScrapeMessage(context.Background(), rec, "ingest.ti.events", env); err != nil {
		t.Fatal(err)
	}
	if rec.subject != "ingest.ti.events" {
		t.Fatalf("subject %q", rec.subject)
	}
	if len(rec.envs) != 1 {
		t.Fatalf("published %d envelopes", len(rec.envs))
	}
	if rec.envs[0].Kind != commit.KindTIIoC {
		t.Fatalf("kind %s", rec.envs[0].Kind)
	}
	if err := rec.envs[0].Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPublishIngest_multiEnvelope(t *testing.T) {
	e1, err := commit.NewEnvelope(commit.SourceTI, commit.KindTIIoC, "ti:ioc:1", map[string]string{"k": "1"})
	if err != nil {
		t.Fatal(err)
	}
	e2, err := commit.NewEnvelope(commit.SourceTI, commit.KindTIIoC, "ti:ioc:2", map[string]string{"k": "2"})
	if err != nil {
		t.Fatal(err)
	}
	rec := &recordPublisher{}
	if err := PublishIngest(context.Background(), rec, "ingest.ti.events", []*commit.Envelope{e1, e2}); err != nil {
		t.Fatal(err)
	}
	if rec.subject != "ingest.ti.events" {
		t.Fatalf("subject %q", rec.subject)
	}
	if len(rec.envs) != 2 {
		t.Fatalf("published %d envelopes", len(rec.envs))
	}
}

type failAfterPublisher struct {
	n   int
	err error
}

func (f *failAfterPublisher) PublishJSON(_ context.Context, _ string, _ *commit.Envelope) error {
	f.n++
	if f.n >= 2 {
		return f.err
	}
	return nil
}

func TestPublishIngest_publishError(t *testing.T) {
	e1, err := commit.NewEnvelope(commit.SourceTI, commit.KindTIIoC, "ti:ioc:1", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	e2, err := commit.NewEnvelope(commit.SourceTI, commit.KindTIIoC, "ti:ioc:2", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	pubErr := errors.New("publish failed")
	pub := &failAfterPublisher{err: pubErr}
	err = PublishIngest(context.Background(), pub, "ingest.events", []*commit.Envelope{e1, e2})
	if !errors.Is(err, pubErr) {
		t.Fatalf("err %v", err)
	}
	if pub.n != 2 {
		t.Fatalf("publish calls %d", pub.n)
	}
}

func TestProcessScrapeMessage_unknownSource(t *testing.T) {
	env := &harvest.Envelope{
		SchemaVersion: harvest.CurrentSchemaVersion,
		Source:        harvest.SourceBrowser,
		Kind:          harvest.KindBrowserInspectRaw,
		ContentKey:    "browser:https://example.com",
		ScrapedAt:     "2026-05-15T12:00:00Z",
		Payload:       json.RawMessage(`{"url":"https://example.com"}`),
	}
	rec := &recordPublisher{}
	err := ProcessScrapeMessage(context.Background(), rec, "ingest.events", env)
	if err == nil {
		t.Fatal("expected unknown source error")
	}
	if len(rec.envs) != 0 {
		t.Fatalf("published %d envelopes", len(rec.envs))
	}
}
