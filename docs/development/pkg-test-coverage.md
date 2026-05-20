# pkg unit test coverage (Veil)

Baseline and wave tracking for [pkg unit tests master plan](../.cursor/plans/pkg_unit_tests_master.plan.md) and [T3 master plan](../.cursor/plans/pkg_unit_tests_t3_master.plan.md).

## Gates

| Command | Scope |
|---------|--------|
| `make test-pkg-all` | All `pkg/` Go packages (root module `./...` + `pkg/engage`, `api`, `auth`, `mcp`, `exec`) |
| `make test-pkg-cover` | Same modules + `scripts/test/pkg-cover.sh` (T0 NOTEST + T2 floors) |
| `make test-pkg-cover-strict` | T3: 100% statement coverage on logic packages (`PKG_COVER_STRICT=1`) |
| `make test-platform-p7` | Includes `test-pkg-cover-strict` + layer bus slices |

## Tier definitions

| Tier | Criterion |
|------|-----------|
| T0 | Every package has `*_test.go` or types-only `entity_test` |
| T1 | Exported pure helpers have table or round-trip tests |
| T2 | Logic packages ≥70% statement coverage (`go test -cover`) |
| T3 | Logic packages **100%** (`make test-pkg-cover-strict`) |

## Waves complete (2026-05-20)

Merged to `main`: W0–W7 (T2) and W8–W14 (T3). CI gate for platform P7: `make test-pkg-cover-strict`.

| Wave | Branch | Focus |
|------|--------|--------|
| W0 | `platform/pkg-tests-w0-harness` | `test-pkg-all`, this doc |
| W1 | `platform/pkg-tests-w1-entity` | vuln, lola, playbook/domain, engage/job |
| W2 | `platform/pkg-tests-w2-wire` | harvest, commit idempotency |
| W3 | `platform/pkg-tests-w3-playbook` | procedure, index, framework, cataloglink |
| W4 | `platform/pkg-tests-w4-engage` | report, contract, events publisher |
| W5 | `platform/pkg-tests-w5-auth-mcp` | auth enforcer, mcp rpc/tools/framed |
| W6 | `platform/pkg-tests-w6-misc` | decision, report, natsjet |
| W7 | `platform/pkg-tests-w7-ci` | `test-pkg-cover`, `scripts/test/pkg-cover.sh`, manifest |

### T2 floors (`scripts/test/pkg-cover.sh`)

| Default | Exceptions |
|---------|------------|
| ≥70% statements | `pkg/mcp` ≥60%; `pkg/exec` ≥50%; `pkg/auth/keycloak` ≥45% |
| T0 allowlist | `*/domain` data-only packages, `engage/contract` — `[no statements]` OK |

### T3 strict (`scripts/test/pkg-cover-strict.sh`)

| Default | Exceptions |
|---------|------------|
| **100%** statements | none (all lowered T2 floors removed) |
| T0 allowlist | same as T2 |

## T3 waves (W8–W14) — done

| Wave | Branch | Focus |
|------|--------|--------|
| W8 | `platform/pkg-tests-w8-strict-gate` | `test-pkg-cover-strict`, T3 docs |
| W9 | `platform/pkg-tests-w9-quick` | commit, domain, ti, engage/events, api, playbook/index |
| W10 | `platform/pkg-tests-w10-decision-report` | decision, report |
| W11 | `platform/pkg-tests-w11-nats-playbook` | natsjet, playbook |
| W12 | `platform/pkg-tests-w12-mcp-auth` | mcp, auth/keycloak |
| W13 | `platform/pkg-tests-w13-exec` | exec sandbox |
| W14 | `platform/pkg-tests-w14-ci` | P7 → strict, master plan sign-off |

T3 exclude list: empty (no packages require Docker/live Keycloak in unit tests).

Regenerate coverage snapshot:

```bash
make test-pkg-cover-strict
# or:
cd pkg && env -u GOWORK go test ./... -coverprofile=/tmp/pkg-cover.out
go tool cover -func=/tmp/pkg-cover.out | tail -1
```
