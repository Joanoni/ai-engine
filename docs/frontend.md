# Frontend

## Tech Stack

| Technology | Version | Purpose |
|---|---|---|
| React | 19 | UI framework |
| TypeScript | — | Type safety |
| Vite | — | Build tool, dev server |
| Plain CSS | — | Styling (no UI library) |
| `@xyflow/react` | v12 | Agent graph canvas |
| `@dagrejs/dagre` | — | Automatic graph layout |
| `recharts` | — | Analytics charts |

The frontend is built with `npm run build` inside `src/frontend/`, outputs to `src/backend/frontend-dist/`, and is embedded in the Go binary via `src/backend/embed.go`.

---

## Layout

Three-column layout:

```
┌──────────┬──────────────────────────────┬─────────────┐
│          │                              │             │
│ Sidebar  │      Cockpit Area            │   Mission   │
│  240px   │  (graph top / terminal bot)  │   Panel     │
│ collapse │  drag-resizable split        │   360px     │
│  Ctrl+B  │                              │             │
└──────────┴──────────────────────────────┴─────────────┘
```

When Analytics is active (`Ctrl+Shift+A`), the Cockpit Area is replaced by `AnalyticsPanel`.

---

## Component Tree

### Root

- **[`App.tsx`](../src/frontend/src/App.tsx)** — Root component. Wires all layout sections, provides `ReactFlowProvider`, manages `showAnalytics` state, registers keyboard shortcuts via `useKeyboardShortcuts`.

### Layout

- **[`layout/Sidebar.tsx`](../src/frontend/src/components/layout/Sidebar.tsx)** — Session history list (fetched from `GET /sessions`), New Mission button, Analytics toggle button (📊, purple accent when active), version badge (fetched from `GET /version`).

- **[`layout/CockpitArea.tsx`](../src/frontend/src/components/layout/CockpitArea.tsx)** — Contains the agent graph (top) and live terminal (bottom) with a drag-resizable vertical split. When Analytics is active, this component is replaced by `AnalyticsPanel` at the `App` level.

- **[`layout/MissionPanel.tsx`](../src/frontend/src/components/layout/MissionPanel.tsx)** — Connection status indicator, prompt editor, launch button, task progress, agent roster, quick stats.

- **[`layout/ResizeHandle.tsx`](../src/frontend/src/components/layout/ResizeHandle.tsx)** — Drag handle for the vertical split between graph and terminal. Emits `mousemove`/`mouseup` events consumed by `useResizable`.

### Graph

- **[`graph/AgentGraph.tsx`](../src/frontend/src/components/graph/AgentGraph.tsx)** — React Flow canvas. Line grid background (`BackgroundVariant.Lines`, gap=40, lineWidth=0.5). No MiniMap. `fitView` called via `onInit` callback (100ms delay) and `useEffect` on node changes (150ms delay). `ResizeObserver` on the container calls `fitView` (50ms debounce) when the container size changes.

- **[`graph/LeaderNode.tsx`](../src/frontend/src/components/graph/LeaderNode.tsx)** — Glassmorphism card, hexagonal "L" badge with purple gradient, top accent line, `pulse-glow-purple` animation when running.

- **[`graph/ExecutorNode.tsx`](../src/frontend/src/components/graph/ExecutorNode.tsx)** — Pill shape (`border-radius: 40px`), circular "E" badge with blue gradient, bottom status bar colored by state.

- **[`graph/AnimatedEdge.tsx`](../src/frontend/src/components/graph/AnimatedEdge.tsx)** — Glow halo path + animated dashes + two staggered SVG `<animateMotion>` particles when the connected agent is active. Static line when idle.

### Terminal

- **[`terminal/LiveTerminal.tsx`](../src/frontend/src/components/terminal/LiveTerminal.tsx)** — JetBrains Mono font, scanline CRT overlay, auto-scroll to bottom on new events, Clear button (filters display without clearing graph state), Export button (downloads `.jsonl`).

- **[`terminal/TerminalLine.tsx`](../src/frontend/src/components/terminal/TerminalLine.tsx)** — Color-coded by event type: `[SESSION]` cyan, `[AGENT]` violet, `[TOOL]` yellow, `[RESULT]` green, `[ERROR]` red. Timestamps use `event.receivedAt ?? event.timestamp ?? ''` as fallback for historical events.

### Mission Panel

- **[`mission/PromptEditor.tsx`](../src/frontend/src/components/mission/PromptEditor.tsx)** — Auto-resize textarea, `Ctrl+Enter` to submit, focus glow.

- **[`mission/LaunchButton.tsx`](../src/frontend/src/components/mission/LaunchButton.tsx)** — 3→2→1→🚀 countdown animation before firing the WebSocket message.

- **[`mission/TaskProgress.tsx`](../src/frontend/src/components/mission/TaskProgress.tsx)** — Parses the Markdown checklist from `tasks.updated` events (`[x]` done, `[-]` in progress, `[ ]` pending). Renders a progress bar and item list.

- **[`mission/AgentRoster.tsx`](../src/frontend/src/components/mission/AgentRoster.tsx)** — Live list of all agents seen in the current session. Each entry shows: status dot (idle/running/done/error), agent name, type badge (L/E), tool call count.

- **[`mission/QuickStats.tsx`](../src/frontend/src/components/mission/QuickStats.tsx)** — Total tool calls, unique agents active, live session duration counter (increments every second while session is running).

### Drawers

- **[`drawers/AgentDetailDrawer.tsx`](../src/frontend/src/components/drawers/AgentDetailDrawer.tsx)** — Slides in from the right when a graph node is clicked (`animation: slide-in-right 250ms ease`). Shows per-agent event timeline (started, tool calls, results, finished). Returns `null` when no agent is selected (no off-screen ghost). Close via `×` button or `Escape`.

### Analytics

- **[`analytics/AnalyticsPanel.tsx`](../src/frontend/src/components/analytics/AnalyticsPanel.tsx)** — Top-level container managing 3-level navigation: `project` → `session` → `agent`. Breadcrumb navigation. Calls `loadSessionLogs` when a session is selected.

- **[`analytics/ProjectView.tsx`](../src/frontend/src/components/analytics/ProjectView.tsx)** — 6 stat cards (Total Missions, Done, Errors, Input Tokens, Output Tokens, Total Cost) from `GET /tokens`. Horizontal bar chart (cost per session, colored by status). Status donut chart (done/error/running). Sessions table (prompt, status, started at, duration, cost) with clickable rows.

- **[`analytics/SessionView.tsx`](../src/frontend/src/components/analytics/SessionView.tsx)** — Header with back button, full prompt, status badge, duration, cost from `GET /sessions/{id}/tokens`. Agent swimlane timeline (div-based horizontal bars, percentage of total session duration). Agent summary table (name, type, LLM turns, tool calls, tokens, cost, avg tool duration) with clickable rows. Tool usage bar chart (call count per tool, colored by success rate). Cost waterfall (stacked bar per agent).

- **[`analytics/AgentView.tsx`](../src/frontend/src/components/analytics/AgentView.tsx)** — Header with agent name, type badge (LEADER/EXECUTOR), model, summary stats. Tokens per turn sparkline (`LineChart`, input=accent, output=purple). Avg tool duration bar chart. Turn-by-turn accordion: stop reason, token counts, tool call count, consecutive errors. Expandable sections: 📋 System Prompt (4-tab: Engine Context / Workspace Tree L4 / Agent Role / Task Context), 💬 Message History, 🔧 Tools Available, 🤖 LLM Response, ⚡ Tool Executions. Expand All / Collapse All button.

---

## Hooks

- **[`useWebSocket.ts`](../src/frontend/src/hooks/useWebSocket.ts)** — WebSocket connection management. Auto-reconnect (max 10 attempts, 3s interval). Exposes `connect`, `send`, `wsEvents`, `connectionStatus`, `latency`, `sessionStartTime`. Sends `ping` heartbeat to measure latency.

- **[`useAgentGraph.ts`](../src/frontend/src/hooks/useAgentGraph.ts)** — Derives `{ nodes, edges }` from the live event stream and the static `/agents` response. `toNodeType()` maps `"leader"` → `"leaderNode"`, `"executor"` → `"executorNode"` for React Flow custom node registration. Uses `@dagrejs/dagre` for automatic layout (`NODE_WIDTH=260`, `NODE_HEIGHT=100`, `nodesep=80`, `ranksep=120`).

- **[`useSessionHistory.ts`](../src/frontend/src/hooks/useSessionHistory.ts)** — Fetches `GET /sessions` and `GET /sessions/{id}/events`. Maps `timestamp → receivedAt` for historical events: `receivedAt: e.receivedAt || e.timestamp || ''`.

- **[`useResizable.ts`](../src/frontend/src/hooks/useResizable.ts)** — Manages the vertical split ratio between graph and terminal via `mousemove`/`mouseup` on `document`.

- **[`useKeyboardShortcuts.ts`](../src/frontend/src/hooks/useKeyboardShortcuts.ts)** — Global `keydown` handler. Shortcuts: `Ctrl+B` (toggle sidebar), `Escape` (close drawer), `Ctrl+Shift+A` (toggle analytics).

- **[`useAnalytics.ts`](../src/frontend/src/hooks/useAnalytics.ts)** — Fetches `GET /sessions/{id}/logs`. Exposes `sessionLogs`, `loadingLogs`, `loadSessionLogs(id)`, `clearSessionLogs`.

---

## Types

- **[`types/events.ts`](../src/frontend/src/types/events.ts)** — `EngineEvent` interface (all event fields including `timestamp?: string`), `EventType` union.

- **[`types/graph.ts`](../src/frontend/src/types/graph.ts)** — `AgentStatus` (`"idle" | "running" | "done" | "error"`), `AgentType` (`"leaderNode" | "executorNode"`), `AgentNodeData`, `StaticAgent` (with `type: "leader" | "executor"` matching the API response).

- **[`types/session.ts`](../src/frontend/src/types/session.ts)** — `SessionMeta` interface (id, prompt, startedAt, finishedAt, status).

- **[`types/logs.ts`](../src/frontend/src/types/logs.ts)** — `LogEntry`, `SystemLayers`, `ContentLog`, `MessageLog`, `ToolCallEntry`, `ToolLog`, `SessionLogs`.

---

## Design System

CSS variables defined in [`src/frontend/src/index.css`](../src/frontend/src/index.css):

| Variable | Value | Usage |
|---|---|---|
| `--bg-base` | `#080c14` | Page background |
| `--bg-surface` | `#0d1117` | Card/panel background |
| `--bg-elevated` | `#161b22` | Elevated surfaces |
| `--accent` | `#58a6ff` | Blue accent (executor nodes, links) |
| `--purple` | `#bc8cff` | Purple accent (leader nodes, analytics) |
| `--success` | `#3fb950` | Done/success states |
| `--error` | `#f85149` | Error states |
| `--text-primary` | `#e6edf3` | Primary text |
| `--text-muted` | `#8b949e` | Secondary/muted text |

Typography: **Inter** (UI text) + **JetBrains Mono** (terminal), loaded from Google Fonts in `index.html`.

Keyframe animations: `node-appear`, `edge-glow-pulse`, `particle-trail`, `pulse-glow`, `pulse-glow-purple`, `pulse-border`, `dash-flow`, `slide-in-right`.

---

## Development

```cmd
cd src/frontend
npm run dev
```

The Vite dev server proxies `/ws` → `ws://localhost:8080` so the frontend connects to a locally running `ai-engine` binary.

```cmd
npm run build
```

Outputs to `src/backend/frontend-dist/`. The Go binary embeds this directory via `src/backend/embed.go` at compile time.

---

## Analytics Toggle

The Analytics panel replaces the Cockpit Area when active. Toggle via:
- `Ctrl+Shift+A` keyboard shortcut
- The 📊 Analytics button in the Sidebar

When active, the Sidebar button is highlighted with the purple accent color.
