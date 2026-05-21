# Agents

## Agent Types

There are two agent types: **leader** and **executor**. The type is derived automatically from the filesystem — no `type` field is needed in `agent.json`.

| Type | Filesystem rule | Role |
|---|---|---|
| **leader** | Directory contains subdirectories that each have a `system_prompt.md` | Coordinates a team of sub-agents |
| **executor** | Directory has no subdirectories with `system_prompt.md` (leaf node) | Performs concrete tasks using tools |

**Swarmito** is the root leader — the only agent that communicates directly with the user. All other agents are invoked by their parent leader via the `create_chat` tool.

---

## 6-Layer System Prompt

Every LLM call composes a system prompt from six layers, assembled in this order:

| Layer | Name | Source | Scope |
|---|---|---|---|
| 1 | Engine Context | `.ai-engine/engine_context.md` | All agents |
| 2 | Role Template | `.ai-engine/role_swarmito.md` / `role_leader.md` / `role_executor.md` | Matched by agent type |
| 3 | Agent Role | `.ai-engine/agents/{name}/system_prompt.md` | This agent only |
| 4 | Team | Auto-generated from `AgentNode.Children` | Leaders only (empty for executors) |
| 5 | Dynamic Context | Workspace file tree (recomputed per LLM call) | All agents |
| 6 | Task Context | `.ai-engine/chats/{session-id}/{agent}/task_context.md` | Written by parent leader via `set_task_context` before `create_chat` |

### Layer 4 — Team section format

For leaders, Layer 3 is generated as:

```markdown
## Team
- **backend-executor**: Implements the Go REST API backend
- **frontend-executor**: Implements the single-file HTML frontend
```

The description comes from `agent.json`. If `description` is absent, the agent name is used as the fallback.

Executors receive an empty string for Layer 4.

### Layer 5 — Dynamic Context

Layer 4 is composed by the `dyncontext.Registry`, which runs all enabled providers in order and concatenates their output separated by `---`. Two providers are built in:

**`WorkspaceTreeProvider`** (`"workspace_tree"`) — walks the workspace directory tree before every LLM call and injects the current file tree as Markdown. Maximum depth: 6. Ignored directories: `.git`, `node_modules`, `vendor`, `dist`, `build`, `__pycache__`, `.venv`, `venv`, `.ai-engine`. The output includes an explicit instruction not to call `list_files` for directories already visible in the tree.

**`MemoryProvider`** (`"memory"`) — reads all `.md` files from `.ai-engine/memory/` sorted by filename and renders them as a `# Agent Memory` Markdown block. Returns empty string if the directory does not exist or contains no `.md` files. All I/O errors are non-fatal. A file written by `write_memory` in turn N is already present in the system prompt in turn N+1.

Both providers are recomputed before every LLM call. Enabled providers are controlled by `dynamic_context.providers` in `config.json`.

### Layer 6 — Task Context

Written by the parent leader using `set_task_context` before calling `create_chat`. Contains the specific task description and any relevant context for the sub-agent. Stored at `.ai-engine/chats/{session-id}/{agent-name}/task_context.md`.

---

## Tool Sets

### Leader tools

| Tool | Description |
|---|---|
| `create_chat` | Invoke a direct child agent by name with a message |
| `set_task_context` | Write task context for a child agent before invoking it |
| `create_task_file` | Create a Markdown checklist file to track tasks |
| `update_task_file` | Update the task checklist (requires prior `create_task_file`) |
| `list_files` | List files in a directory |
| `read_file` | Read a file (with optional offset/limit) |
| `finish_work` | Signal completion and return a result to the parent |
| `write_memory` | Write or overwrite a Markdown file in `.ai-engine/memory/` |
| `update_memory` | Apply search/replace blocks to an existing memory file (same format as `apply_diff`) |
| `delete_memory` | Delete a file from `.ai-engine/memory/` |

### Executor tools

| Tool | Description |
|---|---|
| `run_terminal_command` | Execute a shell command in the persistent shell |
| `list_files` | List files in a directory |
| `read_file` | Read a file (with optional offset/limit) |
| `write_file` | Write content to a file |
| `apply_diff` | Apply a search/replace diff to a file |
| `search_files` | Search for a regex pattern across files |
| `delete_file` | Delete a file |
| `finish_work` | Signal completion and return a result to the parent |
| `write_memory` | Write or overwrite a Markdown file in `.ai-engine/memory/` |
| `update_memory` | Apply search/replace blocks to an existing memory file (same format as `apply_diff`) |
| `delete_memory` | Delete a file from `.ai-engine/memory/` |

Tool access is enforced by `internal/tools.Registry` — executors cannot call leader tools even if the LLM requests them.

---

## Memory

The memory system provides cross-session persistence for agents. All `.md` files in `.ai-engine/memory/` are injected into every agent's system prompt as part of Layer 4 on every LLM turn.

### Memory tools

| Tool | Available to | Description |
|---|---|---|
| `write_memory` | Leaders + Executors | Creates or overwrites a `.md` file in `.ai-engine/memory/`. The `filename` must not contain path separators; `.md` is appended automatically if missing. |
| `update_memory` | Leaders + Executors | Applies one or more search/replace blocks to an existing memory file without rewriting it entirely. Same `diff` format as `apply_diff`. |
| `delete_memory` | Leaders + Executors | Deletes a file from `.ai-engine/memory/`. Non-fatal if the file does not exist. |

### Memory scope

Memory is **global per workspace** — shared across all agents and all sessions. A file written by `backend-executor` in session A is visible to `frontend-leader` in session B.

### Memory lifecycle

- Memory files persist indefinitely until explicitly deleted via `delete_memory`.
- The `.ai-engine/memory/` directory is created by `ai-engine init` and also by `write_memory` on first use.
- To disable memory injection, remove `"memory"` from `dynamic_context.providers` in `config.json`.

### When to use memory

Agents should write to memory when they:
- Make an architectural decision
- Establish a project convention (naming, folder structure, tech stack)
- Discover a non-obvious fact about the codebase
- Define an API contract, data schema, or interface
- Complete a significant unit of work

---

## Leader Lifecycle

```
Leader receives objective
        │
        ▼
create_task_file  ←── creates Markdown checklist
        │
        ▼
┌── for each task ──────────────────────────────┐
│                                               │
│   set_task_context  ←── write task details    │
│        │                                      │
│        ▼                                      │
│   create_chat  ←── invoke child agent         │
│        │                                      │
│        ▼                                      │
│   wait for child finish_work                  │
│        │                                      │
│        ▼                                      │
│   update_task_file  ←── mark task done        │
│                                               │
└───────────────────────────────────────────────┘
        │
        ▼
finish_work  ←── return result to parent
```

---

## Executor Lifecycle

```
Executor receives task (via task_context.md)
        │
        ▼
Read task context (read_file, list_files)
        │
        ▼
┌── tool loop ──────────────────────────────────┐
│                                               │
│   run_terminal_command / write_file /         │
│   apply_diff / search_files / etc.            │
│                                               │
└───────────────────────────────────────────────┘
        │
        ▼
finish_work  ←── return result to parent leader
```

---

## Full Execution Flow

```
User prompt (WebSocket user.message)
        │
        ▼
   swarmito  (root leader)
   ├── set_task_context → backend-leader
   ├── create_chat → backend-leader
   │       ├── set_task_context → backend-executor
   │       ├── create_chat → backend-executor
   │       │       └── run_terminal_command, write_file, ...
   │       │           finish_work → backend-leader
   │       └── update_task_file
   │           finish_work → swarmito
   │
   ├── set_task_context → frontend-leader
   ├── create_chat → frontend-leader
   │       ├── set_task_context → frontend-executor
   │       ├── create_chat → frontend-executor
   │       │       └── write_file, run_terminal_command, ...
   │       │           finish_work → frontend-leader
   │       └── update_task_file
   │           finish_work → swarmito
   │
   └── finish_work → session ends
```

---

## Error Handling

### Tool errors

Tool execution errors are returned to the agent as a `tool_result` with `is_error: true`. The agent can inspect the error and retry with a corrected call.

### Consecutive error limit

`max_tool_retries` consecutive tool errors (across all tools in a turn) cause the session to terminate with an error event. The default is `3` (configurable in `config.json`).

### Total tool call limit

`max_tool_calls` total tool calls across the entire session cause the session to terminate. The default is `50` (configurable in `config.json`).

### Nudge limit

If the LLM returns a response with no tool calls (plain text only), the runner injects a nudge message reminding the agent to use tools. After 5 consecutive nudges without tool calls, the session terminates.

### Batch protocol

All tool results in a batch are collected before any termination check. This satisfies the Anthropic API requirement that every `tool_use` block has a corresponding `tool_result` in the next message.

If `finish_work` is called in the middle of a batch (i.e., other tool calls appear after it in the same response), the remaining calls receive `"skipped: finish_work already called"` as their result, and the loop continues to collect all results before processing the finish.

### Infinite recursion guard

`create_chat` checks the agent call stack (passed via `context.Value`). If the target agent is already in the stack, the call returns an error — preventing `agent-a → agent-b → agent-a` loops.
