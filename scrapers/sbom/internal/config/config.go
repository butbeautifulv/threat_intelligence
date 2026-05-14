package config

import (
	"os"
	"strconv"
	"strings"
)

// IngestModeDirect writes to Neo4j from the scraper process.
// IngestModeNATS publishes envelopes to JetStream; Neo4j is updated by ingest-worker.
const (
	IngestModeDirect = "direct"
	IngestModeNATS   = "nats"
)

// Config holds sbom scraper settings (env + defaults).
type Config struct {
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPass     string
	Neo4jDB       string
	Sources       []string
	MaxCVE        int
	MaxGHSA       int
	GHSAMinYear   int
	IngestMode    string
	NATSURL       string
	NATSSubject   string
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func FromEnv() *Config {
	mode := strings.ToLower(strings.TrimSpace(getenv("INGEST_MODE", IngestModeDirect)))
	if mode != IngestModeNATS {
		mode = IngestModeDirect
	}
	src := getenv("SBOM_SOURCES", "osv,ghsa")
	var parts []string
	for _, s := range strings.Split(src, ",") {
		s = strings.TrimSpace(strings.ToLower(s))
		if s != "" {
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		parts = []string{"osv", "ghsa"}
	}
	return &Config{
		Neo4jURI:    getenv("NEO4J_URI", "neo4j://localhost:7687"),
		Neo4jUser:   getenv("NEO4J_USER", "neo4j"),
		Neo4jPass:   getenv("NEO4J_PASS", "neo4jpassword"),
		Neo4jDB:     getenv("NEO4J_DB", "neo4j"),
		Sources:     parts,
		MaxCVE:      getenvInt("SBOM_MAX_CVES", 200),
		MaxGHSA:     getenvInt("SBOM_MAX_GHSA", 100),
		GHSAMinYear: getenvInt("SBOM_GHSA_MIN_YEAR", 2017),
		IngestMode:  mode,
		NATSURL:     getenv("NATS_URL", "nats://localhost:4222"),
		NATSSubject: getenv("SBOM_NATS_SUBJECT", "ingest.appsec.sbom"),
	}
}
