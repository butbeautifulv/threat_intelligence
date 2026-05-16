package components

import (
	"context"

	"github.com/butbeautifulv/veil/engage/serve/internal/events"
)

type findingBridge struct {
	pub *events.Publisher
}

func (b findingBridge) PublishFinding(ctx context.Context, tool, target, title, severity, description string) error {
	return b.pub.PublishFinding(ctx, events.FindingEvent{
		Tool: tool, Target: target, Title: title, Severity: severity, Description: description,
	})
}
