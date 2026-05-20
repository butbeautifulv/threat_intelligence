---
name: pkg domain umbrella master
overview: Wave 1 — meta-layer pkg/domain (contour, source, ref, taxonomy, aliases). No wire/schema changes.
status: done
baseline_date: "2026-05-20"
parent_plan: .cursor/plans/pkg_domain_umbrella_113018bf.plan.md
---

# pkg/domain umbrella — master plan

## Status table

| Phase | Branch | Agent | Status | Merge SHA | DoD |
|-------|--------|-------|--------|-----------|-----|
| P0 | `platform/domain-p0-model-doc` | subagent P0 | done | `24fcca94` | docs/architecture/pkg-domain-model.md |
| P1–P6 | `platform/domain-umbrella` | subagent P1-P6 | done | `f6bdfd49` | pkg/domain + gates |
| MERGE | `main` | orchestrator | done | `f6bdfd49` | P0 fast-forward + P1–P6 merge; `make test-pkg-domain` green |

## Constraints

- No cross-layer Go imports; no harvest/commit JSON field renames.
- Do not edit engage catalog, `.external/`, unrelated playbook WIP on main.

## Critic gate

Each phase: diff &lt; ~150 LOC Go; `make test-pkg-domain` green (verified on merge).
