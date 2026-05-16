package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/natsjet"
	natsgo "github.com/nats-io/nats.go"
)

// engageAuditWire matches engage/serve/internal/events.AuditEvent JSON.
type engageAuditWire struct {
	Source  string    `json:"source"`
	Tool    string    `json:"tool"`
	Target  string    `json:"target"`
	Subject string    `json:"subject"`
	Success bool      `json:"success"`
	At      time.Time `json:"at"`
}

// engageFindingWire matches engage/serve/internal/events.FindingEvent JSON.
type engageFindingWire struct {
	Tool        string `json:"tool"`
	Target      string `json:"target"`
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// EnsureEngageEventsStream creates the ENGAGE_EVENTS stream for audit/finding subjects.
func EnsureEngageEventsStream(js natsgo.JetStreamContext) error {
	return natsjet.EnsureStream(js, "ENGAGE_EVENTS", []string{"engage.events.>"})
}

// RunEngageEventsConsumer pulls engage JetStream events and publishes ingest envelopes.
func RunEngageEventsConsumer(ctx context.Context, log *slog.Logger, natsURL, engageFilter, ingestToolRunSubject, ingestFindingSubject string) error {
	if log == nil {
		log = slog.Default()
	}
	pub, err := ConnectJetStreamAndStream(natsURL)
	if err != nil {
		return err
	}
	defer pub.Close()
	js := pub.conn.JS
	if err := EnsureEngageEventsStream(js); err != nil {
		return fmt.Errorf("engage events stream: %w", err)
	}
	if engageFilter == "" {
		engageFilter = "engage.events.>"
	}
	if ingestToolRunSubject == "" {
		ingestToolRunSubject = "ingest.engage.tool_run"
	}
	if ingestFindingSubject == "" {
		ingestFindingSubject = "ingest.engage.finding"
	}
	sub, err := js.PullSubscribe(engageFilter, "engage-events-bridge", natsgo.BindStream("ENGAGE_EVENTS"))
	if err != nil {
		return fmt.Errorf("pull subscribe: %w", err)
	}
	defer sub.Unsubscribe()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		msgs, err := sub.Fetch(10, natsgo.MaxWait(2*time.Second))
		if err != nil {
			if err == natsgo.ErrTimeout {
				continue
			}
			return err
		}
		for _, m := range msgs {
			if err := handleEngageMsg(ctx, pub, ingestToolRunSubject, ingestFindingSubject, m); err != nil {
				log.Warn("engage event", slog.Any("err", err))
				_ = m.Nak()
				continue
			}
			_ = m.Ack()
		}
	}
}

func handleEngageMsg(ctx context.Context, pub *JetStreamPublisher, ingestToolRunSubject, ingestFindingSubject string, m *natsgo.Msg) error {
	if strings.Contains(m.Subject, ".finding") {
		return handleEngageFinding(ctx, pub, ingestFindingSubject, m)
	}
	return handleEngageToolRun(ctx, pub, ingestToolRunSubject, m)
}

func handleEngageToolRun(ctx context.Context, pub *JetStreamPublisher, ingestSubject string, m *natsgo.Msg) error {
	var wire engageAuditWire
	if err := json.Unmarshal(m.Data, &wire); err != nil {
		return err
	}
	if strings.TrimSpace(wire.Tool) == "" {
		return fmt.Errorf("empty tool")
	}
	at := wire.At.UTC()
	if at.IsZero() {
		at = time.Now().UTC()
	}
	atStr := at.Format(time.RFC3339)
	payload, err := json.Marshal(commit.EngageToolRunPayload{
		Tool:    wire.Tool,
		Target:  wire.Target,
		Subject: wire.Subject,
		Success: wire.Success,
		At:      atStr,
	})
	if err != nil {
		return err
	}
	env := &commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageToolRun,
		IdempotencyKey: commit.EngageToolRunIdempotencyKey(wire.Tool, wire.Target, atStr),
		Payload:        payload,
	}
	return pub.PublishJSON(ctx, ingestSubject, env)
}

func handleEngageFinding(ctx context.Context, pub *JetStreamPublisher, ingestSubject string, m *natsgo.Msg) error {
	var wire engageFindingWire
	if err := json.Unmarshal(m.Data, &wire); err != nil {
		return err
	}
	if strings.TrimSpace(wire.Title) == "" {
		return fmt.Errorf("empty finding title")
	}
	payload, err := json.Marshal(commit.EngageFindingPayload{
		Tool:        wire.Tool,
		Target:      wire.Target,
		Title:       wire.Title,
		Severity:    wire.Severity,
		Description: wire.Description,
	})
	if err != nil {
		return err
	}
	env := &commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageFinding,
		IdempotencyKey: commit.EngageFindingIdempotencyKey(wire.Tool, wire.Target, wire.Title),
		Payload:        payload,
	}
	return pub.PublishJSON(ctx, ingestSubject, env)
}
