package tools

import (
	"path/filepath"
	"testing"
)

func TestLoadCatalog_merge(t *testing.T) {
	root := filepath.Join("..", "..", "catalog")
	specs, err := LoadCatalog(
		filepath.Join(root, "tools.live.yaml"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) < 5 {
		t.Fatalf("expected >=5 live tools, got %d", len(specs))
	}
	reg := NewRegistry(specs)
	if _, ok := reg.Get("nmap_scan"); !ok {
		t.Fatal("missing nmap_scan")
	}
}
