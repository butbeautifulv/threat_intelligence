#!/usr/bin/env bash
# T3 gate: 100% statement coverage on pipeline logic packages.
export PIPELINE_COVER_STRICT=1
exec "$(cd "$(dirname "$0")" && pwd)/pipeline-cover.sh"
