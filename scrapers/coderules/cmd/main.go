package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"coderules/internal/cwe"
	gh "coderules/internal/github"
	neo4jstore "coderules/internal/storage/neo4j"

	"gopkg.in/yaml.v3"
)

func main() {
	sources := flag.String("sources", getenvStr("CODERULES_SOURCES", "cwe,semgrep,codeql"), "comma-separated: cwe, semgrep, codeql")
	maxCWE := flag.Int("max-cwe", getenvInt("CODERULES_MAX_CWE", 5000), "max CWE weakness records from MITRE catalog")
	maxSemgrep := flag.Int("max-semgrep", getenvInt("CODERULES_MAX_SEMGREP", 80), "max Semgrep YAML rules")
	maxCodeQL := flag.Int("max-codeql", getenvInt("CODERULES_MAX_CODEQL", 60), "max CodeQL .ql files")
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

	src := map[string]bool{}
	for _, s := range strings.Split(*sources, ",") {
		src[strings.TrimSpace(strings.ToLower(s))] = true
	}

	if src["cwe"] {
		log.Println("ingesting CWE catalog (MITRE zip)…")
		if err := cwe.IngestFromMITRE(ctx, st, *maxCWE); err != nil {
			log.Fatalf("cwe: %v", err)
		}
	}
	if src["semgrep"] {
		if err := ingestSemgrep(ctx, st, *maxSemgrep); err != nil {
			log.Fatalf("semgrep: %v", err)
		}
	}
	if src["codeql"] {
		if err := ingestCodeQL(ctx, st, *maxCodeQL); err != nil {
			log.Fatalf("codeql: %v", err)
		}
	}
	log.Println("coderules scrape done")
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

func ingestSemgrep(ctx context.Context, st *neo4jstore.Store, limit int) error {
	log.Println("ingesting Semgrep community rules (subset)…")
	g := gh.NewClient()
	const owner, repo = "semgrep", "semgrep-rules"
	seeds := []string{"python", "javascript", "java", "go", "csharp", "dockerfile", "yaml", "bash"}
	var q []string
	for _, s := range seeds {
		q = append(q, s)
	}
	n := 0
	for len(q) > 0 && n < limit {
		dir := q[0]
		q = q[1:]
		items, err := g.ListDir(ctx, owner, repo, dir)
		if err != nil {
			continue
		}
		for _, it := range items {
			if n >= limit {
				break
			}
			if it.Type == "dir" && !strings.HasPrefix(it.Name, ".") {
				q = append(q, it.Path)
				continue
			}
			if it.Type != "file" || (!strings.HasSuffix(it.Name, ".yml") && !strings.HasSuffix(it.Name, ".yaml")) {
				continue
			}
			raw, err := g.FetchText(ctx, it.DownloadURL)
			if err != nil {
				continue
			}
			var root map[string]any
			if err := yaml.Unmarshal(raw, &root); err != nil {
				continue
			}
			ruleID, title := semgrepMeta(root, it.Name)
			id := neo4jstore.StableID("semgrep", it.Path)
			md := fmt.Sprintf("# %s\n\n**path:** `%s`\n\n```yaml\n%s\n```\n", title, it.Path, string(raw))
			lang := strings.Split(it.Path, "/")[0]
			if err := st.UpsertSemgrepRule(ctx, id, it.Path, title, lang, md); err != nil {
				return err
			}
			for _, cw := range semgrepCWES(root) {
				if err := st.LinkSemgrepRuleToCWE(ctx, id, cw); err != nil {
					return err
				}
			}
			_ = ruleID
			n++
		}
	}
	log.Printf("semgrep rules ingested: %d", n)
	return nil
}

func semgrepMeta(root map[string]any, fileName string) (id, title string) {
	if v, ok := root["id"].(string); ok && v != "" {
		id = v
	}
	if rules, ok := root["rules"].([]any); ok && len(rules) > 0 {
		if rm, ok := rules[0].(map[string]any); ok {
			if id == "" {
				if s, ok := rm["id"].(string); ok {
					id = s
				}
			}
			if msg, ok := rm["message"].(string); ok && strings.TrimSpace(msg) != "" {
				title = strings.TrimSpace(msg)
			}
		}
	}
	if title == "" {
		title = firstNonEmpty(id, fileName)
	}
	if id == "" {
		id = title
	}
	return id, title
}

var cweTokenRe = regexp.MustCompile(`(?i)CWE-\d+`)

func semgrepCWES(root map[string]any) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(s string) {
		for _, m := range cweTokenRe.FindAllString(s, -1) {
			u := strings.ToUpper(m)
			if _, ok := seen[u]; ok {
				continue
			}
			seen[u] = struct{}{}
			out = append(out, u)
		}
	}
	walkMeta := func(meta map[string]any) {
		if meta == nil {
			return
		}
		if v, ok := meta["cwe"]; ok {
			switch t := v.(type) {
			case string:
				add(t)
			case []any:
				for _, x := range t {
					if s, ok := x.(string); ok {
						add(s)
					}
				}
			}
		}
	}
	if rules, ok := root["rules"].([]any); ok {
		for _, r := range rules {
			rm, ok := r.(map[string]any)
			if !ok {
				continue
			}
			if md, ok := rm["metadata"].(map[string]any); ok {
				walkMeta(md)
			}
		}
	}
	return out
}

func ingestCodeQL(ctx context.Context, st *neo4jstore.Store, limit int) error {
	log.Println("ingesting CodeQL queries (subset)…")
	g := gh.NewClient()
	const path = "javascript/ql/src/Security/CWE-079"
	items, err := g.ListDir(ctx, "github", "codeql", path)
	if err != nil {
		return err
	}
	n := 0
	for _, it := range items {
		if n >= limit {
			break
		}
		if it.Type != "file" || !strings.HasSuffix(it.Name, ".ql") {
			continue
		}
		raw, err := g.FetchText(ctx, it.DownloadURL)
		if err != nil {
			continue
		}
		body := string(raw)
		name := it.Name
		id := neo4jstore.StableID("codeql", it.Path)
		md := fmt.Sprintf("# %s\n\n**path:** `%s`\n\n```ql\n%s\n```\n", name, it.Path, body)
		if err := st.UpsertCodeQLRule(ctx, id, it.Path, name, "javascript", md); err != nil {
			return err
		}
		for _, cw := range codeqlCWES(body) {
			if err := st.LinkCodeQLRuleToCWE(ctx, id, cw); err != nil {
				return err
			}
		}
		n++
	}
	log.Printf("codeql rules ingested: %d", n)
	return nil
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func codeqlCWES(body string) []string {
	lines := strings.Split(body, "\n")
	n := len(lines)
	if n > 100 {
		n = 100
	}
	chunk := strings.Join(lines[:n], "\n")
	seen := map[string]struct{}{}
	var out []string
	for _, m := range cweTokenRe.FindAllString(chunk, -1) {
		u := strings.ToUpper(m)
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}
