package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"sbom/internal/ghsa"
	"sbom/internal/osv"
	neo4jstore "sbom/internal/storage/neo4j"
)

func main() {
	sources := flag.String("sources", envOr("SBOM_SOURCES", "osv,ghsa"), "comma-separated: osv, ghsa")
	maxCVE := flag.Int("max-cves", envInt("SBOM_MAX_CVES", 200), "max CVEs from graph to enrich via OSV")
	maxGHSA := flag.Int("max-ghsa", envInt("SBOM_MAX_GHSA", 100), "max GHSA advisories to ingest")
	minYear := flag.Int("ghsa-min-year", envInt("SBOM_GHSA_MIN_YEAR", 2017), "minimum year when walking GHSA paths")
	flag.Parse()

	ctx := context.Background()
	cfg := neo4jstore.Config{
		URI:      getenv("NEO4J_URI", "neo4j://localhost:7687"),
		Username: getenv("NEO4J_USER", "neo4j"),
		Password: getenv("NEO4J_PASS", "neo4jpassword"),
		Database: getenv("NEO4J_DB", "neo4j"),
	}
	st, err := neo4jstore.New(ctx, cfg)
	if err != nil {
		log.Fatalf("neo4j: %v", err)
	}
	defer st.Close(ctx)
	if err := st.EnsureSchema(ctx); err != nil {
		log.Fatalf("schema: %v", err)
	}

	src := map[string]bool{}
	for _, s := range strings.Split(*sources, ",") {
		src[strings.TrimSpace(strings.ToLower(s))] = true
	}

	if src["osv"] {
		if err := runOSV(ctx, st, *maxCVE); err != nil {
			log.Fatalf("osv: %v", err)
		}
	}
	if src["ghsa"] {
		if err := runGHSA(ctx, st, *maxGHSA, *minYear); err != nil {
			log.Fatalf("ghsa: %v", err)
		}
	}
	log.Println("sbom scrape done")
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func envOr(k, def string) string {
	return getenv(k, def)
}

func envInt(k string, def int) int {
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

func runOSV(ctx context.Context, st *neo4jstore.Store, maxCVE int) error {
	cves, err := st.ListCVEs(ctx, maxCVE)
	if err != nil {
		return err
	}
	oc := osv.NewClient()
	for i, cve := range cves {
		doc, err := oc.GetVuln(ctx, cve)
		if err != nil {
			log.Printf("osv %s: %v", cve, err)
			continue
		}
		id, _ := doc["id"].(string)
		aff, _ := doc["affected"].([]any)
		var packs []map[string]any
		for _, a := range aff {
			if m, ok := a.(map[string]any); ok {
				packs = append(packs, m)
			}
		}
		if err := st.UpsertFromOSVVuln(ctx, id, cve, packs); err != nil {
			return fmt.Errorf("store osv %s: %w", cve, err)
		}
		if (i+1)%20 == 0 {
			log.Printf("osv progress %d/%d", i+1, len(cves))
		}
		time.Sleep(150 * time.Millisecond)
	}
	return nil
}

func runGHSA(ctx context.Context, st *neo4jstore.Store, maxGHSA, minYear int) error {
	token := os.Getenv("GITHUB_TOKEN")
	gc := ghsa.NewClient(token)
	paths, err := gc.CollectAdvisoryPaths(ctx, maxGHSA, minYear)
	if err != nil {
		return err
	}
	for i, p := range paths {
		doc, err := gc.FetchAdvisoryJSON(ctx, p)
		if err != nil {
			log.Printf("ghsa fetch %s: %v", p, err)
			continue
		}
		if err := st.UpsertGHSA(ctx, doc); err != nil {
			return fmt.Errorf("store ghsa %s: %w", p, err)
		}
		if (i+1)%10 == 0 {
			log.Printf("ghsa progress %d/%d", i+1, len(paths))
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}
