# Workspace Structure

## Status

🟢 Defined — structure approved. Ready for implementation.

---

## Overview

The AI Engine operates inside a **workspace** — any directory on the user's machine. All engine configuration, agent definitions, and runtime data are stored inside a single `.ai-engine/` folder within the workspace. This keeps the engine self-contained and avoids polluting the user's project.

The engine **reads configuration on every interaction**, enabling hot-reload of agent definitions and settings without restarting the engine.

---

## Folder Structure

```
{workspace}/
├── (user's project files...)
└── .ai-engine/
    ├── config.json                  # Global engine configuration
    ├── .env                         # Optional — ANTHROPIC_API_KEY (not committed)
    ├── engine_context.md            # Optional — Layer 1 of the 3-layer system prompt (injected into all agents)
    ├── agents/
    │   ├── swarmito/
    │   │   ├── agent.json           # Agent metadata and configuration
    │   │   └── system_prompt.md     # Agent system prompt — Layer 2 (role)
    │   ├── leader-a/
    │   │   ├── agent.json
    │   │   └── system_prompt.md
    │   ├── executor-d/
    │   │   ├── agent.json
    │   │   └── system_prompt.md
    │   └── (one folder per agent...)
    ├── chats/                       # Generated at runtime — do not edit manually
    │   └── {session-id}/
    │       ├── swarmito/
    │       │   ├── tasks.md         # Task file for this agent in this session
    │       │   └── task_context.md  # Layer 3 of system prompt — written by set_task_context tool
    │       └── leader-a/
    │           ├── tasks.md
    │           └── task_context.md
    ├── history/                     # Archived task_context.md files after agent finishes
    │   └── {session-id}/
    │       └── {agent-name}/
    │           └── task_context.md
    └── logs/                        # Chat logs — one file per agent per session
        └── {session-id}/
            └── {agent-name}/
                └── chat.jsonl       # JSONL log of all LLM turns and tool calls
```

---

## File Specifications

### `config.json`

Global engine settings. Applied to all agents unless overridden at the agent level.

```json
{
  "provider": "anthropic",
  "default_model": "claude-sonnet-4-5",
  "root_agent": "swarmito",
  "port": 8080,
  "max_tool_retries": 3,
  "max_tool_calls": 50
}
```

| Field | Type | Description |
|---|---|---|
| `provider` | string | LLM provider to use (`anthropic`, extensible to others) |
| `default_model` | string | Default model identifier for all agents |
| `root_agent` | string | Name of the root agent (entry point of the tree) |
| `port` | number | WebSocket server port. Default: `8080` |
| `max_tool_retries` | number | Max consecutive tool errors before session terminates. Default: `3` |
| `max_tool_calls` | number | Max total tool calls per agent before session terminates (runaway loop guard). Default: `50` |

---

### `agents/{name}/agent.json`

Defines the agent's metadata, role in the tree, and optional model override.

```json
{
  "name": "leader-a",
  "type": "leader",
  "team": ["executor-d", "executor-e", "executor-f"],
  "model": "claude-opus-4"
}
```

| Field | Type | Description |
|---|---|---|
| `name` | string | Unique agent identifier. Must match the folder name. |
| `type` | string | `leader` or `executor` |
| `team` | string[] | List of agent names directly below this agent. Empty `[]` for executors. |
| `model` | string | *(optional)* Overrides the global `default_model` for this agent. |

---

### `agents/{name}/system_prompt.md`

The agent's system prompt written in Markdown. Loaded and sent as the `system` parameter on every LLM call.

The prompt must follow this structure:

```markdown
# [Leader-only section]
## Team
- **executor-d**: brief description of this agent's role
- **executor-e**: brief description of this agent's role

# Agent Description
Who this agent is and what it does.

# Skills
(Following Anthropic's skill definition pattern)

# Task
What this agent is expected to accomplish.
```

> The engine does **not** inject the workspace path into the system prompt. Agents must never reference absolute paths.

---

### `chats/{session-id}/{agent-name}/tasks.md`

Generated at runtime by leader agents using the `create_task_file` / `update_task_file` tools. Contains the task list for a specific agent in a specific session.

Format: Markdown checklist.

```markdown
# Tasks — leader-a — session abc123

- [x] Task 1: scaffold the project structure
- [-] Task 2: implement the data model (in progress)
- [ ] Task 3: write unit tests
```

> This file is written by the agent and read by the engine for observability. It is also streamed to the frontend via WebSocket events.

---

## Design Principles

1. **Hot-reload:** The engine reads `.ai-engine/` on every interaction. Changes to agent definitions or config take effect immediately without restarting.
2. **Separation of concerns:** Agent metadata (`agent.json`) is separate from agent behavior (`system_prompt.md`).
3. **JSON for config:** All structured data uses JSON for easy serialization to the frontend via WebSocket.
4. **Markdown for prompts and tasks:** Human-readable, easy to edit, and renderable in the frontend.
5. **Path isolation:** Agents never see or reference the workspace absolute path. The engine resolves all paths internally.
