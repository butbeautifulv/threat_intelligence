package workflow

import "context"

// FindingBus publishes scan findings to the cross-layer event bus.
type FindingBus interface {
	PublishFinding(ctx context.Context, tool, target, title, severity, description string) error
}
