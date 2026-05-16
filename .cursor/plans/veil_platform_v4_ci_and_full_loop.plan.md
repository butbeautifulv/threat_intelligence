---
name: Veil Platform v4 CI and Full Loop
overview: "После v3: закрепить platform gates в CI; опционально v4b — scrape в closed-loop smoke."
todos:
  - id: p4a-ci-platform
    content: "P4a: GHA workflow test-platform-p0 + closed-loop on main"
    status: completed
  - id: p4b-scrape-loop
    content: "P4b: optional scrape→pipeline slice in platform smoke (heavy)"
    status: pending
isProject: false
---

# Veil Platform v4 — CI gates + full loop extension

## Context

Platform v3 (P0–P3) merged on `main`. P3 pilot validates **act → learn → remember → decide** without scrape.

## Phases

| Phase | Branch | DoD |
|-------|--------|-----|
| P4a | `platform/p4a-ci-platform` | `.github/workflows/platform.yml`; PR → `test-platform-p0`; `main` push → `test-platform-closed-loop` |
| P4b | `platform/p4b-scrape-loop-smoke` | TBD: minimal harvest publish + graph hit (Docker, long) |

## Agent chain

Follow [veil-agent-documentation.mdc](../.cursor/rules/veil-agent-documentation.mdc) on merge.
