package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"
	"ingestpub"
	"sbom/internal/feeds/ghsa"
	"sbom/internal/feeds/osv"
	"sbom/internal/repository"
)

// Options overrides env defaults for one run (from CLI flags).
type Options struct {
	Sources     []string
	MaxCVE      int
	MaxGHSA     int
	GHSAMinYear int
	IngestMode  string
	NATSURL     string
	NATSSubject string
}

// Runner orchestrates SBOM fetch → Neo4j or NATS.
type Runner struct {
	log    *slog.Logger
	store  repository.SBOMWriter
	pub    *ingestpub.JetStreamPublisher
	opt    Options
}

func NewRunner(log *slog.Logger, store repository.SBOMWriter, pub *ingestpub.JetStreamPublisher, opt Options) *Runner {
	return &Runner{log: log, store: store, pub: pub, opt: opt}
}

func (r *Runner) sourceEnabled(name string) bool {
	for _, s := range r.opt.Sources {
		if s == name {
			return true
		}
	}
	return false
}

func (r *Runner) Run(ctx context.Context) error {
	if r.opt.IngestMode == "nats" && r.pub == nil {
		return fmt.Errorf("usecase: INGEST_MODE=nats requires NATS publisher")
	}
	if r.sourceEnabled("osv") {
		if err := r.runOSV(ctx); err != nil {
			return err
		}
	}
	if r.sourceEnabled("ghsa") {
		if err := r.runGHSA(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runOSV(ctx context.Context) error {
	cves, err := r.store.ListCVEs(ctx, r.opt.MaxCVE)
	if err != nil {
		return err
	}
	oc := osv.NewClient()
	for i, cve := range cves {
		doc, err := oc.GetVuln(ctx, cve)
		if err != nil {
			r.log.Warn("osv fetch", slog.String("cve", cve), slog.String("err", err.Error()))
			continue
		}
		id, _ := doc["id"].(string)
		aff, _ := doc["affected"].([]any)
		var packs []map[string]any
		for _, a := range aff {
			if m, ok := a.(map[string]any); ok {
				packs = append(packs, m)
			}
		}
		if r.opt.IngestMode == "nats" {
			key := "sbom:osv:" + cve
			env, err := ingestv1.NewEnvelope(ingestv1.SourceSBOM, ingestv1.KindSBOMOSVRecord, key, ingestv1.SBOMOSVPayload{
				OSVID: id, CVE: cve, Affected: packs,
			})
			if err != nil {
				return err
			}
			if err := r.pub.PublishJSON(ctx, r.opt.NATSSubject, env); err != nil {
				return fmt.Errorf("publish osv %s: %w", cve, err)
			}
		} else {
			if err := r.store.UpsertFromOSVVuln(ctx, id, cve, packs); err != nil {
				return fmt.Errorf("store osv %s: %w", cve, err)
			}
		}
		if (i+1)%20 == 0 {
			r.log.Info("osv progress", slog.Int("done", i+1), slog.Int("total", len(cves)))
		}
		time.Sleep(150 * time.Millisecond)
	}
	return nil
}

func (r *Runner) runGHSA(ctx context.Context) error {
	gc := ghsa.NewClientFromEnv()
	paths, err := gc.CollectAdvisoryPaths(ctx, r.opt.MaxGHSA, r.opt.GHSAMinYear)
	if err != nil {
		return err
	}
	for i, p := range paths {
		doc, err := gc.FetchAdvisoryJSON(ctx, p)
		if err != nil {
			r.log.Warn("ghsa fetch", slog.String("path", p), slog.String("err", err.Error()))
			continue
		}
		if r.opt.IngestMode == "nats" {
			env, err := ingestv1.NewEnvelope(ingestv1.SourceSBOM, ingestv1.KindSBOMGHSADocument, ingestv1.SBOMGHSAIdempotencyKey(p), ingestv1.SBOMGHSAPathPayload{Path: p, Doc: doc})
			if err != nil {
				return err
			}
			if err := r.pub.PublishJSON(ctx, r.opt.NATSSubject, env); err != nil {
				return fmt.Errorf("publish ghsa %s: %w", p, err)
			}
		} else {
			if err := r.store.UpsertGHSA(ctx, doc); err != nil {
				return fmt.Errorf("store ghsa %s: %w", p, err)
			}
		}
		if (i+1)%10 == 0 {
			r.log.Info("ghsa progress", slog.Int("done", i+1), slog.Int("total", len(paths)))
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}
