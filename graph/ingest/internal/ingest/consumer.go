package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/butbeautifulv/veil/graph/ingest/internal/components"
	coderulesneo "github.com/butbeautifulv/veil/graph/ingest/internal/appsec/coderules"
	nucleineo "github.com/butbeautifulv/veil/graph/ingest/internal/appsec/nuclei"
	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/nats-io/nats.go"
)

// RunPullLoop consumes JetStream messages until ctx is canceled.
func RunPullLoop(ctx context.Context, log *slog.Logger, sub *nats.Subscription, batch int, maxWait time.Duration, rt *components.Runtime) error {
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
			if err := handleMsg(ctx, log, m, rt); err != nil {
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

func handleMsg(ctx context.Context, log *slog.Logger, m *nats.Msg, rt *components.Runtime) error {
	var env commit.Envelope
	if err := json.Unmarshal(m.Data, &env); err != nil {
		return fmt.Errorf("decode envelope: %w", err)
	}
	if err := env.Validate(); err != nil {
		return err
	}
	if err := validateEnvelopeSource(&env); err != nil {
		return err
	}

	switch env.Source {
	case commit.SourceTI:
		return rt.Apply.TI(ctx, &env)
	case commit.SourceVuln:
		return rt.Apply.Vuln(ctx, &env)
	case commit.SourceLola:
		return rt.Apply.Lola(ctx, &env)
	case commit.SourceDS:
		return rt.Apply.DS(ctx, &env)
	}

	switch env.Kind {
	case commit.KindSBOMOSVRecord:
		var p commit.SBOMOSVPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return rt.SBOM.UpsertFromOSVVuln(ctx, p.OSVID, p.CVE, p.Affected)
	case commit.KindSBOMGHSADocument:
		var p commit.SBOMGHSAPathPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		if p.Doc == nil {
			return fmt.Errorf("ghsa: empty doc")
		}
		return rt.SBOM.UpsertGHSA(ctx, p.Doc)
	case commit.KindCoderulesCWERow:
		var p commit.CoderulesCWEPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return rt.CR.UpsertCWECatalog(ctx, p.ID, p.Name, p.Description, p.Status)
	case commit.KindCoderulesSemgrep:
		var p commit.CoderulesSemgrepPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		md := fmt.Sprintf("# %s\n\n**path:** `%s`\n\n```yaml\n%s\n```\n", p.Title, p.Path, p.RawYAML)
		id := coderulesneo.StableID("semgrep", p.Path)
		if err := rt.CR.UpsertSemgrepRule(ctx, id, p.Path, p.Title, p.Language, md); err != nil {
			return err
		}
		for _, cw := range p.CWEs {
			if err := rt.CR.LinkSemgrepRuleToCWE(ctx, id, cw); err != nil {
				return err
			}
		}
		return nil
	case commit.KindCoderulesCodeQL:
		var p commit.CoderulesCodeQLPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		md := fmt.Sprintf("# %s\n\n**path:** `%s`\n\n```ql\n%s\n```\n", p.Name, p.Path, p.Body)
		id := coderulesneo.StableID("codeql", p.Path)
		if err := rt.CR.UpsertCodeQLRule(ctx, id, p.Path, p.Name, p.Lang, md); err != nil {
			return err
		}
		for _, cw := range p.CWEs {
			if err := rt.CR.LinkCodeQLRuleToCWE(ctx, id, cw); err != nil {
				return err
			}
		}
		return nil
	case commit.KindNucleiTemplate:
		var p commit.NucleiTemplatePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		id := nucleineo.StableID("nuclei", p.Path)
		md := fmt.Sprintf("# %s\n\n**id:** `%s`  \n**path:** `%s`\n\n```yaml\n%s\n```\n", p.Name, p.TemplateID, p.Path, p.RawYAML)
		return rt.Nuclei.UpsertNucleiTemplate(ctx, id, p.TemplateID, p.Path, p.Name, p.Severity, p.TagsJSON, p.CVE, p.CWE, md)
	default:
		log.Warn("unknown kind", slog.String("kind", env.Kind), slog.String("source", env.Source))
		return nil
	}
}

func validateEnvelopeSource(e *commit.Envelope) error {
	switch e.Kind {
	case commit.KindSBOMOSVRecord, commit.KindSBOMGHSADocument:
		if e.Source != commit.SourceSBOM {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceSBOM, e.Source)
		}
	case commit.KindCoderulesCWERow, commit.KindCoderulesSemgrep, commit.KindCoderulesCodeQL:
		if e.Source != commit.SourceCoderules {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceCoderules, e.Source)
		}
	case commit.KindNucleiTemplate:
		if e.Source != commit.SourceNuclei {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceNuclei, e.Source)
		}
	case commit.KindTIIoC, commit.KindTIKEVVulnerability, commit.KindTIReport, commit.KindTICampaign, commit.KindTICluster, commit.KindTIActor,
		commit.KindTILinkCampaignIOC, commit.KindTILinkClusterCampaign, commit.KindTILinkCampaignActor, commit.KindTILinkReportMentionsIOC, commit.KindTIJSONLRecord:
		if e.Source != commit.SourceTI {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceTI, e.Source)
		}
	case commit.KindVulnUpsert, commit.KindVulnMergeExploit:
		if e.Source != commit.SourceVuln {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceVuln, e.Source)
		}
	case commit.KindLolaArtifact, commit.KindLolaLofts, commit.KindLolaAttackTechnique, commit.KindLolaAttackTactic,
		commit.KindLolaMergeTacticTechnique, commit.KindLolaMergeSubtechnique, commit.KindLolaLinkArtifacts:
		if e.Source != commit.SourceLola {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceLola, e.Source)
		}
	case commit.KindDSUpsertSigma, commit.KindDSUpsertYara, commit.KindDSUpsertAtomic, commit.KindDSUpsertCaldera:
		if e.Source != commit.SourceDS {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceDS, e.Source)
		}
	}
	return nil
}
