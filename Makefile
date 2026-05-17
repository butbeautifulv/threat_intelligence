 .PHONY: test-scrape test-scrape-p7c test-pipeline test-pipeline-p7d test-graph test-graph-ingest-p7e test-graph-serve-p7f test-engage-p7g test-graph-serve test-graph-read-smoke test-graph-engage-category test-engage test-engage-ctf test-engage-bugbounty test-engage-cve test-engage-benchmark test-engage-veil-stack-ci test-engage-smoke test-engage-smoke-tool test-engage-compose test-engage-runner-profile test-engage-veil-stack test-engage-decision-parity test-engage-catalog-args test-engage-tool-matrix test-engage-na-matrix test-engage-route-parity test-engage-hardening test-platform-p0 test-platform-p7 test-platform-closed-loop test-platform-full-loop test-platform-p3 test-platform-p4 catalog-engage graph-pack-export graph-pack-build graph-pack-publish test-smoke check-graph-version bump-graph-patch agents-list agents-render deploy-helm-template deploy-ansible-check sync-github-metadata external-clone-agent-store test-agent-eval-registry test-agent-eval-pilot test-agent-eval-paper test-pkg-shared test-pkg-domain

# Shared pkg contracts (harvest, commit, natsjet, auth, engage/events)
test-pkg-shared:
	cd pkg && env -u GOWORK go test ./harvest/... ./commit/... ./natsjet/...
	cd pkg/auth && env -u GOWORK go test ./...
	cd pkg/engage && env -u GOWORK go test ./...

# pkg domain contour (ti/domain, engage contract/toolid, auth httpmiddleware)
test-pkg-domain:
	cd pkg && env -u GOWORK go test ./ti/domain/... ./ti/validate/... ./ti/ids/... ./ti/normalize/...
	cd pkg/engage && env -u GOWORK go test ./contract/... ./toolid/...
	cd pkg/auth && env -u GOWORK go test ./httpmiddleware/...

# P7 gate: pkg + bus + layer unit tests (wave-1 parallel branches merged)
test-platform-p7: test-pkg-domain test-platform-p0 test-scrape-p7c test-pipeline-p7d test-graph-ingest-p7e test-graph-serve-p7f test-engage-p7g

# GOWORK may point at scrape/go.work in the shell; each target uses the matching workspace.
test-platform-p0: test-pkg-shared
	cd pipeline && env GOWORK=$$(pwd)/go.work go test ./connector/nats/...
	cd pipeline && env GOWORK=$$(pwd)/go.work go test ./ned/internal/consumer/... ./ned/internal/dedup/...
	cd graph && env GOWORK=$$(pwd)/go.work go test ./ingest/internal/ingest/...

test-platform-closed-loop:
	chmod +x ./scripts/test/smoke-platform-closed-loop.sh
	./scripts/test/smoke-platform-closed-loop.sh

test-platform-full-loop:
	chmod +x ./scripts/test/smoke-platform-full-loop.sh
	./scripts/test/smoke-platform-full-loop.sh

test-platform-p3: test-platform-p0 test-platform-closed-loop

test-platform-p4: test-platform-p0 test-platform-full-loop

agents-list:
	chmod +x ./scripts/agents/list-manifest.sh
	./scripts/agents/list-manifest.sh

agents-render:
	chmod +x ./scripts/agents/render-task-prompt.sh
	@test -n "$(AGENT)" || (echo "usage: make agents-render AGENT=platform-implementer [PHASE=platform-p4b]" >&2; exit 1)
	./scripts/agents/render-task-prompt.sh "$(AGENT)" $(if $(PHASE),--phase $(PHASE),)

deploy-helm-template:
	@if command -v helm >/dev/null 2>&1; then \
		helm template veil deploy/helm/veil -f deploy/helm/veil/values.yaml \
			-f deploy/helm/veil/values-stage.yaml \
			--set global.imageTag=$${APP_VERSION:-v0.4.5}; \
	else echo "SKIP: helm not installed"; fi

deploy-ansible-check:
	@if command -v ansible-playbook >/dev/null 2>&1; then \
		ansible-playbook deploy/ansible/playbooks/site.yml -i deploy/ansible/inventories/stage --syntax-check; \
	else echo "SKIP: ansible-playbook not installed"; fi

sync-github-metadata:
	chmod +x ./scripts/housekeeping/sync-github-metadata.sh
	./scripts/housekeeping/sync-github-metadata.sh

external-clone-agent-store:
	chmod +x ./scripts/external/clone-agent-store.sh
	./scripts/external/clone-agent-store.sh

test-agent-eval-registry:
	python3 ./scripts/eval/agent-eval-registry-audit.py

test-agent-eval-pilot:
	chmod +x ./scripts/eval/gaia/run-pilot.sh ./scripts/eval/gaia/solvers/stub.sh
	./scripts/eval/gaia/run-pilot.sh

test-agent-eval-paper:
	chmod +x ./scripts/eval/gaia/run-paper-examples.sh ./scripts/eval/gaia/solvers/stub.sh
	./scripts/eval/gaia/run-paper-examples.sh

pentest-veil-mcp:
	chmod +x ./scripts/eval/pentest-veil-mcp.sh
	./scripts/eval/pentest-veil-mcp.sh

pentest-veil-dual:
	chmod +x ./scripts/eval/run-dual-veil-pentest.sh
	./scripts/eval/run-dual-veil-pentest.sh

pentest-veil-prod-aggressive:
	chmod +x ./scripts/eval/run-dual-veil-pentest-secure.sh
	./scripts/eval/run-dual-veil-pentest-secure.sh

test-scrape:
	cd pkg && env -u GOWORK go test ./harvest/... ./commit/...
	cd scrape/pkg && env -u GOWORK go test ./...
	cd scrape && env GOWORK=$$(pwd)/go.work go test ./connector/... ./harvest/...
	cd scrape/harvest && env GOWORK=$$(dirname $$(pwd))/go.work go build -o /dev/null ./cmd/scrape_worker

# P7c slice: TI feeds/helpers + shared scrape feeds (see veil_platform_p7 plan)
test-scrape-p7c:
	cd scrape && env GOWORK=$$(pwd)/go.work go test ./harvest/internal/sources/ti/... ./harvest/internal/sources/lola/... ./harvest/internal/sources/ds/... ./harvest/internal/feeds/...
	cd scrape/pkg && env -u GOWORK go test ./proxypool/...

test-pipeline-p7d:
	cd pipeline && env GOWORK=$$(pwd)/go.work go test ./pkg/ti/normalize/... ./pkg/nvd/map/... ./ned/internal/sources/ds/... ./ned/internal/sources/lola/...

test-graph-ingest-p7e:
	cd graph && env GOWORK=$$(pwd)/go.work go test ./ingest/internal/ingest/... ./ingest/internal/sources/ti/... ./ingest/internal/sources/vuln/...

test-graph-serve-p7f:
	cd graph && env GOWORK=$$(pwd)/go.work go test ./serve/internal/usecase/... ./connector/query/...

test-engage-p7g:
	cd engage && env GOWORK=$$(pwd)/go.work go test ./serve/internal/usecase/tools/... ./serve/internal/security/... ./serve/internal/usecase/intelligence/...

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

test-engage-hardening:
	chmod +x ./scripts/test/engage-hardening-selftest.sh ./scripts/engage/hardening-compose-audit.py ./scripts/engage/hardening-framework-audit.py
	./scripts/test/engage-hardening-selftest.sh

test-engage-framework-audit:
	python3 ./scripts/engage/hardening-framework-audit.py

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
	./scripts/engage/check-catalog-parity.sh
	$(MAKE) test-engage-catalog-args

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
