# API Reference

## WebSocket

**Endpoint:** `ws://localhost:{port}/ws`

The WebSocket connection is bidirectional. The frontend sends messages to the engine and receives events from the engine over the same connection. Only one session may be active per connection at a time.

---

### Messages sent by the frontend

Message envelope:

```json
{
  "event": "...",
  "session_id": "...",
  "payload": {}
}
```

Note: the frontend uses `"event"` (not `"type"`) as the envelope field name.

#### `user.message`

Start a new session with a user prompt.

```json
{
  "event": "user.message",
  "payload": {
    "text": "Build a REST API with Go..."
  }
}
```

`session_id` is not required — the engine generates a new UUID v4 session ID. If a session is already running on this connection, the engine returns an `error` event instead of starting a new session.

#### `ping`

Heartbeat. The server performs no action — used by the frontend to keep the connection alive and measure latency.

```json
{
  "event": "ping"
}
```

---

### Events received from the engine

Event envelope:

```json
{
  "type": "...",
  "session_id": "...",
  "agent_name": "...",
  "timestamp": "2026-05-19T01:00:00Z",
  "payload": {}
}
```

Note: the engine uses `"type"` (not `"event"`) as the envelope field name. `timestamp` is RFC3339 UTC, auto-filled by `Bus.Publish()`.

#### `session.started`

Emitted when a session begins.

```json
{
  "type": "session.started",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2026-05-19T01:00:00Z",
  "payload": {}
}
```

#### `session.finished`

Emitted when the root agent calls `finish_work` and the session ends successfully.

```json
{
  "type": "session.finished",
  "session_id": "...",
  "timestamp": "...",
  "payload": {
    "result": "All tasks completed successfully."
  }
}
```

#### `agent.started`

Emitted when an agent begins execution.

```json
{
  "type": "agent.started",
  "session_id": "...",
  "agent_name": "backend-executor",
  "timestamp": "...",
  "payload": {
    "message": "Starting backend-executor"
  }
}
```

#### `agent.finished`

Emitted when an agent calls `finish_work`.

```json
{
  "type": "agent.finished",
  "session_id": "...",
  "agent_name": "backend-executor",
  "timestamp": "...",
  "payload": {
    "result": "Go REST API implemented at backend/main.go"
  }
}
```

#### `tool.called`

Emitted when an agent calls a tool. The full tool input is not included in the event — it is available in the agent's `chat.jsonl` log.

```json
{
  "type": "tool.called",
  "session_id": "...",
  "agent_name": "backend-executor",
  "timestamp": "...",
  "payload": {
    "tool": "run_terminal_command",
    "id": "toolu_01ABC..."
  }
}
```

#### `tool.result`

Emitted after a tool execution completes.

```json
{
  "type": "tool.result",
  "session_id": "...",
  "agent_name": "backend-executor",
  "timestamp": "...",
  "payload": {
    "tool": "run_terminal_command",
    "id": "toolu_01ABC...",
    "result": "go: creating new module..."
  }
}
```

#### `tasks.updated`

Emitted when a leader calls `create_task_file` or `update_task_file`. The payload contains the full current Markdown checklist content.

```json
{
  "type": "tasks.updated",
  "session_id": "...",
  "agent_name": "swarmito",
  "timestamp": "...",
  "payload": {
    "content": "- [x] Backend API\n- [-] Frontend\n- [ ] Testing"
  }
}
```

#### `error`

Emitted when the session terminates due to an error (tool retry limit, tool call cap, nudge limit, or internal error).

```json
{
  "type": "error",
  "session_id": "...",
  "timestamp": "...",
  "payload": {
    "error": "max_tool_retries exceeded: 3 consecutive errors"
  }
}
```

---

## HTTP Endpoints

All endpoints return JSON. All responses include `Access-Control-Allow-Origin: *`.

### `GET /health`

Health check.

**Response:** `200 OK`
```
ok
```

---

### `GET /version`

Returns the binary version.

**Response:**
```json
{
  "version": "0.0.26"
}
```

---

### `GET /agents`

Returns the flat agent list for graph seeding. Called by the frontend on load to pre-populate the agent graph before any session starts.

**Response:**
```json
{
  "agents": [
    {
      "name": "swarmito",
      "type": "leader",
      "description": "Root orchestrator",
      "parent": ""
    },
    {
      "name": "backend-leader",
      "type": "leader",
      "description": "Coordinates backend implementation",
      "parent": "swarmito"
    },
    {
      "name": "backend-executor",
      "type": "executor",
      "description": "Implements the Go REST API",
      "parent": "backend-leader"
    }
  ]
}
```

`type` values are `"leader"` or `"executor"` (not the React Flow node type names).

---

### `GET /sessions`

Returns all sessions sorted by `startedAt` descending.

**Response:** `SessionMeta[]`
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "prompt": "Build a REST API...",
    "startedAt": "2026-05-19T01:00:00Z",
    "finishedAt": "2026-05-19T01:05:00Z",
    "status": "done"
  }
]
```

`status` values: `"running"`, `"done"`, `"error"`.

---

### `GET /sessions/{id}/events`

Returns the persisted events for a session as a JSON array. Used for session replay. The session ID is validated as a UUID v4 before use.

**Response:** `Event[]`
```json
[
  {
    "type": "session.started",
    "session_id": "550e8400-...",
    "timestamp": "2026-05-19T01:00:00Z",
    "payload": {}
  }
]
```

---

### `GET /sessions/{id}/tokens`

Returns token usage and estimated cost for a specific session.

**Response:**
```json
{
  "session_id": "550e8400-...",
  "input_tokens": 45230,
  "output_tokens": 8910,
  "total_tokens": 54140,
  "estimated_cost_usd": 0.2701,
  "updated_at": "2026-05-19T01:05:00Z"
}
```

---

### `GET /sessions/{id}/logs`

Returns the JSONL chat logs for all agents in a session. Used by the Analytics BI Panel.

**Response:**
```json
{
  "agents": {
    "swarmito": [
      { "ts": "...", "turn": 0, "role": "agent_init", ... },
      { "ts": "...", "turn": 1, "role": "llm_request", ... }
    ],
    "backend-executor": [
      { "ts": "...", "turn": 0, "role": "agent_init", ... }
    ]
  }
}
```

Returns `{ "agents": {} }` if the logs directory does not exist for the session. See [`docs/chat-log-format.md`](./chat-log-format.md) for the full `LogEntry` structure.

---

### `GET /tokens`

Returns project-level aggregate token usage across all sessions.

**Response:**
```json
{
  "input_tokens": 312450,
  "output_tokens": 67890,
  "total_tokens": 380340,
  "estimated_cost_usd": 1.9523,
  "session_count": 12,
  "last_updated_at": "2026-05-19T01:05:00Z"
}
```
