package workflow

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
)

func TestRunPlaybook_reconnaissance(t *testing.T) {
	path := filepath.Join("..", "..", "..", "playbooks", "bugbounty.yaml")
	list, err := LoadPlaybooks(path)
	if err != nil {
		t.Fatal(err)
	}
	pb, ok := FindPlaybook(list, "reconnaissance")
	if !ok {
		t.Fatal("playbook not found")
	}
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "httpx", Binary: "httpx", Enabled: true},
	})
	intel := &intelligence.Service{Registry: reg, Engine: intelligence.DefaultDecisionEngine()}
	svc := &Service{Intel: intel}
	out := svc.RunPlaybook(context.Background(), "", pb, "https://example.com", false)
	if out["playbook"] != "reconnaissance" {
		t.Fatalf("unexpected playbook %v", out["playbook"])
	}
}
