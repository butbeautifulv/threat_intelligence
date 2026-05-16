# syntax=docker/dockerfile:1
# veil-engage MCP: stdio (default) + optional Streamable HTTP (ENGAGE_MCP_HTTP_ENABLED=1).
FROM golang:1.25-bookworm AS build
WORKDIR /build
COPY pkg/ pkg/
COPY engage/ engage/
ENV GOWORK=/build/engage/go.work
ENV CGO_ENABLED=0
WORKDIR /build/engage/serve
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/mcp ./cmd/mcp

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=build /out/mcp /mcp
COPY engage/serve/catalog /app/catalog
USER nonroot:nonroot
EXPOSE 8892
HEALTHCHECK --interval=20s --timeout=5s --start-period=30s --retries=3 \
  CMD ["/mcp", "healthcheck"]
ENTRYPOINT ["/mcp"]
