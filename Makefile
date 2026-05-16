.PHONY: test-scrape test-pipeline test-graph test-graph-serve test-graph-read-smoke test-engage test-engage-smoke catalog-engage graph-pack-export graph-pack-build graph-pack-publish test-smoke check-graph-version bump-graph-patch

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
	cd pipeline && env GOWORK=$$(pwd)/go.work go test ./ned/...
	cd pipeline/ned && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/pipeline_worker

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

test-engage:
	cd engage/serve && env GOWORK=$$(dirname $$(pwd))/go.work go test ./... -count=1
	cd engage/serve && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/api ./cmd/mcp ./cmd/worker

test-engage-smoke:
	./scripts/test/smoke-engage.sh
	./scripts/test/smoke-engage-mcp.sh

test-engage-minimal:
	ENGAGE_TOOLS_MINIMAL=1 ./scripts/engage/enable-catalog-by-category.sh network

test-engage-parity:
	./scripts/engage/check-catalog-parity.sh

catalog-engage:
	python3 ./scripts/engage/extract-legacy-catalog.py

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
