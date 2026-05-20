package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/config"
	"github.com/nats-io/nats.go"
)

type recordIngestPublisher struct {
	subject string
	envs    []*commit.Envelope
}

func (r *recordIngestPublisher) PublishJSON(_ context.Context, subject string, env *commit.Envelope) error {
	r.subject = subject
	r.envs = append(r.envs, env)
	return nil
}

func TestHandleScrapeMsg_invalidJSON(t *testing.T) {
	err := handleScrapeMsg(context.Background(), slog.Default(), &nats.Msg{Data: []byte("{")}, nil, config.Config{})
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestHandleScrapeMsg_invalidEnvelope(t *testing.T) {
	b, _ := json.Marshal(map[string]any{"source": "ti"})
	err := handleScrapeMsg(context.Background(), slog.Default(), &nats.Msg{Data: b}, nil, config.Config{})
	if err == nil {
		t.Fatal("expected validate error")
	}
}

func TestHandleScrapeMsg_matrix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		env        harvest.Envelope
		cfg        config.Config
		wantErr    bool
		wantSubj   string
		wantSource string
		wantKind   string
	}{
		{
			name: "ti ioc raw",
			env: harvest.Envelope{
				SchemaVersion: harvest.CurrentSchemaVersion,
				Source:        harvest.SourceTI,
				Kind:          harvest.KindTIIoCRaw,
				ContentKey:    "ti:ioc:url:https://example.com",
				ScrapedAt:     "2026-05-15T12:00:00Z",
				Payload:       json.RawMessage(`{"type":"url","value":"https://example.com","source":"x"}`),
			},
			wantSubj:   "ingest.events",
			wantSource: commit.SourceTI,
			wantKind:   commit.KindTIIoC,
		},
		{
			name: "nuclei template",
			env: harvest.Envelope{
				SchemaVersion: harvest.CurrentSchemaVersion,
				Source:        harvest.SourceNuclei,
				Kind:          harvest.KindNucleiTemplateRaw,
				ContentKey:    "nuclei:http-missing",
				ScrapedAt:     "2026-05-15T12:00:00Z",
				Payload: mustJSON(t, harvest.NucleiTemplateRaw{
					Path:    "http/http-missing.yaml",
					RawYAML: "id: http-missing\ninfo:\n  name: Missing\n  severity: info\n",
				}),
			},
			cfg: config.Config{
				IngestPublish: "ingest.events",
				DomainSubjects: map[string]string{
					harvest.SourceNuclei: "ingest.nuclei.events",
				},
			},
			wantSubj:   "ingest.nuclei.events",
			wantSource: commit.SourceNuclei,
			wantKind:   commit.KindNucleiTemplate,
		},
		{
			name: "coderules semgrep",
			env: harvest.Envelope{
				SchemaVersion: harvest.CurrentSchemaVersion,
				Source:        harvest.SourceCoderules,
				Kind:          harvest.KindCoderulesSemgrepRaw,
				ContentKey:    "coderules:python/rule.yml",
				ScrapedAt:     "2026-05-15T12:00:00Z",
				Payload: mustJSON(t, harvest.CoderulesSemgrepRaw{
					Path:    "python/rule.yml",
					RawYAML: "rules:\n  - id: py-rule\n    message: test\n    metadata:\n      cwe: CWE-89\n",
				}),
			},
			wantSubj:   "ingest.events",
			wantSource: commit.SourceCoderules,
			wantKind:   commit.KindCoderulesSemgrep,
		},
		{
			name: "sbom osv",
			env: harvest.Envelope{
				SchemaVersion: harvest.CurrentSchemaVersion,
				Source:        harvest.SourceSBOM,
				Kind:          harvest.KindSBOMOSVJSON,
				ContentKey:    "sbom:osv:CVE-2024-1",
				ScrapedAt:     "2026-05-15T12:00:00Z",
				Payload: mustJSON(t, harvest.SBOMOSVRaw{
					OSVID:   "OSV-1",
					CVE:     "CVE-2024-1",
					RawJSON: `{"id":"OSV-1"}`,
				}),
			},
			wantSubj:   "ingest.events",
			wantSource: commit.SourceSBOM,
			wantKind:   commit.KindSBOMOSVRecord,
		},
		{
			name: "ds sigma",
			env: harvest.Envelope{
				SchemaVersion: harvest.CurrentSchemaVersion,
				Source:        harvest.SourceDS,
				Kind:          harvest.KindDSSigmaRaw,
				ContentKey:    "ds:sigma:rules/x.yml",
				ScrapedAt:     "2026-05-15T12:00:00Z",
				Payload: mustJSON(t, harvest.DSSigmaRaw{
					Path:    "rules/x.yml",
					RawYAML: "id: sigma-1\ntitle: Sigma\n",
				}),
			},
			wantSubj:   "ingest.events",
			wantSource: commit.SourceDS,
			wantKind:   commit.KindDSUpsertSigma,
		},
		{
			name: "unknown source",
			env: harvest.Envelope{
				SchemaVersion: harvest.CurrentSchemaVersion,
				Source:        harvest.SourceBrowser,
				Kind:          harvest.KindBrowserInspectRaw,
				ContentKey:    "browser:https://example.com",
				ScrapedAt:     "2026-05-15T12:00:00Z",
				Payload:       json.RawMessage(`{"url":"https://example.com"}`),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b, err := json.Marshal(tt.env)
			if err != nil {
				t.Fatal(err)
			}
			rec := &recordIngestPublisher{}
			cfg := tt.cfg
			if cfg.IngestPublish == "" && len(cfg.DomainSubjects) == 0 {
				cfg.IngestPublish = "ingest.events"
			}
			err = handleScrapeMsg(context.Background(), slog.Default(), &nats.Msg{Data: b}, rec, cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if rec.subject != tt.wantSubj {
				t.Fatalf("subject %q want %q", rec.subject, tt.wantSubj)
			}
			if len(rec.envs) == 0 {
				t.Fatal("expected published envelopes")
			}
			if rec.envs[0].Source != tt.wantSource {
				t.Fatalf("source %s want %s", rec.envs[0].Source, tt.wantSource)
			}
			if rec.envs[0].Kind != tt.wantKind {
				t.Fatalf("kind %s want %s", rec.envs[0].Kind, tt.wantKind)
			}
		})
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
