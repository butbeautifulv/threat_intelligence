# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build
WORKDIR /build
COPY graph/ graph/
COPY pkg/ pkg/
ENV GOWORK=/build/graph/go.work
WORKDIR /build/graph/serve
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates wget \
  && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/api /usr/local/bin/api
USER nobody
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/api"]
