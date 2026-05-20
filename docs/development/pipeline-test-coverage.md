# pipeline unit test coverage (Veil)

Tracking for [pipeline unit tests master plan](../../.cursor/plans/pipeline_unit_tests_master.plan.md) and [pipeline unit tests full plan](../../.cursor/plans/pipeline_unit_tests_full_4e940824.plan.md).

## Gates

| Command | Scope |
|---------|--------|
| `make test-pipeline-all` | `pipeline/pkg`, `pipeline/connector`, `pipeline/ned` (go.work modules) |
| `make test-pipeline-cover` | Same + `scripts/test/pipeline-cover.sh` (T0 + T2 ≥70%) |
| `make test-pipeline-cover-strict` | T3: 100% statements on logic packages |
| `make test-platform-p7` | Includes `test-pipeline-cover-strict` |

## Tier definitions

| Tier | Criterion |
|------|-----------|
| T0 | Every logic package has `*_test.go`; `cmd/*/main` is build-only via `make test-pipeline` |
| T1 | Exported transforms/handlers have table or round-trip tests |
| T3 | Logic packages **100%** (`make test-pipeline-cover-strict`) |

## Sign-off (2026-05-20)

Branch `platform/pipeline-tests-full`: `make test-pipeline-cover-strict` green — all logic packages at 100% statements. P7 gate uses strict pipeline coverage.

## Baseline (pre-T3, main)

| Package | ~coverage |
|---------|-----------|
| `pipeline/pkg/nvd/map` | 100% |
| `pipeline/pkg/nvd/parse` | 75% |
| `pipeline/connector/nats` | 3% |
| `pipeline/ned/internal/sources/ti` | 10% |
| `pipeline/ned/internal/sources/vuln` | 24% |
| `pipeline/ned/internal/transform` | 0% |
| `pipeline/ned/internal/config`, `components` | 0% |

## Waves

| Wave | phase_id | Focus |
|------|----------|--------|
| W0 | `pipeline-tests-w0` | harness, this doc |
| W1 | `pipeline-tests-w1` | nvd + router |
| W2 | `pipeline-tests-w2-ti` | ti transforms |
| W3 | `pipeline-tests-w3-vuln` | vuln + enrich |
| W4 | `pipeline-tests-w4-bus` | appsec parse, dedup, consumer |
| W5 | `pipeline-tests-w5-connector` | connector nats |
| W6 | `pipeline-tests-w6-runtime` | config, components, RunPullLoop |
| W7 | `pipeline-cover-v-*` | gap-fill if strict FAIL |
| W8 | `pipeline-tests-w8-ci` | P7 gate |
