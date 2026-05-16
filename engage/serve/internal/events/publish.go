package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
)

// AuditEvent is published to the Veil pipeline bus when cross-layer events are enabled.
type AuditEvent struct {
	Source  string    `json:"source"`
	Tool    string    `json:"tool"`
	Target  string    `json:"target"`
	Subject string    `json:"subject"`
	Success bool      `json:"success"`
	At      time.Time `json:"at"`
}

// FindingEvent is published when smart-scan discovers vulnerabilities.
type FindingEvent struct {
	Tool        string `json:"tool"`
	Target      string `json:"target"`
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// Publisher sends engage events to NATS JetStream.
type Publisher struct {
	nc             *nats.Conn
	js             nats.JetStreamContext
	auditSubject   string
	findingSubject string
}

func Connect(url, auditSubject string) (*Publisher, error) {
	return ConnectWithSubjects(url, auditSubject, "engage.events.finding")
}

func ConnectWithSubjects(url, auditSubject, findingSubject string) (*Publisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}
	_, _ = js.AddStream(&nats.StreamConfig{Name: "ENGAGE_EVENTS", Subjects: []string{"engage.events.>"}})
	if findingSubject == "" {
		findingSubject = "engage.events.finding"
	}
	return &Publisher{nc: nc, js: js, auditSubject: auditSubject, findingSubject: findingSubject}, nil
}

func (p *Publisher) Close() {
	if p != nil && p.nc != nil {
		_ = p.nc.Drain()
	}
}

func (p *Publisher) PublishAudit(ctx context.Context, e AuditEvent) error {
	if p == nil || p.js == nil {
		return nil
	}
	e.Source = "veil-engage"
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = p.js.Publish(p.auditSubject, b)
	return err
}

func (p *Publisher) PublishFinding(ctx context.Context, e FindingEvent) error {
	if p == nil || p.js == nil {
		return nil
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = p.js.Publish(p.findingSubject, b)
	return err
}
