---
name: pipeline unit tests master
overview: "T3 unit coverage for pipeline/ — waves W0–W8 on branch platform/pipeline-tests-full."
todos:
  - id: w0-harness
    content: "W0: harness + docs"
    status: completed
  - id: w1-nvd-router
    content: "W1: nvd + router"
    status: completed
  - id: w2-ti
    content: "W2: ti transforms"
    status: completed
  - id: w3-vuln
    content: "W3: vuln + enrich"
    status: completed
  - id: w4-bus
    content: "W4: appsec parse + dedup + consumer"
    status: completed
  - id: w5-connector
    content: "W5: connector nats"
    status: completed
  - id: w6-runtime
    content: "W6: config + components + pull loop"
    status: completed
  - id: w7-gap
    content: "W7: gap-fill (components, appsec/parse)"
    status: completed
  - id: w8-ci
    content: "W8: CI + sign-off"
    status: completed
isProject: false
---

# pipeline unit tests — master

Branch: `platform/pipeline-tests-full`

Detail: [pipeline_unit_tests_full_4e940824.plan.md](pipeline_unit_tests_full_4e940824.plan.md)

## Sign-off (2026-05-20)

`make test-pipeline-cover-strict` green — all logic packages 100% statements.

Gate: `make test-platform-p7` includes `test-pipeline-cover-strict`.

Harness: `test-pipeline-all`, `scripts/test/pipeline-cover.sh`, [pipeline-test-coverage.md](../../docs/development/pipeline-test-coverage.md).
