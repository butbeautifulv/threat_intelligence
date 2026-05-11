import neo4j from "neo4j-driver";
import { NextResponse } from "next/server";

function asStringArray(v: unknown): string[] | undefined {
  if (!Array.isArray(v)) return undefined;
  const out = v.filter((x) => typeof x === "string") as string[];
  return out.length ? out : undefined;
}

function shortCpe(uri: string): string {
  // Keep it readable: show vendor:product:version-ish tail.
  // cpe:2.3:a:vendor:product:version:...
  const parts = uri.split(":");
  if (parts.length >= 7) {
    const vendor = parts[3] || "?";
    const product = parts[4] || "?";
    const version = parts[5] || "*";
    return `${vendor}:${product}:${version}`;
  }
  return uri;
}

function deriveTitle(labels: string[], props: Record<string, unknown>, fallbackId: string) {
  const label0 = labels[0] || "";
  if (label0 === "CPE") {
    const uri = typeof props.uri === "string" ? props.uri : "";
    if (uri) return `CPE ${shortCpe(uri)}`;
  }
  if (label0 === "Vulnerability") {
    const cve = typeof props.cve === "string" ? props.cve : "";
    if (cve) return cve;
  }
  if (label0 === "CWE") {
    const id = typeof props.id === "string" ? props.id : "";
    if (id) return id;
  }
  if (label0 === "IOC") {
    const value = typeof props.value === "string" ? props.value : "";
    const t = typeof props.type === "string" ? props.type : "";
    if (value && t) return `${t}:${value}`;
  }
  return (
    (typeof props.title === "string" && props.title) ||
    (typeof props.name === "string" && props.name) ||
    (typeof props.id === "string" && props.id) ||
    (label0 ? `${label0} ${fallbackId}` : fallbackId)
  );
}

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

    const nodes: Array<{
      id: string;
      labels: string[];
      title: string;
      markdown?: string | null;
      kind?: string;
      tags?: string[];
    }> = [];
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
        const kind = labels[0] || "Node";
        const title = deriveTitle(labels, props, id);
        const tags = asStringArray(props.tags);
        nodes.push({
          id,
          labels,
          title,
          markdown: typeof props.markdown === "string" ? (props.markdown as string) : null,
          kind,
          tags,
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

    // Add category nodes to improve layout (Obsidian-style “hubs”).
    // Also makes browsing less chaotic when there are many CPE/leaf nodes.
    const categoryIds = new Set<string>();
    for (const n of nodes) {
      // Never create a category hub for category nodes themselves.
      if (n.labels?.includes("Category") || n.kind === "Category") continue;
      const kind = n.kind || n.labels?.[0] || "Node";
      if (kind === "Category") continue;
      const catId = `category:${kind}`;
      if (!categoryIds.has(catId)) {
        categoryIds.add(catId);
        nodes.push({
          id: catId,
          labels: ["Category"],
          title: kind,
          kind: "Category",
        });
      }
      const edgeId = `cat:${catId}->${n.id}`;
      if (!edgeSeen.has(edgeId) && !n.id.startsWith("category:")) {
        edgeSeen.add(edgeId);
        edges.push({ id: edgeId, type: "CATEGORY", source: catId, target: n.id });
      }
    }

    return NextResponse.json({ nodes, edges });
  } finally {
    await session.close();
    await driver.close();
  }
}

