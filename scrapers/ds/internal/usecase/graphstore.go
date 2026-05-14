package usecase

import "context"

// graphStore is implemented by Neo4j storage and NATS publishers.
type graphStore interface {
	EnsureSchema(ctx context.Context) error
	UpsertSigmaRule(ctx context.Context, id, title, level, logProduct, logService, tagsJSON, markdown, source string) error
	UpsertYaraRule(ctx context.Context, id, name, author, tagsJSON, markdown, source string) error
	UpsertAtomicTest(ctx context.Context, id, name, tactic, technique, execName, execCmd, markdown, source string) error
	UpsertCalderaAbility(ctx context.Context, id, name, tactic, techniqueID, markdown, source string) error
}
