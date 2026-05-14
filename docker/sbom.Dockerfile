# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY . .
WORKDIR /src/scrapers/sbom
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sbom ./cmd

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/sbom /usr/local/bin/sbom
COPY --from=build /src/scrapers/sbom/fixtures/cve_list_seed.txt /fixtures/cve_list_seed.txt
USER nobody
ENTRYPOINT ["/usr/local/bin/sbom"]
