# syntax=docker/dockerfile:1
# Toolbox image for subprocess tools (nmap, nuclei, …). Not distroless.
FROM golang:1.25-bookworm AS pd
RUN go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest && \
    go install github.com/projectdiscovery/httpx/cmd/httpx@latest && \
    go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest && \
    go install github.com/projectdiscovery/katana/cmd/katana@latest && \
    go install github.com/projectdiscovery/naabu/v2/cmd/naabu@latest && \
    go install github.com/projectdiscovery/dnsx/cmd/dnsx@latest && \
    go install github.com/lc/gau/v2/cmd/gau@latest && \
    go install github.com/tomnomnom/waybackurls@latest && \
    go install github.com/hahwul/dalfox/v2@latest && \
    go install github.com/owasp-amass/amass/v4/...@master && \
    go install github.com/ffuf/ffuf/v2@latest && \
    go install github.com/jaeles-project/jaeles@latest && \
    go install github.com/Sh1Yo/x8/cmd/x8@latest

FROM debian:bookworm-slim
ARG APT_MIRROR=
RUN set -eux; \
    if [ -n "${APT_MIRROR}" ]; then \
      sed -i "s|deb.debian.org|${APT_MIRROR}|g" /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
      sed -i "s|deb.debian.org|${APT_MIRROR}|g" /etc/apt/sources.list; \
    fi; \
    for i in 1 2 3; do \
      apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates curl git nmap masscan sqlmap nikto gobuster dirb \
        dnsenum fierce hydra wafw00f enum4linux sslscan testssl.sh \
        whatweb nbtscan binwalk \
        python3 python3-pip \
      && break; \
      echo "apt retry $i" >&2; sleep 5; \
    done; \
    pip3 install --break-system-packages --no-cache-dir arjun dirsearch paramspider enum4linux-ng 2>/dev/null || \
      pip3 install --no-cache-dir arjun dirsearch paramspider enum4linux-ng; \
    rm -rf /var/lib/apt/lists/*
COPY --from=pd /go/bin/nuclei /go/bin/httpx /go/bin/subfinder /go/bin/katana \
  /go/bin/naabu /go/bin/dnsx /go/bin/gau /go/bin/waybackurls /go/bin/dalfox \
  /go/bin/amass /go/bin/ffuf /go/bin/jaeles /go/bin/x8 /usr/local/bin/
ARG FEROX_VERSION=2.11.0
RUN curl -fsSL -o /tmp/ferox.tgz \
    "https://github.com/epi052/feroxbuster/releases/download/v${FEROX_VERSION}/x86_64-unknown-linux-gnu.tar.gz" \
  && tar -xzf /tmp/ferox.tgz -C /usr/local/bin feroxbuster \
  && rm /tmp/ferox.tgz && chmod +x /usr/local/bin/feroxbuster
ARG RUSTSCAN_VERSION=2.4.1
RUN curl -fsSL -o /tmp/rustscan.deb \
    "https://github.com/RustScan/RustScan/releases/download/${RUSTSCAN_VERSION}/rustscan_${RUSTSCAN_VERSION}_amd64.deb" \
  && dpkg -i /tmp/rustscan.deb || apt-get install -yf \
  && rm -f /tmp/rustscan.deb
ARG TRIVY_VERSION=0.58.1
RUN curl -fsSL -o /tmp/trivy.tgz \
    "https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_Linux-64bit.tar.gz" \
  && tar -xzf /tmp/trivy.tgz -C /usr/local/bin trivy \
  && rm /tmp/trivy.tgz && chmod +x /usr/local/bin/trivy
# paramspider expects output dir; wrapper for non-interactive runs
RUN printf '%s\n' '#!/bin/sh' 'exec paramspider "$@"' > /usr/local/bin/paramspider-cli \
  && chmod +x /usr/local/bin/paramspider-cli
RUN useradd -r -u 10001 runner
USER 10001
WORKDIR /tmp/engage
CMD ["sleep", "infinity"]
