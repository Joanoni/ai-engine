# WebSocket Event System

## Status

🟢 Defined — all decisions made. Ready for implementation.

---

## Overview

The AI Engine exposes a **WebSocket server** that streams execution events to connected frontends in real time. This allows the frontend to visualize the full agent tree execution as it happens — which agent is running, what tools are being called, and what the results are.

---

## Connection

```
ws://localhost:{port}/ws
```

The WebSocket connection is **bidirectional**:
- The **frontend sends** messages to the engine (user prompt, user replies to Swarmito).
- The **engine sends** structured execution events to the frontend (agent status, tool calls, task updates, errors).

There is no separate HTTP endpoint for sending prompts — everything goes through the WebSocket.

---

## Event Envelope

All events share a common envelope structure:

```json
{
  "event": "event_type",
  "session_id": "abc123",
  "timestamp": "2026-05-14T14:00:00Z",
  "payload": { ... }
}
```

| Field | Type | Description |
|---|---|---|
| `event` | string | Event type identifier (see below) |
| `session_id` | string | Unique identifier for the current execution session |
| `timestamp` | string | ISO 8601 UTC timestamp |
| `payload` | object | Event-specific data |

---

## Event Types

### Session Events

#### `session.started`
Fired when the user sends a prompt and the engine begins execution.

```json
{
  "event": "session.started",
  "session_id": "abc123",
  "timestamp": "...",
  "payload": {
    "prompt": "Build a REST API for user management"
  }
}
```

#### `session.finished`
Fired when Swarmito calls `finish_work` and the full execution is complete.

```json
{
  "event": "session.finished",
  "session_id": "abc123",
  "timestamp": "...",
  "payload": {
    "result": "The REST API has been implemented. Files created: ..."
  }
}
```

---

### Agent Events

#### `agent.started`
Fired when an agent begins processing (receives its first prompt).

```json
{
  "event": "agent.started",
  "session_id": "abc123",
  "timestamp": "...",
  "payload": {
    "agent": "leader-a",
    "type": "leader",
    "triggered_by": "swarmito"
  }
}
```

#### `agent.finished`
Fired when an agent calls `finish_work`.

```json
{
  "event": "agent.finished",
  "session_id": "abc123",
  "timestamp": "...",
  "payload": {
    "agent": "leader-a",
    "result": "All tasks completed successfully."
  }
}
```

---

### Tool Events

#### `tool.called`
Fired when an agent invokes a tool.

```json
{
  "event": "tool.called",
  "session_id": "abc123",
  "agent_name": "executor-d",
  "timestamp": "...",
  "payload": {
    "tool": "run_terminal_command",
    "id": "toolu_abc123"
  }
}
```

> **Note:** The tool input parameters are not included in the `tool.called` event payload. Only the tool name and call ID are published. The full input is recorded in the JSONL chat log at `.ai-engine/logs/{session-id}/{agent-name}/chat.jsonl`.

#### `tool.result`
Fired when a tool execution completes and the result is returned to the agent.

```json
{
  "event": "tool.result",
  "session_id": "abc123",
  "agent_name": "executor-d",
  "timestamp": "...",
  "payload": {
    "tool": "run_terminal_command",
    "id": "toolu_abc123",
    "result": "go: creating new module: github.com/user/project"
  }
}
```

---

### Task File Events

#### `tasks.updated`
Fired when a leader creates or updates a task file (via `create_task_file`, `update_task_file`) or sets a task context (via `set_task_context`).

```json
{
  "type": "tasks.updated",
  "session_id": "abc123",
  "agent_name": "leader-a",
  "payload": {
    "path": ".ai-engine/chats/abc123/leader-a/tasks.md"
  }
}
```

> **Note:** The payload contains the file path, not the parsed task list. The frontend reads the path for display purposes; the actual task content is in the file on disk.

---

### Error Events

#### `error`
Fired when an unrecoverable error occurs during execution.

```json
{
  "event": "error",
  "session_id": "abc123",
  "timestamp": "...",
  "payload": {
    "agent": "executor-d",
    "message": "Tool execution failed: command not found",
    "fatal": true
  }
}
```

---

## Messages Sent by the Frontend

All messages from the frontend follow the same envelope:

```json
{
  "event": "message_type",
  "session_id": "abc123",
  "payload": { ... }
}
```

### `user.message`
Sent when the user submits a prompt or reply to Swarmito.

```json
{
  "event": "user.message",
  "session_id": "abc123",
  "payload": {
    "text": "Build a REST API for user management"
  }
}
```

> For the first message of a session, `session_id` may be omitted — the engine will generate and return one in the `session.started` event.

---

## Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Direction | Bidirectional | User sends prompts and replies via WebSocket; engine streams events back |
| Token streaming | No | Only structured events (tool calls, status, results) — simpler for V1 |
| Port | Configurable in `config.json` | Avoids conflicts with user's project |
| Event persistence | Not in V1 | Can be added in Phase 4 for session replay |

---

## Decisions

| Question | Decision |
|---|---|
| Default WebSocket port? | **8080** (configurable via `port` field in `config.json`) |
| Event persistence? | **Not in V1.** Deferred to Phase 4. |
