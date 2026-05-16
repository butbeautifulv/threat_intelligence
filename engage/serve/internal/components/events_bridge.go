package components

import (
	"context"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	"github.com/butbeautifulv/veil/engage/serve/internal/events"
)

type eventBridge struct {
	pub *events.Publisher
}

func (b eventBridge) PublishAudit(ctx context.Context, e audit.AuditEvent) error {
	return b.pub.PublishAudit(ctx, events.AuditEvent{
		Tool:    e.Tool,
		Target:  e.Target,
		Subject: e.Subject,
		Success: e.Success,
		At:      e.At,
	})
}
