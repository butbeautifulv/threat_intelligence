#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import shutil
import subprocess
import yaml

ROOT = Path(__file__).resolve().parents[2]
CATALOG = ROOT / "engage/serve/catalog/tools.yaml"
SOURCES = ROOT / "scripts/ops/engage-tools-sources.yaml"
PACKAGES = ROOT / "scripts/ops/engage-tools-packages.yaml"
OUT = ROOT / "docs/engage/engage-tool-install-coverage.md"

BRIDGE_BINARIES = {
    "api", "bugbounty", "ai", "get", "http", "create", "execute", "generate",
    "list", "kube", "browser", "autorecon", "comprehensive", "advanced",
    "analyze", "checkov", "clair", "cloudmapper", "checksec", "clear",
}


def apt_candidate(pkg: str) -> bool:
    cp = subprocess.run(["apt-cache", "policy", pkg], capture_output=True, text=True, check=False)
    for line in cp.stdout.splitlines():
        if line.strip().startswith("Candidate:"):
            v = line.split(":", 1)[1].strip()
            return bool(v and v != "(none)")
    return False


def load_yaml(path: Path) -> dict:
    return yaml.safe_load(path.read_text(encoding="utf-8")) or {}


def main() -> int:
    catalog = load_yaml(CATALOG).get("tools", [])
    sources = load_yaml(SOURCES).get("tools", {})
    packages = load_yaml(PACKAGES).get("tools", {})
    lines = [
        "# Engage Tool Install Coverage",
        "",
        "Auto-generated for install coverage wave.",
        "",
        "```bash",
        "python3 scripts/engage/generate-tool-install-coverage.py",
        "```",
        "",
        "| Tool | Binary | Install Required | Ubuntu/Debian repo | Kali fallback | Upstream fallback | Runtime on this host |",
        "|------|--------|------------------|---------------------|---------------|-------------------|----------------------|",
    ]
    install_required = 0
    ready_now = 0
    repo_ok = 0
    for t in sorted(catalog, key=lambda x: x.get("name", "")):
        name = str(t.get("name", ""))
        binary = str(t.get("binary", ""))
        is_bridge = binary in BRIDGE_BINARIES
        need_install = not is_bridge
        if need_install:
            install_required += 1
        src = sources.get(name, {})
        pkg_map = packages.get(binary, {})
        apt_pkgs = pkg_map.get("apt", []) if isinstance(pkg_map, dict) else []
        repo_status = "n/a"
        if need_install:
            if apt_pkgs:
                ok = all(apt_candidate(p) for p in apt_pkgs)
                repo_status = "ok" if ok else "missing"
                if ok:
                    repo_ok += 1
            else:
                repo_status = "missing"
        kali_status = "n/a"
        if need_install:
            kali_status = "yes" if src.get("kali_pkg_tracker") else "no"
        methods = src.get("preferred_install_methods", []) if isinstance(src, dict) else []
        upstream_status = "n/a"
        if need_install:
            upstream_status = "yes" if any(str(m).startswith(("go:", "cargo:", "pipx:", "gem:")) for m in methods) else "no"
        runtime = "ok" if (binary and shutil.which(binary)) else ("n/a" if is_bridge else "missing")
        if runtime == "ok" or runtime == "n/a":
            ready_now += 1
        lines.append(
            f"| `{name}` | `{binary}` | {'yes' if need_install else 'no (bridge/workflow)'} | {repo_status} | {kali_status} | {upstream_status} | {runtime} |"
        )
    lines.extend(
        [
            "",
            f"- Catalog tools: **{len(catalog)}**",
            f"- Install-required tools: **{install_required}**",
            f"- Ubuntu/Debian repo-ok (install-required subset): **{repo_ok}**",
            f"- Runtime ready on this host (`ok` + `n/a`): **{ready_now}/{len(catalog)}**",
        ]
    )
    OUT.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"wrote {OUT}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
