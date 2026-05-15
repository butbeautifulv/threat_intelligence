# ingest_worker

NATS consumer for `ingest.>` → Neo4j MERGE.

- **Contract:** [../contract/ingestv1](../contract/ingestv1/)
- **Domain writers:** [../sources/](../sources/) via [../workeringest/](../workeringest/)
- **AppSec storage:** [../storage/](../storage/)
- **Deploy:** [../../deploy/graph/compose.yml](../../deploy/graph/compose.yml)

```bash
cd graph && go build -o bin/ingest_worker ./ingest_worker/cmd
```
