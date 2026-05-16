package audit

import (
	"context"
	"log/slog"
	"time"

	engageevents "github.com/butbeautifulv/veil/pkg/engage/events"
	"github.com/butbeautifulv/veil/engage/serve/internal/telemetry"
)

// EventPublisher optional cross-layer bus (e.g. NATS).
type EventPublisher interface {
	PublishAudit(ctx context.Context, e engageevents.AuditEvent) error
}

// Logger records tool invocations to slog and optional JSONL store.
type Logger struct {
	log    *slog.Logger
	store  Appender
	events EventPublisher
}

func New(l *slog.Logger) *Logger {
	return &Logger{log: l}
}

func NewWithStore(l *slog.Logger, store Appender) *Logger {
	return &Logger{log: l, store: store}
}

func (a *Logger) SetEventPublisher(p EventPublisher) {
	a.events = p
}

func (a *Logger) ToolRun(subject, tool, target, jobID string, success bool, errMsg string) {
	at := time.Now().UTC()
	if a != nil && a.log != nil {
		a.log.Info("engage tool run",
			slog.String("subject", subject),
			slog.String("tool", tool),
			slog.String("target", target),
			slog.String("job_id", jobID),
			slog.Bool("success", success),
			slog.String("error", errMsg),
			slog.Time("at", at),
		)
	}
	if a != nil && a.store != nil {
		_ = a.store.Append(Event{
			Subject: subject,
			Tool:    tool,
			Target:  target,
			JobID:   jobID,
			Success: success,
			Error:   errMsg,
			At:      at,
		})
	}
	if a != nil && a.events != nil {
		_ = a.events.PublishAudit(context.Background(), engageevents.AuditEvent{
			Tool: tool, Target: target, Subject: subject, Success: success, At: at,
		})
	}
	telemetry.RecordAuditEvent()
}
