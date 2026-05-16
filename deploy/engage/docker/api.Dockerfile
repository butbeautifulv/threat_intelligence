# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build
WORKDIR /build
COPY pkg/ pkg/
COPY engage/ engage/
ENV GOWORK=/build/engage/go.work
WORKDIR /build/engage/serve
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
COPY engage/serve/catalog /app/catalog
USER nonroot:nonroot
EXPOSE 8890
HEALTHCHECK CMD ["/api", "healthcheck"]
ENTRYPOINT ["/api"]
