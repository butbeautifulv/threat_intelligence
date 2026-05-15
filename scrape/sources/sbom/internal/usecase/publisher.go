package usecase

import "context"

// rawPublisher publishes scrapev1 payloads (factory.DomainPublisher implements this).
type rawPublisher interface {
	Publish(ctx context.Context, kind, contentKey string, payload any) error
}
