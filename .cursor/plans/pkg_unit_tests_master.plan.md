---
name: pkg unit tests master
overview: Full pkg unit test coverage — W0–W7 (T2 floors), T3 W8–W14 (100% strict gate on logic packages).
status: done
parent_plan: .cursor/plans/pkg_unit_tests_full_165670eb.plan.md
---

# pkg unit tests — master plan

## Status

| Wave | Branch | Status | Merge SHA |
|------|--------|--------|-----------|
| W0 | `platform/pkg-tests-w0-harness` | done | `f90b390b` |
| W1 | `platform/pkg-tests-w1-entity` | done | `72d2e238` |
| W2 | `platform/pkg-tests-w2-wire` | done | (merge) |
| W3 | `platform/pkg-tests-w3-playbook` | done | `eb2e9fdb` |
| W4 | `platform/pkg-tests-w4-engage` | done | `78df5a94` |
| W5 | `platform/pkg-tests-w5-auth-mcp` | done | (merge) |
| W6 | `platform/pkg-tests-w6-misc` | done | `3928765f` |
| W7 | `platform/pkg-tests-w7-ci` | done | `make test-pkg-cover` |
| W8–W14 | T3 waves (see `pkg_unit_tests_t3_master.plan.md`) | done | `make test-pkg-cover-strict` |
| MERGE | `main` | done | `make test-pkg-cover-strict` green |
| VERIFY | `main` | done | 2026-05-20 — all m0–m6 pass; see `pkg_unit_tests_verify_100.plan.md` |

## DoD

`make test-pkg-cover` green (T2); `make test-pkg-cover-strict` green (T3); T0 no NOTEST packages. `test-platform-p7` uses strict gate.

W0–W7 detail: archived in `.cursor/plans/pkg_unit_tests_full_165670eb.plan.md` (reference only).
