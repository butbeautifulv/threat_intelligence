package neo4jstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/butbeautifulv/threat_intelligence/graph/neo4j"
	"lola/internal/domain"
	"lola/internal/repository"
)

type Store struct {
	client *neo4j.Client
}

var _ repository.LolaRepository = (*Store)(nil)

type Config = neo4j.Config

func New(ctx context.Context, cfg Config) (*Store, error) {
	c, err := neo4j.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Store{client: c}, nil
}

func (s *Store) Close(ctx context.Context) error { return s.client.Close(ctx) }

func (s *Store) EnsureSchema(ctx context.Context) error {
	return neo4j.EnsureConstraints(ctx, s.client, []string{
		`CREATE CONSTRAINT lola_artifact_id IF NOT EXISTS FOR (n:LolaArtifact) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT lola_command_id IF NOT EXISTS FOR (n:Command) REQUIRE n.id IS UNIQUE`,
	})
}

func (s *Store) UpsertArtifact(ctx context.Context, source string, a *domain.Artifact) error {
	if a.Name == "" {
		return fmt.Errorf("artifact name required")
	}
	id := fmt.Sprintf("%s:%s", source, slug(a.Name))
	now := time.Now().UTC().Format(time.RFC3339Nano)

	osJSON, _ := json.Marshal(a.OS)
	sigmaJSON, _ := json.Marshal(a.Detection.Sigma)
	yaraJSON, _ := json.Marshal(a.Detection.Yara)
	pathsJSON, _ := json.Marshal(a.Paths)
	resJSON, _ := json.Marshal(a.Resources)

	md := renderArtifactMarkdown(source, a)

	params := map[string]any{
		"id":          id,
		"source":      source,
		"name":        a.Name,
		"description": a.Description,
		"mitreID":     a.MitreID,
		"category":    a.Category,
		"privileges":  a.Privileges,
		"os":          string(osJSON),
		"sigma":       string(sigmaJSON),
		"yara":        string(yaraJSON),
		"paths":       string(pathsJSON),
		"resources":   string(resJSON),
		"markdown":    md,
		"updatedAt":   now,
	}

	q := `
MERGE (n:LolaArtifact {id: $id})
SET n.source = $source,
    n.name = $name,
    n.description = $description,
    n.mitreID = $mitreID,
    n.category = $category,
    n.privileges = $privileges,
    n.os = $os,
    n.sigma = $sigma,
    n.yara = $yara,
    n.paths = $paths,
    n.resources = $resources,
    n.markdown = $markdown,
    n.updatedAt = $updatedAt
`
	if err := s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	}); err != nil {
		return err
	}

	for i, cmd := range a.Commands {
		cid := commandID(id, i, cmd.Command)
		cp := map[string]any{
			"cid":         cid,
			"artifactId":  id,
			"command":     cmd.Command,
			"description": cmd.Description,
			"usecase":     cmd.Usecase,
			"category":    cmd.Category,
			"privileges":  cmd.Privileges,
			"mitreID":     cmd.MitreID,
			"os":          joinJSON(cmd.OS),
			"tags":        joinJSON(cmd.Tags),
			"updatedAt":   now,
		}
		q2 := `
MERGE (c:Command {id: $cid})
SET c.command = $command,
    c.description = $description,
    c.usecase = $usecase,
    c.category = $category,
    c.privileges = $privileges,
    c.mitreID = $mitreID,
    c.os = $os,
    c.tags = $tags,
    c.updatedAt = $updatedAt
WITH c
MATCH (a:LolaArtifact {id: $artifactId})
MERGE (a)-[r:HAS_COMMAND]->(c)
SET r.updatedAt = $updatedAt
`
		if err := s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
			_, err := tx.Run(ctx, q2, cp)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

func slug(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r+('a'-'A'))
		default:
			if len(out) > 0 && out[len(out)-1] != '_' {
				out = append(out, '_')
			}
		}
	}
	if len(out) == 0 {
		return "unknown"
	}
	outStr := string(out)
	outStr = strings.Trim(outStr, "_")
	if outStr == "" {
		return "unknown"
	}
	return outStr
}

func commandID(artifactID string, idx int, cmd string) string {
	h := fmt.Sprintf("%s#%d#%s", artifactID, idx, cmd)
	// short stable id without crypto import for simplicity
	var x uint64 = 14695981039346656037
	for _, b := range []byte(h) {
		x ^= uint64(b)
		x *= 1099511628211
	}
	return fmt.Sprintf("cmd:%s:%016x", slug(artifactID), x)
}

func joinJSON(ss []string) string {
	b, _ := json.Marshal(ss)
	return string(b)
}

func renderArtifactMarkdown(source string, a *domain.Artifact) string {
	var b []byte
	b = append(b, fmt.Sprintf("# %s\n\n", a.Name)...)
	b = append(b, fmt.Sprintf("**Source:** `%s`  \n", source)...)
	if a.Category != "" {
		b = append(b, fmt.Sprintf("**Category:** %s  \n", a.Category)...)
	}
	if a.MitreID != "" {
		b = append(b, fmt.Sprintf("**MITRE:** %s  \n", a.MitreID)...)
	}
	if len(a.OS) > 0 {
		b = append(b, fmt.Sprintf("**OS:** %v  \n\n", a.OS)...)
	}
	b = append(b, "## Description\n\n"...)
	b = append(b, a.Description...)
	b = append(b, "\n\n## Commands\n\n"...)
	for _, c := range a.Commands {
		b = append(b, "### Command\n\n"...)
		b = append(b, "```\n"...)
		b = append(b, c.Command...)
		b = append(b, "\n```\n\n"...)
		if c.Description != "" {
			b = append(b, c.Description...)
			b = append(b, "\n\n"...)
		}
	}
	if len(a.Detection.Sigma) > 0 {
		b = append(b, "## Sigma (links)\n\n"...)
		for _, s := range a.Detection.Sigma {
			b = append(b, fmt.Sprintf("- %s\n", s)...)
		}
		b = append(b, "\n"...)
	}
	if len(a.Paths) > 0 {
		b = append(b, "## Paths\n\n"...)
		for _, p := range a.Paths {
			b = append(b, fmt.Sprintf("- `%s`\n", p)...)
		}
	}
	return string(b)
}
