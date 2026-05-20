# pkg/exec

Shared subprocess layer for Engage (and optional discovery). Veil stack overview: [README.md](../../README.md#architecture).

Cross-layer subprocess execution: legacy `docker exec` sandbox ([sandbox.go](sandbox.go)) when `ENGAGE_EXECUTION_PROFILE=docker-exec`; default **client-native** uses host `runLocal` only. Optional `ENGAGE_PATH_EXTRA`, timeouts, and process tracking.

When `ENGAGE_EXECUTION_PROFILE=client-native`, `Executor.Run` **never** uses the docker sandbox even if misconfigured (defense in depth); subprocesses use the host `PATH` (after `filterEnv` / `mergeEngagePathExtra`).

## When to use

- **Engage** tool catalog runs (allowlisted binaries, audit hooks).
- **Discovery** rare subprocess work (e.g. `git clone`, headless fetcher) behind build tag `discoveryexec` and a dedicated fetcher container profile.

## When not to use

- **Plain HTTP / GitHub raw feeds** — use `discovery/harvest` `feeds.Client` and `discovery/pkg/githubraw` (ledger-backed GET). No subprocess, no sandbox overhead.
- **Pipeline / graph / knowledge** — these layers should not spawn catalog tools; they consume NATS/harvest envelopes only.
- **Discovery browser crawl** — Playwright sidecar under `discovery/cmd/browser-agent`; HTTP + harvest publish in `discovery/browser` (engage proxies via `DISCOVERY_BROWSER_URL`).

`pkg/exec` must not import `discovery/`, `pipeline/`, `knowledge/`, or `engage/`.
