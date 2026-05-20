package config

import (
	"os"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/pkg/harvest"
)

func TestLoad_defaults(t *testing.T) {
	os.Clearenv()
	c := Load()
	if c.NATSURL != "nats://localhost:4222" {
		t.Fatalf("NATSURL %q", c.NATSURL)
	}
	if c.ScrapeStream != "SCRAPE" || c.ScrapeDurable != "pipeline-worker" {
		t.Fatal("scrape defaults")
	}
	if c.ScrapeSubject != "scrape.>" || c.IngestPublish != "ingest.events" {
		t.Fatal("subject defaults")
	}
	if c.Batch != 10 {
		t.Fatalf("batch %d", c.Batch)
	}
	if c.MaxWait != 5*time.Second {
		t.Fatalf("maxwait %v", c.MaxWait)
	}
}

func TestLoad_envOverrides(t *testing.T) {
	t.Setenv("NATS_URL", "nats://nats:4222")
	t.Setenv("NATS_SCRAPE_STREAM", "S1")
	t.Setenv("NATS_SCRAPE_DURABLE", "d1")
	t.Setenv("NATS_SCRAPE_SUBSCRIBE_SUBJECT", "scrape.ti.>")
	t.Setenv("NATS_INGEST_PUBLISH_SUBJECT", "ingest.all")
	t.Setenv("PIPELINE_BATCH", "3")
	t.Setenv("PIPELINE_MAX_WAIT", "2s")
	t.Setenv("TI_INGEST_SUBJECT", "ingest.ti.custom")

	c := Load()
	if c.NATSURL != "nats://nats:4222" || c.ScrapeStream != "S1" || c.ScrapeDurable != "d1" {
		t.Fatal("nats/scrape overrides")
	}
	if c.ScrapeSubject != "scrape.ti.>" || c.IngestPublish != "ingest.all" {
		t.Fatal("subject overrides")
	}
	if c.Batch != 3 || c.MaxWait != 2*time.Second {
		t.Fatal("batch/maxwait")
	}
	if c.DomainSubjects[harvest.SourceTI] != "ingest.ti.custom" {
		t.Fatal("domain subject map")
	}
}

func TestLoad_invalidIntAndDurationFallback(t *testing.T) {
	t.Setenv("PIPELINE_BATCH", "nope")
	t.Setenv("PIPELINE_MAX_WAIT", "not-a-duration")
	c := Load()
	if c.Batch != 10 {
		t.Fatalf("batch %d", c.Batch)
	}
	if c.MaxWait != 5*time.Second {
		t.Fatalf("maxwait %v", c.MaxWait)
	}
}

func TestIngestSubjectFor(t *testing.T) {
	c := Config{
		IngestPublish: "ingest.events",
		DomainSubjects: map[string]string{
			harvest.SourceTI: "ingest.ti.events",
		},
	}
	if got := c.IngestSubjectFor(harvest.SourceTI); got != "ingest.ti.events" {
		t.Fatalf("ti %q", got)
	}
	if got := c.IngestSubjectFor(harvest.SourceVuln); got != "ingest.events" {
		t.Fatalf("default %q", got)
	}
}
