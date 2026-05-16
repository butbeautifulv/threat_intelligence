# syntax=docker/dockerfile:1
# Dev/lab API image with Docker CLI for ENGAGE_RUNNER_MODE=docker (profile runner).
# Mounts /var/run/docker.sock — high privilege; not for production distroless deploy.
FROM golang:1.25-bookworm AS build
WORKDIR /build
COPY pkg/ pkg/
COPY engage/ engage/
ENV GOWORK=/build/engage/go.work
WORKDIR /build/engage/serve
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/api ./cmd/api

FROM docker:27-cli AS dockercli
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=dockercli /usr/local/bin/docker /usr/local/bin/docker
COPY --from=build /out/api /api
COPY engage/serve/catalog /app/catalog
WORKDIR /app
EXPOSE 8890
HEALTHCHECK CMD ["/api", "healthcheck"]
ENTRYPOINT ["/api"]
