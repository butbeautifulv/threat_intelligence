FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY pkg/ pkg/
COPY pipeline/connector/ pipeline/connector/
COPY pipeline/engage-events/ pipeline/engage-events/
WORKDIR /src/pipeline/engage-events
RUN CGO_ENABLED=0 go build -o /engage-events-worker ./cmd/worker

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /engage-events-worker /engage-events-worker
ENTRYPOINT ["/engage-events-worker"]
