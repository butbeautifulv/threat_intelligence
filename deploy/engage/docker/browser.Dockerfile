# syntax=docker/dockerfile:1
# Playwright + Chromium browser sidecar for web catalog tools.
FROM mcr.microsoft.com/playwright:v1.49.1-jammy
WORKDIR /app
COPY engage/serve/cmd/browser-agent/package.json engage/serve/cmd/browser-agent/index.mjs ./
RUN npm install --omit=dev && npx playwright install chromium
ENV ENGAGE_BROWSER_LISTEN=:8910
EXPOSE 8910
HEALTHCHECK CMD curl -fsS http://127.0.0.1:8910/health || exit 1
CMD ["node", "index.mjs"]
