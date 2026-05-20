# Engage client-native tools on Linux (multi-distro)

Engage runs catalog tools as **host subprocesses** (see [engage-mcp-topology.md](engage-mcp-topology.md)). This runbook covers installing common CLIs with the **system package manager** and verifying them with preflight.

## Prerequisites

- **Python 3** with **PyYAML** (for reading `scripts/ops/engage-tools-packages.yaml`):
  - Debian/Ubuntu: `sudo apt-get install -y python3-yaml`
  - Fedora: `sudo dnf install -y python3-pyyaml`
  - Arch: `sudo pacman -S --needed python-yaml`

## Package map

Authoritative mapping of profile → tool → distro packages:

- [`../scripts/ops/engage-tools-packages.yaml`](../scripts/ops/engage-tools-packages.yaml)

Profiles: `minimal`, `recommended`, `full` (see file). Some tools have **empty** package lists on certain distros (for example `masscan` on Alpine); install those manually or skip them in that environment.

## Plan-only vs install

From the repo root:

```bash
# Show the command the installer would run (no packages installed)
make engage-install-plan

# Or: ./scripts/ops/install-engage-host-tools.sh --plan --profile minimal
```

To **install** packages (uses `sudo`):

```bash
make engage-install-host-tools

# Or pick profile explicitly:
ENGAGE_INSTALL_PROFILE=minimal ./scripts/ops/install-engage-host-tools.sh --yes
```

Override the YAML path with `ENGAGE_TOOLS_PACKAGES_YAML` if you maintain a forked map.

## Preflight

```bash
./scripts/engage/preflight-client-tools.sh
./scripts/engage/preflight-client-tools.sh --profile minimal
./scripts/engage/preflight-client-tools.sh --profile full --json
```

Environment: `ENGAGE_PREFLIGHT_PROFILE`, `ENGAGE_TOOLS_PACKAGES_YAML`.

## Red-vs-blue lab (optional)

After `engage-api` is running as **victim** on `:8891`, run the harness:

```bash
export ENGAGE_VICTIM_URL=http://127.0.0.1:8891
make test-engage-red-blue
```

Details and legal scope: [engage-red-blue-lab.md](engage-red-blue-lab.md).

## See also

- [engage-client-dependencies.md](engage-client-dependencies.md) — what must exist on `PATH`
- [../engage/README.md](../engage/README.md) — dev compose vs client-native
