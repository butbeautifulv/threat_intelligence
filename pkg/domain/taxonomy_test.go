package domain

import "testing"

func TestVeilCategoryValid(t *testing.T) {
	for _, c := range CommonCategories() {
		if !c.Valid() {
			t.Fatalf("%q should be valid", c)
		}
	}
	if VeilCategory("not-a-category").Valid() {
		t.Fatal("unknown category should be invalid")
	}
}

func TestSubdomainFamilyNonEmpty(t *testing.T) {
	if len(SubdomainFamily) == 0 {
		t.Fatal("SubdomainFamily should be populated")
	}
	for sub, cats := range SubdomainFamily {
		if sub == "" || len(cats) == 0 {
			t.Fatalf("bad subdomain entry %q -> %v", sub, cats)
		}
		if cats[0] != CategoryPlaybook {
			t.Fatalf("%q should include playbook category", sub)
		}
	}
}
