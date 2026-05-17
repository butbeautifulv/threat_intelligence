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

func TestLoadCatalog_productionMergeOrder(t *testing.T) {
	root := filepath.Join("..", "..", "catalog")
	catalog := filepath.Join(root, "tools.yaml")
	live := filepath.Join(root, "tools.live.yaml")
	enabled := filepath.Join(root, "tools.enabled.yaml")

	specs, err := LoadCatalog(catalog, live, enabled)
	if err != nil {
		t.Fatal(err)
	}
	reg := NewRegistry(specs)
	enabledCount := len(reg.List())
	if enabledCount < 103 {
		t.Fatalf("expected >=103 subprocess-enabled tools after live overlay, got %d", enabledCount)
	}
	s, ok := reg.Get("nmap_scan")
	if !ok || !s.Enabled {
		t.Fatal("nmap_scan should be enabled via tools.live.yaml overlay")
	}
	if len(reg.ListAll()) < 158 {
		t.Fatalf("expected full catalog >=158 names, got %d", len(reg.ListAll()))
	}
}
