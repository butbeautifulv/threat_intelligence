# syntax=docker/dockerfile:1
# Toolbox image for subprocess tools (nmap, nuclei, …). Not distroless.
# Targets: engage-runner (default tier-1), engage-runner-full (P9g heavy stack).
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
    go install github.com/Sh1Yo/x8/cmd/x8@latest && \
    go install github.com/tomnomnom/anew@latest && \
    go install github.com/hakluke/hakrawler@latest && \
    go install github.com/tomnomnom/qsreplace@latest

FROM debian:bookworm-slim AS runner-os
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
        arp-scan exiftool foremost steghide xxd binutils \
        samba-common-bin responder hashpump dotdotpwn xsser \
        python3 python3-pip python3-venv \
      && break; \
      echo "apt retry $i" >&2; sleep 5; \
    done; \
    pip3 install --break-system-packages --no-cache-dir \
      arjun dirsearch paramspider enum4linux-ng \
      netexec wfuzz uro smbmap ROPgadget ropper volatility3 pwntools scoutsuite jwt_tool \
      2>/dev/null || pip3 install --no-cache-dir \
      arjun dirsearch paramspider enum4linux-ng \
      netexec wfuzz uro smbmap ROPgadget ropper volatility3 pwntools scoutsuite jwt_tool; \
    git clone --depth 1 https://github.com/docker/docker-bench-security.git /opt/docker-bench; \
    git clone --depth 1 https://github.com/niklasb/libc-database.git /opt/libc-database; \
    git clone --depth 1 https://github.com/RhinoSecurityLabs/pacu.git /opt/pacu; \
    rm -rf /var/lib/apt/lists/*
COPY --from=pd /go/bin/nuclei /go/bin/httpx /go/bin/subfinder /go/bin/katana \
  /go/bin/naabu /go/bin/dnsx /go/bin/gau /go/bin/waybackurls /go/bin/dalfox \
  /go/bin/amass /go/bin/ffuf /go/bin/jaeles /go/bin/x8 \
  /go/bin/anew /go/bin/hakrawler /go/bin/qsreplace /usr/local/bin/
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
RUN printf '%s\n' '#!/bin/sh' 'exec paramspider "$@"' > /usr/local/bin/paramspider-cli \
  && chmod +x /usr/local/bin/paramspider-cli
ARG TERRASCAN_VERSION=1.19.9
RUN mkdir -p /opt/terrascan/bin \
  && curl -fsSL -o /tmp/terrascan.tgz \
    "https://github.com/tenable/terrascan/releases/download/v${TERRASCAN_VERSION}/terrascan_${TERRASCAN_VERSION}_Linux_x86_64.tar.gz" \
  && tar -xzf /tmp/terrascan.tgz -C /opt/terrascan/bin terrascan \
  && rm -f /tmp/terrascan.tgz && chmod +x /opt/terrascan/bin/terrascan
ARG PWNINIT_VERSION=3.3.1
RUN curl -fsSL -o /usr/local/bin/pwninit \
    "https://github.com/levitatingpineapple/pwninit/releases/download/v${PWNINIT_VERSION}/pwninit-v${PWNINIT_VERSION}-x86_64-unknown-linux-gnu" \
  && chmod +x /usr/local/bin/pwninit
COPY deploy/engage/docker/wrappers/ /usr/local/bin/
RUN set -eux; \
  chmod +x /usr/local/bin/engage-stub /usr/local/bin/arp /usr/local/bin/docker \
    /usr/local/bin/scout /usr/local/bin/volatility3 /usr/local/bin/msfvenom \
    /usr/local/bin/one /usr/local/bin/libc /usr/local/bin/jwt /usr/local/bin/pwntools \
    /usr/local/bin/ropgadget /usr/local/bin/netexec /usr/local/bin/pacu \
    /usr/local/bin/prowler /usr/local/bin/pwninit /usr/local/bin/terrascan /usr/local/bin/zap \
    /usr/local/bin/engage-python-install /usr/local/bin/engage-python-exec; \
  for b in correlate delete detect discover display error format intelligent \
    modify monitor optimize pause research resume select server terminate test threat \
    vulnerability falco graphql kube kube-hunter kube-bench checkov clair; do \
    ln -sf engage-stub "/usr/local/bin/$b"; \
  done; \
  ln -sf docker /usr/local/bin/docker-bench-security
ENV ENGAGE_PYTHON_BASE=/tmp/engage/pyenv
RUN useradd -r -u 10001 runner

FROM runner-os AS engage-runner
USER 10001
WORKDIR /tmp/engage
CMD ["sleep", "infinity"]

# P9g: tier-1 + heavy stack (headless wrappers). Build: --target engage-runner-full
FROM engage-runner AS engage-runner-full
USER root
RUN set -eux; \
    apt-get update && apt-get install -y --no-install-recommends \
      hashcat john gdb radare2 metasploit-framework \
      default-jre-headless ruby ruby-dev build-essential zlib1g-dev libxml2-dev libcurl4-openssl-dev libssl-dev pkg-config \
      unzip wget \
    && gem install wpscan --no-document \
    && mkdir -p /opt/wpscan \
    && WPSCAN_GEM="$(command -v wpscan)" && cp "$WPSCAN_GEM" /opt/wpscan/wpscan-bin \
    && pip3 install --break-system-packages --no-cache-dir angr volatility3 \
    && rm -rf /var/lib/apt/lists/*
ARG BURP_VERSION=2024.11.2
RUN mkdir -p /opt/burp \
  && wget -q -O /opt/burp/burpsuite_community.jar \
    "https://portswigger-cdn.net/burp/releases/download?product=community&version=${BURP_VERSION}&type=Jar"
ARG GHIDRA_VERSION=11.2.1
RUN wget -q -O /tmp/ghidra.zip \
    "https://github.com/NationalSecurityAgency/ghidra/releases/download/Ghidra_${GHIDRA_VERSION}_build/ghidra_${GHIDRA_VERSION}_PUBLIC_20241105.zip" \
  && unzip -q /tmp/ghidra.zip -d /opt \
  && mv /opt/ghidra_* /opt/ghidra \
  && rm -f /tmp/ghidra.zip
COPY deploy/engage/docker/wrappers/ /usr/local/bin/
RUN chmod +x /usr/local/bin/burpsuite /usr/local/bin/ghidra /usr/local/bin/hashcat \
    /usr/local/bin/john /usr/local/bin/gdb /usr/local/bin/metasploit /usr/local/bin/angr \
    /usr/local/bin/radare2 /usr/local/bin/volatility /usr/local/bin/wpscan
USER 10001
