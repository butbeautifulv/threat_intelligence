package transform_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/harvest"
	tidomain "github.com/butbeautifulv/veil/pkg/ti/domain"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/transform"
)

func TestScrapeToIngest_routes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	nvdRaw, err := os.ReadFile(filepath.Join("..", "..", "..", "pkg", "nvd", "parse", "testdata", "nvd_page_min.json"))
	if err != nil {
		t.Fatal(err)
	}
	vulnPayload, err := json.Marshal(harvest.VulnNVDPage{RawJSON: string(nvdRaw)})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		source string
		kind   string
		payload any
	}{
		{
			name:   harvest.SourceDS,
			source: harvest.SourceDS,
			kind:   harvest.KindDSSigmaRaw,
			payload: harvest.DSSigmaRaw{
				Path: "rules/test.yml",
				RawYAML: `id: router-sigma-1
title: Router Sigma
level: medium
`,
			},
		},
		{
			name:   harvest.SourceTI,
			source: harvest.SourceTI,
			kind:   harvest.KindTIIoCRaw,
			payload: tidomain.IOC{
				Type:   tidomain.IOCURL,
				Value:  "https://example.com/ioc",
				Source: "test",
			},
		},
		{
			name:   harvest.SourceVuln,
			source: harvest.SourceVuln,
			kind:   harvest.KindVulnNVDPage,
			payload: json.RawMessage(vulnPayload),
		},
		{
			name:   harvest.SourceLola,
			source: harvest.SourceLola,
			kind:   harvest.KindLolaArtifactRaw,
			payload: harvest.LolaArtifactRaw{
				Source:  "misp",
				Path:    "obj/1.json",
				RawBody: `{"name":"router-artifact","type":"malware"}`,
			},
		},
		{
			name:   harvest.SourceSBOM,
			source: harvest.SourceSBOM,
			kind:   harvest.KindSBOMOSVJSON,
			payload: harvest.SBOMOSVRaw{
				OSVID:   "OSV-router",
				CVE:     "CVE-2024-2000",
				RawJSON: `{"id":"OSV-router","affected":[{"package":{"ecosystem":"npm","name":"pkg"}}]}`,
			},
		},
		{
			name:   harvest.SourceCoderules,
			source: harvest.SourceCoderules,
			kind:   harvest.KindCoderulesCWERaw,
			payload: harvest.CoderulesCWERaw{
				ID: "CWE-79", Name: "XSS", Description: "d", Status: "Stable",
			},
		},
		{
			name:   harvest.SourceNuclei,
			source: harvest.SourceNuclei,
			kind:   harvest.KindNucleiTemplateRaw,
			payload: harvest.NucleiTemplateRaw{
				Path:    "/tmp/router.yaml",
				RawYAML: "id: router-nuclei\ninfo:\n  name: Router\n  severity: info\n",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var payload []byte
			switch p := tt.payload.(type) {
			case json.RawMessage:
				payload = p
			default:
				payload, err = json.Marshal(p)
				if err != nil {
					t.Fatal(err)
				}
			}
			env := &harvest.Envelope{
				SchemaVersion: harvest.CurrentSchemaVersion,
				Source:        tt.source,
				Kind:          tt.kind,
				Payload:       payload,
			}
			out, err := transform.ScrapeToIngest(ctx, env)
			if err != nil {
				t.Fatalf("ScrapeToIngest(%s): %v", tt.source, err)
			}
			if len(out) == 0 {
				t.Fatalf("ScrapeToIngest(%s): expected envelopes", tt.source)
			}
		})
	}
}

func TestScrapeToIngest_unknownSource(t *testing.T) {
	t.Parallel()
	env := &harvest.Envelope{
		SchemaVersion: harvest.CurrentSchemaVersion,
		Source:        "unknown-source",
		Kind:          "scrape_unknown",
		Payload:       json.RawMessage(`{}`),
	}
	_, err := transform.ScrapeToIngest(context.Background(), env)
	if err == nil {
		t.Fatal("expected error for unknown source")
	}
	if !strings.Contains(err.Error(), `unknown source "unknown-source"`) {
		t.Fatalf("error: %v", err)
	}
}
