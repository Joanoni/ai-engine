# Workspace

The workspace is the directory where you run `ai-engine`. All engine state lives under `.ai-engine/` inside this directory.

## Folder Structure

```
.ai-engine/
├── .env                            ← API keys (never committed)
├── config.json                     ← engine configuration
├── model-pricing.json              ← token cost table per model
├── engine_context.md               ← Layer 1: shared instructions for all agents
├── role_swarmito.md                ← Layer 2: role template for the root orchestrator
├── role_leader.md                  ← Layer 2: role template for leader agents
├── role_executor.md                ← Layer 2: role template for executor agents
│
├── agents/                         ← agent definitions (filesystem = hierarchy)
│   └── swarmito/
│       ├── agent.json              ← optional: description, model
│       ├── system_prompt.md        ← Layer 2: agent role
│       ├── backend-leader/
│       │   ├── agent.json
│       │   ├── system_prompt.md
│       │   └── backend-executor/
│       │       ├── agent.json
│       │       └── system_prompt.md
│       └── frontend-leader/
│           ├── agent.json
│           ├── system_prompt.md
│           └── frontend-executor/
│               └── system_prompt.md
│
├── memory/                         ← persistent agent memory (injected into all agent prompts)
│   └── *.md
│
├── sessions/                       ← runtime: session persistence
│   └── {session-id}/
│       ├── meta.json
│       ├── events.jsonl
│       └── tokens.json
│
├── chats/                          ← runtime: task context per agent per session
│   └── {session-id}/
│       └── {agent-name}/
│           └── task_context.md
│
├── logs/                           ← runtime: JSONL chat logs
│   └── {session-id}/
│       └── {agent-name}/
│           └── chat.jsonl
│
└── tokens.json                     ← runtime: project-level token aggregate
```

---

## `config.json`

Full field reference:

| Field | Type | Required | Description |
|---|---|---|---|
| `provider` | string | yes | LLM provider. Currently only `"anthropic"` is supported. |
| `default_model` | string | yes | Model name used by agents that do not specify their own `model` in `agent.json`. Example: `"claude-sonnet-4-6"`. |
| `port` | number | yes | HTTP/WebSocket server port. Default: `8080`. |
| `max_tool_retries` | number | no | Maximum consecutive tool errors before the session terminates. Default: `3`. |
| `max_tool_calls` | number | no | Maximum total tool calls across the entire session before termination. Default: `50`. |
| `dynamic_context.providers` | string[] | no | List of dynamic context provider names to enable. Supported values: `"workspace_tree"`, `"memory"`. |

Example:

```json
{
  "provider": "anthropic",
  "default_model": "claude-sonnet-4-6",
  "port": 8080,
  "max_tool_retries": 3,
  "max_tool_calls": 50,
  "dynamic_context": {
    "providers": ["workspace_tree", "memory"]
  }
}
```

---

## `memory/`

All `.md` files placed in `.ai-engine/memory/` are injected into every agent's system prompt as Layer 4 dynamic context. The content is recomputed live before every LLM call — a file written in turn N is already present in turn N+1.

- Files are sorted alphabetically by filename and rendered as a `# Agent Memory` Markdown block.
- Memory is **global per workspace** — shared across all agents and sessions.
- Managed via the `write_memory`, `update_memory`, and `delete_memory` tools.
- The directory is created automatically by `ai-engine init` and by `write_memory` on first use.
- To disable memory injection, remove `"memory"` from `dynamic_context.providers` in `config.json`.

---

## `agent.json`

Each agent directory may contain an optional `agent.json`. Only two fields are meaningful:

| Field | Type | Required | Description |
|---|---|---|---|
| `description` | string | no | Short description of the agent's role. Used in the auto-generated Layer 3 Team section for the parent leader. Falls back to the agent name if absent. |
| `model` | string | no | Override the model for this specific agent. If absent, `default_model` from `config.json` is used. |

Any other fields (`name`, `type`, `team`) are silently ignored for backward compatibility.

Example:

```json
{
  "description": "Implements the Go REST API backend",
  "model": "claude-opus-4"
}
```

---

## `model-pricing.json`

Maps model names to token costs. Used by `internal/pricing` to compute `estimated_cost_usd` in token reports.

Structure:

```json
{
  "claude-sonnet-4-6": {
    "input_per_million": 3.00,
    "output_per_million": 15.00,
    "currency": "USD"
  },
  "claude-opus-4": {
    "input_per_million": 15.00,
    "output_per_million": 75.00,
    "currency": "USD"
  }
}
```

| Field | Type | Description |
|---|---|---|
| `input_per_million` | number | Cost in `currency` per 1,000,000 input tokens |
| `output_per_million` | number | Cost in `currency` per 1,000,000 output tokens |
| `currency` | string | Currency code (e.g. `"USD"`) |

If a model is not found in the pricing map, `CalcCost` returns `0` and logs a warning.

---

## Role Template Files

Three role template files provide generic behavioral guidelines injected as Layer 2 of the system prompt, before the agent-specific `system_prompt.md`:

| File | Applied to |
|---|---|
| `role_swarmito.md` | The root orchestrator (`swarmito`) |
| `role_leader.md` | Any agent whose directory contains child agents (leaders) |
| `role_executor.md` | Any agent with no child agents (executors) |

These files are created by `ai-engine init` and can be customised per workspace. If a file does not exist, the layer is silently skipped.

---

## `system_prompt.md`

Each agent's `system_prompt.md` defines its role (Layer 3 of the system prompt). There is no enforced schema — write it as plain Markdown. Recommended structure:

```markdown
# Role

You are [agent name], responsible for [responsibility].

## Responsibilities

- [responsibility 1]
- [responsibility 2]

## Guidelines

- [guideline 1]
- [guideline 2]
```

Keep it focused on the agent's specific domain. The engine automatically injects the role template (Layer 2), team roster (Layer 4), workspace file tree (Layer 5), and task context (Layer 6) — you do not need to describe those in `system_prompt.md`.

---

## Agent Hierarchy Rule

The engine determines agent type purely from the filesystem:

- A directory under `.ai-engine/agents/` that contains **subdirectories with `system_prompt.md`** → **leader**
- A directory that has **no such subdirectories** → **executor**

No `type` field is needed. The hierarchy can be arbitrarily deep.

---

## Runtime-Generated Files

These files and directories are created by the engine at runtime. Do not edit them manually while a session is running.

| Path | Description |
|---|---|
| `.ai-engine/sessions/{id}/meta.json` | Session metadata: id, prompt, startedAt, finishedAt, status |
| `.ai-engine/sessions/{id}/events.jsonl` | All events for the session, one JSON object per line |
| `.ai-engine/sessions/{id}/tokens.json` | Per-session token usage and estimated cost |
| `.ai-engine/chats/{id}/{agent}/task_context.md` | Task context written by parent leader before invoking child |
| `.ai-engine/logs/{id}/{agent}/chat.jsonl` | Full JSONL chat log for the agent (see `docs/chat-log-format.md`) |
| `.ai-engine/tokens.json` | Project-level aggregate token usage across all sessions |

---

## Hot-Reload

The engine re-reads the entire `.ai-engine/` directory on every session start. This includes:

- `config.json` — port, model, limits
- `engine_context.md` — Layer 1 shared instructions
- All `agent.json` and `system_prompt.md` files — agent definitions and roles

Changes take effect on the next session without restarting the binary.
