#!/usr/bin/env bash
# Shared smoke helpers — source from scripts/test/smoke-*.sh
smoke_skip_no_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "[smoke] SKIP: docker not available" >&2
    exit 0
  fi
}

smoke_wait_http() {
  local url=$1 max=${2:-60} i=0
  while [[ $i -lt $max ]]; do
    if curl -sf "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
    i=$((i + 1))
  done
  echo "[smoke] timeout waiting for $url" >&2
  return 1
}
