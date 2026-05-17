package tools

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

func TestResolveCatalogName(t *testing.T) {
	reg := NewRegistry([]tool.Spec{
		{Name: "nmap_scan", Category: toolid.CategoryNetwork, Enabled: true},
	})
	if got := ResolveCatalogName("nmap", reg); got != "nmap_scan" {
		t.Fatalf("got %q", got)
	}
	if got := ResolveCatalogName("nmap_scan", reg); got != "nmap_scan" {
		t.Fatalf("got %q", got)
	}
}
