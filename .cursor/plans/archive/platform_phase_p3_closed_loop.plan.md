---
name: Platform P3 Closed Loop Pilot
overview: Document and smoke-test act→learn→remember→decide for web-host target class on veil stack.
todos:
  - id: p3-docs
    content: docs/architecture/platform-closed-loop-pilot.md
    status: completed
  - id: p3-smoke
    content: scripts/test/smoke-platform-closed-loop.sh + make target
    status: completed
isProject: false
---

# P3 — closed-loop pilot

## DoD

- [x] Target class **web host** documented with NATS/HTTP flow
- [x] `make test-platform-closed-loop` — Docker smoke (SKIP without daemon)
- [x] Uses engage `target-graph` (P2) after tool run + events bridge

## Branch

`platform/p3-closed-loop`
