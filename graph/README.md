# Graph layer

Consumes `ingest.>`, MERGE into Neo4j; HTTP API and MCP read Bolt.

- **Worker:** [ingest_worker/](ingest_worker/)
- **Sources (writers):** [sources/](sources/) (ti, vuln, lola, ds)
- **Storage (AppSec):** [storage/](storage/)
- **API / MCP:** [api/](api/), [mcp/](mcp/)
- **Neo4j client:** [neo4jclient/](neo4jclient/)
- **Contract:** [contract/ingestv1/](contract/ingestv1/)
- **Build:** `cd graph && go build ./ingest_worker/... ./api/...`
- **Deploy:** [../deploy/graph/compose.yml](../deploy/graph/compose.yml)
