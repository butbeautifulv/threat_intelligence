---
name: Engage Phase 12
overview: "Playbooks execution, pipeline engage.events consumer, veil-graph intelligence depth, attack chain execution with params, ops closure."
todos:
  - id: engage-r57-playbook-run
    content: "R57: POST /api/playbooks/{name}/run + MCP execution from bugbounty.yaml"
    status: completed
  - id: engage-r58-pipeline-consumer
    content: "R58: pkg/commit engage envelope + pipeline NATS consumer + compose/smoke"
    status: completed
  - id: engage-r59-graph-intel
    content: "R59: veil-graph depth for correlate/discover_attack_chains/ai_vulnerability_assessment + MCP"
    status: completed
  - id: engage-r60-pattern-depth
    content: "R60: Expand attack_patterns Params + pass params through CreateAttackChain execution"
    status: completed
  - id: engage-r61-ops-closure
    content: "R61: ENGAGE_PDF_ENGINE=wkhtml, CI metrics/webhook, Postgres audit recent/export"
    status: completed
  - id: engage-r62-catalog-ci
    content: "R62 (optional): Recategorize misc intelligence tools + expand CI tool matrix"
    status: completed
isProject: false
---

# Engage Phase 12 — complete

See implementation in repo; audit plan: [engage_phase_12_c80ea11e.plan.md](engage_phase_12_c80ea11e.plan.md).
