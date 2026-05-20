---
name: Engage Phase 30 Decommission
overview: "Phase 30 (R148–R150): veil-engage-only runbook, .external deprecation, audit sign-off."
todos:
  - id: p30-r148-runbook
    content: "R148: mcp-agents.md dual-MCP → veil-engage migration"
    status: completed
  - id: p30-r149-external
    content: "R149: external-hexstrike.md deprecation checklist"
    status: completed
  - id: p30-r150-signoff
    content: "R150: engage-audit-report + master plan sign-off"
    status: completed
isProject: false
---

# Phase 30 — Decommission HexStrike reference (R148–R150)

**Ветка:** `engage/phase-30-decommission`  
**Только docs** — параллельно с 28/29 без конфликтов кода.

## R148 — MCP runbook

[docs/agents/mcp-agents.md](docs/agents/mcp-agents.md): steps to disable Flask HexStrike MCP, use `veil-engage` only, env vars, Cursor config.

## R149 — External deprecation

[docs/external/external-hexstrike.md](docs/external/external-hexstrike.md): when `.external/` optional; what still needs legacy (parity extract only).

## R150 — Sign-off

- Update [docs/engage/engage-audit-report.md](docs/engage/engage-audit-report.md) migration sign-off section.
- Master plan v2 frontmatter todos p24–p30 status.

## DoD

- [x] Team can operate without `:8888` Flask per runbook
- [x] No code changes required (docs-only PR)

## Verify

Docs review only; `make test-engage` unchanged (sanity if any script touched).
