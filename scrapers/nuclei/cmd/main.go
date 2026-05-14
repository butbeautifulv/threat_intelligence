package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	gh "nuclei/internal/github"
	neo4jstore "nuclei/internal/storage/neo4j"

	"gopkg.in/yaml.v3"
)

func main() {
	max := flag.Int("max", getenvInt("NUCLEI_MAX", 120), "max templates to ingest")
	years := flag.String("years", getenvStr("NUCLEI_YEARS", "2023,2024"), "comma-separated CVE year folders under http/cves/")
	flag.Parse()

	ctx := context.Background()
	st, err := neo4jstore.New(ctx, neo4jstore.Config{
		URI:      getenvStr("NEO4J_URI", "neo4j://localhost:7687"),
		Username: getenvStr("NEO4J_USER", "neo4j"),
		Password: getenvStr("NEO4J_PASS", "neo4jpassword"),
		Database: getenvStr("NEO4J_DB", "neo4j"),
	})
	if err != nil {
		log.Fatalf("neo4j: %v", err)
	}
	defer st.Close(ctx)
	if err := st.EnsureSchema(ctx); err != nil {
		log.Fatalf("schema: %v", err)
	}

	g := gh.NewClient()
	const owner, repo = "projectdiscovery", "nuclei-templates"
	n := 0
	for _, y := range strings.Split(*years, ",") {
		if n >= *max {
			break
		}
		year := strings.TrimSpace(y)
		if year == "" {
			continue
		}
		base := "http/cves/" + year
		items, err := g.ListDir(ctx, owner, repo, base)
		if err != nil {
			log.Printf("list %s: %v", base, err)
			continue
		}
		for _, it := range items {
			if n >= *max {
				break
			}
			if it.Type != "file" || (!strings.HasSuffix(it.Name, ".yaml") && !strings.HasSuffix(it.Name, ".yml")) {
				continue
			}
			raw, err := g.FetchText(ctx, it.DownloadURL)
			if err != nil {
				continue
			}
			tplID, name, sev, tags, cve, cwe, err := parseNuclei(raw)
			if err != nil || tplID == "" {
				continue
			}
			id := neo4jstore.StableID("nuclei", it.Path)
			md := fmt.Sprintf("# %s\n\n**id:** `%s`  \n**path:** `%s`\n\n```yaml\n%s\n```\n", name, tplID, it.Path, string(raw))
			if err := st.UpsertNucleiTemplate(ctx, id, tplID, it.Path, name, sev, tags, cve, cwe, md); err != nil {
				log.Fatalf("store: %v", err)
			}
			n++
		}
	}
	log.Printf("nuclei templates ingested: %d", n)
}

func getenvStr(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
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

func parseNuclei(raw []byte) (tplID, name, severity, tagsJSON, cveID, cweID string, err error) {
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return "", "", "", "", "", "", err
	}
	tplID, _ = root["id"].(string)
	tagsJSON = "[]"
	info, _ := root["info"].(map[string]any)
	if info != nil {
		name, _ = info["name"].(string)
		severity, _ = info["severity"].(string)
		tagsJSON = "[]"
		if tg, ok := info["tags"]; ok {
			tags := normalizeTags(tg)
			b, _ := json.Marshal(tags)
			tagsJSON = string(b)
			if tagsJSON == "null" {
				tagsJSON = "[]"
			}
		}
		class, _ := info["classification"].(map[string]any)
		if class != nil {
			if s, ok := class["cve-id"].(string); ok {
				cveID = strings.TrimSpace(strings.ToUpper(s))
			}
			if s, ok := class["cwe-id"].(string); ok {
				cweID = strings.TrimSpace(s)
			}
		}
	}
	if cveID == "" && strings.HasPrefix(strings.ToUpper(tplID), "CVE-") {
		cveID = strings.ToUpper(tplID)
	}
	return tplID, name, severity, tagsJSON, cveID, cweID, nil
}

func normalizeTags(v any) []string {
	switch t := v.(type) {
	case string:
		var out []string
		for _, p := range strings.Split(t, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	case []any:
		var out []string
		for _, x := range t {
			if s, ok := x.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	default:
		return nil
	}
}
