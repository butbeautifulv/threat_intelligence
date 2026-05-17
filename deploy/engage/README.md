# Engage deployment

Compose stacks for the Engage offensive tooling layer.

## Runner images

| Image | Dockerfile target | Profile | Approx. size / RAM |
|-------|-------------------|---------|-------------------|
| `engage-runner` | `engage-runner` | `runner` | ~2–3 GB disk; 512 MB–1 GB RAM typical |
| `engage-runner-full` | `engage-runner-full` | `runner-full` | ~8–12 GB disk; **4–8 GB RAM** recommended (Ghidra, Metasploit, angr) |

Tier-1 CLI: [docker/runner.Dockerfile](docker/runner.Dockerfile) target `engage-runner`.

**P9g heavy stack** (Burp JAR, Ghidra, hashcat, john, gdb, Metasploit, angr, radare2, volatility3, wpscan): same Dockerfile, target `engage-runner-full`. Headless wrappers live in [docker/wrappers/](docker/wrappers/).

```bash
# Slim runner (default)
docker compose -f deploy/engage/compose.yml --profile runner up -d --build engage-runner

# Full port heavy stack
docker compose -f deploy/engage/compose.yml --profile runner-full up -d --build engage-runner-full
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

Docs: [docs/engage-tools.md](../../docs/engage-tools.md) · [docs/engage-runtime.md](../../docs/engage-runtime.md)
