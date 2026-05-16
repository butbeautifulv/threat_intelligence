package workflow

import (
	"path/filepath"
	"testing"
)

func TestLoadPlaybooks(t *testing.T) {
	path := filepath.Join("..", "..", "..", "playbooks", "bugbounty.yaml")
	list, err := LoadPlaybooks(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 5 {
		t.Fatalf("expected playbooks, got %d", len(list))
	}
	if _, ok := FindPlaybook(list, "reconnaissance"); !ok {
		t.Fatal("missing reconnaissance playbook")
	}
}
