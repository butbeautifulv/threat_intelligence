package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearch_diskImaging(t *testing.T) {
	root, err := RepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	idx := filepath.Join(root, DefaultIndexRel)
	if _, err := os.Stat(idx); err != nil {
		t.Skip("cyber-skills.json missing; run make skills-index")
	}
	cat, err := Open(idx)
	if err != nil {
		t.Fatal(err)
	}
	hits := cat.Search("disk imaging", "", 5)
	if len(hits) == 0 {
		t.Fatal("expected hits")
	}
	found := false
	for _, h := range hits {
		if h.ID == "acquiring-disk-image-with-dd-and-dcfldd" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected acquiring-disk-image slug in top hits, got %v", hits[0].ID)
	}
}

func TestByTechnique_nonEmpty(t *testing.T) {
	cat, err := Default()
	if err != nil {
		t.Skip(err)
	}
	// T1059.001 is heavily referenced in corpus
	hits := cat.ByTechnique("T1059.001")
	if len(hits) == 0 {
		t.Fatal("expected skills for T1059.001")
	}
}
