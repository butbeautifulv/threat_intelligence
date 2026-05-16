# syntax=docker/dockerfile:1
# Toolbox image for subprocess tools (nmap, nuclei, …). Not distroless.
FROM golang:1.25-bookworm AS pd
RUN go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest && \
    go install github.com/projectdiscovery/httpx/cmd/httpx@latest && \
    go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl nmap \
  && rm -rf /var/lib/apt/lists/*
COPY --from=pd /go/bin/nuclei /go/bin/httpx /go/bin/subfinder /usr/local/bin/
ARG FEROX_VERSION=2.11.0
RUN curl -fsSL -o /tmp/ferox.tgz \
    "https://github.com/epi052/feroxbuster/releases/download/v${FEROX_VERSION}/x86_64-unknown-linux-gnu.tar.gz" \
  && tar -xzf /tmp/ferox.tgz -C /usr/local/bin feroxbuster \
  && rm /tmp/ferox.tgz && chmod +x /usr/local/bin/feroxbuster
RUN useradd -r -u 10001 runner
USER 10001
WORKDIR /tmp/engage
CMD ["sleep", "infinity"]
