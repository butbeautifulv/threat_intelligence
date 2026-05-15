package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"
	"ingestpub"

	"coderules/internal/feeds/cwe"
	gh "coderules/internal/feeds/github"
	"coderules/internal/parse"

	"gopkg.in/yaml.v3"
)

// Options for one run (CLI overrides env).
type Options struct {
	Sources              []string
	MaxCWE, MaxSemgrep   int
	MaxCodeQL            int
	NATSURL, NATSSubject string
}

// Runner runs coderules ingest via NATS (ingest-worker → Neo4j).
type Runner struct {
	log *slog.Logger
	pub *ingestpub.JetStreamPublisher
	opt Options
}

func NewRunner(log *slog.Logger, pub *ingestpub.JetStreamPublisher, opt Options) *Runner {
	return &Runner{log: log, pub: pub, opt: opt}
}

func (r *Runner) enabled(name string) bool {
	for _, s := range r.opt.Sources {
		if s == name {
			return true
		}
	}
	return false
}

func (r *Runner) Run(ctx context.Context) error {
	if r.pub == nil {
		return fmt.Errorf("coderules: NATS publisher required")
	}
	if r.enabled("cwe") {
		r.log.Info("ingesting CWE catalog (MITRE zip)…")
		b := &cweNATSBridge{ctx: ctx, pub: r.pub, sub: r.opt.NATSSubject}
		if err := cwe.StreamMITRE(ctx, b, r.opt.MaxCWE); err != nil {
			return err
		}
	}
	if r.enabled("semgrep") {
		if err := r.runSemgrep(ctx); err != nil {
			return err
		}
	}
	if r.enabled("codeql") {
		if err := r.runCodeQL(ctx); err != nil {
			return err
		}
	}
	return nil
}

type cweNATSBridge struct {
	ctx context.Context
	pub *ingestpub.JetStreamPublisher
	sub string
}

func (b *cweNATSBridge) UpsertCWECatalog(ctx context.Context, cweID, name, description, status string) error {
	env, err := ingestv1.NewEnvelope(ingestv1.SourceCoderules, ingestv1.KindCoderulesCWERow, ingestv1.CoderulesCWEIdempotencyKey(cweID), ingestv1.CoderulesCWEPayload{
		ID: cweID, Name: name, Description: description, Status: status,
	})
	if err != nil {
		return err
	}
	return b.pub.PublishJSON(b.ctx, b.sub, env)
}

func (r *Runner) runSemgrep(ctx context.Context) error {
	r.log.Info("ingesting Semgrep community rules (subset)…")
	g := gh.NewClient()
	const owner, repo = "semgrep", "semgrep-rules"
	seeds := []string{"python", "javascript", "java", "go", "csharp", "dockerfile", "yaml", "bash"}
	var q []string
	for _, s := range seeds {
		q = append(q, s)
	}
	n := 0
	for len(q) > 0 && n < r.opt.MaxSemgrep {
		dir := q[0]
		q = q[1:]
		items, err := g.ListDir(ctx, owner, repo, dir)
		if err != nil {
			continue
		}
		for _, it := range items {
			if n >= r.opt.MaxSemgrep {
				break
			}
			if it.Type == "dir" && !strings.HasPrefix(it.Name, ".") {
				q = append(q, it.Path)
				continue
			}
			if it.Type != "file" || (!strings.HasSuffix(it.Name, ".yml") && !strings.HasSuffix(it.Name, ".yaml")) {
				continue
			}
			raw, err := g.FetchText(ctx, it.DownloadURL)
			if err != nil {
				continue
			}
			var root map[string]any
			if err := yaml.Unmarshal(raw, &root); err != nil {
				continue
			}
			ruleID, title := parse.SemgrepMeta(root, it.Name)
			lang := strings.Split(it.Path, "/")[0]
			cwes := parse.SemgrepCWES(root)
			env, err := ingestv1.NewEnvelope(ingestv1.SourceCoderules, ingestv1.KindCoderulesSemgrep, ingestv1.CoderulesSemgrepIdempotencyKey(it.Path), ingestv1.CoderulesSemgrepPayload{
				Path: it.Path, Language: lang, RuleID: ruleID, Title: title, RawYAML: string(raw), CWEs: cwes,
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
	r.log.Info("semgrep rules ingested", slog.Int("count", n))
	return nil
}

func (r *Runner) runCodeQL(ctx context.Context) error {
	r.log.Info("ingesting CodeQL queries (subset)…")
	g := gh.NewClient()
	const path = "javascript/ql/src/Security/CWE-079"
	items, err := g.ListDir(ctx, "github", "codeql", path)
	if err != nil {
		return err
	}
	n := 0
	for _, it := range items {
		if n >= r.opt.MaxCodeQL {
			break
		}
		if it.Type != "file" || !strings.HasSuffix(it.Name, ".ql") {
			continue
		}
		raw, err := g.FetchText(ctx, it.DownloadURL)
		if err != nil {
			continue
		}
		body := string(raw)
		name := it.Name
		cwes := parse.CodeQLCWES(body)
		env, err := ingestv1.NewEnvelope(ingestv1.SourceCoderules, ingestv1.KindCoderulesCodeQL, ingestv1.CoderulesCodeQLIdempotencyKey(it.Path), ingestv1.CoderulesCodeQLPayload{
			Path: it.Path, Name: name, Lang: "javascript", Body: body, CWEs: cwes,
		})
		if err != nil {
			return err
		}
		if err := r.pub.PublishJSON(ctx, r.opt.NATSSubject, env); err != nil {
			return err
		}
		n++
	}
	r.log.Info("codeql rules ingested", slog.Int("count", n))
	return nil
}
