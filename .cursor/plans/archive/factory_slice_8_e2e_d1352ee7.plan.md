---
name: Factory slice 8 E2E
overview: "УСТАРЕЛ / БРАКОВАН: ошибочно смешал E2E с graph-pack release v0.3.2. Не исполнять. Актуальный план — factory_slice_8_v2_e2e_refactor.plan.md."
todos:
  - id: e2e-smoke-script
    content: "Чеклист/scripts/smoke-scrape-e2e.sh: compose ×2, NATS lag, crawl_resource SQL, Cypher"
    status: cancelled
  - id: release-env-doc
    content: "docs/architecture/threatintel-runtime.md: release profile env (NVD_MAX_PAGES, SCRAPE_FORCE_REFETCH, …)"
    status: cancelled
  - id: export-pack-032
    content: export-graph-cypher + GRAPH_PACK_VERSION=v0.3.2 build; sha256 ≠ b4fd360a
    status: cancelled
  - id: gh-release-bootstrap
    content: gh release v0.3.2-graph-pack; docker/graph-bootstrap.sh DEFAULT_PACK_URL
    status: cancelled
  - id: api-mcp-smoke
    content: curl /health, /v1/categories; MCP smoke; graph-dedup --dry-run
    status: cancelled
  - id: veil-master-slice8
    content: Создать factory_slice_8_e2e_release.plan.md; veil_refactor срез 8 done, фаза E done
    status: cancelled
isProject: false
---

# Срез 8 v1 — БРАКОВАН (не исполнять)

**Статус:** superseded by [factory_slice_8_v2_e2e_refactor.plan.md](factory_slice_8_v2_e2e_refactor.plan.md)

**Почему бракован:**

- Смешал **Veil E (работающий пайплайн)** с **graph-pack release** (`gh release`, `DEFAULT_PACK_URL`, sha256 gate) — релиз **не делаем**.
- Не закрывал главную цель: **структура `ingest/` заиграла красками**, snake_case имён, **удаление legacy** после зелёного E2E.

Оставлен только как историческая запись. Все todos **cancelled**.
