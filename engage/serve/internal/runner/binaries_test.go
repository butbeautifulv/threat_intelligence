package runner

import "testing"

func TestCatalogBinaries_P9iCoverage(t *testing.T) {
	t.Parallel()
	for _, name := range []string{
		"anew", "arp", "exiftool", "netexec", "wfuzz", "volatility3", "zap",
	} {
		if !IsCatalogBinary(name) {
			t.Fatalf("missing catalog binary %q", name)
		}
	}
	if len(CatalogBinaries) < 100 {
		t.Fatalf("CatalogBinaries too small: %d", len(CatalogBinaries))
	}
}
