# Getting Started

This guide walks you through setting up AI Engine in a new project directory from scratch.

## Prerequisites

- `ai-engine` binary (from `bin\latest\ai-engine.exe` or a release)
- An Anthropic API key

## Step 1 — Scaffold the workspace

Place the `ai-engine` binary in your project directory and run:

```cmd
ai-engine init
```

This creates the `.ai-engine/` folder with the following structure:

```
.ai-engine/
├── .env                        ← API key goes here
├── config.json                 ← engine configuration
├── model-pricing.json          ← token cost table
└── agents/
    └── swarmito/
        ├── agent.json
        └── system_prompt.md
```

`ai-engine init` is idempotent — running it again on an existing workspace skips files that already exist.

## Step 2 — Set your API key

Open `.ai-engine/.env` and add your Anthropic API key:

```
ANTHROPIC_API_KEY=sk-ant-...
```

Alternatively, set `ANTHROPIC_API_KEY` as a system environment variable before running the binary.

## Step 3 — Configure the engine

Open `.ai-engine/config.json`. The scaffolded default looks like:

```json
{
  "provider": "anthropic",
  "default_model": "claude-sonnet-4-6",
  "port": 8080,
  "max_tool_retries": 3,
  "max_tool_calls": 50,
  "dynamic_context": {
    "providers": ["workspace_tree"]
  }
}
```

Adjust `port` if `8080` is already in use. Adjust `default_model` to any Claude model available on your API key. See [`docs/workspace.md`](./workspace.md) for the full field reference.

## Step 4 — Define your agents

Agents are defined by the filesystem hierarchy under `.ai-engine/agents/`. Each directory that contains a `system_prompt.md` is an agent. A directory with subdirectories that also contain `system_prompt.md` files is automatically a **leader**; a directory with no such subdirectories is an **executor**.

A minimal two-level tree looks like:

```
.ai-engine/agents/
└── swarmito/
    ├── agent.json
    ├── system_prompt.md          ← Swarmito's role
    ├── backend-leader/
    │   ├── system_prompt.md      ← backend leader role
    │   └── backend-executor/
    │       └── system_prompt.md  ← backend executor role
    └── frontend-leader/
        ├── system_prompt.md      ← frontend leader role
        └── frontend-executor/
            └── system_prompt.md  ← frontend executor role
```

Each `agent.json` is optional and may contain only `description` and/or `model`:

```json
{
  "description": "Implements the Go REST API backend",
  "model": "claude-opus-4"
}
```

If `model` is omitted, the agent uses `default_model` from `config.json`.

Write each `system_prompt.md` to describe the agent's role, responsibilities, and any domain-specific instructions. See [`docs/agents.md`](./agents.md) for the full agent model documentation.

## Step 5 — Start the server

From your project directory:

```cmd
ai-engine
```

The engine prints its startup log and begins listening:

```
workspace : /path/to/your/project
version   : 0.0.26
port      : 8080
```

## Step 6 — Open the frontend

Open your browser and navigate to:

```
http://localhost:8080
```

The embedded React frontend loads automatically. The connection status indicator in the Mission Panel shows **Connected** when the WebSocket handshake succeeds.

## Step 7 — Launch a mission

Type your objective in the prompt editor and click **Launch** (or press `Ctrl+Enter`). The engine:

1. Generates a new session ID (UUID v4).
2. Starts Swarmito with your prompt.
3. Swarmito decomposes the objective and delegates to sub-agents.
4. All events stream to the frontend in real time via WebSocket.
5. The session ends when Swarmito calls `finish_work`.

The Agent Graph panel shows the live status of every agent. The Live Terminal shows all events as they occur. The Task Progress panel tracks the Markdown checklist maintained by leaders.

## Next steps

- [`docs/agents.md`](./agents.md) — understand the agent model, tool sets, and execution flow
- [`docs/workspace.md`](./workspace.md) — full configuration reference
- [`docs/api.md`](./api.md) — WebSocket and HTTP API reference
- [`docs/architecture.md`](./architecture.md) — system internals
