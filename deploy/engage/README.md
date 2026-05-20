# Engage deployment

Compose stacks for the Engage offensive tooling layer.

Default runtime is host-path `client-native` (`deploy/engage/compose.yml`): run `engage-api` / `engage-mcp` / `engage-worker` without a toolbox runner container. Use [`compose.runner.yml`](compose.runner.yml) only for legacy lab/CI docker-exec scenarios.

## Runner images (legacy lab / CI)

| Image | Dockerfile target | Profile | Approx. size / RAM |
|-------|-------------------|---------|-------------------|
| `engage-runner` | `engage-runner` | `runner` | ~2–3 GB disk; 512 MB–1 GB RAM typical |
| `engage-runner-full` | `engage-runner-full` | `runner-full` | ~8–12 GB disk; **4–8 GB RAM** recommended (Ghidra, Metasploit, angr) |

Tier-1 CLI: [docker/runner.Dockerfile](docker/runner.Dockerfile) target `engage-runner`.

**P9g heavy stack** (Burp JAR, Ghidra, hashcat, john, gdb, Metasploit, angr, radare2, volatility3, wpscan): same Dockerfile, target `engage-runner-full`. Headless wrappers live in [docker/wrappers/](docker/wrappers/).

```bash
# Slim runner (legacy lab profile)
docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.runner.yml --profile runner up -d --build engage-runner

# Full port heavy stack
docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.runner.yml --profile runner-full up -d --build engage-runner-full
export ENGAGE_RUNNER_CONTAINER=engage-runner-full
export ENGAGE_RUNNER_IMAGE=engage-runner-full
export ENGAGE_RUNNER_PROFILE=full
./scripts/engage/list-runner-binaries.sh
```

Lab overlay with docker exec: [compose.runner.yml](compose.runner.yml).

## `ENGAGE_RUNNER_PROFILE=full`

Use the full runner when catalog tools need the P9g heavy stack (Burp, Ghidra, hashcat, Metasploit, angr, etc.):

```bash
export ENGAGE_RUNNER_PROFILE=full
export ENGAGE_RUNNER_IMAGE=engage-runner-full
export ENGAGE_RUNNER_CONTAINER=engage-runner-full
```

Compose profile `runner-full` builds the same image as Dockerfile target `engage-runner-full`. Local verification (skips if Docker is unavailable):

```bash
make test-engage-runner-full-smoke
# or: ./scripts/test/smoke-engage-runner-full.sh
```

**P11a executable matrix in runner** (`scripts/test/engage-executable-matrix-runner.sh`): builds `engage-runner-full`, runs `go run ./cmd/executable-matrix` inside the image with `ENGAGE_MATRIX_IN_RUNNER=1` (skips host stub PATH layer; uses image `/usr/local/bin`). Local: `make test-engage-executable-matrix-runner`. CI: set repository variable or workflow env `ENGAGE_MATRIX_IN_RUNNER=1` to enable the `engage-runner-executable-matrix` job in [`.github/workflows/engage.yml`](../../.github/workflows/engage.yml).

**P10d cloud security smoke** (`scripts/test/smoke-engage-runner-full.sh`): after the P9g heavy-stack checks, verifies cloud subprocess tools on the runner image:

| Tool | Check |
|------|--------|
| `prowler` | `--version` / `--help`, or engage-stub JSON |
| `scout` / `scout-suite` | ScoutSuite `--help` via wrapper |
| `pacu` | `/opt/pacu` CLI or stub |
| `terrascan` | `version` / `--help` (`/opt/terrascan/bin`) |
| `netexec` / `nxc` | `--help` |
| `docker` / `docker-bench-security` | CIS bench script under `/opt/docker-bench` |
| `kube-hunter`, `kube-bench`, `checkov`, `clair`, `falco`, `kube` | engage-stub placeholders (catalog / bridge until full install) |

**P11c Python subprocess smoke** (same script, after cloud checks): catalog tools `install_python_package` / `execute_python_script` via `engage-python-install` and `engage-python-exec`. Creates venv `p11c-smoke` under `ENGAGE_PYTHON_BASE` (`/tmp/engage/pyenv` in the image), pip-installs `requests`, then runs an inline `print('hello …')` script.

| Binary | Check |
|--------|--------|
| `engage-python-install` | `--env p11c-smoke --package requests` → `engage-python-install: ok` |
| `engage-python-exec` | `--env p11c-smoke --script "print('hello …')"` → stdout contains hello line |

Docs: [docs/engage/engage-tools.md](../../docs/engage/engage-tools.md) · [docs/engage/engage-runtime.md](../../docs/engage/engage-runtime.md)
