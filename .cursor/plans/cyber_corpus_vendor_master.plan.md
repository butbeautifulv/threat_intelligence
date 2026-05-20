---
name: Cyber corpus vendor master
overview: "Вендоринг Anthropic Cybersecurity Skills: mappings SOT в pkg/playbook/corpus, skills в corpus/, framework API. Статус выполнения на main."
todos:
  - id: v0-mappings-sot
    content: "V0: pkg/playbook/corpus/mappings + ATTRIBUTION + docs/cyber-domain-model.md"
    status: completed
  - id: v1-skills-mirror
    content: "V1: corpus/anthropic-cybersecurity-skills/skills + corpus-import.sh"
    status: completed
  - id: v2-path-switch
    content: "V2: generate-cyber-skills-index + pkg/playbook/index corpus_path"
    status: completed
  - id: v3-framework-pkg
    content: "V3: pkg/playbook/framework Navigator layer"
    status: completed
  - id: v4-framework-api
    content: "V4: veil-api/MCP framework read endpoints"
    status: completed
  - id: master-plan-file
    content: "Master plan status table"
    status: completed
isProject: true
---

# Cyber corpus vendor — master plan

Prerequisite: [anthropic_skills_knowledge plan](anthropic_skills_knowledge_c05c1c9c.plan.md) MCP read path (done).

## Status

| Phase | Branch (recommended) | Status | Notes |
|-------|----------------------|--------|-------|
| V0 mappings SOT | `feat/playbook-v0-mappings-sot` | **done** | `pkg/playbook/corpus/mappings/` |
| V1 skills mirror | `feat/playbook-v1-corpus-skills` | **done** | 754 SKILL.md under `corpus/.../skills/` |
| V2 path switch | `feat/playbook-v2-corpus-paths` | **done** | `corpus_path`, `make corpus-import` |
| V3 framework pkg | `feat/pkg-playbook-framework` | **done** | `pkg/playbook/framework` |
| V4 framework API | `feat/knowledge-framework-read` | **done** | HTTP + MCP `playbook_framework` |
| V5 graph (optional) | `feat/knowledge-framework-graph` | pending | playbook_seed / CSF edges |
| V6 engage (optional) | `engage/phase-corpus-domain-hints` | partial | playbook_hints exist |

## Layout (split)

| Path | Role |
|------|------|
| [pkg/playbook/corpus/mappings/](../../pkg/playbook/corpus/mappings/) | Veil SOT — MITRE/NIST/OWASP |
| [corpus/anthropic-cybersecurity-skills/skills/](../../corpus/anthropic-cybersecurity-skills/skills/) | Procedure mirror |
| [docs/skills-index/cyber-skills.json](../../docs/skills-index/cyber-skills.json) | Generated index |

## Verify

```bash
make check-corpus-mappings
make check-skills-index
cd pkg && go test ./playbook/...
make test-knowledge-serve
```

Detail: [cyber_corpus_vendor_b52b05d3.plan.md](cyber_corpus_vendor_b52b05d3.plan.md).
