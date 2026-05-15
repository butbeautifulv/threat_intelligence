package commit

import "testing"

func TestIOCNodeID(t *testing.T) {
	id, err := IOCNodeID("ti:ioc:abc123")
	if err != nil {
		t.Fatal(err)
	}
	if id != "abc123" {
		t.Fatalf("got %q want abc123", id)
	}
}

func TestIOCLinkNodeID(t *testing.T) {
	id, err := IOCLinkNodeID("ti:lc:camp1:deadbeef")
	if err != nil {
		t.Fatal(err)
	}
	if id != "deadbeef" {
		t.Fatalf("got %q want deadbeef", id)
	}
}

func TestTILinkSuffix(t *testing.T) {
	if got := TILinkSuffix("ti:lca:camp:actorhash"); got != "actorhash" {
		t.Fatalf("got %q", got)
	}
}
