package toolid

import "testing"

func TestCategory_Valid(t *testing.T) {
	valid := []Category{
		CategoryNetwork, CategoryWeb, CategoryCloud, CategoryBinary,
		CategoryAuth, CategoryOSINT, CategoryCTF, CategoryIntel,
	}
	for _, c := range valid {
		if !c.Valid() {
			t.Fatalf("%q should be valid", c)
		}
	}
}

func TestCategory_Valid_rejectsUnknown(t *testing.T) {
	bad := []Category{"", "unknown", Category("exploit")}
	for _, c := range bad {
		if c.Valid() {
			t.Fatalf("%q should be invalid", c)
		}
	}
}
