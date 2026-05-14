# Security

## Reporting a vulnerability

If you believe you have found a security vulnerability in this repository (for example in the HTTP API, MCP server, or a dependency choice), please report it responsibly:

- Prefer **GitHub Security Advisories** for this repository (private disclosure), if enabled for the project.
- Otherwise, open a **private** issue or contact the maintainers through a channel they publish on the repository or organization profile.

Please do **not** file public issues with exploit details until a fix is available.

## Scope

This project runs **Neo4j**, **Go services**, and optional **scrapers** with network access to third-party APIs. Harden deployments with secrets management, non-default passwords, network policies, and regular image updates—not only application patches.

## Supported versions

Security fixes are applied on a best-effort basis on the default development branch. There is no separate LTS branch unless documented explicitly in the future.
