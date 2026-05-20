// Seed CyberSkill nodes and HAS_PLAYBOOK edges from docs/skills-index/cyber-skills.json.
package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	graphneo4j "github.com/butbeautifulv/veil/knowledge/connector/neo4j"
	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
	pbprocedure "github.com/butbeautifulv/veil/pkg/playbook/procedure"
	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cat, err := pbindex.Open("")
	if err != nil {
		log.Fatal(err)
	}
	procCat, _ := pbprocedure.Open("")
	procByID := map[string]int{}
	if procCat != nil {
		for _, p := range procCat.Meta().Procedures {
			procByID[p.ID] = p.StepCount
		}
	}
	cfg := graphneo4j.Config{
		URI:      envOr("NEO4J_URI", "neo4j://localhost:7687"),
		Username: envOr("NEO4J_USERNAME", "neo4j"),
		Password: envOr("NEO4J_PASSWORD", "neo4jpassword"),
		Database: envOr("NEO4J_DATABASE", "neo4j"),
	}
	client, err := graphneo4j.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close(ctx)

	if err := graphneo4j.EnsureConstraints(ctx, client, []string{
		`CREATE CONSTRAINT cyber_skill_id IF NOT EXISTS FOR (n:CyberSkill) REQUIRE n.id IS UNIQUE`,
	}); err != nil {
		log.Fatal(err)
	}

	meta := cat.Meta()
	linked := 0
	for _, s := range meta.Skills {
		steps := procByID[s.ID]
		hasStructured := steps > 0
		if err := client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
			_, err := tx.Run(ctx, `
MERGE (sk:CyberSkill {id: $id})
SET sk.title = $title, sk.subdomain = $subdomain, sk.source = 'anthropic-cyber-skills',
    sk.stepCount = $stepCount, sk.hasStructured = $hasStructured, sk.updatedAt = datetime()`,
				map[string]any{
					"id": s.ID, "title": s.Name, "subdomain": s.Subdomain,
					"stepCount": steps, "hasStructured": hasStructured,
				})
			return err
		}); err != nil {
			log.Printf("skill %s: %v", s.ID, err)
			continue
		}
		for _, tid := range s.AttackIDs {
			err := client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
				_, err := tx.Run(ctx, `
MATCH (t:AttackTechnique {id: $tid}), (sk:CyberSkill {id: $sid})
MERGE (t)-[r:HAS_PLAYBOOK]->(sk)
SET r.updatedAt = datetime()`,
					map[string]any{"tid": strings.ToUpper(tid), "sid": s.ID})
				return err
			})
			if err == nil {
				linked++
			}
		}
	}
	log.Printf("seeded %d skills, %d HAS_PLAYBOOK edges", len(meta.Skills), linked)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
