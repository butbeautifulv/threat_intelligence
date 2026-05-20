package parse

import (
	"strings"
	"testing"
)

func TestSemgrepMeta(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		root     map[string]any
		fileName string
		wantID   string
		wantTitle string
	}{
		{
			name:     "root id and rule message",
			root:     map[string]any{"id": "root-rule"},
			fileName: "ignored.yml",
			wantID:   "root-rule",
			wantTitle: "root-rule",
		},
		{
			name: "rule id and message",
			root: map[string]any{
				"rules": []any{
					map[string]any{
						"id":      "semgrep-id",
						"message": "  SQL injection  ",
					},
				},
			},
			fileName:  "python/rule.yml",
			wantID:    "semgrep-id",
			wantTitle: "SQL injection",
		},
		{
			name: "fallback to filename",
			root: map[string]any{
				"rules": []any{
					map[string]any{"message": "   "},
				},
			},
			fileName:  "go/empty-msg.yml",
			wantID:    "go/empty-msg.yml",
			wantTitle: "go/empty-msg.yml",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id, title := SemgrepMeta(tt.root, tt.fileName)
			if id != tt.wantID || title != tt.wantTitle {
				t.Fatalf("id=%q title=%q want id=%q title=%q", id, title, tt.wantID, tt.wantTitle)
			}
		})
	}
}

func TestSemgrepCWES(t *testing.T) {
	t.Parallel()
	root := map[string]any{
		"rules": []any{
			map[string]any{
				"metadata": map[string]any{"cwe": "CWE-79 CWE-79"},
			},
			map[string]any{
				"metadata": map[string]any{"cwe": "CWE-79: XSS"},
			},
			map[string]any{
				"metadata": map[string]any{"cwe": []any{"cwe-89", "CWE-79", 99}},
			},
			map[string]any{"metadata": map[string]any{"cwe": 42}},
		},
	}
	got := SemgrepCWES(root)
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
	if got[0] != "CWE-79" || got[1] != "CWE-89" {
		t.Fatalf("order/content: %v", got)
	}
}

func TestSemgrepCWES_noRules(t *testing.T) {
	if got := SemgrepCWES(map[string]any{}); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestSemgrepCWES_metadataWithoutCWEKey(t *testing.T) {
	root := map[string]any{
		"rules": []any{
			map[string]any{"metadata": map[string]any{"severity": "high"}},
		},
	}
	if got := SemgrepCWES(root); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestSemgrepCWES_rulesNotSlice(t *testing.T) {
	if got := SemgrepCWES(map[string]any{"rules": "x"}); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestSemgrepCWES_dedupWithinMetadataString(t *testing.T) {
	root := map[string]any{
		"rules": []any{
			map[string]any{
				"metadata": map[string]any{"cwe": "CWE-80 CWE-80"},
			},
		},
	}
	got := SemgrepCWES(root)
	if len(got) != 1 || got[0] != "CWE-80" {
		t.Fatalf("got %v", got)
	}
}

func TestSemgrepCWES_nonMapMetadata(t *testing.T) {
	root := map[string]any{
		"rules": []any{
			map[string]any{"metadata": "not-a-map"},
		},
	}
	if got := SemgrepCWES(root); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestSemgrepCWES_cweStringOnly(t *testing.T) {
	root := map[string]any{
		"rules": []any{
			map[string]any{
				"metadata": map[string]any{"cwe": "cwe-611"},
			},
		},
	}
	got := SemgrepCWES(root)
	if len(got) != 1 || got[0] != "CWE-611" {
		t.Fatalf("got %v", got)
	}
}

func TestSemgrepCWES_ignoresNonStringCWEValues(t *testing.T) {
	root := map[string]any{
		"rules": []any{
			map[string]any{
				"metadata": map[string]any{"cwe": []any{42, true}},
			},
			map[string]any{
				"metadata": map[string]any{"cwe": map[string]any{"id": "CWE-1"}},
			},
		},
	}
	if got := SemgrepCWES(root); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestSemgrepCWES_nilMetadataMap(t *testing.T) {
	var nilMeta map[string]any
	root := map[string]any{
		"rules": []any{
			map[string]any{"metadata": nilMeta},
		},
	}
	if got := SemgrepCWES(root); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestSemgrepCWES_skipsInvalidRules(t *testing.T) {
	root := map[string]any{
		"rules": []any{
			"not-a-map",
			map[string]any{"metadata": nil},
			map[string]any{"id": "no-cwe-metadata"},
		},
	}
	if got := SemgrepCWES(root); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestSemgrepMeta_emptyRulesSlice(t *testing.T) {
	id, title := SemgrepMeta(map[string]any{"rules": []any{}}, "f.yml")
	if id != "f.yml" || title != "f.yml" {
		t.Fatalf("id=%q title=%q", id, title)
	}
}

func TestSemgrepMeta_nonMapRuleEntry(t *testing.T) {
	root := map[string]any{"rules": []any{map[string]any{"id": "r1", "message": "msg"}, "bad"}}
	id, title := SemgrepMeta(root, "file.yml")
	if id != "r1" || title != "msg" {
		t.Fatalf("id=%q title=%q", id, title)
	}
}

func TestSemgrepMeta_ruleIDWhenRootEmpty(t *testing.T) {
	root := map[string]any{
		"rules": []any{
			map[string]any{"id": "from-rule"},
		},
	}
	id, title := SemgrepMeta(root, "file.yml")
	if id != "from-rule" || title != "from-rule" {
		t.Fatalf("id=%q title=%q", id, title)
	}
}

func TestFirstNonEmpty_skipsBlanks(t *testing.T) {
	if got := firstNonEmpty("", "  ", "ok"); got != "ok" {
		t.Fatalf("got %q", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestCodeQLCWES(t *testing.T) {
	t.Parallel()
	body := strings.Repeat("line\n", 50) + "// CWE-502 and cwe-787 in query\n" + strings.Repeat("x\n", 200)
	got := CodeQLCWES(body)
	if len(got) != 2 || got[0] != "CWE-502" || got[1] != "CWE-787" {
		t.Fatalf("got %v", got)
	}
}

func TestCodeQLCWES_dedup(t *testing.T) {
	got := CodeQLCWES("CWE-22\nCWE-22\n")
	if len(got) != 1 || got[0] != "CWE-22" {
		t.Fatalf("got %v", got)
	}
}

func TestCodeQLCWES_empty(t *testing.T) {
	if got := CodeQLCWES(""); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}
