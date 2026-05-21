# Architecture

## Core Concepts

**Workspace** — the directory where the user runs `ai-engine`. All agent definitions, configuration, logs, and session data live under `.ai-engine/` inside this directory. The engine never reads or writes outside the workspace.

**Agent tree** — a hierarchy of agents defined by the filesystem structure under `.ai-engine/agents/`. The tree is loaded into memory once per session start as a tree of `AgentNode` structs. Swarmito is always the root.

**Session lifecycle** — a session begins when the frontend sends a `user.message` WebSocket event. The engine generates a UUID v4 session ID, starts the root agent (Swarmito), and streams all events back to the frontend. The session ends when Swarmito calls `finish_work` or a terminal error occurs. All session data is persisted to `.ai-engine/sessions/{id}/`.

---

## System Diagram

```
Browser
  │
  │  WebSocket ws://localhost:{port}/ws
  │  HTTP      http://localhost:{port}/...
  ▼
┌─────────────────────────────────────────────────────┐
│  internal/server  (HTTP + WebSocket)                │
│  ┌─────────────┐   ┌──────────────────────────────┐ │
│  │  Event Bus  │◄──│  Per-connection session mutex │ │
│  └──────┬──────┘   └──────────────────────────────┘ │
└─────────┼───────────────────────────────────────────┘
          │ Publish / Subscribe
          ▼
┌─────────────────────────────────────────────────────┐
│  internal/agent  (Runner)                           │
│                                                     │
│  6-layer system prompt composition                  │
│  Tool dispatch loop                                 │
│  Token tracking + nudge limit                       │
│                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────┐ │
│  │  LLM Provider│  │Tool Executor │  │  Sandbox  │ │
│  │  (Anthropic) │  │  (Registry)  │  │  (Shell)  │ │
│  └──────────────┘  └──────────────┘  └───────────┘ │
└─────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────┐
│  Support packages                                   │
│  internal/chatlog      internal/sessionstore        │
│  internal/tokenstore   internal/pricing             │
│  internal/registry     internal/dyncontext          │
└─────────────────────────────────────────────────────┘
```

---

## Go Packages

### `cmd/ai-engine`

Entry point. Parses the subcommand (`init` or none), resolves the workspace path (current working directory), instantiates all dependencies, and wires them together before starting the server. Injects the version string (set at compile time via `-ldflags "-X main.Version=..."`) into the server.

### `internal/config`

Loads `.ai-engine/config.json` and `.ai-engine/.env`. The `.env` parser strips surrounding single/double quotes from values. Validates that `provider` and `default_model` are non-empty. Config is re-read on every session start (hot-reload) — no restart required after editing `.ai-engine/`.

### `internal/registry`

Builds the in-memory `AgentNode` tree from the filesystem. Key functions:

- `LoadTree() (*AgentNode, error)` — recursively walks `.ai-engine/agents/swarmito/`, reads `agent.json` (optional), and builds the full tree. A directory with subdirectories containing `system_prompt.md` becomes a leader node; without such subdirectories it becomes an executor node.
- `FindNode(root *AgentNode, name string) *AgentNode` — BFS search by agent name.
- `LoadAgentTree() ([]AgentTreeNode, error)` — returns a flat list of all nodes for the `/agents` HTTP endpoint.
- `LoadEngineContext() (string, error)` — reads `.ai-engine/engine_context.md`.
- `LoadRoleTemplate(agentType, agentName) (string, error)` — reads the appropriate `role_swarmito.md`, `role_leader.md`, or `role_executor.md` from `.ai-engine/`. Returns `("", nil)` if the file does not exist.

### `internal/llm`

Defines the `LLMProvider` interface and shared types (`Request`, `Response`, `Message`, `ContentBlock`, `ToolDefinition`, `ToolCall`). `ToolDefinition.InputSchema` is `json.RawMessage`.

#### `internal/llm/anthropic`

Implements `LLMProvider` for the Anthropic Messages API. Uses `MaxOutputTokens` from `llm.Request` with a fallback of 16000.

### `internal/agent`

- `Agent` struct — holds the `*registry.AgentNode`, session ID, `RoleTemplate` string, and associated runtime dependencies.
- `Runner` — the execution loop. On each turn: composes the 6-layer system prompt, calls `provider.Send()`, dispatches all tool calls in the response as a batch, collects all results before any termination check, then checks retry/finish conditions. Tracks `consecutiveErrors`, `consecutiveNudges`, and `totalToolCalls`.

### `internal/chatlog`

JSONL logger. One file per agent per session at `.ai-engine/logs/{session}/{agent}/chat.jsonl`. Thread-safe (`sync.Mutex`). All writes are non-fatal — logging errors are printed to stderr and never terminate the session. See [`docs/chat-log-format.md`](./chat-log-format.md) for the full format reference.

### `internal/dyncontext`

Defines the `DynamicContextProvider` interface. Two providers are built in:

- **`WorkspaceTreeProvider`** — walks the workspace directory tree (max depth 6), ignores `.git`, `node_modules`, `vendor`, `dist`, `build`, `__pycache__`, `.venv`, `venv`, `.ai-engine`, and returns a Markdown-formatted file tree.
- **`MemoryProvider`** — reads all `.md` files from `.ai-engine/memory/` sorted by filename and renders them as a `# Agent Memory` Markdown block. Returns empty string if the directory does not exist or contains no `.md` files. All I/O errors are non-fatal (logged to stderr, empty string returned).

Both providers are recomputed before every LLM call (live recomputation). The `Registry` filters providers by the `dynamic_context.providers` list in `config.json`; if the list is empty, all providers are enabled.

### `internal/pricing`

Reads `.ai-engine/model-pricing.json`. `CalcCost(model, inputTokens, outputTokens)` returns the estimated cost in USD. Logs a warning when the model is not found in the pricing map.

### `internal/tokenstore`

Persists token usage to two locations:
- `.ai-engine/tokens.json` — project-level aggregate (all sessions)
- `.ai-engine/sessions/{id}/tokens.json` — per-session totals

### `internal/sessionstore`

Persists session metadata and events:
- `.ai-engine/sessions/{id}/meta.json` — id, prompt, startedAt, finishedAt, status
- `.ai-engine/sessions/{id}/events.jsonl` — one JSON event per line

The `events.jsonl` file handle is opened at `StartSession` and kept open for the duration of the session (closed at `FinishSession`), eliminating per-event open/close overhead. Thread-safe via `sync.Mutex`.

### `internal/tools`

Defines the `Tool` interface (`Name()`, `Description()`, `InputSchema()`, `Execute(ctx, input)`). All tool implementations live here. `Registry` maps tool names to handlers and enforces the leader/executor tool sets — executors cannot call leader tools even if the LLM requests them.

**Leader tools:** `create_chat`, `set_task_context`, `create_task_file`, `update_task_file`, `list_files`, `read_file`, `finish_work`, `write_memory`, `update_memory`, `delete_memory`

**Executor tools:** `run_terminal_command`, `list_files`, `read_file`, `write_file`, `apply_diff`, `search_files`, `delete_file`, `finish_work`, `write_memory`, `update_memory`, `delete_memory`

### `internal/sandbox`

Path isolation and shell management.

**`sandbox.go`** — `Sandbox` type. All agent file paths are resolved relative to the workspace root. Uses `filepath.Rel` + `strings.HasPrefix(rel, "..")` to detect traversal attempts — correctly handles Windows case-insensitive paths.

**`shell.go`** — `Shell` type. A persistent `cmd.exe` (Windows) or `sh` (Unix) process that lives for the duration of a single agent execution. Uses sentinel-based output reading (`echo __AI_ENGINE_CMD_DONE_7f3a9b2c__`) to detect end of command output. Timeout via `time.After`. `Close()` is safe to call multiple times.

**`shell_windows.go`** — Windows Job Object (`JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`). Assigns the shell process to the job on creation; closing the job handle kills the entire process tree, including background processes started with `&`.

**`shell_unix.go`** — No-op stubs for `attachJobObject`/`closeJobObject`. On Unix, `exec.CommandContext` with `SIGKILL` handles process group cleanup.

### `internal/session`

UUID v4 session ID generation.

### `internal/events`

`Event` struct with `Type`, `SessionID`, `AgentName`, `Timestamp` (auto-filled by `Publish()` as RFC3339 UTC), and `Payload`. `Bus` supports `Subscribe(handler) SubscriptionID` and `Unsubscribe(id)` — handlers are stored in a `map[SubscriptionID]Handler`. `Publish` is non-blocking toward subscribers (uses `select { case: default: }` for the WebSocket forwarding handler).

### `internal/scaffold`

Implements `ai-engine init`. Writes the `.ai-engine/` skeleton using embedded templates. `writeNew` is idempotent — skips files that already exist.

### `internal/server`

HTTP + WebSocket server. Serves the embedded React frontend at `/`. All HTTP endpoints return JSON with `Access-Control-Allow-Origin: *`. Per-connection `sessionMu sync.Mutex` + `sessionActive bool` prevents concurrent sessions on the same WebSocket connection. Graceful shutdown via `signal.NotifyContext` on SIGINT/SIGTERM.

---

## Key Design Decisions

### 1. Workspace path isolation

All agent file operations are resolved relative to the workspace root via `filepath.Rel`. Agents never see absolute paths. Traversal attempts (`../`) are rejected by the sandbox.

### 2. Tool-only execution

All agent actions are explicit tool calls. A plain text LLM response (no tool calls, `stop_reason != "end_turn"`) triggers a nudge message injected into the conversation. After 5 consecutive nudges without tool calls, the session terminates.

### 3. Hot-reload

`.ai-engine/` is re-read on every session start. Changes to `config.json`, agent definitions, `system_prompt.md` files, and `engine_context.md` take effect immediately without restarting the binary.

### 4. Provider abstraction

The `LLMProvider` interface decouples the agent runner from any specific LLM API. V1 implements Anthropic. Adding a new provider requires only implementing the interface.

### 5. Sequential execution

Leaders process sub-agents sequentially (one `create_chat` at a time). There is no parallel agent execution in V1.

### 6. 6-layer system prompt

Every LLM call composes a system prompt from six layers in order. See [`docs/agents.md`](./agents.md) for the full layer specification.

### 7. Persistent shell per agent

Each agent execution gets one persistent shell process. `cd` commands persist between `run_terminal_command` calls within the same agent. The Windows Job Object ensures all child processes (including backgrounded ones) are killed when the agent finishes.

### 8. Graceful shutdown

`signal.NotifyContext` on SIGINT/SIGTERM cancels the root context, which propagates to all in-flight agent goroutines via `context.Context`.

### 9. Batch tool result protocol

The Anthropic API requires every `tool_use` block in a response to have a corresponding `tool_result` in the next message. The runner collects all tool results before performing any termination check. If `finish_work` is called in the middle of a batch, remaining tool calls receive `"skipped: finish_work already called"` as their result.
