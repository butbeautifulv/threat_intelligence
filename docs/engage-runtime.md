# Engage runtime

**Engage** is the fourth Veil layer: authorized offensive security tooling, workflows, and reports. It is separate from the graph read path.

## Threat model

| Risk | Mitigation |
|------|------------|
| Unauthenticated tool execution | Keycloak JWT, `VEIL_REQUIRE_AUTH`, nginx TLS |
| Collateral damage / illegal use | Lab/VPN only; RBAC roles `veil-engage-runner`, `veil-engage-admin` |
| Graph exfiltration via engage | Engage uses veil-api with service account (`veil-reader`), not direct Neo4j |

## Ports (dev)

| Service | Port |
|---------|------|
| engage-api | 8890 |
| engage-mcp HTTP (optional) | 8892 |
| nginx (secure overlay) | 8443 |

## Environment

| Variable | Default | Role |
|----------|---------|------|
| `ENGAGE_API_LISTEN` | `:8890` | API bind |
| `ENGAGE_CATALOG_PATH` | `catalog/tools.yaml` | Tool registry |
| `ENGAGE_RUNNER_WORKDIR` | `/tmp/engage` | Subprocess cwd |
| `ENGAGE_RUNNER_MODE` | `local` | `local` or `docker` (exec in `engage-runner` container) |
| `ENGAGE_RUNNER_CONTAINER` | — | Container name when `ENGAGE_RUNNER_MODE=docker` |
| `ENGAGE_VEIL_API_URL` | `http://localhost:8090` | Graph read API |
| `ENGAGE_VEIL_CLIENT_ID` / `SECRET` / `TOKEN_URL` | — | OAuth2 client credentials |
| `AUTH_ENABLED` | `0` | Keycloak JWT |

## Compose

```bash
docker compose -f deploy/engage/compose.yml up -d --build engage-api
curl -sS http://localhost:8890/health | jq .
```

Secure overlay: `deploy/engage/compose.secure.yml` + `deploy/profiles/secure-engage.env`.

## MCP

- **stdio:** `veil-engage` — [examples/mcp/engage.stdio.json.example](../examples/mcp/engage.stdio.json.example)
- **HTTP (optional):** `ENGAGE_MCP_HTTP_ENABLED=1` on `:8892`, or [engage.http.json.example](../examples/mcp/engage.http.json.example)

Engage is a greenfield Go rewrite of the MIT tool server in `.external/` (attribution in [engage/NOTICE.hexstrike](../engage/NOTICE.hexstrike)).

## Related

- [engage-legacy-parity.md](engage-legacy-parity.md)
- [deploy-secure.md](deploy-secure.md) (graph)
- [auth-keycloak.md](auth-keycloak.md)
