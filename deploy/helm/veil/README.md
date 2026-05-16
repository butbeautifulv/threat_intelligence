# Veil Helm chart (control plane)

Kubernetes workloads: graph `api`, `engage-api`, autoscaled `pipeline_worker` / `ingest_worker`, `scrape` CronJob.

```bash
helm template veil . -f values.yaml -f values-stage.yaml \
  --set global.imageTag=v0.4.5
```

Stateful services (Neo4j, NATS) default to Ansible+Compose data nodes — set `global.natsUrl` / `global.neo4jUri` accordingly.
