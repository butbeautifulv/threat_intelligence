import neo4j from "neo4j-driver";
import { NextResponse } from "next/server";

function asStringArray(v: unknown): string[] | undefined {
  if (!Array.isArray(v)) return undefined;
  const out = v.filter((x) => typeof x === "string") as string[];
  return out.length ? out : undefined;
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

function toMarkdown(props: Record<string, unknown>) {
  const md = typeof props.markdown === "string" ? props.markdown : null;
  if (md) return md;
  const lines: string[] = ["```yaml"];
  for (const [k, v] of Object.entries(props)) {
    if (k === "markdown") continue;
    lines.push(`${k}: ${JSON.stringify(v)}`);
  }
  lines.push("```");
  return lines.join("\n");
}

export async function GET(_: Request, ctx: { params: Promise<{ id: string }> }) {
  const { id } = await ctx.params;

  // Synthetic category nodes are not stored in Neo4j.
  if (id.startsWith("category:")) {
    const kind = id.slice("category:".length) || "Node";
    const descriptions: Record<string, string> = {
      Vulnerability:
        "CVE / vulnerability entries. These nodes usually connect to **CWE** (weakness) and **CPE** (affected products).",
      CWE: "Common Weakness Enumeration — vulnerability classes (e.g. `CWE-79`).",
      CPE:
        "Common Platform Enumeration — affected product identifiers. The UI shortens them to `vendor:product:version` for readability.",
      IOC: "Indicators of Compromise — IPs/domains/URLs/hashes referenced by reports/campaigns.",
      Actor: "Threat actors (if ingested).",
      Campaign: "Campaign entities (if ingested).",
      Report: "Reports/articles (if ingested).",
      SigmaRule: "Sigma detections (if ingested).",
      YaraRule: "YARA signatures (if ingested).",
      AtomicTest: "Atomic Red Team tests (if ingested).",
    };

    const body = descriptions[kind] ?? "Category hub node used to organize the graph view.";
    const node = {
      id,
      labels: ["Category"],
      title: kind,
      markdown: `# ${kind}\n\n${body}\n`,
      kind: "Category",
      tags: undefined as string[] | undefined,
    };
    return NextResponse.json({ node, markdown: node.markdown });
  }

  const driver = getDriver();
  const session = driver.session({ database: process.env.NEO4J_DB || "neo4j" });
  try {
    const res = await session.run(
      `
      MATCH (n)
      WHERE elementId(n) = $id
      RETURN n
      LIMIT 1
      `,
      { id }
    );

    if (!res.records.length) {
      return NextResponse.json({ error: "not_found" }, { status: 404 });
    }

    const n = res.records[0].get("n");
    const labels = (n.labels ?? []) as string[];
    const props = (n.properties ?? {}) as Record<string, unknown>;
    const title =
      (typeof props.title === "string" && props.title) ||
      (typeof props.name === "string" && props.name) ||
      (typeof props.id === "string" && props.id) ||
      (labels[0] ? `${labels[0]} ${id}` : id);
    const tags = asStringArray(props.tags);

    const node = {
      id: n.elementId as string,
      labels,
      title,
      markdown: typeof props.markdown === "string" ? (props.markdown as string) : null,
      kind: labels[0] || "Node",
      tags,
    };
    return NextResponse.json({ node, markdown: toMarkdown(props) });
  } finally {
    await session.close();
    await driver.close();
  }
}

