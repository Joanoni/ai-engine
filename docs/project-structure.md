# Project Structure

## Status

🟢 Phase 4 complete — full stack operational. Binary v0.0.12. Agent graph fully functional with custom glassmorphism nodes, animated edges, correct node type mapping, and dynamic fitView centering.

---

## Overview

The AI Engine is a Go application. The project follows standard Go conventions with a clear separation between the public API surface (`internal/` packages) and the entry point (`cmd/`).

The Go module name is: `github.com/swarmit/ai-engine`

---

## Folder Structure

```
ai-engine/                          # Repository root
├── cmd/
│   └── ai-engine/
│       └── main.go                 # Entry point — starts WebSocket server, loads config
│
├── internal/
│   ├── config/
│   │   └── config.go               # Loads and parses .ai-engine/config.json
│   │
│   ├── registry/
│   │   └── registry.go             # Agent Registry — loads agent.json + system_prompt.md
│   │
│   ├── llm/
│   │   ├── provider.go             # LLMProvider interface definition
│   │   └── anthropic/
│   │       └── anthropic.go        # Anthropic implementation of LLMProvider
│   │
│   ├── agent/
│   │   ├── agent.go                # Agent type definitions (Leader, Executor)
│   │   ├── runner.go               # Agent execution loop (think → tool call → repeat)
│   │   └── chat.go                 # Chat session: message history management
│   │
│   ├── chatlog/
│   │   └── logger.go               # Chat logger — writes JSONL entries to .ai-engine/logs/{session}/{agent}/chat.jsonl
│   │
│   ├── tools/
│   │   ├── registry.go             # Tool registry — maps tool names to handlers; defines leader/executor tool sets
│   │   ├── tool.go                 # Tool interface definition
│   │   ├── finish_work.go          # finish_work tool (shared — leader + executor)
│   │   ├── create_chat.go          # create_chat tool (leader) — runs sub-agent via SubAgentRunner
│   │   ├── create_task_file.go     # create_task_file tool (leader)
│   │   ├── update_task_file.go     # update_task_file tool (leader)
│   │   ├── set_task_context.go     # set_task_context tool (leader) — writes Layer 3 of sub-agent system prompt
│   │   ├── run_terminal_command.go # run_terminal_command tool (executor) — workdir + timeout support
│   │   ├── list_files.go           # list_files tool (executor + leader)
│   │   ├── read_file.go            # read_file tool (executor + leader)
│   │   ├── write_file.go           # write_file tool (executor)
│   │   ├── apply_diff.go           # apply_diff tool (executor)
│   │   ├── search_files.go         # search_files tool (executor)
│   │   └── delete_file.go          # delete_file tool (executor)
│   │
│   ├── sandbox/
│   │   ├── sandbox.go              # Workspace Sandbox — resolves agent paths to workspace
│   │   ├── shell.go                # Shell — persistent cmd.exe/sh process per agent execution
│   │   ├── shell_windows.go        # Windows Job Object — kills entire process tree on close
│   │   └── shell_unix.go           # Unix stubs (no-op; Unix kills process group via SIGKILL)
│   │
│   ├── session/
│   │   └── session.go              # Session Manager — creates/tracks sessions
│   │
│   ├── events/
│   │   └── bus.go                  # Event Bus — publishes events to WebSocket clients
│   │
│   ├── scaffold/
│   │   ├── scaffold.go             # Init command — creates .ai-engine/ workspace skeleton
│   │   └── templates/
│   │       ├── engine_context.md        # Embedded template for engine_context.md
│   │       ├── swarmito_agent.json      # Embedded template for swarmito/agent.json
│   │       └── swarmito_system_prompt.md # Embedded template for swarmito/system_prompt.md
│   │
│   └── server/
│       └── server.go               # WebSocket + HTTP server — handles connections and routing
│
├── frontend/                       # React + TypeScript SPA — Frontend v2 "Mission Control" (branch frontend-v2-test2)
│   ├── src/
│   │   ├── main.tsx                # Vite entry point
│   │   ├── App.tsx                 # Root component — 3-column layout, keyboard shortcuts, ReactFlowProvider
│   │   ├── App.css                 # Minimal layout-only styles
│   │   ├── index.css               # CSS variables, global reset, keyframe animations
│   │   ├── components/
│   │   │   ├── layout/
│   │   │   │   ├── Sidebar.tsx         # Collapsible sidebar — session history + New Mission
│   │   │   │   ├── CockpitArea.tsx     # Center area — graph + resize handle + terminal
│   │   │   │   ├── MissionPanel.tsx    # Right panel — all mission controls
│   │   │   │   └── ResizeHandle.tsx    # Drag handle for vertical split
│   │   │   ├── graph/
│   │   │   │   ├── AgentGraph.tsx      # React Flow canvas — dot grid, MiniMap, Controls
│   │   │   │   ├── LeaderNode.tsx      # Custom node — glassmorphism, hexagonal badge, glow
│   │   │   │   ├── ExecutorNode.tsx    # Custom node — pill shape, circular badge
│   │   │   │   └── AnimatedEdge.tsx    # Custom edge — SVG particle animation
│   │   │   ├── terminal/
│   │   │   │   ├── LiveTerminal.tsx    # Terminal panel — scanline, auto-scroll, export .jsonl
│   │   │   │   └── TerminalLine.tsx    # Single terminal line — color-coded by event type
│   │   │   ├── mission/
│   │   │   │   ├── PromptEditor.tsx    # Auto-resize textarea, Ctrl+Enter to submit
│   │   │   │   ├── LaunchButton.tsx    # Countdown 3→2→1→🚀 before firing
│   │   │   │   ├── TaskProgress.tsx    # Markdown checklist parser + progress bar
│   │   │   │   ├── AgentRoster.tsx     # Live agent list with status + tool call count
│   │   │   │   └── QuickStats.tsx      # Tool calls, agents, live duration counter
│   │   │   └── drawers/
│   │   │       └── AgentDetailDrawer.tsx  # Slide-in drawer — per-agent event timeline
│   │   ├── hooks/
│   │   │   ├── useWebSocket.ts         # WebSocket + auto-reconnect + latency + sessionStartTime
│   │   │   ├── useAgentGraph.ts        # Derives { nodes, edges } from EngineEvent[] (dagre layout)
│   │   │   ├── useSessionHistory.ts    # localStorage session persistence (max 20)
│   │   │   ├── useResizable.ts         # Drag-to-resize vertical split ratio
│   │   │   └── useKeyboardShortcuts.ts # Global keyboard shortcuts (Ctrl+B, Escape)
│   │   └── types/
│   │       ├── events.ts               # TypeScript types for engine events
│   │       ├── graph.ts                # AgentNode, AgentEdge, AgentStatus, AgentType
│   │       └── session.ts              # SessionRecord for localStorage persistence
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts              # Dev proxy: /ws → ws://localhost:8080
│
├── .ai-engine/                     # NOT committed — created by user in their workspace
│   └── (see docs/workspace-structure.md)
│
├── go.mod
├── go.sum
├── README.md
├── embed.go                        # Embeds frontend/dist into the Go binary at build time
├── build.cmd                       # Windows build script: npm build → go build → outputs binary
└── projects/                       # Example workspaces (generated by agents)
    ├── todo-app/                   # Todo app — first real use case
    │   ├── ai-engine.exe           # Compiled binary for this workspace
    │   ├── .ai-engine/             # Engine config, agent definitions, runtime data
    │   ├── backend/                # Go REST API (generated by agents)
    │   └── frontend/               # HTML frontend (generated by agents)
    ├── notes-app/                  # Notes app — second example workspace
    │   ├── ai-engine.exe
    │   ├── .ai-engine/
    │   ├── backend/
    │   └── frontend/
    ├── quiz-app/                   # Quiz App — multiple-choice quiz with scoring and leaderboard
    │   ├── ai-engine.exe           # Binary v0.0.1+
    │   ├── .ai-engine/             # Config: claude-sonnet-4-6, port 8080, max_tool_calls 60
    │   ├── backend/                # Go REST API port 8082 — questions, sessions, leaderboard
    │   └── frontend/               # Single-file HTML — start screen, quiz flow, result, leaderboard
    ├── expense-tracker/            # Expense Tracker — personal finance tracker (binary v0.0.4, Frontend v2)
    │   ├── ai-engine.exe           # Binary v0.0.4 with Frontend v2 "Mission Control" embedded
    │   ├── .ai-engine/             # Config: claude-sonnet-4-6, port 8080, max_tool_calls 60
    │   ├── backend/                # Go REST API port 8083 — expenses CRUD + summary by category
    │   └── frontend/               # Single-file HTML — dark UI, expense form, filter, CSS bar chart
    └── kanban-board/               # Kanban Board — drag-and-drop task manager (binary v0.0.14, 4-level hierarchy)
        ├── ai-engine.exe           # Binary v0.0.14
        ├── .ai-engine/             # Config: claude-sonnet-4-6, port 8080, max_tool_calls 60
        ├── backend/                # Go REST API port 8081 — cards CRUD + move between columns
        └── frontend/               # Single-file HTML — dark UI, three columns, drag-and-drop
```

---

## Package Responsibilities

### `cmd/ai-engine`
Entry point. Responsibilities:
1. Use the current working directory (CWD) as the workspace path.
2. Dispatch subcommands: `init` (scaffold workspace), `help`, or default (start server).
3. Load `.ai-engine/.env` via `config.LoadEnv` (optional).
4. Instantiate all dependencies: `Sandbox`, `LLMProvider`, `Registry`, `SessionManager`, `EventBus`.
5. Load `config.json` for startup info and port.
6. Embed `frontend/dist` via `embed.go` and serve it at `/`.
7. Start the WebSocket server via `internal/server`. Listens for SIGINT/SIGTERM for graceful shutdown.

### `internal/config`
Parses `.ai-engine/config.json` into a `Config` struct. Called on every session start (hot-reload).

### `internal/registry`
Loads agent definitions from `.ai-engine/agents/{name}/`. Returns `AgentDefinition` structs containing metadata (`agent.json`) and system prompt (`system_prompt.md`). Called on demand — not cached — to support hot-reload.

### `internal/llm`
Defines the `LLMProvider` interface:
```go
type LLMProvider interface {
    Send(ctx context.Context, req Request) (Response, error)
}
```
`Request` contains: system prompt, message history, available tools.
`Response` contains: text content (if any) and tool calls (if any).

### `internal/llm/anthropic`
Implements `LLMProvider` using the Anthropic Messages API. Handles tool call serialization/deserialization per Anthropic's format.

### `internal/chatlog`
Writes structured JSONL logs to `.ai-engine/logs/{session-id}/{agent-name}/chat.jsonl`. Each line is a `LogEntry` with fields for role (`user`, `assistant`, `tool_result`, `error`, `finish`), turn number, tool call details, and results. The logger is opened at agent start and closed when the agent finishes. Errors are non-fatal — logging failures do not terminate the session.

### `internal/agent`
Core agent execution logic:
- `agent.go`: defines `Agent`, `AgentType` (leader/executor), `AgentDefinition`.
- `runner.go`: the agent execution loop — sends messages to LLM, processes tool calls, loops until `finish_work` is called. Composes the 3-layer system prompt on every LLM call. Enforces `max_tool_retries` (consecutive errors) and `max_tool_calls` (total calls) limits. Archives `task_context.md` to `.ai-engine/history/` on agent finish.
- `chat.go`: manages the message history for a single chat session between two entities.

### `internal/tools`
Each tool is a separate file implementing a common `Tool` interface:
```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() json.RawMessage   // JSON Schema for the tool's input
    Execute(ctx context.Context, input json.RawMessage) (string, error)
}
```
The `registry.go` maps tool names to `Tool` implementations and returns the tool list for a given agent type.

Leader tool set: `create_chat`, `set_task_context`, `create_task_file`, `update_task_file`, `list_files`, `read_file`, `finish_work`.

Executor tool set: `run_terminal_command`, `list_files`, `read_file`, `write_file`, `apply_diff`, `search_files`, `delete_file`, `finish_work`.

`SubAgentRunner` is a function type injected into `create_chat` to avoid a circular import between `tools` and `agent` packages.

### `internal/sandbox`
Wraps all filesystem and terminal operations. Receives a path or command from an agent (relative), prepends the workspace absolute path, and executes. Prevents path traversal attacks (e.g., `../../etc/passwd`).

### `internal/sandbox` — Shell

`shell.go` defines the `Shell` type: a persistent `cmd.exe` (Windows) or `sh` (Unix) process that stays alive for the duration of a single agent execution. All `run_terminal_command` calls within an agent share the same shell process, so working directory and environment variables are preserved between calls.

On Windows, `shell_windows.go` creates a **Windows Job Object** (`JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`) and assigns the shell process to it. When `Shell.Close()` is called (or the timeout kills the shell), the OS automatically terminates all child processes in the job tree — including background processes started with `&`. This eliminates the orphan process problem that caused ports to remain occupied after timeout.

On Unix, `shell_unix.go` provides no-op stubs — Unix already kills the process group correctly via `SIGKILL` when the parent is killed.

### `internal/session`
Creates sessions with unique IDs (UUID v4). Tracks active sessions. Each session holds a reference to its root agent chat.

### `internal/events`
Defines the `Event` struct and all event type constants. The `Bus` broadcasts events to all registered WebSocket connections. Thread-safe.

### `internal/scaffold`
Implements the `ai-engine init` subcommand. Creates the `.ai-engine/` workspace skeleton with minimum required files: `config.json`, `.env`, `engine_context.md`, `agents/swarmito/agent.json`, and `agents/swarmito/system_prompt.md`. Templates are embedded in the binary via `//go:embed`. Never overwrites existing files — safe to run in an already-initialised workspace.

### `internal/server`
Sets up the HTTP + WebSocket server using `net/http` + `gorilla/websocket`. Handles:
- `/ws` — WebSocket endpoint. Incoming `user.message` events create a new session and start agent execution in a goroutine. Outgoing events from the Event Bus are forwarded to the connected client via a buffered write channel (serialises concurrent writes).
- `/health` — HTTP health check endpoint (returns `200 ok`).
- `/` — Serves the embedded React frontend (`frontend/dist`).
- Hot-reloads `config.json` and `engine_context.md` on every `user.message`.
- Builds the `SubAgentRunner` closure (recursive) and `tools.Registry` per session.

---

## CLI

The binary is invoked as `ai-engine [command]`.

| Command | Behaviour |
|---|---|
| *(no argument)* | Starts the WebSocket + HTTP server using the current directory as workspace |
| `init` | Scaffolds a new `.ai-engine/` workspace skeleton in the current directory. Never overwrites existing files. Prints next-steps instructions on success. |
| `help`, `--help`, `-h` | Prints usage information and exits |
| *(unknown argument)* | Prints an error and usage, then exits with code 1 |

### `ai-engine init` output

On success, prints:

```
Initialising AI Engine workspace at: {path}

Workspace initialised successfully.

Next steps:
  1. Open .ai-engine/.env and set your ANTHROPIC_API_KEY.
  2. Open .ai-engine/config.json and set provider and default_model.
  3. Add your agents under .ai-engine/agents/.
  4. Run ai-engine to start the server.
```

Files created by `init` (only if they do not already exist):
- `.ai-engine/config.json`
- `.ai-engine/.env`
- `.ai-engine/engine_context.md` (from embedded template)
- `.ai-engine/agents/swarmito/agent.json` (from embedded template)
- `.ai-engine/agents/swarmito/system_prompt.md` (from embedded template)

---

## Key Interfaces

```go
// LLM provider abstraction
type LLMProvider interface {
    Send(ctx context.Context, req llm.Request) (llm.Response, error)
}

// Tool abstraction
type Tool interface {
    Name() string
    Description() string
    InputSchema() json.RawMessage
    Execute(ctx context.Context, input json.RawMessage) (string, error)
}
```

---

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `ANTHROPIC_API_KEY` | Anthropic API key | Required (no default) |

> The engine also loads a `.env` file from the workspace root (if present). `ANTHROPIC_API_KEY` can be defined there instead of as a system environment variable.

---

## Go Module

```
module github.com/swarmit/ai-engine

go 1.22
```

Key dependencies:
- `github.com/gorilla/websocket v1.5.3` — WebSocket server
- `github.com/anthropics/anthropic-sdk-go v0.2.0-alpha.4` — Anthropic LLM client
- `github.com/google/uuid v1.6.0` — session ID generation
- `github.com/tidwall/gjson`, `sjson`, `match`, `pretty` — indirect dependencies of the Anthropic SDK
