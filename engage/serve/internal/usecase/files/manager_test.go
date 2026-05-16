package files

import (
	"testing"
)

func TestManager_createListDelete(t *testing.T) {
	m, err := NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Create("reports/a.txt", "hello", false); err != nil {
		t.Fatal(err)
	}
	list, err := m.List("reports")
	if err != nil {
		t.Fatal(err)
	}
	files, _ := list["files"].([]string)
	if len(files) != 1 || files[0] != "a.txt" {
		t.Fatalf("list: %v", list)
	}
	if _, err := m.Delete("reports/a.txt"); err != nil {
		t.Fatal(err)
	}
}

func TestManager_rejectsTraversal(t *testing.T) {
	m, err := NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Create("../escape.txt", "x", false); err == nil {
		t.Fatal("expected traversal error")
	}
}
