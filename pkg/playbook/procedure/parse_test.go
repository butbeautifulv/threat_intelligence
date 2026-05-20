package procedure

import (
	"os"
	"path/filepath"
	"testing"

	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
)

func TestParseSkillMarkdown_diskImaging(t *testing.T) {
	root, err := pbindex.RepoRoot()
	if err != nil {
		t.Skip(err)
	}
	path := filepath.Join(root, "corpus/anthropic-cybersecurity-skills/skills/acquiring-disk-image-with-dd-and-dcfldd/SKILL.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skip(err)
	}
	spec := ParseSkillMarkdown("acquiring-disk-image-with-dd-and-dcfldd", "digital-forensics", nil, nil, string(raw))
	if len(spec.WhenToUse) < 2 {
		t.Fatalf("when_to_use: %d", len(spec.WhenToUse))
	}
	if len(spec.Steps) < 2 {
		t.Fatalf("steps: %d", len(spec.Steps))
	}
}

func TestCatalog_open(t *testing.T) {
	cat, err := Open("")
	if err != nil {
		t.Skip(err)
	}
	if len(cat.file.Procedures) < 700 {
		t.Fatalf("procedures: %d", len(cat.file.Procedures))
	}
	spec, err := cat.GetSpec("acquiring-disk-image-with-dd-and-dcfldd")
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Steps) == 0 {
		t.Fatal("expected steps in spec")
	}
}
