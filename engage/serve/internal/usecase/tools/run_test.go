package tools

import (
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

func TestMergeParameters_targetAliasesURL(t *testing.T) {
	spec := tool.Spec{
		Name: "gobuster_scan",
		Parameters: []tool.Param{
			{Name: "target", Required: true},
			{Name: "url", Required: true},
			{Name: "mode", Default: "dir"},
		},
	}
	out := mergeParameters(spec, contract.ToolRunRequest{
		Target: "http://example.com",
		Parameters: map[string]string{
			"mode": "dir",
		},
	})
	if out["target"] != "http://example.com" {
		t.Fatalf("target = %q", out["target"])
	}
	if out["url"] != "http://example.com" {
		t.Fatalf("url alias = %q", out["url"])
	}
	if out["mode"] != "dir" {
		t.Fatalf("mode = %q", out["mode"])
	}
}

func TestMergeParameters_targetAliasesDomain(t *testing.T) {
	spec := tool.Spec{
		Name: "amass_scan",
		Parameters: []tool.Param{
			{Name: "target", Required: true},
			{Name: "domain", Required: true},
		},
	}
	out := mergeParameters(spec, contract.ToolRunRequest{Target: "example.com"})
	if out["domain"] != "example.com" {
		t.Fatalf("domain alias = %q", out["domain"])
	}
}

func TestMergeParameters_explicitURLNotOverwritten(t *testing.T) {
	spec := tool.Spec{
		Parameters: []tool.Param{
			{Name: "target", Required: true},
			{Name: "url", Required: true},
		},
	}
	out := mergeParameters(spec, contract.ToolRunRequest{
		Target: "http://fallback.com",
		Parameters: map[string]string{
			"url": "http://explicit.com",
		},
	})
	if out["url"] != "http://explicit.com" {
		t.Fatalf("url = %q, want explicit", out["url"])
	}
}
