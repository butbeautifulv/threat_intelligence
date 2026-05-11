package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"mcp/internal/connector/neo4jconn"
	"mcp/internal/domain"
)

type QueryUsecase struct {
	neo    *neo4jconn.Connector
	logger *slog.Logger
}

func NewQueryUsecase(neo *neo4jconn.Connector, logger *slog.Logger) *QueryUsecase {
	return &QueryUsecase{neo: neo, logger: logger}
}

func (u *QueryUsecase) ListKinds(ctx context.Context) ([]string, error) {
	res, err := u.neo.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		r, err := tx.Run(ctx, `MATCH (n) UNWIND labels(n) AS l RETURN l AS label, count(*) AS c ORDER BY c DESC`, nil)
		if err != nil {
			return nil, err
		}
		type row struct {
			Label string
			C     int64
		}
		var out []row
		for r.Next(ctx) {
			out = append(out, row{Label: r.Record().Values[0].(string), C: r.Record().Values[1].(int64)})
		}
		if err := r.Err(); err != nil {
			return nil, err
		}
		// Prefer first label only semantics (same as panel kind).
		// We still return all labels; agent can decide.
		kinds := make([]string, 0, len(out))
		for _, x := range out {
			if x.Label == "" {
				continue
			}
			kinds = append(kinds, x.Label)
		}
		return kinds, nil
	})
	if err != nil {
		return nil, err
	}
	kinds := res.([]string)
	sort.Strings(kinds)
	return kinds, nil
}

func (u *QueryUsecase) NodesByKind(ctx context.Context, kind string, limit, offset int) ([]domain.Node, error) {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return nil, fmt.Errorf("kind is empty")
	}
	if limit <= 0 || limit > 5000 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	q := fmt.Sprintf(`
MATCH (n:%s)
WITH n SKIP $offset LIMIT $limit
RETURN elementId(n) AS id, labels(n) AS labels, properties(n) AS props
`, safeLabel(kind))

	res, err := u.neo.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		r, err := tx.Run(ctx, q, map[string]any{"limit": limit, "offset": offset})
		if err != nil {
			return nil, err
		}
		var out []domain.Node
		for r.Next(ctx) {
			rec := r.Record()
			id, _ := rec.Get("id")
			labels, _ := rec.Get("labels")
			props, _ := rec.Get("props")
			out = append(out, domain.Node{
				ElementID: toString(id),
				Labels:    toStringSlice(labels),
				Props:     toMap(props),
			})
		}
		return out, r.Err()
	})
	if err != nil {
		return nil, err
	}
	return res.([]domain.Node), nil
}

func (u *QueryUsecase) GetNode(ctx context.Context, id string) (*domain.Node, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id is empty")
	}
	q := `
MATCH (n)
WHERE elementId(n) = $id OR n.id = $id OR n.cve = $id OR n.uri = $id OR n.link = $id
RETURN elementId(n) AS id, labels(n) AS labels, properties(n) AS props
LIMIT 1
`
	res, err := u.neo.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		r, err := tx.Run(ctx, q, map[string]any{"id": id})
		if err != nil {
			return nil, err
		}
		if !r.Next(ctx) {
			if err := r.Err(); err != nil {
				return nil, err
			}
			return (*domain.Node)(nil), nil
		}
		rec := r.Record()
		n := &domain.Node{
			ElementID: toString(mustGet(rec, "id")),
			Labels:    toStringSlice(mustGet(rec, "labels")),
			Props:     toMap(mustGet(rec, "props")),
		}
		return n, r.Err()
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return res.(*domain.Node), nil
}

func (u *QueryUsecase) Neighbors(ctx context.Context, id string, depth, limit int) (*domain.Graph, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id is empty")
	}
	if depth <= 0 || depth > 3 {
		depth = 1
	}
	if limit <= 0 || limit > 5000 {
		limit = 500
	}
	q := fmt.Sprintf(`
MATCH (seed)
WHERE elementId(seed) = $id OR seed.id = $id OR seed.cve = $id OR seed.uri = $id OR seed.link = $id
WITH seed
MATCH p=(seed)-[r*1..%d]-(n)
WITH collect(DISTINCT seed) + collect(DISTINCT n) AS ns, relationships(p) AS rs
UNWIND ns AS node
WITH DISTINCT node, rs
RETURN elementId(node) AS id, labels(node) AS labels, properties(node) AS props, rs AS rels
LIMIT $limit
`, depth)

	res, err := u.neo.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		r, err := tx.Run(ctx, q, map[string]any{"id": id, "limit": limit})
		if err != nil {
			return nil, err
		}

		nodeSeen := map[string]domain.Node{}
		edgeSeen := map[string]domain.Edge{}

		for r.Next(ctx) {
			rec := r.Record()
			nid := toString(mustGet(rec, "id"))
			if nid != "" {
				if _, ok := nodeSeen[nid]; !ok {
					nodeSeen[nid] = domain.Node{
						ElementID: nid,
						Labels:    toStringSlice(mustGet(rec, "labels")),
						Props:     toMap(mustGet(rec, "props")),
					}
				}
			}

			relsAny, _ := rec.Get("rels")
			rels, _ := relsAny.([]any)
			for _, rr := range rels {
				m, ok := rr.(map[string]any)
				if !ok {
					continue
				}
				rid := toString(m["id"])
				if rid == "" {
					// relationship objects from neo4j driver won't have "id" here.
					continue
				}
				if _, ok := edgeSeen[rid]; ok {
					continue
				}
				edgeSeen[rid] = domain.Edge{
					ElementID: rid,
					Type:      toString(m["type"]),
					Source:    toString(m["source"]),
					Target:    toString(m["target"]),
				}
			}
		}

		// If rels didn't materialize, do a second pass for edges (reliable).
		// This is intentionally simple: return nodes + edges directly around seed up to depth.
		nodes := make([]domain.Node, 0, len(nodeSeen))
		for _, n := range nodeSeen {
			nodes = append(nodes, n)
		}

		// Fallback edge query (most stable).
		edgesRes, err := tx.Run(ctx, fmt.Sprintf(`
MATCH (seed)
WHERE elementId(seed) = $id OR seed.id = $id OR seed.cve = $id OR seed.uri = $id OR seed.link = $id
MATCH p=(seed)-[r*1..%d]-(n)
UNWIND relationships(p) AS rel
RETURN DISTINCT elementId(rel) AS id, type(rel) AS type, elementId(startNode(rel)) AS source, elementId(endNode(rel)) AS target
LIMIT $limit
`, depth), map[string]any{"id": id, "limit": limit})
		if err != nil {
			return nil, err
		}
		var edges []domain.Edge
		for edgesRes.Next(ctx) {
			rec := edgesRes.Record()
			edges = append(edges, domain.Edge{
				ElementID: toString(mustGet(rec, "id")),
				Type:      toString(mustGet(rec, "type")),
				Source:    toString(mustGet(rec, "source")),
				Target:    toString(mustGet(rec, "target")),
			})
		}
		if err := edgesRes.Err(); err != nil {
			return nil, err
		}

		return &domain.Graph{Nodes: nodes, Edges: edges}, r.Err()
	})
	if err != nil {
		return nil, err
	}
	return res.(*domain.Graph), nil
}

func (u *QueryUsecase) Search(ctx context.Context, query string, kind string, limit int) ([]domain.Node, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is empty")
	}
	if limit <= 0 || limit > 2000 {
		limit = 50
	}

	kind = strings.TrimSpace(kind)
	var q string
	params := map[string]any{"q": strings.ToLower(query), "limit": limit}
	if kind != "" {
		q = fmt.Sprintf(`
MATCH (n:%s)
WHERE
  (exists(n.title) AND toLower(n.title) CONTAINS $q) OR
  (exists(n.name) AND toLower(n.name) CONTAINS $q) OR
  (exists(n.id) AND toLower(n.id) CONTAINS $q) OR
  (exists(n.cve) AND toLower(n.cve) CONTAINS $q) OR
  (exists(n.value) AND toLower(n.value) CONTAINS $q) OR
  (exists(n.uri) AND toLower(n.uri) CONTAINS $q) OR
  (exists(n.link) AND toLower(n.link) CONTAINS $q)
RETURN elementId(n) AS id, labels(n) AS labels, properties(n) AS props
LIMIT $limit
`, safeLabel(kind))
	} else {
		q = `
MATCH (n)
WHERE
  (exists(n.title) AND toLower(n.title) CONTAINS $q) OR
  (exists(n.name) AND toLower(n.name) CONTAINS $q) OR
  (exists(n.id) AND toLower(n.id) CONTAINS $q) OR
  (exists(n.cve) AND toLower(n.cve) CONTAINS $q) OR
  (exists(n.value) AND toLower(n.value) CONTAINS $q) OR
  (exists(n.uri) AND toLower(n.uri) CONTAINS $q) OR
  (exists(n.link) AND toLower(n.link) CONTAINS $q)
RETURN elementId(n) AS id, labels(n) AS labels, properties(n) AS props
LIMIT $limit
`
	}

	res, err := u.neo.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		r, err := tx.Run(ctx, q, params)
		if err != nil {
			return nil, err
		}
		var out []domain.Node
		for r.Next(ctx) {
			rec := r.Record()
			out = append(out, domain.Node{
				ElementID: toString(mustGet(rec, "id")),
				Labels:    toStringSlice(mustGet(rec, "labels")),
				Props:     toMap(mustGet(rec, "props")),
			})
		}
		return out, r.Err()
	})
	if err != nil {
		return nil, err
	}
	return res.([]domain.Node), nil
}

func safeLabel(label string) string {
	// defensive: keep only alnum and underscore.
	var b strings.Builder
	for _, r := range label {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "Node"
	}
	return out
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toStringSlice(v any) []string {
	a, ok := v.([]any)
	if !ok {
		if ss, ok := v.([]string); ok {
			return ss
		}
		return nil
	}
	out := make([]string, 0, len(a))
	for _, x := range a {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func toMap(v any) map[string]any {
	m, ok := v.(map[string]any)
	if ok {
		return m
	}
	// Some neo4j values come back as map[string]interface{} already;
	// but we keep a safe fallback for unexpected types.
	b, err := json.Marshal(v)
	if err != nil {
		return map[string]any{"value": fmt.Sprintf("%v", v)}
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]any{"value": fmt.Sprintf("%v", v)}
	}
	return out
}

func mustGet(rec *driver.Record, key string) any {
	v, _ := rec.Get(key)
	return v
}

