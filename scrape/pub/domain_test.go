package scrapepub

import (
	"testing"

	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"
)

func TestNewDomainPublisher_trimsSubject(t *testing.T) {
	dp := NewDomainPublisher(nil, scrapev1.SourceDS, "  scrape.ds.events  ")
	if dp.Subject != "scrape.ds.events" {
		t.Fatalf("subject = %q", dp.Subject)
	}
	if dp.Source != scrapev1.SourceDS {
		t.Fatalf("source = %q", dp.Source)
	}
}
