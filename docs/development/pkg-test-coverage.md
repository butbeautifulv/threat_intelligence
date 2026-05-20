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

## Wave 1 complete (2026-05-20)

Merged to `main`: W0 harness, W1–W6 parallel test waves. Gate: `make test-pkg-all` green.

| Wave | Branch | Focus |
|------|--------|--------|
| W0 | `platform/pkg-tests-w0-harness` | `test-pkg-all`, this doc |
| W1 | `platform/pkg-tests-w1-entity` | vuln, lola, playbook/domain, engage/job |
| W2 | `platform/pkg-tests-w2-wire` | harvest, commit idempotency |
| W3 | `platform/pkg-tests-w3-playbook` | procedure, index, framework, cataloglink |
| W4 | `platform/pkg-tests-w4-engage` | report, contract, events publisher |
| W5 | `platform/pkg-tests-w5-auth-mcp` | auth enforcer, mcp rpc/tools/framed |
| W6 | `platform/pkg-tests-w6-misc` | decision, report, natsjet |

T0 (every package has tests): satisfied. T2 (≥70% logic packages): re-check with `go test -cover` when changing code.

Regenerate coverage snapshot:

```bash
cd pkg && env -u GOWORK go test ./... -coverprofile=/tmp/pkg-cover.out
go tool cover -func=/tmp/pkg-cover.out | tail -1
```
