.PHONY: test-scrape test-pipeline test-graph test-graph-serve test-graph-read-smoke test-graph-engage-category test-engage test-engage-ctf test-engage-bugbounty test-engage-cve test-engage-benchmark test-engage-veil-stack-ci test-engage-smoke test-engage-smoke-tool test-engage-compose test-engage-runner-profile test-engage-veil-stack test-engage-decision-parity test-engage-catalog-args test-engage-tool-matrix test-engage-na-matrix test-engage-route-parity catalog-engage graph-pack-export graph-pack-build graph-pack-publish test-smoke check-graph-version bump-graph-patch

# GOWORK may point at scrape/go.work in the shell; each target uses the matching workspace.
test-scrape:
	cd pkg && env -u GOWORK go test ./harvest/... ./commit/...
	cd scrape/pkg && env -u GOWORK go test ./...
	cd scrape && env GOWORK=$$(pwd)/go.work go test ./connector/... ./harvest/...
	cd scrape/harvest && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/scrape_worker

test-pipeline:
	cd pkg && env -u GOWORK go test ./harvest/... ./commit/... ./ti/...
	cd pipeline/pkg && env -u GOWORK go test ./nvd/...
	cd pipeline && env GOWORK=$$(pwd)/go.work go build -o /dev/null ./connector/...
	cd pipeline && env GOWORK=$$(pwd)/go.work go test ./connector/...
	cd pipeline && env GOWORK=$$(pwd)/go.work go test ./ned/...
	cd pipeline/ned && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/pipeline_worker
	cd pipeline/engage-events && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/worker

test-graph:
	cd pkg && env -u GOWORK go test ./commit/... ./ti/...
	cd graph && env GOWORK=$$(pwd)/go.work go build -o /dev/null ./connector/...
	cd graph/ingest && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/ingest_worker
	cd graph/serve && env GOWORK=$$(dirname $$(pwd))/go.work go test ./...
	cd graph/serve && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/api ./cmd/mcp

test-graph-serve:
	cd graph/serve && env GOWORK=$$(dirname $$(pwd))/go.work go test ./... -race -count=1

test-graph-read-smoke:
	./scripts/test/smoke-graph-read.sh

test-graph-engage-category:
	chmod +x ./scripts/test/smoke-graph-engage-category.sh
	./scripts/test/smoke-graph-engage-category.sh

test-engage-ctf:
	cd engage/serve && env GOWORK=$$(dirname $$(pwd))/go.work go test ./internal/usecase/ctf/... -count=1 -run Golden
	chmod +x ./scripts/test/smoke-ctf-web.sh ./scripts/test/smoke-ctf-pwn.sh
	./scripts/test/smoke-ctf-web.sh
	./scripts/test/smoke-ctf-pwn.sh

test-engage-bugbounty:
	cd engage/serve && env GOWORK=$$(dirname $$(pwd))/go.work go test ./internal/usecase/bugbounty/... -count=1 -run Golden
	chmod +x ./scripts/test/smoke-bugbounty-recon.sh ./scripts/test/smoke-bugbounty-recon-execute.sh
	./scripts/test/smoke-bugbounty-recon.sh
	./scripts/test/smoke-bugbounty-recon-execute.sh

test-engage-cve:
	chmod +x ./scripts/test/smoke-cve-monitor.sh
	./scripts/test/smoke-cve-monitor.sh

test-engage-benchmark:
	chmod +x ./scripts/benchmark/engage-hexstrike-parity.sh
	./scripts/benchmark/engage-hexstrike-parity.sh

test-engage-catalog-args:
	chmod +x ./scripts/engage/check-catalog-args.sh
	./scripts/engage/check-catalog-args.sh

test-engage-tool-matrix:
	chmod +x ./scripts/test/smoke-engage-tool-matrix.sh
	./scripts/test/smoke-engage-tool-matrix.sh

test-engage-tool-matrix-strict:
	chmod +x ./scripts/test/smoke-engage-tool-matrix.sh
	ENGAGE_TOOL_MATRIX_STRICT=1 ENGAGE_TOOL_MATRIX_MIN=30 ./scripts/test/smoke-engage-tool-matrix.sh

test-engage:
	cd engage/serve && env GOWORK=$$(dirname $$(pwd))/go.work go test ./... -count=1
	cd engage/serve && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/api ./cmd/mcp ./cmd/worker ./cmd/browser-agent

test-engage-smoke:
	./scripts/test/smoke-engage.sh
	./scripts/test/smoke-engage-mcp.sh

test-engage-smoke-tool:
	chmod +x ./scripts/test/smoke-engage-tool.sh
	./scripts/test/smoke-engage-tool.sh

test-engage-minimal:
	ENGAGE_TOOLS_MINIMAL=1 ./scripts/engage/enable-catalog-by-category.sh network

test-engage-parity:
	./scripts/engage/check-catalog-parity.sh

test-engage-route-parity:
	python3 ./scripts/engage/check-route-parity.py

test-engage-compose:
	chmod +x ./scripts/test/smoke-engage-compose.sh
	./scripts/test/smoke-engage-compose.sh

test-engage-runner-profile:
	chmod +x ./scripts/test/smoke-engage-runner-profile.sh
	./scripts/test/smoke-engage-runner-profile.sh

test-engage-browser:
	chmod +x ./scripts/test/smoke-engage-browser.sh
	./scripts/test/smoke-engage-browser.sh

test-engage-redis-workers:
	chmod +x ./scripts/test/smoke-engage-redis-workers.sh
	./scripts/test/smoke-engage-redis-workers.sh

test-engage-secure:
	chmod +x ./scripts/test/smoke-engage-secure.sh
	./scripts/test/smoke-engage-secure.sh

test-engage-metrics:
	chmod +x ./scripts/test/smoke-engage-metrics.sh
	./scripts/test/smoke-engage-metrics.sh

test-engage-keycloak:
	chmod +x ./scripts/test/smoke-engage-keycloak.sh
	./scripts/test/smoke-engage-keycloak.sh

test-engage-events-pipeline:
	chmod +x ./scripts/test/smoke-engage-events-pipeline.sh
	./scripts/test/smoke-engage-events-pipeline.sh

test-engage-veil-stack:
	chmod +x ./scripts/test/smoke-veil-engage-stack.sh
	./scripts/test/smoke-veil-engage-stack.sh

test-engage-veil-stack-ci:
	chmod +x ./scripts/test/smoke-veil-engage-stack-ci.sh
	./scripts/test/smoke-veil-engage-stack-ci.sh

test-engage-decision-parity:
	./scripts/engage/check-decision-parity.sh

test-engage-na-matrix:
	python3 ./scripts/engage/generate-tools-na-matrix.py --check

catalog-engage:
	python3 ./scripts/engage/extract-legacy-catalog.py
	python3 ./scripts/engage/generate-tools-live.py
	python3 ./scripts/engage/generate-tools-na-matrix.py

test-engage-tools: catalog-engage test-engage-catalog-args test-engage-tool-matrix

graph-pack-export:
	./scripts/graph-pack/export-cypher.sh

graph-pack-build:
	./scripts/graph-pack/build.sh $(GRAPH_PACK_VERSION)

graph-pack-publish:
	./scripts/release/publish-graph-pack.sh

test-smoke:
	./scripts/test/smoke-scrape-e2e.sh

check-graph-version:
	./scripts/release/check-graph-version-bump.sh

bump-graph-patch:
	./scripts/release/bump-graph-version.sh patch
