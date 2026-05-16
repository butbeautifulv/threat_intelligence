package components

import (
	"context"

	engageevents "github.com/butbeautifulv/veil/pkg/engage/events"
)

type findingBridge struct {
	pub *engageevents.Publisher
}

func (b findingBridge) PublishFinding(ctx context.Context, tool, target, title, severity, description string) error {
	return b.pub.PublishFinding(ctx, engageevents.FindingEvent{
		Tool: tool, Target: target, Title: title, Severity: severity, Description: description,
	})
}
