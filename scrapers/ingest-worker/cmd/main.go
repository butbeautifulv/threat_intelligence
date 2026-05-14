package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
	"ingestpub"

	coderulesneo "coderules/storage/neo4j"
	nucleineo "nuclei/storage/neo4j"
	sbomneo "sbom/storage/neo4j"
)

func main() {
	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := run(rootCtx, log); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("shutdown complete")
}

func run(rootCtx context.Context, log *slog.Logger) error {
	natsURL := getenv("NATS_URL", "nats://localhost:4222")
	stream := getenv("NATS_INGEST_STREAM", "INGEST")
	durable := getenv("NATS_DURABLE", "ingest-worker")
	subject := getenv("NATS_SUBSCRIBE_SUBJECT", "ingest.appsec.>")
	batch := getenvInt("INGEST_BATCH", 10)
	maxWait := getenvDuration("INGEST_MAX_WAIT", 5*time.Second)

	neoCfg := sbomneo.Config{
		URI:      getenv("NEO4J_URI", "neo4j://localhost:7687"),
		Username: getenv("NEO4J_USER", "neo4j"),
		Password: getenv("NEO4J_PASS", "neo4jpassword"),
		Database: getenv("NEO4J_DB", "neo4j"),
	}

	sbomSt, err := sbomneo.New(rootCtx, neoCfg)
	if err != nil {
		return fmt.Errorf("neo4j sbom: %w", err)
	}
	defer sbomSt.Close(rootCtx)
	crSt, err := coderulesneo.New(rootCtx, coderulesneo.Config(neoCfg))
	if err != nil {
		return fmt.Errorf("neo4j coderules: %w", err)
	}
	defer crSt.Close(rootCtx)
	nuSt, err := nucleineo.New(rootCtx, nucleineo.Config(neoCfg))
	if err != nil {
		return fmt.Errorf("neo4j nuclei: %w", err)
	}
	defer nuSt.Close(rootCtx)

	for _, fn := range []func(context.Context) error{
		sbomSt.EnsureSchema, crSt.EnsureSchema, nuSt.EnsureSchema,
	} {
		if err := fn(rootCtx); err != nil {
			return fmt.Errorf("schema: %w", err)
		}
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("jetstream: %w", err)
	}
	if err := ingestpub.EnsureAppSecStream(js); err != nil {
		return fmt.Errorf("stream: %w", err)
	}

	sub, err := js.PullSubscribe(subject, durable, nats.BindStream(stream))
	if err != nil {
		return fmt.Errorf("pull subscribe stream=%s: %w", stream, err)
	}
	defer sub.Unsubscribe()

	log.Info("ingest-worker started", slog.String("nats", natsURL), slog.String("stream", stream), slog.String("durable", durable), slog.String("subject", subject))

	eg, ctx := errgroup.WithContext(rootCtx)
	eg.Go(func() error {
		return runPullLoop(ctx, log, sub, batch, maxWait, sbomSt, crSt, nuSt)
	})
	return eg.Wait()
}

func runPullLoop(ctx context.Context, log *slog.Logger, sub *nats.Subscription, batch int, maxWait time.Duration, sbomSt *sbomneo.Store, crSt *coderulesneo.Store, nuSt *nucleineo.Store) error {
	for {
		select {
		case <-ctx.Done():
			log.Info("consumer stopped")
			return nil
		default:
		}
		msgs, err := sub.Fetch(batch, nats.MaxWait(maxWait))
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				continue
			}
			log.Warn("fetch", slog.String("err", err.Error()))
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(time.Second):
			}
			continue
		}
		for _, m := range msgs {
			if err := handleMsg(ctx, log, m, sbomSt, crSt, nuSt); err != nil {
				log.Warn("message", slog.String("err", err.Error()))
				_ = m.NakWithDelay(2 * time.Second)
				continue
			}
			if err := m.Ack(); err != nil {
				log.Warn("ack", slog.String("err", err.Error()))
			}
		}
	}
}

func validateEnvelopeSource(e *ingestv1.Envelope) error {
	switch e.Kind {
	case ingestv1.KindSBOMOSVRecord, ingestv1.KindSBOMGHSADocument:
		if e.Source != ingestv1.SourceSBOM {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, ingestv1.SourceSBOM, e.Source)
		}
	case ingestv1.KindCoderulesCWERow, ingestv1.KindCoderulesSemgrep, ingestv1.KindCoderulesCodeQL:
		if e.Source != ingestv1.SourceCoderules {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, ingestv1.SourceCoderules, e.Source)
		}
	case ingestv1.KindNucleiTemplate:
		if e.Source != ingestv1.SourceNuclei {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, ingestv1.SourceNuclei, e.Source)
		}
	}
	return nil
}

func handleMsg(ctx context.Context, log *slog.Logger, m *nats.Msg, sbomSt *sbomneo.Store, crSt *coderulesneo.Store, nuSt *nucleineo.Store) error {
	var env ingestv1.Envelope
	if err := json.Unmarshal(m.Data, &env); err != nil {
		return fmt.Errorf("decode envelope: %w", err)
	}
	if err := env.Validate(); err != nil {
		return err
	}
	if err := validateEnvelopeSource(&env); err != nil {
		return err
	}
	switch env.Kind {
	case ingestv1.KindSBOMOSVRecord:
		var p ingestv1.SBOMOSVPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return sbomSt.UpsertFromOSVVuln(ctx, p.OSVID, p.CVE, p.Affected)
	case ingestv1.KindSBOMGHSADocument:
		var p ingestv1.SBOMGHSAPathPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		if p.Doc == nil {
			return fmt.Errorf("ghsa: empty doc")
		}
		return sbomSt.UpsertGHSA(ctx, p.Doc)
	case ingestv1.KindCoderulesCWERow:
		var p ingestv1.CoderulesCWEPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return crSt.UpsertCWECatalog(ctx, p.ID, p.Name, p.Description, p.Status)
	case ingestv1.KindCoderulesSemgrep:
		var p ingestv1.CoderulesSemgrepPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		md := fmt.Sprintf("# %s\n\n**path:** `%s`\n\n```yaml\n%s\n```\n", p.Title, p.Path, p.RawYAML)
		id := coderulesneo.StableID("semgrep", p.Path)
		if err := crSt.UpsertSemgrepRule(ctx, id, p.Path, p.Title, p.Language, md); err != nil {
			return err
		}
		for _, cw := range p.CWEs {
			if err := crSt.LinkSemgrepRuleToCWE(ctx, id, cw); err != nil {
				return err
			}
		}
		return nil
	case ingestv1.KindCoderulesCodeQL:
		var p ingestv1.CoderulesCodeQLPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		md := fmt.Sprintf("# %s\n\n**path:** `%s`\n\n```ql\n%s\n```\n", p.Name, p.Path, p.Body)
		id := coderulesneo.StableID("codeql", p.Path)
		if err := crSt.UpsertCodeQLRule(ctx, id, p.Path, p.Name, p.Lang, md); err != nil {
			return err
		}
		for _, cw := range p.CWEs {
			if err := crSt.LinkCodeQLRuleToCWE(ctx, id, cw); err != nil {
				return err
			}
		}
		return nil
	case ingestv1.KindNucleiTemplate:
		var p ingestv1.NucleiTemplatePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		id := nucleineo.StableID("nuclei", p.Path)
		md := fmt.Sprintf("# %s\n\n**id:** `%s`  \n**path:** `%s`\n\n```yaml\n%s\n```\n", p.Name, p.TemplateID, p.Path, p.RawYAML)
		return nuSt.UpsertNucleiTemplate(ctx, id, p.TemplateID, p.Path, p.Name, p.Severity, p.TagsJSON, p.CVE, p.CWE, md)
	default:
		// Ack and skip: forward-compatible with new kinds until worker is extended.
		log.Warn("unknown kind", slog.String("kind", env.Kind))
		return nil
	}
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getenvDuration(k string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
