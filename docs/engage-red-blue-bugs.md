# Engage red-vs-blue lab — bug log (field)

Use this file during the **field** phase of the install + lab track. One row (or subsection) per reproducible issue.

**Lab scope:** only [engage-red-blue-lab.md](engage-red-blue-lab.md) — authorized localhost / lab victim.

| Date | Distro (ID / version) | Profile | Symptom | Repro steps | Severity | Fixed in |
|------|------------------------|---------|---------|-------------|----------|----------|
| 2026-05-20 | Ubuntu (apt) | recommended | Several tools unavailable in base apt (`httpx`, `nuclei`, `subfinder`, `amass`, `feroxbuster`) caused partial install gap | `./scripts/ops/install-engage-host-tools.sh --plan --profile recommended` + `./scripts/engage/preflight-client-tools.sh --profile recommended --json` | Medium | 6cd739f + working tree (`--fallback` + sources registry) |
| 2026-05-20 | Ubuntu (apt) | recommended | Fallback emitted noisy warnings for tools without upstream method entries | `./scripts/ops/install-engage-host-tools.sh --plan --fallback --profile recommended` | Low | pending (populate source methods for additional tools incrementally) |

## Notes

- Attach `curl -v` redacted traces if useful; never paste production tokens.
- Link PRs with `engage/fix-pXX-<slug>` when resolved.
