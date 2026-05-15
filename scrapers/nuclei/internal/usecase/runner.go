package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"
	"ingestpub"

	gh "nuclei/internal/feeds/github"
	"nuclei/internal/parse"
)

// Options for one run (CLI overrides env).
type Options struct {
	MaxTemplates         int
	YearsCSV             string
	NATSURL, NATSSubject string
}

// Runner runs nuclei template ingest via NATS (ingest-worker → Neo4j).
type Runner struct {
	log *slog.Logger
	pub *ingestpub.JetStreamPublisher
	opt Options
}

func NewRunner(log *slog.Logger, pub *ingestpub.JetStreamPublisher, opt Options) *Runner {
	return &Runner{log: log, pub: pub, opt: opt}
}

func (r *Runner) Run(ctx context.Context) error {
	if r.pub == nil {
		return fmt.Errorf("nuclei: NATS publisher required")
	}

	g := gh.NewClient()
	const owner, repo = "projectdiscovery", "nuclei-templates"
	n := 0
	for _, y := range strings.Split(r.opt.YearsCSV, ",") {
		if n >= r.opt.MaxTemplates {
			break
		}
		year := strings.TrimSpace(y)
		if year == "" {
			continue
		}
		base := "http/cves/" + year
		items, err := g.ListDir(ctx, owner, repo, base)
		if err != nil {
			r.log.Warn("list dir", slog.String("base", base), slog.String("err", err.Error()))
			continue
		}
		for _, it := range items {
			if n >= r.opt.MaxTemplates {
				break
			}
			if it.Type != "file" || (!strings.HasSuffix(it.Name, ".yaml") && !strings.HasSuffix(it.Name, ".yml")) {
				continue
			}
			raw, err := g.FetchText(ctx, it.DownloadURL)
			if err != nil {
				continue
			}
			p, err := parse.ParseYAML(raw)
			if err != nil || p.TemplateID == "" {
				continue
			}
			env, err := ingestv1.NewEnvelope(ingestv1.SourceNuclei, ingestv1.KindNucleiTemplate, ingestv1.NucleiTemplateIdempotencyKey(it.Path), ingestv1.NucleiTemplatePayload{
				Path: it.Path, TemplateID: p.TemplateID, Name: p.Name, Severity: p.Severity, TagsJSON: p.TagsJSON,
				CVE: p.CVE, CWE: p.CWE, RawYAML: string(raw),
			})
			if err != nil {
				return err
			}
			if err := r.pub.PublishJSON(ctx, r.opt.NATSSubject, env); err != nil {
				return err
			}
			n++
		}
	}
	r.log.Info("nuclei templates ingested", slog.Int("count", n))
	return nil
}
