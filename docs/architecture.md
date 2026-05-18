# Architecture

## Status

🟢 Phase 4 complete — engine operational. Binary v0.0.12. Agent graph fully functional with custom glassmorphism nodes (`LeaderNode`, `ExecutorNode`), animated edges with dual particles and glow halo, correct node type mapping from `/agents` endpoint, and dynamic `fitView` centering.

---

## Goals

- Provide a runtime capable of orchestrating a **tree of AI agents**.
- Be modular and extensible: new agent types, tools, and LLM providers should be easy to add.
- Be observable: execution is streamed in real time to a frontend via WebSocket.
- **V1 scope:** implement the minimum viable engine to support the defined agent tree and workflow. Designed for future extensibility.

---

## Core Concepts

| Term | Definition |
|---|---|
| **workspace** | The directory on the user's machine where the engine operates. Agents are unaware of its absolute path. |
| **Swarmito** | The top-level agent that serves as the interface between the user and the agent tree. |
| **chat** | A conversation context between two entities (two agents, or user ↔ Swarmito). |
| **leader** | An agent that has a team below it in the hierarchy (non-leaf node). |
| **executor** | An agent that performs concrete tasks (leaf node). |
| **task file** | A Markdown checklist file (per agent, per session) containing the list of tasks to be executed. |

---

## Tech Stack

| Layer | Technology | Notes |
|---|---|---|
| Runtime | **Go** | Single binary, native concurrency, good for long-running services |
| LLM Provider (V1) | **Anthropic (Claude)** | Implemented behind a `LLMProvider` interface for future expansion |
| Frontend Interface | **WebSocket** | Bidirectional, port `8080` (configurable), structured JSON events |
| Config Format | **JSON** | Easy to serialize and send to frontend |
| Agent Prompts | **Markdown** | Human-readable, easy to edit, renderable in frontend |

---

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                           AI Engine                             │
│                                                                 │
│   Frontend ◀──────── WebSocket :8080 ──────────▶ User          │
│       │                    │                                    │
│       │              ┌─────▼──────┐                            │
│       │              │  Session   │                            │
│       │              │  Manager   │                            │
│       │              └─────┬──────┘                            │
│       │                    │                                    │
│       │         ┌──────────▼──────────┐                        │
│       │         │  Swarmito (root)    │                        │
│       │         │  leader             │                        │
│       │         └──────────┬──────────┘                        │
│       │              ┌─────┴──────┐                            │
│       │              ▼            ▼                            │
│       │         Leader A      Leader B  ...                    │
│       │         /   |   \                                      │
│       │        D    E    F  (executors)                        │
│       │                                                        │
│  ┌────▼────────────────────────────────────────────────────┐   │
│  │                      Engine Core                        │   │
│  │                                                         │   │
│  │  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐  │   │
│  │  │ Chat Manager│  │ Tool Executor│  │  LLM Client   │  │   │
│  │  └─────────────┘  └──────────────┘  │  (interface)  │  │   │
│  │                                     └───────┬───────┘  │   │
│  │  ┌──────────────────────┐                   │          │   │
│  │  │  Workspace Sandbox   │           ┌───────▼───────┐  │   │
│  │  │  (path isolation)    │           │   Anthropic   │  │   │
│  │  └──────────────────────┘           │   Provider    │  │   │
│  │                                     └───────────────┘  │   │
│  │  ┌──────────────────────┐                              │   │
│  │  │   Event Bus          │  ──▶ WebSocket stream        │   │
│  │  └──────────────────────┘                              │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Components

### Engine Core

| Component | Responsibility |
|---|---|
| **Session Manager** | Creates and tracks execution sessions. Generates session IDs (UUID v4). |
| **Chat Manager** | Creates and manages chat sessions between agents or between user and Swarmito. Reads agent definitions from `.ai-engine/` on each interaction (hot-reload). |
| **LLM Client** | Provider-agnostic interface (`LLMProvider`) for sending messages and receiving tool calls. V1 implementation: Anthropic. Detects `max_tokens` truncation and returns an error. |
| **Tool Executor** | Receives tool call requests from agents, validates them, executes them, and returns results. Enforces `max_tool_retries` (consecutive errors) and `max_tool_calls` (total calls) limits. |
| **Workspace Sandbox** | Intercepts all file/terminal operations and resolves paths relative to the configured workspace directory. Agents never see absolute paths. Prevents path traversal attacks. |
| **Agent Registry** | Loads agent definitions from `.ai-engine/agents/` on demand. Parses `agent.json` and `system_prompt.md`. Also loads optional `engine_context.md`. |
| **Event Bus** | Publishes structured JSON events to all connected WebSocket clients in real time. Thread-safe. |
| **Chat Logger** | Writes structured JSONL logs to `.ai-engine/logs/{session-id}/{agent-name}/chat.jsonl`. Records every LLM turn, tool call, tool result, and finish event. Non-fatal — logging errors do not terminate the session. |

---

## Key Design Decisions

### 1. Workspace Path Isolation
Agents **never know the absolute path** of the workspace. All file and terminal commands issued by agents are treated as relative to the workspace root. The engine intercepts and resolves paths before execution. Path traversal attacks (e.g., `../../etc/passwd`) are rejected.

### 2. Tool-Only Execution
Agents operate **exclusively through tools**. There is no free-form text output that triggers side effects — all actions are explicit tool calls. If the LLM returns a plain text response with no tool calls, the engine injects a nudge message to keep the loop going.

### 3. Hot-Reload Configuration
The engine reads `.ai-engine/` on every interaction. Agent definitions, system prompts, and config can be modified without restarting the engine.

### 4. Provider Abstraction
The LLM client is defined as a Go interface (`LLMProvider`). V1 implements Anthropic. Adding a new provider requires only a new implementation of the interface.

### 5. Bidirectional WebSocket
The frontend communicates exclusively via WebSocket on port `8080`. User prompts are sent as `user.message` events; the engine streams back structured execution events. A `/health` HTTP endpoint is also available.

### 6. Error Strategy: Retry with Configurable Limits
Tool errors are fed back to the agent as tool results (`is_error: true`). The agent can retry. If `max_tool_retries` consecutive errors occur, or `max_tool_calls` total calls are exceeded, the session terminates. Both limits are configurable in `config.json`.

### 7. Sequential Execution (V1)
Leaders process tasks and agent chats sequentially. Parallel execution is a future enhancement enabled by Go's native concurrency primitives.

### 8. 3-Layer System Prompt
Every agent's system prompt is composed at runtime from three layers: (1) `engine_context.md` — shared engine instructions; (2) `system_prompt.md` — agent role and skills; (3) `task_context.md` — per-session task written by the leader via `set_task_context`. This separates stable role definitions from dynamic task assignments.

### 9. Structured Chat Logging
Every agent's conversation is logged to `.ai-engine/logs/{session-id}/{agent-name}/chat.jsonl`. Each line is a JSON object capturing the turn, role, tool calls, results, and errors. Logging is non-fatal — failures do not affect agent execution.

### 10. Graceful Shutdown
The server listens for `SIGINT`/`SIGTERM` via `signal.NotifyContext`. On signal, `http.Server.Shutdown` is called, allowing in-flight requests to complete before exit.
