# syntax=docker/dockerfile:1
# Toolbox image for subprocess tools (nmap, nuclei, …). Not distroless.
FROM golang:1.25-bookworm AS pd
RUN go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest && \
    go install github.com/projectdiscovery/httpx/cmd/httpx@latest && \
    go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest && \
    go install github.com/ffuf/ffuf/v2@latest

FROM debian:bookworm-slim
ARG APT_MIRROR=
RUN set -eux; \
    if [ -n "${APT_MIRROR}" ]; then \
      sed -i "s|deb.debian.org|${APT_MIRROR}|g" /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
      sed -i "s|deb.debian.org|${APT_MIRROR}|g" /etc/apt/sources.list; \
    fi; \
    for i in 1 2 3; do \
      apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates curl nmap masscan sqlmap nikto gobuster \
      && break; \
      echo "apt retry $i" >&2; sleep 5; \
    done; \
    rm -rf /var/lib/apt/lists/*
COPY --from=pd /go/bin/nuclei /go/bin/httpx /go/bin/subfinder /go/bin/ffuf /usr/local/bin/
ARG FEROX_VERSION=2.11.0
RUN curl -fsSL -o /tmp/ferox.tgz \
    "https://github.com/epi052/feroxbuster/releases/download/v${FEROX_VERSION}/x86_64-unknown-linux-gnu.tar.gz" \
  && tar -xzf /tmp/ferox.tgz -C /usr/local/bin feroxbuster \
  && rm /tmp/ferox.tgz && chmod +x /usr/local/bin/feroxbuster
RUN useradd -r -u 10001 runner
USER 10001
WORKDIR /tmp/engage
CMD ["sleep", "infinity"]
