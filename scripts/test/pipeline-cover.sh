#!/usr/bin/env bash
# Enforce pipeline unit-test tiers: T0 (tests present), T2/T3 (coverage floors).
# Usage: ./scripts/test/pipeline-cover.sh
# T3: PIPELINE_COVER_STRICT=1 or pipeline-cover-strict.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PIPELINE="${ROOT}/pipeline"
GOWORK="${PIPELINE}/go.work"

if [[ "${PIPELINE_COVER_STRICT:-}" == "1" ]]; then
	DEFAULT_FLOOR=100
	_LOG_PREFIX="[pipeline-cover-strict]"
else
	DEFAULT_FLOOR=70
	_LOG_PREFIX="[pipeline-cover]"
fi

# T0: thin entrypoints — build gate in make test-pipeline, not statement coverage.
build_only=(
	github.com/butbeautifulv/veil/pipeline/ned/cmd/pipeline_worker
	github.com/butbeautifulv/veil/pipeline/engage-events/cmd/worker
)

is_build_only() {
	local pkg="$1"
	for b in "${build_only[@]}"; do
		[[ "${pkg}" == "${b}" ]] && return 0
	done
	return 1
}

log() { printf '%s %s\n' "${_LOG_PREFIX}" "$*"; }
fail() { log "FAIL: $*"; exit 1; }

check_notest() {
	local dir="$1"
	local use_gowork="$2"
	local mod_label="$3"
	local missing
	if [[ "${use_gowork}" == "1" ]]; then
		missing="$(cd "${dir}" && env GOWORK="${GOWORK}" go list -f '{{if not .TestGoFiles}}{{if .GoFiles}}{{if not .Main}}{{.ImportPath}}{{end}}{{end}}{{end}}' ./... 2>/dev/null | grep -v '^$' || true)"
	else
		missing="$(cd "${dir}" && env -u GOWORK go list -f '{{if not .TestGoFiles}}{{if .GoFiles}}{{if not .Main}}{{.ImportPath}}{{end}}{{end}}{{end}}' ./... 2>/dev/null | grep -v '^$' || true)"
	fi
	local filtered=""
	while IFS= read -r pkg; do
		[[ -z "${pkg}" ]] && continue
		is_build_only "${pkg}" && continue
		filtered+="${pkg}"$'\n'
	done <<< "${missing}"
	if [[ -n "$(echo -n "${filtered}" | tr -d '\n')" ]]; then
		fail "${mod_label}: packages without *_test.go:\n${filtered}"
	fi
}

declare -a COVER_LINES=()

run_cover() {
	local dir="$1"
	local use_gowork="$2"
	local label="$3"
	check_notest "${dir}" "${use_gowork}" "${label}"
	if [[ "${use_gowork}" == "1" ]]; then
		while IFS= read -r line; do
			COVER_LINES+=("${line}")
		done < <(cd "${dir}" && env GOWORK="${GOWORK}" go test ./... -cover 2>&1 | grep -E '^(ok|FAIL)\s+' || true)
	else
		while IFS= read -r line; do
			COVER_LINES+=("${line}")
		done < <(cd "${dir}" && env -u GOWORK go test ./... -cover 2>&1 | grep -E '^(ok|FAIL)\s+' || true)
	fi
}

run_cover "${PIPELINE}/pkg" 0 "pipeline/pkg"
run_cover "${PIPELINE}/connector" 1 "pipeline/connector"
run_cover "${PIPELINE}/ned" 1 "pipeline/ned"

for line in "${COVER_LINES[@]}"; do
	[[ "${line}" == FAIL* ]] && fail "tests failed: ${line}"
	pkg="$(echo "${line}" | awk '{print $2}')"
	rest="$(echo "${line}" | cut -d$'\t' -f3-)"
	if is_build_only "${pkg}"; then
		log "T0 ${pkg} (build-only)"
		continue
	fi
	if [[ "${rest}" == *"[no statements]"* ]]; then
		log "T0 ${pkg} (no statements)"
		continue
	fi
	if [[ "${rest}" != *"coverage:"* ]]; then
		continue
	fi
	pct="$(echo "${rest}" | sed -n 's/.*coverage: \([0-9.]*\)%.*/\1/p')"
	if [[ -z "${pct}" ]]; then
		fail "${pkg}: could not parse coverage from: ${line}"
	fi
	awk -v p="${pct}" -v f="${DEFAULT_FLOOR}" 'BEGIN { exit (p+0 >= f+0) ? 0 : 1 }' || fail "${pkg}: ${pct}% < floor ${DEFAULT_FLOOR}%"
	if [[ "${PIPELINE_COVER_STRICT:-}" == "1" ]]; then
		log "T3 ${pkg} ${pct}%"
	else
		log "T2 ${pkg} ${pct}%"
	fi
done

if [[ "${PIPELINE_COVER_STRICT:-}" == "1" ]]; then
	log "OK — pipeline T3 coverage gates passed"
else
	log "OK — pipeline coverage gates passed"
fi
