# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY . .
WORKDIR /src/scrapers/lola
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/lola ./cmd

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/lola /usr/local/bin/lola
USER nobody
ENTRYPOINT ["/usr/local/bin/lola"]
