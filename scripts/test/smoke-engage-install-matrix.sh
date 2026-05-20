#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

echo "[matrix] host distro check"
./scripts/ops/install-engage-host-tools.sh --plan --profile recommended --policy repo-first >/tmp/engage-plan-repo.txt || true
./scripts/ops/install-engage-host-tools.sh --plan --profile recommended --policy upstream-fallback >/tmp/engage-plan-upstream.txt || true
./scripts/engage/preflight-client-tools.sh --profile recommended --json || true

if ! command -v docker >/dev/null 2>&1; then
  echo "[matrix] docker not found: skip debian container validation"
  exit 0
fi

echo "[matrix] debian stable dry-run in container"
docker run --rm -v "$ROOT":"$ROOT" -w "$ROOT" debian:stable-slim bash -lc '
  set -euo pipefail
  if ! apt-get update >/dev/null; then
    echo "[matrix] skip debian stable check: apt index unavailable"
    exit 0
  fi
  if ! apt-get install -y python3 python3-yaml >/dev/null; then
    echo "[matrix] skip debian stable check: python deps unavailable"
    exit 0
  fi
  ./scripts/ops/install-engage-host-tools.sh --plan --profile recommended --policy repo-first || true
'
echo "[matrix] done"
