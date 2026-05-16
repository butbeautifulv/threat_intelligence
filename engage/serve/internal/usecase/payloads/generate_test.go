package payloads

import (
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/files"
)

func TestGenerate_buffer(t *testing.T) {
	fm, err := files.NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	res, err := Generate(fm, Request{Type: "buffer", Size: 10, Pattern: "AB"})
	if err != nil {
		t.Fatal(err)
	}
	if res["created"] != true {
		t.Fatalf("res: %v", res)
	}
}

func TestGenerate_rejectsOversize(t *testing.T) {
	fm, _ := files.NewManager(t.TempDir())
	_, err := Generate(fm, Request{Size: maxPayloadSize + 1})
	if err == nil {
		t.Fatal("expected error")
	}
}
