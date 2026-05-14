package config

import (
	"os"
	"strconv"
	"strings"
)

const IngestModeDirect = "direct"
const IngestModeNATS = "nats"

type Config struct {
	Neo4jURI, Neo4jUser, Neo4jPass, Neo4jDB string
	Sources                                []string
	MaxCWE, MaxSemgrep, MaxCodeQL          int
	IngestMode                             string
	NATSURL, NATSSubject                   string
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
	src := getenv("CODERULES_SOURCES", "cwe,semgrep,codeql")
	var parts []string
	for _, s := range strings.Split(src, ",") {
		s = strings.TrimSpace(strings.ToLower(s))
		if s != "" {
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		parts = []string{"cwe", "semgrep", "codeql"}
	}
	return &Config{
		Neo4jURI:    getenv("NEO4J_URI", "neo4j://localhost:7687"),
		Neo4jUser:   getenv("NEO4J_USER", "neo4j"),
		Neo4jPass:   getenv("NEO4J_PASS", "neo4jpassword"),
		Neo4jDB:     getenv("NEO4J_DB", "neo4j"),
		Sources:     parts,
		MaxCWE:      getenvInt("CODERULES_MAX_CWE", 5000),
		MaxSemgrep:  getenvInt("CODERULES_MAX_SEMGREP", 80),
		MaxCodeQL:   getenvInt("CODERULES_MAX_CODEQL", 60),
		IngestMode:  mode,
		NATSURL:     getenv("NATS_URL", "nats://localhost:4222"),
		NATSSubject: getenv("CODERULES_NATS_SUBJECT", "ingest.appsec.coderules"),
	}
}
