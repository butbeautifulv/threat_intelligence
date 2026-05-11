package neo4jstore

import (
	"context"
	"fmt"
	"time"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/butbeautifulv/threat_intelligence/graph/neo4j"
	"vuln/internal/domain"
	"vuln/internal/repository"
)

type Store struct {
	client *neo4j.Client
}

var _ repository.VulnerabilityRepository = (*Store)(nil)

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
		`CREATE CONSTRAINT vuln_cve IF NOT EXISTS FOR (n:Vulnerability) REQUIRE n.cve IS UNIQUE`,
		`CREATE CONSTRAINT cwe_id IF NOT EXISTS FOR (n:CWE) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT cpe_uri IF NOT EXISTS FOR (n:CPE) REQUIRE n.uri IS UNIQUE`,
	})
}

func (s *Store) Save(ctx context.Context, v *domain.Vulnerability) error {
	return s.Upsert(ctx, v)
}

func (s *Store) FindByCVE(ctx context.Context, id string) (*domain.Vulnerability, error) {
	// Minimal implementation for interface completeness.
	// For now, only returns whether the node exists (and summary fields).
	var out domain.Vulnerability
	q := `
MATCH (v:Vulnerability {cve: $cve})
RETURN v.cve AS cve, v.summary AS summary, v.cvss_base AS cvss_base, v.cvss_vector AS cvss_vector, v.cvss_version AS cvss_version
LIMIT 1
`
	params := map[string]any{"cve": id}

	sess := s.client.Session(ctx)
	defer sess.Close(ctx)

	rec, err := sess.ExecuteRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, q, params)
		if err != nil {
			return nil, err
		}
		if !res.Next(ctx) {
			if err := res.Err(); err != nil {
				return nil, err
			}
			return nil, nil
		}
		r := res.Record()
		if v, ok := r.Get("cve"); ok {
			out.CVE, _ = v.(string)
		}
		if v, ok := r.Get("summary"); ok {
			out.Summary, _ = v.(string)
		}
		if v, ok := r.Get("cvss_base"); ok {
			bv, ok := v.(float64)
			if !ok {
				bvInt, ok2 := v.(int64)
				if ok2 {
					bv = float64(bvInt)
					ok = true
				}
			}
			if ok {
			if out.CVSS == nil {
				out.CVSS = &domain.CVSS{}
			}
			out.CVSS.Base = bv
			}
		}
		if v, ok := r.Get("cvss_vector"); ok {
			vec, _ := v.(string)
			if out.CVSS == nil {
				out.CVSS = &domain.CVSS{}
			}
			out.CVSS.Vector = vec
		}
		if v, ok := r.Get("cvss_version"); ok {
			ver, _ := v.(string)
			if out.CVSS == nil {
				out.CVSS = &domain.CVSS{}
			}
			out.CVSS.Version = ver
		}
		return &out, nil
	})
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, fmt.Errorf("vulnerability not found: %s", id)
	}
	v, ok := rec.(*domain.Vulnerability)
	if !ok {
		return nil, fmt.Errorf("unexpected record type")
	}
	return v, nil
}

func (s *Store) Upsert(ctx context.Context, v *domain.Vulnerability) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	params := map[string]any{
		"cve":        v.CVE,
		"id":         v.ID,
		"summary":    v.Summary,
		"source":     "nvd",
		"updatedAt":  now,
		"cwes":       v.CWE,
		"cpes":       cpeURIs(v.CPEs),
		"cvss_base":  any(nil),
		"cvss_vec":   any(nil),
		"cvss_ver":   any(nil),
		"markdown":   "# " + v.CVE + "\n\n" + v.Summary,
	}
	if v.CVSS != nil {
		params["cvss_base"] = v.CVSS.Base
		params["cvss_vec"] = v.CVSS.Vector
		params["cvss_ver"] = v.CVSS.Version
	}

	q := `
MERGE (v:Vulnerability {cve: $cve})
SET v.id = $id,
    v.summary = $summary,
    v.markdown = $markdown,
    v.source = $source,
    v.updatedAt = $updatedAt
FOREACH (_ IN CASE WHEN $cvss_base IS NULL THEN [] ELSE [1] END |
  SET v.cvss_base = $cvss_base,
      v.cvss_vector = $cvss_vec,
      v.cvss_version = $cvss_ver
)

WITH v
UNWIND (CASE WHEN $cwes IS NULL THEN [] ELSE $cwes END) AS cweId
WITH v, cweId WHERE cweId IS NOT NULL AND cweId <> ""
MERGE (cwe:CWE {id: cweId})
MERGE (v)-[:HAS_CWE]->(cwe)

WITH v
UNWIND (CASE WHEN $cpes IS NULL THEN [] ELSE $cpes END) AS cpeUri
WITH v, cpeUri WHERE cpeUri IS NOT NULL AND cpeUri <> ""
MERGE (cpe:CPE {uri: cpeUri})
MERGE (v)-[:AFFECTS]->(cpe)
`

	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func cpeURIs(cpes []domain.CPE) []string {
	if len(cpes) == 0 {
		return nil
	}
	out := make([]string, 0, len(cpes))
	for _, c := range cpes {
		if c.URI != "" {
			out = append(out, c.URI)
		}
	}
	return out
}

