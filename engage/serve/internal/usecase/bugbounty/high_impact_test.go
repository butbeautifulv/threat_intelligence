package bugbounty

import "testing"

func TestSortedPriorityVulns(t *testing.T) {
	got := SortedPriorityVulns([]string{"xss", "rce", "csrf"})
	if len(got) == 0 {
		t.Fatal("empty")
	}
	if got[0] != "rce" {
		t.Fatalf("order %v", got)
	}
}

func TestTestScenariosFor_sqli(t *testing.T) {
	s := testScenariosFor("sqli")
	if len(s) == 0 {
		t.Fatal("expected scenarios")
	}
}
