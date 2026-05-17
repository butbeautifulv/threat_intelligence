package tooldispatch

import (
	"context"
	"path/filepath"
	"testing"
)

func TestProbeTool_bridgeWorkflow(t *testing.T) {
	root := filepath.Join("..", "..", "..", "catalog")
	catalog := filepath.Join(root, "tools.yaml")
	d, err := NewMatrixDispatcher(catalog)
	if err != nil {
		t.Fatal(err)
	}
	r := ProbeTool(context.Background(), d, "get_telemetry")
	if !r.Pass {
		t.Fatalf("get_telemetry: %s", r.Error)
	}
	if r.Kind != "bridge" {
		t.Fatalf("kind = %q, want bridge", r.Kind)
	}
}

func TestRunMatrix_catalogCount(t *testing.T) {
	root := filepath.Join("..", "..", "..", "catalog")
	catalog := filepath.Join(root, "tools.yaml")
	d, err := NewMatrixDispatcher(catalog)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, 160)
	for _, s := range d.Runner.Registry.ListAll() {
		names = append(names, s.Name)
	}
	if len(names) < 158 {
		t.Fatalf("catalog registry has %d tools, want >=158", len(names))
	}
}
