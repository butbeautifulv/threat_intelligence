import neo4j from "neo4j-driver";
import { NextResponse } from "next/server";

function env(name: string, fallback?: string) {
  const v = process.env[name] ?? fallback;
  if (!v) throw new Error(`Missing env: ${name}`);
  return v;
}

function getDriver() {
  const uri = env("NEO4J_URI", "neo4j://localhost:7687");
  const user = env("NEO4J_USER", "neo4j");
  const pass = env("NEO4J_PASS", "neo4jpassword");
  return neo4j.driver(uri, neo4j.auth.basic(user, pass));
}

export async function GET() {
  const driver = getDriver();
  const session = driver.session({ database: process.env.NEO4J_DB || "neo4j" });
  try {
    const res = await session.run(
      `
      MATCH (n)
      WITH n LIMIT 400
      OPTIONAL MATCH (n)-[r]->(m)
      RETURN n, collect(distinct { id: elementId(r), type: type(r), source: elementId(n), target: elementId(m) }) AS rels
      `
    );

    const nodes: Array<{ id: string; labels: string[]; title: string; markdown?: string | null }> = [];
    const edges: Array<{ id: string; type: string; source: string; target: string }> = [];
    const nodeSeen = new Set<string>();
    const edgeSeen = new Set<string>();

    for (const rec of res.records) {
      const n = rec.get("n");
      const id = n.elementId as string;
      if (!nodeSeen.has(id)) {
        nodeSeen.add(id);
        const labels = (n.labels ?? []) as string[];
        const props = (n.properties ?? {}) as Record<string, unknown>;
        const title =
          (typeof props.title === "string" && props.title) ||
          (typeof props.name === "string" && props.name) ||
          (typeof props.id === "string" && props.id) ||
          (labels[0] ? `${labels[0]} ${id}` : id);
        nodes.push({
          id,
          labels,
          title,
          markdown: typeof props.markdown === "string" ? (props.markdown as string) : null,
        });
      }

      const rels = rec.get("rels") as Array<{ id: string; type: string; source: string; target: string | null }>;
      for (const r of rels) {
        if (!r?.id || !r.target) continue;
        if (edgeSeen.has(r.id)) continue;
        edgeSeen.add(r.id);
        edges.push({ id: r.id, type: r.type, source: r.source, target: r.target });
      }
    }

    return NextResponse.json({ nodes, edges });
  } finally {
    await session.close();
    await driver.close();
  }
}

