# Engage — agentic AI threats and Veil mitigations

Maps **OWASP GenAI Agentic / Red Team landscapes** and **DAF MLSO** runtime practices to concrete controls in this repository. For secured infrastructure running active countermeasures.

## Threat model (agent + MCP + tools)

| Threat | Example | Veil mitigation | Control ID |
|--------|---------|-----------------|------------|
| **Prompt injection → tool abuse** | Agent asked to run `curl` exfiltration | Catalog-only `POST /api/tools/{name}`; RBAC `PermEngageToolRun` | VEIL-ENG-001 |
| **Privilege escalation via shell** | `/api/command` with `id; curl …` | Deny raw in prod; metachar block; use tools API | VEIL-ENG-005 |
| **SSRF / metadata theft** | Target `http://169.254.169.254/` | `ENGAGE_TARGET_GUARD=block` (prod default) | VEIL-ENG-007 |
| **Runner host compromise** | Tool escapes to API host | `ENGAGE_RUNNER_MODE=docker` + isolated `engage-runner` | VEIL-ENG-002 |
| **Unauthenticated MCP** | Open HTTP MCP on :8892 | `ENGAGE_MCP_HTTP_AUTH_STRICT=1`, Keycloak | VEIL-ENG-003 |
| **Unbounded agent delegation** | Spawn sub-agents with full catalog | Single MCP server; no dynamic agent creation (DAF MLSO T-PROD-RUN-ML-1-2) | design |
| **Data exfil via graph** | MCP reads entire TI graph | Separate **veil-mcp** (read) vs **veil-engage** (exec); RBAC roles | [mcp-agents.md](../agents/mcp-agents.md) |
| **Supply-chain tool binary** | Tampered `nmap` on PATH | Runner image rebuild; catalog `binary:` pinned in YAML | VEIL-ENG-002 |

## DAF MLSO practices adopted

| Practice ID | Requirement (summary) | Veil implementation |
|-------------|----------------------|---------------------|
| T-PROD-RUN-ML-1-1 | AI/agents not privileged; MCP/API access least privilege | Docker runner, RBAC, optional target guard |
| T-PROD-RUN-ML-1-2 | Agents cannot spawn agents | No agent-delegation API |
| T-PROD-RUN-ML-2-2 | No LLM output to shell/exec/eval | No `exec`/`eval` endpoints; command allowlist |
| T-PROD-RUN-ML-2-4 | Agent API access minimal | Tool catalog + auth per request |

Practices marked **human-in-the-loop** (T-PROD-RUN-ML-1-3) are **operational** — use job queues + operator approval in your SOAR, not enforced in code today.

## OWASP agentic lifecycle alignment

| Phase | Veil touchpoint |
|-------|-----------------|
| Govern | [veil-controls.yaml](../deploy/security/veil-controls.yaml), secure profiles |
| Develop & experiment | Dev compose; `ENGAGE_TARGET_GUARD=off` locally if needed |
| Test & evaluate | `make test-engage-hardening` (safe self-test, not full red-team) |
| Deploy | `secure-engage.env`, nginx TLS overlay |
| Operate / Monitor | Audit log, `engage.events` → graph ingest |

Full red-team of upstream LLM clients is **out of repo scope** — use OWASP landscape tools against your agent host; Veil hardens the **tool execution plane**.

## Configuration (secured prod)

```bash
# deploy/profiles/secure-engage.env (excerpt)
ENGAGE_ENV=prod
ENGAGE_DENY_RAW_COMMAND=1
ENGAGE_RUNNER_MODE=docker
ENGAGE_TARGET_GUARD=block
ENGAGE_MCP_HTTP_AUTH_STRICT=1
VEIL_REQUIRE_AUTH=1
```

## Lab exception

Internal bug-bounty or lab scans against RFC1918 targets require explicit override:

```bash
ENGAGE_TARGET_GUARD=off   # only on isolated lab engage-api, never on shared prod
```

Document exceptions in your change ticket; critic should reject prod profile changes that disable the guard.

## Related

- [engage-hardening.md](engage-hardening.md)
- [external-security-frameworks.md](../external/external-security-frameworks.md)
- [mcp-agents.md](../agents/mcp-agents.md)
