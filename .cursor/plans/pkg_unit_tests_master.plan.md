---
name: pkg unit tests master
overview: Full pkg unit test coverage — W0 harness through W6 parallel waves, W7 CI optional.
status: active
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
| MERGE | `main` | done | `make test-pkg-all` green |

## DoD

`make test-pkg-all` green; T0 no NOTEST packages; wire/playbook/auth targets per wave.
