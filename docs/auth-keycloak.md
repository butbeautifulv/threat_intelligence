# Keycloak authentication and RBAC (graph read)

Optional JWT authentication for the **HTTP API** (`:8090`) and **MCP stdio** (`veil-mcp`). **Disabled by default** (`AUTH_ENABLED=0`).

Active Directory is integrated **in Keycloak** (User Federation → LDAP). Veil only validates OIDC access tokens and realm/client roles in the JWT.

## Quick reference

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTH_ENABLED` | `0` | Enable JWT validation |
| `RBAC_ENABLED` | `0` | Require Keycloak roles (only if auth on) |
| `KEYCLOAK_ISSUER` | — | Realm issuer URL, e.g. `https://keycloak.example/realms/veil` |
| `KEYCLOAK_AUDIENCE` | — | API client ID (`veil-api`); also checks `azp` |
| `KEYCLOAK_CLIENT_ID` | `veil-api` | Client for `resource_access.<client>.roles` |
| `RBAC_ROLE_READER` | `veil-reader` | Read graph via API/MCP |
| `RBAC_ROLE_ADMIN` | `veil-admin` | Same as reader today (admin ops later) |
| `MCP_ACCESS_TOKEN` | — | JWT for MCP stdio when `AUTH_ENABLED=1` |

## Keycloak setup

1. Create realm **`veil`** (or use existing).
2. Create client **`veil-api`**:
   - Access Type: `confidential` (server) or `public` + PKCE (user agents).
   - Valid redirect URIs as needed for your IdP flows.
3. Create realm roles: **`veil-reader`**, **`veil-admin`** (or match `RBAC_ROLE_*` env).
4. Assign roles to users or groups (AD groups can map via LDAP federation + mapper).
5. Ensure access token includes roles:
   - Realm roles → `realm_access.roles`
   - Client roles → `resource_access.veil-api.roles` (use client role mapper).

### AD (LDAP) federation

In Keycloak Admin: **User Federation** → add LDAP/Active Directory provider → sync users/groups → map AD groups to realm roles (e.g. `Group Mapper` or `Role mapper`).

Veil does not talk to AD directly.

## RBAC matrix

| Role | Permission |
|------|------------|
| `veil-reader` | `graph:read` — all `/v1/*` endpoints and MCP tools |
| `veil-admin` | Same as reader (reserved for future admin APIs) |

With `RBAC_ENABLED=0` and `AUTH_ENABLED=1`, any valid JWT is accepted.

## HTTP API

```bash
# Dev token (password grant — disable in production)
TOKEN=$(curl -sS -X POST "$KEYCLOAK_ISSUER/protocol/openid-connect/token" \
  -d "client_id=veil-api" \
  -d "client_secret=YOUR_SECRET" \
  -d "grant_type=password" \
  -d "username=user" \
  -d "password=pass" | jq -r .access_token)

curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8090/v1/categories
```

`GET /health` stays **unauthenticated** (compose liveness).

Responses: **401** invalid/missing token, **403** missing role when RBAC on.

## MCP stdio

1. Obtain an access token (same Keycloak client / user as API).
2. Set `MCP_ACCESS_TOKEN` in the MCP server env (see [examples/mcp/mcp.auth.json.example](../examples/mcp/mcp.auth.json.example)).
3. Set `AUTH_ENABLED=1` (and `RBAC_ENABLED=1` if needed).

`initialize` and `tools/list` work without a token; **`tools/call`** requires a valid JWT.

## Compose example (commented)

In [deploy/knowledge/compose.yml](../deploy/knowledge/compose.yml) under `api`:

```yaml
# AUTH_ENABLED: "1"
# RBAC_ENABLED: "1"
# KEYCLOAK_ISSUER: https://keycloak.example/realms/veil
# KEYCLOAK_AUDIENCE: veil-api
```

## MCP Streamable HTTP

When `MCP_HTTP_ENABLED=1`, the MCP binary listens on `MCP_HTTP_LISTEN` (default `:8091`) and uses the **same Bearer JWT middleware** as the REST API on `POST /mcp`. Stdio continues to use `MCP_ACCESS_TOKEN` for `tools/call`.

See [mcp-agents.md](mcp-agents.md) for curl examples and client configuration.

## Related

- [mcp-agents.md](mcp-agents.md) — MCP client configuration
- [threatintel-runtime.md](threatintel-runtime.md) — ports and services
- [SECURITY.md](../SECURITY.md) — reporting vulnerabilities
