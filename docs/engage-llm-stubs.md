# Engage `ai_*` tools — stub policy (P11b)

Legacy HexStrike MCP names several tools with an `ai_` prefix. **Veil engage does not call an external LLM for any of them today.** Behavior is deterministic: pattern-based payloads, ranked catalog tools, graph-backed assessment, or minimal JSON stubs. A future LLM provider hook is **out of scope** for the Go sign-off (see [engage_hexstrike_post_p10_signoff.plan.md](../.cursor/plans/archive/engage_hexstrike_post_p10_signoff.plan.md)).

## Dispatch order (HTTP + MCP)

All catalog tools share one chain in [`tooldispatch.Dispatch`](../engage/serve/internal/usecase/tooldispatch/dispatch.go):

1. Playbook by tool name
2. **`tryAgentTool`** — [`agent_tools.go`](../engage/serve/internal/usecase/tooldispatch/agent_tools.go)
3. Bridge workflow binary (`binary: ai`, …) when a handler exists
4. **`callIntelBridge`** — [`bridge_handlers.go`](../engage/serve/internal/usecase/tooldispatch/bridge_handlers.go) when `IsIntelBridgeTool` (includes entries in `intelBridgeHandlers`)
5. Subprocess runner (`tools.Runner`)

`ai_*` names never reach subprocess: catalog `binary: ai` is a workflow placeholder ([`engage-tools-na-matrix.md`](engage-tools-na-matrix.md) → `bridge_api`).

## Catalog `ai_*` tools

| Tool | Catalog `binary` | Today | Dispatch path | Implementation | Future LLM (out of scope) |
|------|------------------|-------|---------------|----------------|---------------------------|
| `ai_generate_payload` | `ai` | **Deterministic stub** | `tryAgentTool` | [`payloads.Generate`](../engage/serve/internal/usecase/payloads/) (buffer/cyclic/random; `note`: not LLM). Requires files manager (`ENGAGE_FILES_DIR`). | Model-generated exploit strings / context-aware payloads |
| `ai_generate_attack_suite` | `ai` | **Deterministic stub** | `tryAgentTool` | `Intel.CreateAttackChain` + optional sample payload from `payloads.Generate` (`note`: patterns + ranked tools, not LLM). | LLM-authored multi-step attack narratives |
| `ai_reconnaissance_workflow` | `ai` | **Minimal stub** | `tryAgentTool` | JSON: `tool`, `target`, `success`, `note` pointing to `intelligent_smart_scan` / `ai_generate_payload`. | Agentic recon planning + tool selection |
| `ai_test_payload` | `ai` | **Minimal stub** | `tryAgentTool` | Same minimal stub as recon workflow. | LLM-driven payload mutation / response analysis |
| `ai_vulnerability_assessment` | `ai` | **Deterministic intel bridge** | `callIntelBridge` → `intelBridgeHandlers` | [`AIVulnerabilityAssessment`](../engage/serve/internal/usecase/intelligence/graph_intel.go): target analysis, ranked tool selection, parallel catalog runs, veil-graph context (`assessment_mode`: `deterministic_ranked_scan`). | Natural-language vuln narrative / prioritization |

### Related non-`ai_*` entry points

| API / tool | Role vs `ai_*` |
|------------|----------------|
| `POST /api/payloads/generate` | Same generator as `ai_generate_payload` (HTTP, not MCP name) |
| `POST /api/intelligence/smart-scan` | Preferred substitute for recon-style workflows |
| `discover_attack_chains` | Intel bridge; attack-chain graph (no `ai_` prefix) |
| `intelligent_smart_scan` | Playbook / intelligence route (see [engage-legacy-parity.md](engage-legacy-parity.md)) |

## Executable matrix

[`isAgentToolName`](../engage/serve/internal/usecase/tooldispatch/executable_matrix.go) classifies all `ai_generate_*` plus `ai_reconnaissance_workflow` and `ai_test_payload` as **bridge** (no runner binary). `ai_vulnerability_assessment` is bridge via `intelBridgeHandlers`, not `tryAgentTool`.

## Verification

```bash
make test-engage-executable-matrix   # 158/158 callable (bridge + subprocess)
make test-engage-bridge-coverage     # intel bridge handlers incl. ai_vulnerability_assessment
```

Responses include `"note"` / `"assessment_mode"` fields where behavior is intentionally non-LLM — agents and golden tests should not treat these tools as model-backed.
