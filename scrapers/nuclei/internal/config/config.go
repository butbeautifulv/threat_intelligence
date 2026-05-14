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
	MaxTemplates                            int
	YearsCSV                                string
	IngestMode                              string
	NATSURL, NATSSubject                    string
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
	return &Config{
		Neo4jURI:     getenv("NEO4J_URI", "neo4j://localhost:7687"),
		Neo4jUser:    getenv("NEO4J_USER", "neo4j"),
		Neo4jPass:    getenv("NEO4J_PASS", "neo4jpassword"),
		Neo4jDB:      getenv("NEO4J_DB", "neo4j"),
		MaxTemplates: getenvInt("NUCLEI_MAX", 120),
		YearsCSV:     getenv("NUCLEI_YEARS", "2023,2024"),
		IngestMode:   mode,
		NATSURL:      getenv("NATS_URL", "nats://localhost:4222"),
		NATSSubject:  getenv("NUCLEI_NATS_SUBJECT", "ingest.appsec.nuclei"),
	}
}
