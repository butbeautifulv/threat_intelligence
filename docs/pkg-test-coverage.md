# pkg unit test coverage (Veil)

Baseline and wave tracking for [pkg unit tests master plan](../.cursor/plans/pkg_unit_tests_master.plan.md).

## Gates

| Command | Scope |
|---------|--------|
| `make test-pkg-all` | All `pkg/` Go packages (root module `./...` + `pkg/engage`, `api`, `auth`, `mcp`, `exec`) |
| `make test-platform-p7` | Includes `test-pkg-all` + layer bus slices |

## Tier definitions

| Tier | Criterion |
|------|-----------|
| T0 | Every package has `*_test.go` or types-only `entity_test` |
| T1 | Exported pure helpers have table or round-trip tests |
| T2 | Logic packages ≥70% statement coverage (`go test -cover`) |

## Baseline gaps (2026-05-20)

| Package | Tests | Cover (approx) | Wave |
|---------|-------|----------------|------|
| `pkg/vuln/domain` | none | — | W1 |
| `pkg/lola/domain` | none | — | W1 |
| `pkg/playbook/domain` | none | — | W1 |
| `engage/domain/job` | none | — | W1 |
| `pkg/commit` | partial | 39% | W2 |
| `pkg/harvest` | partial | 64% | W2 |
| `pkg/playbook/procedure` | partial | 5% | W3 |
| `pkg/playbook/index` | partial | 23% | W3 |
| `pkg/playbook/framework` | partial | 30% | W3 |
| `pkg/playbook/cataloglink` | partial | 0% | W3 |
| `engage/domain/report` | const only | 0% | W4 |
| `engage/events` | JSON only | publisher 0% | W4 |
| `pkg/auth` | partial | 52% | W5 |
| `pkg/mcp` | partial | 32% | W5 |
| `pkg/decision` | partial | 44% | W6 |
| `pkg/report` | partial | 68% | W6 |
| `pkg/natsjet` | partial | 61% | W6 |

Regenerate coverage snapshot:

```bash
cd pkg && env -u GOWORK go test ./... -coverprofile=/tmp/pkg-cover.out
go tool cover -func=/tmp/pkg-cover.out | tail -1
```
