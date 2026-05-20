#!/usr/bin/env python3
"""Audit Veil repo against deploy/security/veil-controls.yaml (JCSF/DAF/OWASP mappings)."""
from __future__ import annotations

import re
import sys
from pathlib import Path

try:
    import yaml
except ImportError:
    print("pip install pyyaml", file=sys.stderr)
    sys.exit(2)

ROOT = Path(__file__).resolve().parents[2]
CONTROLS = ROOT / "deploy/security/veil-controls.yaml"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8", errors="replace")


def check_engage_deny_raw_in_secure_profile() -> tuple[bool, str]:
    p = ROOT / "deploy/profiles/secure-engage.env"
    body = read(p)
    if "ENGAGE_DENY_RAW_COMMAND=1" not in body:
        return False, "secure-engage.env missing ENGAGE_DENY_RAW_COMMAND=1"
    return True, "ok"


def check_engage_runner_mode_docker_in_secure_profile() -> tuple[bool, str]:
    body = read(ROOT / "deploy/profiles/secure-engage.env")
    if "ENGAGE_RUNNER_MODE=docker" not in body:
        return False, "secure-engage.env should set ENGAGE_RUNNER_MODE=docker"
    if "ENGAGE_EXECUTION_PROFILE=docker-exec" not in body:
        return False, "secure-engage.env should set ENGAGE_EXECUTION_PROFILE=docker-exec with docker runner"
    return True, "ok"


def check_engage_require_auth_in_secure_profile() -> tuple[bool, str]:
    body = read(ROOT / "deploy/profiles/secure-engage.env")
    if "VEIL_REQUIRE_AUTH=1" not in body:
        return False, "missing VEIL_REQUIRE_AUTH=1"
    return True, "ok"


def check_engage_mcp_auth_strict() -> tuple[bool, str]:
    body = read(ROOT / "deploy/profiles/secure-engage.env")
    if "ENGAGE_MCP_HTTP_AUTH_STRICT=1" not in body:
        return False, "missing ENGAGE_MCP_HTTP_AUTH_STRICT=1"
    return True, "ok"


def check_engage_target_guard_in_secure_profile() -> tuple[bool, str]:
    body = read(ROOT / "deploy/profiles/secure-engage.env")
    if "ENGAGE_TARGET_GUARD=block" not in body and "ENGAGE_ENV=prod" in body:
        # prod defaults to block in code; profile may omit explicit key
        return True, "ok (prod default block in code)"
    if "ENGAGE_TARGET_GUARD=block" in body:
        return True, "ok"
    return False, "recommend ENGAGE_TARGET_GUARD=block in secure-engage.env"


def check_code_command_metachar_guard() -> tuple[bool, str]:
    p = ROOT / "engage/serve/internal/usecase/command/runner.go"
    if "ContainsShellMetacharacters" not in read(p):
        return False, "command runner missing metachar guard"
    return True, "ok"


def check_code_securityhttp_headers() -> tuple[bool, str]:
    p = ROOT / "engage/serve/internal/transport/securityhttp/middleware.go"
    body = read(p)
    for h in ("Content-Security-Policy", "Permissions-Policy", "Strict-Transport-Security"):
        if h not in body:
            return False, f"missing {h}"
    return True, "ok"


def check_code_target_guard() -> tuple[bool, str]:
    if not (ROOT / "engage/serve/internal/security/target_guard.go").exists():
        return False, "target_guard.go missing"
    return True, "ok"


def check_code_files_path_traversal_test() -> tuple[bool, str]:
    p = ROOT / "engage/serve/internal/usecase/files/manager_test.go"
    if "rejectsTraversal" not in read(p):
        return False, "files traversal test missing"
    return True, "ok"


def check_code_audit_logger_present() -> tuple[bool, str]:
    if not (ROOT / "engage/serve/internal/audit/log.go").exists():
        return False, "audit log.go missing"
    return True, "ok"


def check_repo_deploy_hybrid_layout() -> tuple[bool, str]:
    needed = [
        ROOT / "deploy/terraform/README.md",
        ROOT / "deploy/ansible/playbooks/site.yml",
        ROOT / "deploy/helm/veil/Chart.yaml",
        ROOT / "docs/deploy/deploy-platform-hybrid.md",
    ]
    missing = [str(p.relative_to(ROOT)) for p in needed if not p.exists()]
    if missing:
        return False, "missing: " + ", ".join(missing)
    return True, "ok"


def check_compose_graph_secure_overlay_exists() -> tuple[bool, str]:
    if not (ROOT / "deploy/knowledge/compose.secure.yml").exists():
        return False, "compose.secure.yml missing"
    return True, "ok"


def check_workflow_engage_hardening() -> tuple[bool, str]:
    body = read(ROOT / ".github/workflows/engage.yml")
    if "internal/security" not in body:
        return False, "engage.yml missing security tests"
    return True, "ok"


def check_workflow_engage_secure() -> tuple[bool, str]:
    if not (ROOT / ".github/workflows/engage-secure.yml").exists():
        return False, "engage-secure.yml missing"
    return True, "ok"


def check_script_graph_pack_checksum() -> tuple[bool, str]:
    if not (ROOT / "scripts/graph-pack/build.sh").exists():
        return False, "graph pack build script missing"
    return True, "ok"


def check_eval_gaia_pilot_harness() -> tuple[bool, str]:
    needed = [
        ROOT / "eval/gaia/fixtures/pilot/metadata.jsonl",
        ROOT / "scripts/eval/gaia/score.py",
        ROOT / "scripts/eval/gaia/run-pilot.sh",
    ]
    missing = [str(p.relative_to(ROOT)) for p in needed if not p.exists()]
    if missing:
        return False, "missing: " + ", ".join(missing)
    return True, "ok"


def check_eval_gaia_data_gitignored() -> tuple[bool, str]:
    gi = read(ROOT / "eval/gaia/.gitignore")
    if "data/" not in gi:
        return False, "eval/gaia/.gitignore should ignore data/"
    return True, "ok"


def check_workflow_agent_eval_pilot() -> tuple[bool, str]:
    p = ROOT / ".github/workflows/agent-eval.yml"
    if not p.exists():
        return False, "agent-eval.yml missing"
    body = read(p)
    if "test-agent-eval-pilot" not in body:
        return False, "agent-eval.yml should run make test-agent-eval-pilot"
    return True, "ok"


VERIFY = {
    "engage_deny_raw_in_secure_profile": check_engage_deny_raw_in_secure_profile,
    "engage_runner_mode_docker_in_secure_profile": check_engage_runner_mode_docker_in_secure_profile,
    "engage_require_auth_in_secure_profile": check_engage_require_auth_in_secure_profile,
    "engage_mcp_auth_strict": check_engage_mcp_auth_strict,
    "engage_target_guard_in_secure_profile": check_engage_target_guard_in_secure_profile,
    "code_command_metachar_guard": check_code_command_metachar_guard,
    "code_securityhttp_headers": check_code_securityhttp_headers,
    "code_target_guard": check_code_target_guard,
    "code_files_path_traversal_test": check_code_files_path_traversal_test,
    "code_audit_logger_present": check_code_audit_logger_present,
    "repo_deploy_hybrid_layout": check_repo_deploy_hybrid_layout,
    "compose_graph_secure_overlay_exists": check_compose_graph_secure_overlay_exists,
    "workflow_engage_hardening": check_workflow_engage_hardening,
    "workflow_engage_secure": check_workflow_engage_secure,
    "script_graph_pack_checksum": check_script_graph_pack_checksum,
    "eval_gaia_pilot_harness": check_eval_gaia_pilot_harness,
    "eval_gaia_data_gitignored": check_eval_gaia_data_gitignored,
    "workflow_agent_eval_pilot": check_workflow_agent_eval_pilot,
}


def main() -> int:
    if not CONTROLS.exists():
        print(f"missing {CONTROLS}", file=sys.stderr)
        return 2
    doc = yaml.safe_load(CONTROLS.read_text(encoding="utf-8"))
    failed = []
    passed = 0
    for ctrl in doc.get("controls", []):
        cid = ctrl["id"]
        for v in ctrl.get("verify", []):
            fn = VERIFY.get(v)
            if fn is None:
                failed.append((cid, v, "unknown verifier"))
                continue
            ok, msg = fn()
            if ok:
                passed += 1
            else:
                failed.append((cid, v, msg))
    for cid, v, msg in failed:
        print(f"FAIL [{cid}] {v}: {msg}")
    print(f"\nframework audit: {passed} passed, {len(failed)} failed")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
