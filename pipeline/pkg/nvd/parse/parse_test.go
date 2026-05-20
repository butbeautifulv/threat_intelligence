package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/butbeautifulv/veil/pipeline/pkg/nvd/parse"
)

func TestParsePage_CWEAndCPE(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "nvd_page_min.json"))
	if err != nil {
		t.Fatal(err)
	}
	vulns, total, err := parse.ParsePage(raw)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Fatalf("totalResults: got %d want 2", total)
	}
	if len(vulns) != 2 {
		t.Fatalf("vulns: got %d want 2", len(vulns))
	}
	v0 := vulns[0]
	if v0.CVE != "CVE-2024-0001" {
		t.Fatalf("cve: %q", v0.CVE)
	}
	if len(v0.CWE) != 1 || v0.CWE[0] != "CWE-79" {
		t.Fatalf("cwe: %#v", v0.CWE)
	}
	if len(v0.CPEs) != 1 || v0.CPEs[0].URI == "" {
		t.Fatalf("cpes: %#v", v0.CPEs)
	}
	if v0.CVSS == nil || v0.CVSS.Base != 9.8 {
		t.Fatalf("cvss: %#v", v0.CVSS)
	}
	v1 := vulns[1]
	if len(v1.CWE) != 1 || v1.CWE[0] != "CWE-89" {
		t.Fatalf("cwe v1: %#v", v1.CWE)
	}
	if len(v1.CPEs) != 1 {
		t.Fatalf("cpes v1: %#v", v1.CPEs)
	}
}

func TestParsePage_table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		raw       string
		wantErr   bool
		wantTotal int
		wantLen   int
		check     func(t *testing.T, vulns []parse.Vulnerability)
	}{
		{
			name:    "malformed JSON",
			raw:     `{not valid json`,
			wantErr: true,
		},
		{
			name:      "empty CVE list",
			raw:       `{"totalResults": 0, "vulnerabilities": []}`,
			wantTotal: 0,
			wantLen:   0,
		},
		{
			name: "skip empty CVE id",
			raw: `{
				"totalResults": 2,
				"vulnerabilities": [
					{"cve": {"id": ""}},
					{"cve": {"id": "CVE-2024-9999", "descriptions": [{"lang": "en", "value": "only id and desc"}]}}
				]
			}`,
			wantTotal: 2,
			wantLen:   1,
			check: func(t *testing.T, vulns []parse.Vulnerability) {
				if vulns[0].CVE != "CVE-2024-9999" {
					t.Fatalf("cve: %q", vulns[0].CVE)
				}
				if vulns[0].Summary != "only id and desc" {
					t.Fatalf("summary: %q", vulns[0].Summary)
				}
			},
		},
		{
			name: "partial fields id only",
			raw: `{
				"totalResults": 1,
				"vulnerabilities": [
					{"cve": {"id": "CVE-2024-0003"}}
				]
			}`,
			wantTotal: 1,
			wantLen:   1,
			check: func(t *testing.T, vulns []parse.Vulnerability) {
				v := vulns[0]
				if v.CVE != "CVE-2024-0003" || v.ID != "CVE-2024-0003" {
					t.Fatalf("id/cve: %#v", v)
				}
				if v.Summary != "" {
					t.Fatalf("summary: %q want empty", v.Summary)
				}
				if len(v.CWE) != 0 {
					t.Fatalf("cwe: %#v want empty", v.CWE)
				}
				if len(v.CPEs) != 0 {
					t.Fatalf("cpes: %#v want empty", v.CPEs)
				}
				if v.CVSS != nil {
					t.Fatalf("cvss: %#v want nil", v.CVSS)
				}
			},
		},
		{
			name: "partial fields summary without cwe cpe cvss",
			raw: `{
				"totalResults": 1,
				"vulnerabilities": [
					{
						"cve": {
							"id": "CVE-2024-0004",
							"descriptions": [{"lang": "fr", "value": "French only"}]
						}
					}
				]
			}`,
			wantTotal: 1,
			wantLen:   1,
			check: func(t *testing.T, vulns []parse.Vulnerability) {
				v := vulns[0]
				if v.Summary != "French only" {
					t.Fatalf("summary: %q", v.Summary)
				}
				if len(v.CWE) != 0 || len(v.CPEs) != 0 || v.CVSS != nil {
					t.Fatalf("expected only summary, got %#v", v)
				}
			},
		},
		{
			name: "missing totalResults defaults to zero",
			raw: `{
				"vulnerabilities": [
					{"cve": {"id": "CVE-2024-0005", "descriptions": [{"lang": "en", "value": "no total field"}]}}
				]
			}`,
			wantTotal: 0,
			wantLen:   1,
		},
		{
			name: "skip non-map vulnerability items",
			raw: `{
				"totalResults": 1,
				"vulnerabilities": [
					"not-a-map",
					{"cve": {"id": "CVE-2024-0006"}}
				]
			}`,
			wantTotal: 1,
			wantLen:   1,
			check: func(t *testing.T, vulns []parse.Vulnerability) {
				if vulns[0].CVE != "CVE-2024-0006" {
					t.Fatalf("cve: %q", vulns[0].CVE)
				}
			},
		},
		{
			name: "cvss v30 fallback when v31 invalid",
			raw: `{
				"totalResults": 1,
				"vulnerabilities": [{
					"cve": {
						"id": "CVE-CVSS-V30",
						"metrics": {
							"cvssMetricV31": ["not-a-map"],
							"cvssMetricV30": [{
								"cvssData": {
									"version": "3.0",
									"vectorString": "CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N",
									"baseScore": 7.5
								}
							}]
						}
					}
				}]
			}`,
			wantTotal: 1,
			wantLen:   1,
			check: func(t *testing.T, vulns []parse.Vulnerability) {
				if vulns[0].CVSS == nil || vulns[0].CVSS.Version != "3.0" || vulns[0].CVSS.Base != 7.5 {
					t.Fatalf("cvss v30: %#v", vulns[0].CVSS)
				}
			},
		},
		{
			name: "english description fallbacks and skips",
			raw: `{
				"totalResults": 3,
				"vulnerabilities": [
					{
						"cve": {
							"id": "CVE-DESC-SKIP",
							"descriptions": [
								"not-a-map",
								{"lang": "en", "value": ""},
								{"lang": "de", "value": "ignored not first"}
							]
						}
					},
					{
						"cve": {
							"id": "CVE-DESC-FALLBACK",
							"descriptions": [
								{"lang": "fr", "value": "French fallback"},
								{"lang": "en", "value": ""}
							]
						}
					},
					{
						"cve": {
							"id": "CVE-DESC-EMPTY",
							"descriptions": []
						}
					},
					{
						"cve": {
							"id": "CVE-DESC-BAD-FIRST",
							"descriptions": ["not-a-map"]
						}
					}
				]
			}`,
			wantTotal: 3,
			wantLen:   4,
			check: func(t *testing.T, vulns []parse.Vulnerability) {
				if vulns[0].Summary != "" {
					t.Fatalf("non-map first desc summary: %q want empty", vulns[0].Summary)
				}
				if vulns[1].Summary != "French fallback" {
					t.Fatalf("fallback summary: %q", vulns[1].Summary)
				}
				if vulns[2].Summary != "" {
					t.Fatalf("empty desc summary: %q", vulns[2].Summary)
				}
				if vulns[3].Summary != "" {
					t.Fatalf("bad first desc summary: %q", vulns[3].Summary)
				}
			},
		},
		{
			name: "weaknesses and configurations skip invalid entries",
			raw: `{
				"totalResults": 1,
				"vulnerabilities": [{
					"cve": {
						"id": "CVE-PARSE-SKIPS",
						"weaknesses": [
							"not-a-map",
							{"description": "not-an-array"},
							{"description": ["not-a-map", {"value": ""}, {"value": "CWE-123"}]}
						],
						"configurations": [
							"not-a-map",
							{"nodes": "not-an-array"},
							{"nodes": [
								"not-a-map",
								{"cpeMatch": "not-an-array"},
								{"cpeMatch": ["not-a-map", {"criteria": "", "cpe23Uri": "cpe:2.3:a:via:cpe23:1.0:*:*:*:*:*:*:*"}]}
							]}
						],
						"metrics": {
							"cvssMetricV31": [],
							"cvssMetricV30": ["not-a-map"]
						}
					}
				}]
			}`,
			wantTotal: 1,
			wantLen:   1,
			check: func(t *testing.T, vulns []parse.Vulnerability) {
				v := vulns[0]
				if len(v.CWE) != 1 || v.CWE[0] != "CWE-123" {
					t.Fatalf("cwe: %#v", v.CWE)
				}
				if len(v.CPEs) != 1 || v.CPEs[0].URI != "cpe:2.3:a:via:cpe23:1.0:*:*:*:*:*:*:*" {
					t.Fatalf("cpes: %#v", v.CPEs)
				}
				if v.CVSS != nil {
					t.Fatalf("cvss want nil, got %#v", v.CVSS)
				}
			},
		},
		{
			name: "pickCVSS nil cvssData and empty fields",
			raw: `{
				"totalResults": 2,
				"vulnerabilities": [
					{
						"cve": {
							"id": "CVE-CVSS-NIL-DATA",
							"metrics": {
								"cvssMetricV31": [{"cvssData": null}]
							}
						}
					},
					{
						"cve": {
							"id": "CVE-CVSS-EMPTY-DATA",
							"metrics": {
								"cvssMetricV31": [{
									"cvssData": {"version": "", "vectorString": "", "baseScore": 0}
								}]
							}
						}
					}
				]
			}`,
			wantTotal: 2,
			wantLen:   2,
			check: func(t *testing.T, vulns []parse.Vulnerability) {
				for _, v := range vulns {
					if v.CVSS != nil {
						t.Fatalf("%s cvss want nil, got %#v", v.CVE, v.CVSS)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			vulns, total, err := parse.ParsePage([]byte(tt.raw))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if total != tt.wantTotal {
				t.Fatalf("totalResults: got %d want %d", total, tt.wantTotal)
			}
			if len(vulns) != tt.wantLen {
				t.Fatalf("vulns: got %d want %d", len(vulns), tt.wantLen)
			}
			if tt.check != nil {
				tt.check(t, vulns)
			}
		})
	}
}
