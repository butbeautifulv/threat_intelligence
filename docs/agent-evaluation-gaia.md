# Agent evaluation — GAIA benchmark

[GAIA](https://huggingface.co/datasets/gaia-benchmark/GAIA) measures **general AI assistants** (tooling, search, multi-step reasoning) with short, unambiguous answers. Veil uses GAIA for **agent/orchestrator quality**, separate from **engage tool parity** (HexStrike catalog) and **security hardening** audits.

| Concern | Artifact | CI default |
|---------|----------|------------|
| Pipeline smoke (no HF) | `eval/gaia/fixtures/pilot/` | `make test-agent-eval-pilot` |
| Full GAIA split | `eval/gaia/data/` (local, gitignored) | `workflow_dispatch` only |
| Security tools / MCP | `make test-engage-parity` | every engage PR |

## Dataset policy (required)

GAIA is **gated** on Hugging Face. You must:

1. Log in and **accept the dataset terms** on the [dataset page](https://huggingface.co/datasets/gaia-benchmark/GAIA).
2. Set `HF_TOKEN` (or `HUGGING_FACE_HUB_TOKEN`) locally.
3. **Never commit** validation/test rows, attachments, or answers to this repo (contamination / leakage).
4. **Do not reshare** splits outside a private/gated HF repo per dataset license.

Parquet layout (2025+): `metadata.parquet`, levels 1–3, columns `task_id`, `Question`, `Level`, `Final answer`, `file_name`, `file_path`.

## Repository layout

```
eval/gaia/
  schema/task.schema.json      # JSONL row shape
  fixtures/pilot/              # synthetic pilot (CI-safe)
  results/                     # generated metrics (gitignored outputs ok locally)
  data/                        # HF download cache (gitignored)
scripts/eval/gaia/
  run-pilot.sh                 # offline pilot run
  score.py                     # normalized exact match
  load_hf.py                   # optional gated download
  solvers/stub.sh              # pilot infra (100% on fixtures)
  solvers/mcp-engage.sh        # hook for real MCP agent (manual)
deploy/eval/gaia.env.example
```

## Commands

```bash
# Registry + pilot (no network, no HF)
make test-agent-eval-registry
make test-agent-eval-pilot

# Optional: download after accepting HF terms
cp deploy/eval/gaia.env.example deploy/eval/gaia.env
# edit HF_TOKEN
pip install huggingface_hub datasets
source deploy/eval/gaia.env
python3 scripts/eval/gaia/load_hf.py

# Run pilot scorer
./scripts/eval/gaia/run-pilot.sh
```

## Scoring

`scripts/eval/gaia/score.py` uses **normalized exact match** (strip, lowercase, light punctuation) aligned with common GAIA leaderboard practice. Predictions are JSONL: `{"task_id":"...","prediction":"..."}`.

## Wiring agents (MCP)

Production evaluation should call your **veil-engage** (and optionally **veil-graph**) MCP client:

1. Load question (+ attachment path under `GAIA_DATA_DIR` when `file_path` is set).
2. Let the agent use allowed tools only (secure profile for anything network-facing).
3. Emit one final short answer per `task_id`.
4. Score with `score.py`; archive metrics under `eval/gaia/results/` locally — do not publish private test answers.

`scripts/eval/gaia/solvers/mcp-engage.sh` is a **placeholder**; implement your client in a private runner or extend the script without weakening engage hardening.

## Controls and critic gate

| Control | Meaning |
|---------|---------|
| `VEIL-EVAL-001` | Pilot fixture + scorer present; CI can detect regressions in eval harness |
| `VEIL-EVAL-002` | GAIA data stays out of git; HF load is opt-in |

Critic/orchestrator should treat **GAIA level-1 pilot pass** as necessary for eval-infra changes, and **full GAIA scores** as release metrics (manual or dispatch workflow), not as a substitute for `make test-engage-hardening`.

## Related

- [external-agent-store.md](external-agent-store.md) — why third-party agent repos are not vendored into runtime
- [engage-agentic-threats.md](engage-agentic-threats.md) — MCP/tool threats during eval runs
- [mcp-agents.md](mcp-agents.md) — MCP setup
