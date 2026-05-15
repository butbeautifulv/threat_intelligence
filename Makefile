.PHONY: contracts test-scrape test-pipeline test-graph

contracts:
	./scripts/gen-contracts.sh

test-scrape:
	cd pkg/nvdparse && go test .
	cd scrape/contract/scrapev1 && go test .
	cd scrape/scrape_worker && go build -o /dev/null .

test-pipeline:
	cd pkg/nvdparse && go test .
	cd pipeline/contract/ingestv1 && go test .
	cd pipeline/pipeline_worker && go test ./...
	cd pipeline/pipeline_worker && go build -o /dev/null ./...

test-graph:
	cd graph/contract/ingestv1 && go test .
	cd graph/ingest_worker && go build -o /dev/null ./...
	cd graph/api && go build -o /dev/null ./...
