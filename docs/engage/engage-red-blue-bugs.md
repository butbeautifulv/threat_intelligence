# Engage red-vs-blue lab — bug log (field)

Use this file during the **field** phase of the install + lab track. One row (or subsection) per reproducible issue.

**Lab scope:** only [engage-red-blue-lab.md](engage-red-blue-lab.md) — authorized localhost / lab victim.

| Date | Distro (ID / version) | Profile | Symptom | Repro steps | Severity | Fixed in |
|------|------------------------|---------|---------|-------------|----------|----------|
| 2026-05-20 | Ubuntu (apt) | recommended | Several tools unavailable in base apt (`httpx`, `nuclei`, `subfinder`, `amass`, `feroxbuster`) caused partial install gap | `./scripts/ops/install-engage-host-tools.sh --plan --profile recommended` + `./scripts/engage/preflight-client-tools.sh --profile recommended --json` | Medium | 6cd739f + working tree (`--fallback` + sources registry) |
| 2026-05-20 | Ubuntu (apt) | recommended | Fallback emitted noisy warnings for tools without upstream method entries | `./scripts/ops/install-engage-host-tools.sh --plan --fallback --profile recommended` | Low | pending (populate source methods for additional tools incrementally) |
| 2026-05-20 | Ubuntu (apt) | recommended | Self-pentest harness passed; victim stayed healthy under aggressive local abuse flow | `ENGAGE_VICTIM_URL=http://127.0.0.1:8891 make test-engage-red-blue` | Info | e3147d3 |
| 2026-05-20 | Ubuntu (apt) | core47 | Core47 install reached 46/47; only `ghidra` remained manual-required due unavailable apt/kali index and heavyweight upstream package | `./scripts/ops/install-engage-host-tools.sh --yes --profile core47 --policy full-auto` + `./scripts/engage/preflight-client-tools.sh --profile core47 --json` | Medium | working tree (source-map/manual remediation) |
| 2026-05-20 | Ubuntu (apt) | core47 | Self-pentest harness passed after core47 installation flow (victim healthy under aggressive abuse scenarios) | `ENGAGE_VICTIM_URL=http://127.0.0.1:8891 make test-engage-red-blue` | Info | working tree |
| 2026-05-20 | Ubuntu (apt) | prod/aggressive self-pentest | MCP+HTTP self-pentest found 2 high findings (`ENG-CATALOG`, `GRAPH-OPEN`), 1 medium (`missing HSTS`); runtime guards for SSRF/traversal/command checks stayed effective | `TARGET_API_URL=http://127.0.0.1:8090 TARGET_ENGAGE_URL=http://127.0.0.1:8891 VEIL_PENTEST_PROFILE=prod VEIL_PENTEST_MODE=aggressive ./scripts/eval/pentest-veil-mcp.sh` + `make test-engage-red-blue` | High | c8f9b89 (see [engage-lab-pentest.md](engage-lab-pentest.md)) |

## Notes

- Attach `curl -v` redacted traces if useful; never paste production tokens.
- Link PRs with `engage/fix-pXX-<slug>` when resolved.
