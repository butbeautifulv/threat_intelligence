---
name: Engage Phase 29 Refactor
overview: "Phase 29 (R143–R147): httpserver split, catalog pipeline, component interfaces, MCP bridge table."
todos:
  - id: p29-r143-router
    content: "R143: split router.go into register_* files (<800 LOC each)"
    status: pending
  - id: p29-r144-catalog
    content: "R144: catalog-engage runs extract+live+na-matrix+parity+args"
    status: pending
  - id: p29-r145-components
    content: "R145: APIComponents interfaces for Intel/CVE/CTF"
    status: pending
  - id: p29-r146-bridge
    content: "R146: intel_bridge name→handler map"
    status: pending
  - id: p29-r147-cleanup
    content: "R147: dead code removal, full test gates"
    status: pending
isProject: false
---

# Phase 29 — Engage refactor (R143–R147)

**Ветка:** `engage/phase-29-refactor`  
**Rebase:** `git fetch && git rebase origin/main` before push (may follow Phase 28 merge).

## R143 — Router split

Split [router.go](engage/serve/internal/transport/httpserver/router.go) (~881 LOC) into:

- `router_intel.go`, `router_ctf.go`, `router_vuln.go`, `router_tools.go`, `router_admin.go` (or similar)
- Keep `router.go` as wiring + `NewRouter` only
- **Target:** no file >800 LOC

## R144 — Catalog pipeline

[Makefile](Makefile) `catalog-engage`:

```makefile
catalog-engage:
	extract + generate-tools-live + generate-tools-na-matrix + check-catalog-parity + check-catalog-args
```

## R145 — Components

[components/api.go](engage/serve/internal/components/api.go): small interfaces `IntelProvider`, `CVEProvider`, `CTFProvider` for tests.

## R146 — MCP bridge

[intel_bridge.go](engage/serve/internal/transport/mcpserver/intel_bridge.go): `map[string]bridgeHandler` instead of giant switch where feasible.

## R147 — Gates

```bash
make test-engage
make test-engage-route-parity
make test-engage-parity
```

## DoD

- [ ] httpserver no file >800 LOC
- [ ] route-parity unchanged
- [ ] `make test-engage` green
