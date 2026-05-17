# pkg/exec

Cross-layer subprocess execution: sandboxed `docker exec`, local runs with filtered env, timeouts, and optional process tracking.

## When to use

- **Engage** tool catalog runs (allowlisted binaries, audit hooks).
- **Discovery** rare subprocess work (e.g. `git clone`, headless fetcher) behind build tag `discoveryexec` and a dedicated fetcher container profile.

## When not to use

- **Plain HTTP / GitHub raw feeds** — use `discovery/harvest` `feeds.Client` and `discovery/pkg/githubraw` (ledger-backed GET). No subprocess, no sandbox overhead.
- **Pipeline / graph / knowledge** — these layers should not spawn catalog tools; they consume NATS/harvest envelopes only.
- **Engage browser tools** — stay on the browser sidecar HTTP path in `engage/serve/internal/runner` until P8g moves browser to discovery.

`pkg/exec` must not import `discovery/`, `pipeline/`, `knowledge/`, or `engage/`.
