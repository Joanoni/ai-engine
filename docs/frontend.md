# Frontend

## Status

🟢 Implemented — Phase 3 + Phase 4 (Agent Graph) complete.

---

## Overview

The AI Engine frontend is a minimal React + TypeScript SPA built with Vite. Its sole purpose is to allow a user to send a prompt to the engine and observe the full agent tree execution in real time via a structured event feed.

It lives inside the same repository at `/frontend` and communicates exclusively with the engine's WebSocket server.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Framework | React 19 + TypeScript |
| Build tool | Vite |
| Styling | Plain CSS (no UI library, no Tailwind) |
| WebSocket | Native browser `WebSocket` API |
| Agent Graph | `@xyflow/react 12` + `@dagrejs/dagre 3` |
| Package manager | npm |

---

## Folder Structure

```
frontend/
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── App.css
│   ├── components/
│   │   ├── PromptInput.tsx
│   │   ├── EventFeed.tsx
│   │   ├── AgentGraph.tsx        # React Flow canvas — renders nodes and edges
│   │   └── AgentGraphNode.tsx    # Custom node — status badge (L/E), label, pulse animation
│   ├── hooks/
│   │   ├── useWebSocket.ts
│   │   └── useAgentGraph.ts      # Derives { nodes, edges } from EngineEvent[] via dagre layout
│   └── types/
│       ├── events.ts
│       └── graph.ts              # AgentNode, AgentEdge, AgentStatus, AgentType
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```

---

## Running Locally

### Prerequisites

- Node.js 18+
- The AI Engine backend running on `localhost:8080` (or the configured port)

### Commands

```bash
cd frontend
npm install
npm run dev       # starts dev server at http://localhost:5173
npm run build     # production build to frontend/dist/
```

### Environment

The WebSocket URL defaults to `ws://localhost:8080/ws`. To override, set the environment variable before building:

```bash
VITE_WS_URL=ws://localhost:9090/ws npm run dev
```

If `VITE_WS_URL` is not set, the app falls back to `ws://localhost:8080/ws`.

---

## WebSocket Protocol

The frontend communicates with the engine exclusively via WebSocket.

### Connection

```
ws://localhost:8080/ws
```

### Message sent by the frontend

All outgoing messages follow this envelope:

```json
{
  "event": "user.message",
  "session_id": "abc123",
  "payload": {
    "text": "Build a REST API for user management"
  }
}
```

> For the first message of a session, `session_id` may be omitted — the engine will generate one and return it in the `session.started` event.

### Events received from the engine

All incoming events follow this envelope:

```json
{
  "type": "event_type",
  "session_id": "abc123",
  "agent_name": "executor-d",
  "payload": { ... }
}
```

> **Note:** The engine uses the field name `"type"` (not `"event"`) in outgoing events. See `internal/events/bus.go` — the `Event` struct uses `Type EventType \`json:"type"\``.

#### Full event type reference

| `type` value | Trigger | Payload fields |
|---|---|---|
| `session.started` | User sends first message | _(none)_ |
| `session.finished` | Swarmito calls `finish_work` | `result: string` |
| `agent.started` | An agent begins execution | `message: string` |
| `agent.finished` | An agent calls `finish_work` | `result: string` |
| `tool.called` | An agent invokes a tool | `tool: string`, `id: string` |
| `tool.result` | A tool execution completes | `tool: string`, `id: string`, `result: string` |
| `tasks.updated` | A leader creates/updates a task file | `content: string` |
| `error` | An unrecoverable error occurs | `error: string` |

---

## Component Specifications

### `useWebSocket.ts`

Custom React hook. Manages the WebSocket lifecycle.

**Interface:**

```ts
interface UseWebSocketReturn {
  events: EngineEvent[];       // all events received in the current session
  isConnected: boolean;        // WebSocket readyState === OPEN
  isRunning: boolean;          // true between session.started and session.finished/error
  sendMessage: (text: string) => void;  // sends a user.message event
  clearEvents: () => void;     // resets the event list
}
```

**Behaviour:**

- Connects to `VITE_WS_URL` (or `ws://localhost:8080/ws`) on mount.
- On `session.started`: sets `isRunning = true`.
- On `session.finished` or `error`: sets `isRunning = false`.
- Appends every received event to the `events` array.
- `sendMessage(text)` sends `{"event":"user.message","payload":{"text":"..."}}`.
- Does **not** auto-reconnect in V1 — if the connection drops, the user must refresh.

---

### `PromptInput.tsx`

**Props:** none (uses `useWebSocket` internally via prop drilling from `App.tsx`)

**Actual props passed from App:**

```ts
interface PromptInputProps {
  onSend: (text: string) => void;
  disabled: boolean;
}
```

**Behaviour:**

- Controlled `<textarea>` for multi-line input.
- "Send" button submits on click or `Ctrl+Enter`.
- Disabled (both textarea and button) when `disabled === true` (i.e., a session is running).
- Clears the textarea after sending.

---

### `EventFeed.tsx`

**Props:**

```ts
interface EventFeedProps {
  events: EngineEvent[];
}
```

**Behaviour:**

- Renders a scrollable list of events, newest at the bottom.
- Auto-scrolls to the bottom when new events arrive.
- Each event row shows: timestamp, event type, agent name (if present), and a summary of the payload.
- Color-coded by event type:
  - `session.started` / `session.finished` → blue
  - `agent.started` / `agent.finished` → purple
  - `tool.called` / `tool.result` → grey
  - `tasks.updated` → teal
  - `error` → red

---

### `types/events.ts`

TypeScript types mirroring the engine's event model.

```ts
export type EventType =
  | 'session.started'
  | 'session.finished'
  | 'agent.started'
  | 'agent.finished'
  | 'tool.called'
  | 'tool.result'
  | 'tasks.updated'
  | 'error';

export interface EngineEvent {
  type: EventType;
  session_id?: string;
  agent_name?: string;
  payload?: Record<string, unknown>;
  // raw timestamp added client-side on receipt
  receivedAt: string;
}
```

---

### `App.tsx`

Root component. Responsibilities:

1. Instantiates `useWebSocket`.
2. Renders `<PromptInput onSend={sendMessage} disabled={isRunning} />`.
3. Renders `<EventFeed events={events} />`.
4. Shows a connection status indicator (`isConnected`).
5. Shows a "Clear" button to call `clearEvents()` when no session is running.
6. Renders `<AgentGraph nodes={nodes} edges={edges} />` (derived from `useAgentGraph(events)`).
7. Shows a "Hide Feed / Show Feed" toggle button to collapse/expand the event feed panel.

---

### `AgentGraph.tsx`

Renders the agent execution tree as an interactive React Flow canvas.

**Props:**
```ts
interface AgentGraphProps {
  nodes: AgentNode[];
  edges: AgentEdge[];
}
```

**Behaviour:**
- When `nodes` is empty, renders a placeholder message: "Send a prompt to see the agent graph".
- Uses `@xyflow/react` with a custom node type `agentNode` mapped to `AgentGraphNode`.
- Includes `<Background>`, `<Controls>`, and `<MiniMap>` panels.
- Nodes are not draggable or connectable (read-only view).
- MiniMap node colour reflects agent status: blue (running), green (done), red (error), grey (idle).
- `ReactFlowProvider` is provided by the parent `App.tsx` component — **not** inside `AgentGraph.tsx`. React Flow v12 requires the provider to be in a parent component for edges to render correctly.

---

### `AgentGraphNode.tsx`

Custom React Flow node component. Memoized.

**Visual design:**
- Leader nodes: rounded rectangle border (`borderRadius: 10px`).
- Executor nodes: pill border (`borderRadius: 30px`).
- Status badge (circle, top-left): shows `L` for leader, `E` for executor. Colour matches status.
- Status colours: idle → grey (`#9ca3af`), running → blue (`#3b82f6`), done → green (`#22c55e`), error → red (`#ef4444`).
- Running nodes have a pulsing border animation (`nodePulse` keyframe).
- Optional `lastMessage` subtitle shown below the label.
- **Handles:** `<Handle type="target" position={Position.Top}>` (incoming edges) and `<Handle type="source" position={Position.Bottom}>` (outgoing edges) are required for React Flow v12 to render edges. Both handles use dynamic color matching the node status and are styled as small dots (`width: 8`, `height: 8`, `border: none`).

---

### `useAgentGraph.ts`

Pure hook. Given the full `EngineEvent[]` list, returns `{ nodes, edges }`. No separate state — the graph is always derived from the event log via `useMemo`.

**Agent type inference:** inferred from agent name — names containing `leader`, `orchestrat`, or `manager` (case-insensitive) are typed as `leader`; all others as `executor`.

**State transitions:**

| Event | Graph update |
|---|---|
| `session.started` | Reset graph (clear all nodes and edges) |
| `agent.started` | Upsert node (status: `running`); add edge from `triggered_by` if present |
| `agent.finished` | Set node status to `done`; stop animation on incoming edges |
| `error` | Set node status to `error` (reads `payload.agent`) |
| `session.finished` | Set all `running` nodes to `done`; stop all edge animations |

**Layout:** Auto-layout via `@dagrejs/dagre`, top-down (`rankdir: TB`), `nodesep: 60`, `ranksep: 80`. Node size: 180×60. Layout is recomputed on every `useMemo` recalculation.

---

### `types/graph.ts`

```ts
export type AgentStatus = 'idle' | 'running' | 'done' | 'error';
export type AgentType = 'leader' | 'executor';

export interface AgentNodeData extends Record<string, unknown> {
  label: string;
  agentType: AgentType;
  status: AgentStatus;
  lastMessage?: string;
}

export interface AgentNode {
  id: string;
  type: 'agentNode';
  data: AgentNodeData;
  position: { x: number; y: number };
}

export interface AgentEdge {
  id: string;
  source: string;
  target: string;
  animated: boolean;
}
```

---

## `vite.config.ts`

Configure the dev server proxy to avoid CORS issues during local development:

```ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
})
```

> With this proxy, the frontend connects to `ws://localhost:5173/ws` in dev, which is forwarded to `ws://localhost:8080/ws`. In production, point `VITE_WS_URL` directly to the engine.

---

## Dependencies

| Package | Version | Purpose |
|---|---|---|
| `react` | `^19.2.6` | UI framework |
| `react-dom` | `^19.2.6` | DOM renderer |
| `@xyflow/react` | `^12.10.2` | Agent graph canvas (React Flow v12) |
| `@dagrejs/dagre` | `^3.0.0` | Directed graph auto-layout for agent tree |

---

## Design Constraints

- **No UI library** — plain CSS only. No Tailwind, no MUI, no Chakra.
- **No token streaming** — the engine does not stream partial text; only structured events are displayed.
- **No session history** — V1 does not persist sessions. Refreshing the page clears the feed.
- **Single connection** — one WebSocket connection per browser tab. Multiple tabs are not a concern for V1.
- **No authentication** — V1 assumes local use only.
