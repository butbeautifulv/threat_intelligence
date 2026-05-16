package neo4jstore

import "testing"

func TestExtractCVEs_dedup(t *testing.T) {
	got := extractCVEs("CVE-2024-1234 in title", "also CVE-2024-1234 and CVE-2023-99999")
	if len(got) != 2 {
		t.Fatalf("len %d: %v", len(got), got)
	}
	if got[0] != "CVE-2024-1234" || got[1] != "CVE-2023-99999" {
		t.Fatalf("%v", got)
	}
}
