package handle

import (
	"context"
	"fmt"

	"github.com/butbeautifulv/threat_intelligence/pipeline/contract/ingestv1"
	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"
	"ingestpub"
)

// Transform scrapev1 → zero or more ingestv1 envelopes.
func Transform(ctx context.Context, env *scrapev1.Envelope) ([]*ingestv1.Envelope, error) {
	switch env.Source {
	case scrapev1.SourceDS:
		return HandleDS(ctx, nil, "", env)
	case scrapev1.SourceTI:
		return HandleTI(ctx, env)
	case scrapev1.SourceVuln:
		return HandleVuln(ctx, env)
	case scrapev1.SourceLola:
		return HandleLola(ctx, env)
	case scrapev1.SourceSBOM:
		return HandleSBOM(ctx, env)
	case scrapev1.SourceCoderules:
		return HandleCoderules(ctx, env)
	case scrapev1.SourceNuclei:
		return HandleNuclei(ctx, env)
	default:
		return nil, fmt.Errorf("pipeline: unknown source %q", env.Source)
	}
}

// ProcessMessage transforms and publishes to ingest JetStream.
func ProcessMessage(ctx context.Context, ingestPub *ingestpub.JetStreamPublisher, ingestSubject string, env *scrapev1.Envelope) error {
	out, err := Transform(ctx, env)
	if err != nil {
		return err
	}
	return PublishIngest(ctx, ingestPub, ingestSubject, out)
}
