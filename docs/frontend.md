# Frontend

## Status

🟢 **Frontend v2 "Mission Control" implemented** — branch `frontend-v2-test2`, binary v0.0.4.

---

## Overview

The AI Engine frontend is a React + TypeScript SPA built with Vite. It serves as a **developer cockpit** for launching agent missions and observing the full agent tree execution in real time.

Two versions exist:

| Version | Branch | Description |
|---|---|---|
| **v1** | `main` | Minimal layout: graph + event feed + prompt input. Single panel. |
| **v2** | `frontend-v2-test2` | Full "Mission Control" redesign: 3-column cockpit, glassmorphism graph, live terminal, session history, resizable panels, agent detail drawer. |

This document describes **v2**. For v1, see git history.

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
| UI Font | `Inter` (Google Fonts) |
| Terminal Font | `JetBrains Mono` (Google Fonts) |

---

## Design System

### Color Palette (CSS variables in `index.css`)

```css
--bg-base:       #080c14;   /* deep navy — main background */
--bg-surface:    #0d1117;   /* cards and panels */
--bg-surface-2:  #161b22;   /* elevated elements */
--border:        #21262d;   /* subtle borders */
--accent:        #58a6ff;   /* electric blue — primary actions */
--success:       #3fb950;   /* green — done */
--warning:       #d29922;   /* yellow — running / tool calls */
--error:         #f85149;   /* red — error */
--purple:        #bc8cff;   /* violet — leader agents */
--text-primary:  #e6edf3;
--text-muted:    #7d8590;
```

### Visual Effects

- **Glassmorphism** on agent node cards: `backdrop-filter: blur(12px)`, semi-transparent background
- **Glow** on active nodes: pulsing `box-shadow` in accent color (`pulse-glow` keyframe)
- **Animated edges**: SVG `stroke-dashoffset` animation while agent is running (`dash-flow` keyframe)
- **Slide-in drawer**: `transform: translateX` transition (`slide-in-right` keyframe)
- **Scanline overlay** in terminal: CSS `repeating-linear-gradient` for CRT effect

---

## Layout

```
┌──────────────┬──────────────────────────────┬──────────────┐
│   SIDEBAR    │       COCKPIT AREA           │   MISSION    │
│   240px      │       flex: 1                │   PANEL      │
│  (collapsible│  ┌──────────────────────┐    │   360px      │
│   Ctrl+B)    │  │   AGENT GRAPH (60%)  │    │              │
│              │  │                      │    │              │
│              │  ├──── ResizeHandle ────┤    │              │
│              │  │  LIVE TERMINAL (40%) │    │              │
│              │  └──────────────────────┘    │              │
└──────────────┴──────────────────────────────┴──────────────┘
```

The vertical split between graph and terminal is drag-resizable.

---

## Folder Structure

```
frontend/src/
├── App.tsx                          # Root component — wires all layout + keyboard shortcuts
├── App.css                          # Minimal layout-only styles
├── index.css                        # CSS variables, global reset, keyframe animations
├── main.tsx                         # Vite entry point (unchanged)
├── components/
│   ├── layout/
│   │   ├── Sidebar.tsx              # Collapsible sidebar — session history + New Mission
│   │   ├── CockpitArea.tsx          # Center area — graph + resize handle + terminal
│   │   ├── MissionPanel.tsx         # Right panel — all mission controls
│   │   └── ResizeHandle.tsx         # Drag handle for vertical split
│   ├── graph/
│   │   ├── AgentGraph.tsx           # React Flow canvas — dot grid, MiniMap, Controls
│   │   ├── LeaderNode.tsx           # Custom node — glassmorphism, hexagonal badge, glow
│   │   ├── ExecutorNode.tsx         # Custom node — pill shape, circular badge
│   │   └── AnimatedEdge.tsx         # Custom edge — SVG particle animation
│   ├── terminal/
│   │   ├── LiveTerminal.tsx         # Terminal panel — scanline, auto-scroll, export
│   │   └── TerminalLine.tsx         # Single terminal line — color-coded by event type
│   ├── mission/
│   │   ├── PromptEditor.tsx         # Auto-resize textarea, Ctrl+Enter to submit
│   │   ├── LaunchButton.tsx         # Countdown 3→2→1→🚀 before firing
│   │   ├── TaskProgress.tsx         # Markdown checklist parser + progress bar
│   │   ├── AgentRoster.tsx          # Live agent list with status + tool call count
│   │   └── QuickStats.tsx           # Tool calls, agents, live duration counter
│   └── drawers/
│       └── AgentDetailDrawer.tsx    # Slide-in drawer — per-agent event timeline
├── hooks/
│   ├── useWebSocket.ts              # WebSocket lifecycle + auto-reconnect + latency
│   ├── useAgentGraph.ts             # Derives { nodes, edges } from EngineEvent[]
│   ├── useSessionHistory.ts         # localStorage session persistence (max 20)
│   ├── useResizable.ts              # Drag-to-resize vertical split ratio
│   └── useKeyboardShortcuts.ts      # Global keyboard shortcuts (Ctrl+B, Escape)
└── types/
    ├── events.ts                    # EngineEvent, EventType
    ├── graph.ts                     # AgentNode, AgentEdge, AgentStatus, AgentType
    └── session.ts                   # SessionRecord for localStorage persistence
```

---

## Component Specifications

### `App.tsx`

Root component. Responsibilities:
1. Instantiates `useWebSocket`, `useAgentGraph`, `useSessionHistory`.
2. Manages sidebar collapsed state and selected agent (for drawer).
3. Registers keyboard shortcuts via `useKeyboardShortcuts`: `Ctrl+B` (sidebar), `Escape` (drawer).
4. Wraps everything in `ReactFlowProvider`.
5. Renders: `<Sidebar>`, `<CockpitArea>`, `<MissionPanel>`, `<AgentDetailDrawer>`.

---

### `components/layout/Sidebar.tsx`

Left sidebar (240px, collapsible via `Ctrl+B`).

**Content:**
- Top: Logo area — icon + "AI Engine" text + "v2" version badge.
- Middle: Session history list (from `useSessionHistory`).
  - Each item: relative timestamp (e.g. "2 min ago"), status badge (running/done/error), first 60 chars of prompt.
  - Active session: left accent border highlight.
  - Clicking a past session loads its events for replay (read-only).
- Bottom: "New Mission" button — clears current session, enables input.

When collapsed: shows only logo icon + session count badge.

---

### `components/layout/CockpitArea.tsx`

Center area. Contains:
- `<AgentGraph>` in the top zone (default 60% height).
- `<ResizeHandle>` drag handle.
- `<LiveTerminal>` in the bottom zone (default 40% height).

Manages the vertical split ratio via `useResizable`.

---

### `components/layout/ResizeHandle.tsx`

Horizontal drag strip (6px tall, full width).

- Background: `--border`; hover: `--accent` at 40% opacity.
- Cursor: `ns-resize`.
- Center indicator: 3 horizontal dots.
- `onMouseDown` → starts drag, updates parent split ratio on `mousemove`, ends on `mouseup`.

---

### `components/layout/MissionPanel.tsx`

Right panel (360px). Sections top-to-bottom:
1. Connection status dot + label + latency.
2. `<PromptEditor>`.
3. `<LaunchButton>`.
4. `<TaskProgress>` — visible only when session is running.
5. `<AgentRoster>` — visible only when agents exist.
6. `<QuickStats>` — visible only when session is running or done.

---

### `components/graph/AgentGraph.tsx`

React Flow canvas. Props: `{ nodes, edges, onNodeClick }`.

- Background: `--bg-base` with line grid (`<Background variant="lines">`, `gap=40`, `lineWidth=0.5`, `color="rgba(33,38,45,0.4)"`).
- `<Controls>` (bottom-left). No `<MiniMap>` — removed because it overlapped nodes in small viewports.
- Radial gradient depth overlay: `radial-gradient(ellipse at 50% 40%, rgba(88,166,255,0.04), transparent 70%)` — subtle depth effect, `pointerEvents: none`.
- Custom node types: `leaderNode` → `LeaderNode`, `executorNode` → `ExecutorNode`.
- Custom edge type: `animatedEdge` → `AnimatedEdge`.
- Empty state: inline SVG network icon + "Launch a mission to see the agent graph".
- `ReactFlowProvider` is in `App.tsx` — **not** inside this component.
- Nodes are read-only (not draggable, not connectable).
- `fitView` is dynamic: `onInit` callback calls `instance.fitView({ padding: 0.4, duration: 400 })` after 100ms delay; `useEffect` re-fits on `nodes` change (150ms delay). This ensures correct centering after DOM layout settles.

---

### `components/graph/LeaderNode.tsx`

Custom React Flow node for leader agents.

- Rounded rectangle (`borderRadius: 14px`), `minWidth: 220px`.
- Glassmorphism: `background: linear-gradient(135deg, rgba(188,140,255,0.08), rgba(13,17,23,0.97))`, `backdrop-filter: blur(16px)`.
- Full-width top accent line: 3px gradient `transparent → borderColor → transparent`.
- Top-left badge: 32×32 hexagon (`clipPath: polygon(...)`) with `linear-gradient(135deg, #bc8cff, #7c3aed)`, "L" letter, `box-shadow: 0 0 8px rgba(188,140,255,0.5)`.
- Agent name (bold, 13px) + status label below (monospace, 10px, color-coded).
- Status-driven border color: `rgba(188,140,255,0.3)` idle → `rgba(188,140,255,0.9)` running → `rgba(63,185,80,0.7)` done → `rgba(248,81,73,0.7)` error.
- Running: `pulse-glow-purple` animation + animated blink dot before status label.
- Hover: `transform: translateY(-2px)` lift via `onMouseEnter`/`onMouseLeave` state.
- Entrance: `node-appear` animation (scale + translateY fade-in).
- Tool call count badge (top-right, from `data.toolCallCount`).
- Handles: `<Handle type="target" position={Position.Top}>` + `<Handle type="source" position={Position.Bottom}>`, styled as 10×10 colored dots.

---

### `components/graph/ExecutorNode.tsx`

Custom React Flow node for executor agents.

- Pill shape (`borderRadius: 40px`), `minWidth: 190px`.
- Glassmorphism: `background: linear-gradient(135deg, rgba(88,166,255,0.06), rgba(13,17,23,0.97))`, `backdrop-filter: blur(16px)`.
- Inner radial glow: `radial-gradient(ellipse at 30% 0%, rgba(88,166,255,0.10), transparent 70%)`.
- Top-left badge: 28×28 circle with `linear-gradient(135deg, #58a6ff, #1d4ed8)`, "E" letter, `box-shadow: 0 0 6px rgba(88,166,255,0.4)`.
- Bottom status bar: 3px height, color-coded by status, `box-shadow` matching status color.
- Agent name (600 weight, 12px) + status label (monospace, 10px, color-coded).
- Running: `pulse-glow` animation + animated blink dot before status label.
- Hover: `transform: translateY(-2px)` lift.
- Entrance: `node-appear` animation.
- Tool call count badge (top-right, from `data.toolCallCount`).
- Handles: same as `LeaderNode`, 8×8.

---

### `components/graph/AnimatedEdge.tsx`

Custom React Flow edge.

- Uses `getBezierPath` from `@xyflow/react`.
- **Base path**: `strokeWidth: 2` (animated) / `1.5` (static), `rgba(88,166,255,0.6)` / `rgba(33,38,45,0.6)`.
- **Glow halo** (animated only): same path, `strokeWidth: 6`, `opacity: 0.15`, `stroke: var(--accent)`, `edge-glow-pulse` animation — creates soft glow around the edge.
- **Animated dashes** (animated only): `stroke-dasharray: 8 12`, `dash-flow` keyframe.
- **Two particles** (animated only): `<circle r="3">` with `<animateMotion>` — first at `dur="1.2s"`, second at `dur="1.8s" begin="0.6s"` — staggered flow effect.

---

### `components/terminal/LiveTerminal.tsx`

Live event log styled as a real terminal.

- Background: `#050810` (darker than base).
- Font: `JetBrains Mono`, 13px.
- Header: "LIVE LOG" title + "Clear" button + "Export" button (downloads events as `.jsonl` via `URL.createObjectURL`).
- Scanline overlay: CSS `repeating-linear-gradient` for CRT effect.
- Auto-scrolls to bottom on new events.
- Renders `<TerminalLine>` per event.

---

### `components/terminal/TerminalLine.tsx`

Single terminal line. Props: `{ event: EngineEvent }`.

**Format:** `HH:MM:SS  [TYPE]  agent-name  › message`

**Color coding:**

| Event type | Color |
|---|---|
| `session.started` / `session.finished` | `--accent` (cyan-blue) |
| `agent.started` / `agent.finished` | `--purple` |
| `tool.called` / `tool.result` | `--warning` (yellow) |
| `tasks.updated` | `#2dd4bf` (teal) |
| `error` | `--error` |

---

### `components/mission/PromptEditor.tsx`

Props: `{ onSend, disabled }`.

- Auto-resize textarea (min 120px, max 240px).
- Background: `--bg-surface-2`, border: `--border`.
- Focus: border → `--accent`.
- Placeholder: "Describe your mission..."
- `Ctrl+Enter` submits.

---

### `components/mission/LaunchButton.tsx`

Props: `{ onClick, disabled, isRunning }`.

- Full-width, 48px height.
- Gradient: `linear-gradient(135deg, #1d4ed8, #58a6ff)`.
- Idle text: "🚀 Launch Mission". Running text: "⏳ Mission Running...".
- On click: countdown animation "3..." → "2..." → "1..." → "🚀 Launching" (600ms via `setTimeout` chains) before calling `onClick`.
- Running state: pulsing border animation (`pulse-border` keyframe).

---

### `components/mission/TaskProgress.tsx`

Props: `{ events: EngineEvent[] }`.

Parses the latest `tasks.updated` event's `content` field (Markdown checklist):
- `[x]` → completed
- `[-]` → in progress
- `[ ]` → pending

Renders:
- Section title: "TASK PROGRESS".
- Progress bar: filled = completed/total, color `--success`.
- Item list with icons (✓ / ⟳ / ○) and truncated text (40 chars).

---

### `components/mission/AgentRoster.tsx`

Props: `{ events: EngineEvent[] }`.

Derives agent list from events. Per agent: name, type (leader/executor), status, tool call count.

- Section title: "AGENT ROSTER".
- Each row: status dot (colored), name, type badge (L/E), tool call count (right-aligned).
- Running agents: subtle highlighted background.

---

### `components/mission/QuickStats.tsx`

Props: `{ events: EngineEvent[], startTime: Date | null }`.

- Section title: "QUICK STATS".
- 3 stat cards: "Tool Calls" (total `tool.called` events), "Agents" (unique agent names), "Duration" (live counter while running, final value when done).

---

### `components/drawers/AgentDetailDrawer.tsx`

Props: `{ agent: string | null, events: EngineEvent[], onClose: () => void }`.

- Width: 420px, full height.
- Background: `--bg-surface`, border-left: `--border`.
- Slide-in animation: `transform: translateX(0)` from `translateX(100%)` (`slide-in-right` keyframe).
- Header: agent name + type badge + close button (×).
- Content: scrollable per-agent event timeline filtered by `event.agent_name === agent`.
- Close: × button or `Escape` key.

---

## Hooks

### `useWebSocket.ts`

Extended interface:

```ts
interface UseWebSocketReturn {
  events: EngineEvent[];
  isConnected: boolean;
  isRunning: boolean;
  connectionStatus: string;        // "Connected" | "Disconnected" | "Reconnecting (N/10)..."
  latency: number | null;          // ms, measured on each event receipt
  sessionStartTime: Date | null;   // set on session.started
  sendMessage: (text: string) => void;
  clearEvents: () => void;
}
```

**Auto-reconnect:** on disconnect, retries every 3 seconds, max 10 attempts. `connectionStatus` reflects reconnect state.

---

### `useAgentGraph.ts`

Derives `{ nodes, edges }` from `EngineEvent[]` and static agent data from `GET /agents`.

**Key fix (v0.0.9):** The `/agents` endpoint returns `"type": "leader"` / `"type": "executor"`. A `toNodeType()` function maps these to `"leaderNode"` / `"executorNode"` before passing to React Flow — without this mapping, React Flow falls back to `"default"` and renders plain white nodes.

- `inferAgentType(name)` — fallback heuristic: names containing `leader`, `orchestrat`, `manager`, `swarmito` → `"leaderNode"`; all others → `"executorNode"`.
- `toNodeType(apiType)` — maps `"leader"` → `"leaderNode"`, `"executor"` → `"executorNode"`, falls back to `inferAgentType`.
- `populateFromStatic()` — seeds the graph from `/agents` before any session starts, using `toNodeType()`.
- Layout constants: `NODE_WIDTH=260`, `NODE_HEIGHT=100`, `nodesep=80`, `ranksep=120` (updated in v0.0.10 to match redesigned node sizes).
- Edge type: `animatedEdge`.

---

### `useSessionHistory.ts`

Manages `localStorage` key `ai-engine-sessions`.

```ts
interface SessionRecord {
  id: string;
  prompt: string;
  startedAt: string;    // ISO timestamp
  finishedAt?: string;
  status: 'running' | 'done' | 'error';
  events: EngineEvent[];
}
```

- `sessions: SessionRecord[]` — all saved sessions (max 20, oldest dropped).
- `startSession(prompt)` — creates new record.
- `saveEvent(event)` — appends to active session.
- `finishSession(status)` — marks as finished.
- `loadSession(id)` — returns record for replay.

---

### `useResizable.ts`

```ts
function useResizable(defaultRatio: number): {
  ratio: number;
  handleMouseDown: (e: MouseEvent) => void;
}
```

Manages vertical split ratio. Uses `mousemove`/`mouseup` on `document`.

---

### `useKeyboardShortcuts.ts`

Registers global `keydown` handlers. Takes a map of `{ key: string, handler: () => void }`.

Registered shortcuts in `App.tsx`:
- `Ctrl+B` — toggle sidebar.
- `Escape` — close agent detail drawer.

---

## Types

### `types/events.ts`

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
  receivedAt: string;   // ISO timestamp added client-side on receipt
}
```

---

### `types/graph.ts`

```ts
export type AgentStatus = 'idle' | 'running' | 'done' | 'error';
export type AgentType = 'leaderNode' | 'executorNode';   // React Flow custom node type names

export interface AgentNodeData extends Record<string, unknown> {
  label: string;
  agentType: AgentType;
  status: AgentStatus;
  lastMessage?: string;
}

export interface AgentNode {
  id: string;
  type: AgentType;
  data: AgentNodeData;
  position: { x: number; y: number };
}

export interface AgentEdge {
  id: string;
  source: string;
  target: string;
  type: 'animatedEdge';
  animated: boolean;
}
```

---

### `types/session.ts`

```ts
export interface SessionRecord {
  id: string;
  prompt: string;
  startedAt: string;
  finishedAt?: string;
  status: 'running' | 'done' | 'error';
  events: EngineEvent[];
}
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

The WebSocket URL defaults to `ws://localhost:8080/ws`. To override:

```bash
VITE_WS_URL=ws://localhost:9090/ws npm run dev
```

---

## WebSocket Protocol

### Message sent by the frontend

```json
{
  "event": "user.message",
  "session_id": "abc123",
  "payload": {
    "text": "Build a REST API for user management"
  }
}
```

> For the first message of a session, `session_id` may be omitted.

### Events received from the engine

```json
{
  "type": "event_type",
  "session_id": "abc123",
  "agent_name": "executor-d",
  "payload": { ... }
}
```

> **Note:** The engine uses `"type"` (not `"event"`) in outgoing events. See [`internal/events/bus.go`](../internal/events/bus.go).

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

## `vite.config.ts`

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
- **Session history in localStorage** — v2 persists sessions client-side (max 20). Engine-side persistence is a future enhancement.
- **Single connection** — one WebSocket connection per browser tab. Multiple tabs are not a concern for V1.
- **No authentication** — V1 assumes local use only.
