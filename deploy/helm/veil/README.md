# Veil Helm chart (control plane)

Kubernetes workloads: graph `api`, `engage-api`, MCP HTTP (when enabled), autoscaled `pipeline_worker` / `ingest_worker`, `scrape` CronJob.

```bash
helm template veil . -f values.yaml -f values-stage.yaml \
  --set global.imageTag=v0.4.5
```

Stateful services (Neo4j, NATS) default to Ansible+Compose data nodes — set `global.natsUrl` / `global.neo4jUri` accordingly. Prod Neo4j: Enterprise **3-core** cluster URI (`neo4j+routing://…`) per [docs/platform-unified-access.md](../../../docs/platform-unified-access.md).

## Stateless scale (4 / 8 / 16)

Align Helm replicas with Compose `VEIL_*_SCALE` (P12c). Allowed values for operator presets: **4**, **8**, **16**.

**Stage (minimal):**

```bash
helm upgrade --install veil . \
  -f values.yaml -f values-stage.yaml \
  --set graph.api.replicas=1 \
  --set engage.api.replicas=1 \
  --set pipeline.worker.replicas=1 \
  --set ingest.worker.replicas=1
```

**Medium (4 replicas — API + engage pools):**

```bash
helm upgrade --install veil . \
  -f values.yaml -f values-prod.yaml \
  --set graph.api.replicas=4 \
  --set engage.api.replicas=4 \
  --set pipeline.worker.replicas=4 \
  --set ingest.worker.replicas=4
```

**High (8 replicas):**

```bash
helm upgrade --install veil . \
  -f values.yaml -f values-prod.yaml \
  --set graph.api.replicas=8 \
  --set engage.api.replicas=8 \
  --set pipeline.worker.replicas=8 \
  --set ingest.worker.replicas=8
```

**Peak (16 replicas, HPA max where enabled):**

```bash
helm upgrade --install veil . \
  -f values.yaml -f values-prod.yaml \
  --set graph.api.replicas=16 \
  --set engage.api.replicas=16 \
  --set pipeline.worker.replicas=16 \
  --set pipeline.worker.hpa.maxReplicas=16 \
  --set ingest.worker.replicas=16 \
  --set ingest.worker.hpa.maxReplicas=16
```

MCP HTTP deployments and unified Ingress are wired in P12c/f; path routing contract: [docs/platform-unified-access.md](../../../docs/platform-unified-access.md).
