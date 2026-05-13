# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY . .
WORKDIR /src/scrapers/vuln
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/vuln ./cmd

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/vuln /usr/local/bin/vuln
USER nobody
ENTRYPOINT ["/usr/local/bin/vuln"]
