package parse

import (
	"strings"
	"testing"
)

func TestParseNucleiYAML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		raw     string
		want    NucleiTemplate
		wantErr bool
	}{
		{
			name: "full metadata",
			raw: `id: CVE-2024-9999
info:
  name: Test Template
  severity: critical
  tags: xss, sqli
  classification:
    cve-id: cve-2024-9999
    cwe-id: CWE-79
`,
			want: NucleiTemplate{
				TemplateID: "CVE-2024-9999",
				Name:       "Test Template",
				Severity:   "critical",
				TagsJSON:   `["xss","sqli"]`,
				CVE:        "CVE-2024-9999",
				CWE:        "CWE-79",
			},
		},
		{
			name: "cve from id prefix",
			raw: `id: CVE-2023-1234
info:
  name: CVE rule
`,
			want: NucleiTemplate{
				TemplateID: "CVE-2023-1234",
				Name:       "CVE rule",
				TagsJSON:   "[]",
				CVE:        "CVE-2023-1234",
			},
		},
		{
			name: "tags array",
			raw: `id: http-test
info:
  tags:
    - auth
    - exposure
`,
			want: NucleiTemplate{
				TemplateID: "http-test",
				TagsJSON:   `["auth","exposure"]`,
			},
		},
		{
			name: "minimal id only",
			raw:  `id: bare-template`,
			want: NucleiTemplate{
				TemplateID: "bare-template",
				TagsJSON:   "[]",
			},
		},
		{
			name:    "invalid yaml",
			raw:     "id: [\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseNucleiYAML([]byte(tt.raw))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.TemplateID != tt.want.TemplateID || got.Name != tt.want.Name ||
				got.Severity != tt.want.Severity || got.TagsJSON != tt.want.TagsJSON ||
				got.CVE != tt.want.CVE || got.CWE != tt.want.CWE {
				t.Fatalf("got %+v want %+v", got, tt.want)
			}
		})
	}
}

func TestNormalizeTags_stringAndArray(t *testing.T) {
	t.Parallel()
	if got := normalizeTags(" a , ,b "); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("string tags: %v", got)
	}
	if got := normalizeTags([]any{" x ", 1, "y"}); len(got) != 2 {
		t.Fatalf("array tags: %v", got)
	}
	if got := normalizeTags(nil); got != nil {
		t.Fatalf("default: %v", got)
	}
}

func TestParseNucleiYAML_tagsNullBecomesEmptyArray(t *testing.T) {
	raw := "id: t\ninfo:\n  tags: null\n"
	got, err := ParseNucleiYAML([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if got.TagsJSON != "[]" {
		t.Fatalf("tags %q", got.TagsJSON)
	}
	if strings.TrimSpace(got.TemplateID) != "t" {
		t.Fatalf("id %q", got.TemplateID)
	}
}
