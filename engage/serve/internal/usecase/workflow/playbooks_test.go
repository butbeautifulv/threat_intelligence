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

func TestLoadAllPlaybooks_includesCTF(t *testing.T) {
	catalog := filepath.Join("..", "..", "..", "catalog", "tools.live.yaml")
	list, err := LoadAllPlaybooks(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := FindPlaybook(list, "ctf-web"); !ok {
		t.Fatal("missing ctf-web playbook")
	}
	if _, ok := FindPlaybook(list, "ctf-pwn"); !ok {
		t.Fatal("missing ctf-pwn playbook")
	}
}
