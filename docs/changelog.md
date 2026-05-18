# Changelog

## [0.0.22] — 2026-05-18

### Fixed

#### Analytics Panel — crash ao abrir (`TypeError: Cannot read properties of undefined (reading 'toFixed')`)

- **Root cause:** mismatch de nome de campo entre backend e frontend. O backend (`tokenstore`) serializa o campo como `estimated_cost_usd`, mas as interfaces `TokenData` em `ProjectView.tsx` e `SessionView.tsx` declaravam `cost_usd`. O campo chegava como `undefined` no frontend, causando `undefined.toFixed()` → crash do React.
- [`frontend/src/components/analytics/ProjectView.tsx`](../frontend/src/components/analytics/ProjectView.tsx) — interface `TokenData.cost_usd` → `estimated_cost_usd`; guard `?.estimated_cost_usd != null` adicionado na StatCard "Total Cost"; uso na linha de fetch de custos por sessão corrigido.
- [`frontend/src/components/analytics/SessionView.tsx`](../frontend/src/components/analytics/SessionView.tsx) — interface `TokenData.cost_usd` → `estimated_cost_usd`; render do custo total da sessão corrigido com guard `?.estimated_cost_usd != null`.
- Testado via Playwright no workspace `projects/budget-tracker/`: ProjectView, SessionView e AgentView renderizando corretamente sem crash.

---

## [0.0.18] — 2026-05-18

### Added

#### Analytics BI Panel — full observability dashboard for sessions and agents

**Backend**

- **`GET /sessions/{id}/logs`** — new endpoint in [`internal/server/server.go`](../internal/server/server.go) that reads per-agent `chat.jsonl` log files from `.ai-engine/logs/{sessionId}/{agentName}/chat.jsonl` and returns them as `{ agents: Record<string, LogEntry[]> }`. Returns empty agents map if the logs directory does not exist.
- New imports: `bytes`, `os`, `path/filepath` added to server package.

**Frontend — new types**

- [`frontend/src/types/logs.ts`](../frontend/src/types/logs.ts) — complete TypeScript type definitions for the log format: `LogEntry`, `SystemLayers`, `ContentLog`, `MessageLog`, `ToolCallEntry`, `ToolLog`, `SessionLogs`.

**Frontend — new hook**

- [`frontend/src/hooks/useAnalytics.ts`](../frontend/src/hooks/useAnalytics.ts) — `useAnalytics()` hook: `loadSessionLogs(id)` fetches `GET /sessions/{id}/logs`, stores result in state, exposes `sessionLogs`, `loadingLogs`, `clearSessionLogs`.

**Frontend — Analytics components (4 new files)**

- [`frontend/src/components/analytics/ProjectView.tsx`](../frontend/src/components/analytics/ProjectView.tsx) — macro project view:
  - 6 stat cards: Total Missions, Done, Errors, Input Tokens, Output Tokens, Total Cost (from `GET /tokens`).
  - Horizontal bar chart (recharts `BarChart`) — cost per session, colored by status (accent=done, error=error), clickable bars navigate to session view.
  - Status donut chart (recharts `PieChart`) — done/error/running distribution with count labels.
  - Sessions table — prompt, status, started at, duration, cost; clickable rows.

- [`frontend/src/components/analytics/SessionView.tsx`](../frontend/src/components/analytics/SessionView.tsx) — session drill-down:
  - Header with back button, full prompt, status badge, duration, cost from `GET /sessions/{id}/tokens`.
  - Agent swimlane timeline — SVG-free div-based horizontal bars showing each agent's time span as a percentage of total session duration.
  - Agent summary table — agent name, type badge (L/E), LLM turns, tool calls, input/output tokens, cost, avg tool duration; clickable rows navigate to agent view.
  - Tool usage bar chart (recharts `BarChart`) — call count per tool, colored by success rate (green ≥90%, yellow ≥50%, red <50%).
  - Cost waterfall — horizontal stacked bar showing each agent's cost contribution as a percentage of total session cost, with legend.

- [`frontend/src/components/analytics/AgentView.tsx`](../frontend/src/components/analytics/AgentView.tsx) — agent drill-down:
  - Header: agent name, type badge (LEADER/EXECUTOR), model, summary stats (turns, tool calls, tokens, cost).
  - Tokens per turn sparkline (recharts `LineChart`, input=accent, output=purple).
  - Avg tool duration bar chart (recharts `BarChart`, per tool name).
  - Turn-by-turn accordion: each turn shows stop reason, token counts, tool call count, consecutive errors. Expandable sections: 📋 System Prompt (4-tab: Engine Context / Workspace Tree L4 / Agent Role / Task Context), 💬 Message History (collapsible list with role + preview), 🔧 Tools Available, 🤖 LLM Response (text + stop reason + tokens), ⚡ Tool Executions (input JSON + output, success/error badge, duration).
  - Expand All / Collapse All button.

- [`frontend/src/components/analytics/AnalyticsPanel.tsx`](../frontend/src/components/analytics/AnalyticsPanel.tsx) — top-level container managing 3-level navigation: `project` → `session` → `agent`. Breadcrumb + Project shortcut button. Calls `loadSessionLogs` on session selection.

**Frontend — integration**

- [`frontend/src/App.tsx`](../frontend/src/App.tsx) — added `showAnalytics` state, `useAnalytics()` hook, `handleToggleAnalytics` callback, `Ctrl+Shift+A` keyboard shortcut. When `showAnalytics=true`, renders `<AnalyticsPanel>` instead of `<CockpitArea>`.
- [`frontend/src/components/layout/Sidebar.tsx`](../frontend/src/components/layout/Sidebar.tsx) — added `onToggleAnalytics` and `showAnalytics` props; new `📊 Analytics` button above the New Mission button, styled with purple accent when active.

### Changed

- Bundle size: 841 KB (721 modules, +33 recharts packages).
- Binary: `ai-engine.exe` v0.0.18.

---

## [0.0.17] — 2026-05-18

### Added

#### `internal/chatlog/logger.go` — Comprehensive debug logging structs

- **`LogEntry`** extended with new roles and fields:
  - `role=agent_init` — written once at agent start (`turn=0`): captures `agent_name`, `agent_type`, `session_id`, `model`.
  - `role=llm_request` — written before every `provider.Send()` call: captures `model`, `system_prompt` (full composed string), `system_layers` (each layer individually), `messages` (full history), `tools` (all definitions), `message_count`, `total_tool_calls_so_far`, `consecutive_errors`.
  - `role=llm_response` — written after every `provider.Send()` call: captures `text`, `tool_calls`, `stop_reason`, `input_tokens`, `output_tokens`. On provider error, written with `stop_reason=error`.
  - `duration_ms` added to `role=tool_result` entries — wall-clock milliseconds for each tool execution.
- **`SystemLayers`** — new struct holding each system prompt layer individually: `engine_context` (Layer 1), `dynamic_context` (Layer 4), `agent_role` (Layer 2), `task_context` (Layer 3).
- **`MessageLog`** / **`ContentLog`** — serializable snapshots of `llm.Message` / `llm.ContentBlock` for embedding the full conversation history in `llm_request` entries.
- **`ToolLog`** — serializable snapshot of `llm.ToolDefinition` (name, description, input_schema) for embedding tool definitions in `llm_request` entries.

#### `internal/agent/runner.go` — Complete LLM turn instrumentation

- `agent_init` entry written immediately after `logger.Open()` — one entry per agent execution.
- `llm_request` entry written before every `provider.Send()` — captures the exact payload sent to the Anthropic API including the full 4-layer system prompt and complete message history.
- `llm_response` entry written after every `provider.Send()` — replaces the former `role=assistant` entry; includes token usage and stop reason.
- `duration_ms` recorded for every tool execution via `time.Now()` / `time.Since()` around `tool.Execute()`.
- Three helper functions added (unexported):
  - [`toMessageLog()`](internal/agent/runner.go) — converts `[]llm.Message` → `[]chatlog.MessageLog`.
  - [`toToolLog()`](internal/agent/runner.go) — converts `[]llm.ToolDefinition` → `[]chatlog.ToolLog`.
  - [`toToolCallEntries()`](internal/agent/runner.go) — converts `[]llm.ToolCall` → `[]chatlog.ToolCallEntry`.
- Former `role=assistant` log entry removed — superseded by `role=llm_response`.
- All log writes remain non-fatal (errors logged to stderr, never terminate the session).

### Changed

- `chat.jsonl` log format is fully backward-compatible — all new fields are additive (`omitempty`).
- Binary: `ai-engine.exe` v0.0.17.

---

## [0.0.16] — 2026-05-18

### Added

#### `projects/flashcard-app/` — New test workspace (Layer 4 validation)
- New example workspace: **Flashcard Study App** — create decks, add cards, study with flip animation, track correct/incorrect stats per deck.
- Binary: `ai-engine.exe` v0.0.16 (first binary with Dynamic Context Injection — Layer 4).
- Config: `claude-sonnet-4-6`, port `8080`, `max_tool_calls: 60`.
- Agent tree: `swarmito` → `backend-leader` → `backend-executor` + `frontend-leader` → `frontend-executor`.
- Backend spec: Go REST API on port `8081` — deck CRUD, card CRUD per deck, study result recording (correct/incorrect), stats per deck.
- Frontend spec: single-file HTML — deck list, deck detail, study mode with card flip animation, stats panel. Dark modern UI.

### Changed

#### Dynamic Context Injection (Layer 4) — first production use
- This is the first binary compiled after implementing Layer 4 (`internal/dyncontext/`).
- `WorkspaceTreeProvider` is active by default — agents receive the current workspace file tree in their system prompt on every LLM call.

---

## [0.0.15] — 2026-05-18

### Added

#### `internal/sandbox/shell.go` — Persistent shell per agent
- New `Shell` type: a persistent `cmd.exe` (Windows) / `sh` (Unix) process that stays alive for the entire duration of a single agent execution.
- Uses `os.Pipe` to merge stdout and stderr into a single reader — avoids the Go `exec` limitation of not being able to use both `StdoutPipe` and `Stderr` simultaneously.
- Sentinel-based output reading: each command is followed by `echo __AI_ENGINE_CMD_DONE_7f3a9b2c__`; the reader collects lines until the sentinel is detected, signalling end of output.
- Timeout via `time.After`: if the sentinel is not received within `timeout_seconds`, the shell process is killed and `[TIMEOUT after Xs]` is prepended to any partial output collected.
- `Shell.Close()` is safe to call multiple times (guarded by `closed` flag + mutex).

#### `internal/sandbox/shell_windows.go` — Windows Job Object
- Build-tagged `//go:build windows`.
- `attachJobObject()`: creates a Windows Job Object via `windows.CreateJobObject`, configures `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`, and assigns the shell process to it via `windows.AssignProcessToJobObject`.
- `closeJobObject()`: closes the Job Object handle, triggering OS-level kill of the entire process tree — including background processes started with `&`.
- Uses `golang.org/x/sys/windows` (new dependency added to `go.mod`).

#### `internal/sandbox/shell_unix.go` — Unix stubs
- Build-tagged `//go:build !windows`.
- No-op implementations of `attachJobObject` and `closeJobObject` — on Unix, `exec.CommandContext` with `SIGKILL` already kills the process group correctly.

### Changed

#### `internal/tools/run_terminal_command.go` — Uses persistent Shell
- `RunTerminalCommand` now holds a `*sandbox.Shell` instead of `*sandbox.Sandbox`.
- `Execute()` delegates to `shell.Exec(command, timeout)` — no longer spawns a new `cmd.exe /C` process per call.
- `workdir` parameter removed from the input schema — agents use `cd` commands to change directories, and the change persists for subsequent calls in the same shell session.
- **Root cause fixed:** background processes started with `&` are now inside the shell's Job Object and are killed when the shell closes, eliminating orphan processes that held ports open after timeout.

#### `internal/tools/registry.go` — Shell injection
- `NewRegistry` now accepts `*sandbox.Shell` as a second parameter (may be `nil` for deferred injection).
- New method `SetShell(shell *sandbox.Shell)` — injects the shell into the `RunTerminalCommand` tool after the registry is created. Called by `Runner.Run()` at agent start.

#### `internal/agent/runner.go` — Shell lifecycle management
- `Run()` now calls `sandbox.NewShell(r.sb.WorkspacePath())` at the start of each agent execution.
- `defer shell.Close()` ensures the shell (and all its child processes via Job Object) is terminated when the agent finishes, errors out, or is cancelled.
- `r.tools.SetShell(shell)` injects the shell into the tool registry before the agent loop begins.

#### `internal/server/server.go` — Registry construction
- Both `tools.NewRegistry` calls now pass `nil` as the shell parameter — the shell is created and injected by `Runner.Run()`, not at registry construction time.

---

## [0.0.14] — 2026-05-18

### Added

#### `projects/kanban-board/` — New test workspace (4-level agent hierarchy)
- New example workspace: **Kanban Board** — a task management app with drag-and-drop support.
- Binary: `ai-engine.exe` v0.0.14.
- Config: `claude-sonnet-4-6`, port `8080`, `max_tool_calls: 60`.
- **4-level agent hierarchy** (deepest tree in any example workspace):
  ```
  swarmito (L1)
  ├── backend-leader (L2)
  │   ├── data-leader (L3) → data-executor (L4)
  │   └── api-leader (L3)  → api-executor (L4)
  └── frontend-leader (L2)
      └── ui-leader (L3)   → ui-executor (L4)
  ```
- Agent responsibilities:
  - `swarmito`: root orchestrator — delegates to backend-leader and frontend-leader
  - `backend-leader`: coordinates data-leader and api-leader in sequence (data first, then API)
  - `data-leader`: coordinates data-executor to implement Go data models and in-memory storage
  - `data-executor`: writes `backend/main.go` data layer (`Card` struct, `Store`, CRUD methods)
  - `api-leader`: coordinates api-executor to implement HTTP handlers on top of the data layer
  - `api-executor`: writes HTTP handlers, CORS middleware, routing, and `main()` on port 8081
  - `frontend-leader`: coordinates ui-leader
  - `ui-leader`: coordinates ui-executor to implement the single-file HTML frontend
  - `ui-executor`: writes `frontend/index.html` with drag-and-drop Kanban UI
- Backend spec: Go REST API on port `8081` — `POST /cards`, `GET /cards`, `PATCH /cards/{id}/move`, `DELETE /cards/{id}`. Standard library only, no external deps.
- Frontend spec: single-file `frontend/index.html`, dark theme (`#0d1117`), three columns (To Do / In Progress / Done), HTML5 drag-and-drop, add card form, delete buttons.
- Workspace initialised via `ai-engine.exe init`.

---

## [0.0.13] — 2026-05-18

### Added

#### `build.cmd` — Patch bump
- Version bumped from `0.0.12` → `0.0.13` via `build.cmd patch`.
- Binary compiled with updated frontend embedded.

---

## [0.0.12] — 2026-05-15

### Fixed

#### `frontend/src/components/graph/AgentGraph.tsx` — Remove MiniMap (overlapping nodes)
- Removed `<MiniMap>` component entirely — it was positioned at `x=672, y=383` directly overlapping the `backend-executor` node, making it appear cut off.
- Increased `fitView` padding from `0.3` → `0.4` for more breathing room around nodes.

---

## [0.0.11] — 2026-05-15

### Fixed

#### `frontend/src/components/graph/AgentGraph.tsx` — fitView not centering graph
- Replaced static `fitView` / `fitViewOptions` props with dynamic `onInit` callback that calls `instance.fitView({ padding: 0.3, duration: 400 })` after 100ms delay (allows DOM to settle before calculating viewport).
- Added `useEffect` that re-fits whenever `nodes` changes (150ms delay) — ensures graph re-centers when new agents appear during a session.

---

## [0.0.10] — 2026-05-15

### Fixed

#### `frontend/src/hooks/useAgentGraph.ts` — Dagre layout constants updated for larger nodes
- `NODE_WIDTH` updated `200` → `260` to match redesigned node widths (~240px).
- `NODE_HEIGHT` updated `70` → `100` to match redesigned node heights (~95px).
- `nodesep` updated `60` → `80`, `ranksep` updated `100` → `120` for more elegant spacing.

#### `frontend/src/components/graph/AgentGraph.tsx` — fitView padding
- `fitViewOptions.padding` updated `0.2` → `0.3`.

---

## [0.0.9] — 2026-05-15

### Fixed

#### `frontend/src/hooks/useAgentGraph.ts` — Node type mapping bug (root cause of plain white nodes)
- **Root cause:** `/agents` endpoint returns `"type": "leader"` / `"type": "executor"`, but React Flow requires `"leaderNode"` / `"executorNode"` to match registered custom node types. `populateFromStatic()` was passing the raw API value directly, causing React Flow to fall back to `"default"` and render plain white-bordered boxes.
- Added `toNodeType()` mapping function: `"leader"` → `"leaderNode"`, `"executor"` → `"executorNode"`.
- `populateFromStatic()` now calls `toNodeType(agent.type)` instead of passing the raw value.

#### `frontend/src/types/graph.ts` — StaticAgent type corrected
- `StaticAgent.type` changed from `AgentType` (`"leaderNode" | "executorNode"`) to `"leader" | "executor"` to match the actual API response format.

### Added

#### Frontend — Agent graph visual redesign (custom nodes now rendering)
- **`LeaderNode.tsx`**: 220px min-width, 32×32 hexagonal badge, inner radial glow, full-width 3px accent line, tool call count badge, animated blink dot on running status, hover lift effect, `node-appear` entrance animation.
- **`ExecutorNode.tsx`**: 190px min-width, 28×28 circular badge, inner radial glow, 3px status bar with `box-shadow`, tool call count badge, animated blink dot on running, hover lift, `node-appear` animation.
- **`AnimatedEdge.tsx`**: Glow halo path (`strokeWidth: 6, opacity: 0.15`, `edge-glow-pulse` animation), base `strokeWidth: 2/1.5`, two particles (`dur="1.2s"` and `dur="1.8s" begin="0.6s"`), static edges use `rgba(33,38,45,0.6)`.
- **`AgentGraph.tsx`**: `BackgroundVariant.Lines` with `gap=40, lineWidth=0.5`, radial gradient depth overlay, empty state with inline SVG network icon.
- **`index.css`**: Added `node-appear`, `edge-glow-pulse`, `particle-trail` keyframe animations.

---

## [0.0.8] — 2026-05-15

### Added

#### `build.cmd` — Version injection via ldflags
- Added `-ldflags "-X main.Version=%NEW_VERSION%"` to the Go build command — version is now injected into the binary at compile time.

#### `cmd/ai-engine/main.go` — Version variable
- Declared `var Version = "dev"` (overridden by ldflags at build time).
- Startup log now prints `Version : x.y.z`.

#### `internal/server/server.go` — `/version` endpoint
- Added `version string` field to `Server`.
- `New()` now accepts `version string` as a parameter.
- New endpoint `GET /version` returns `{"version": "x.y.z"}`.

#### `frontend/src/components/layout/Sidebar.tsx` — Real version badge
- Version badge now fetches `GET /version` on mount and displays the real binary version (e.g. `0.0.8`) instead of the hardcoded `v2`.

---

## [0.0.7] — 2026-05-15

### Added

#### `internal/sessionstore/store.go` — New session persistence package
- New `Store` type that persists session data to `.ai-engine/sessions/{id}/`.
- Writes `meta.json` (id, prompt, startedAt, finishedAt, status) and `events.jsonl` (one JSON event per line).
- Methods: `StartSession`, `AppendEvent`, `FinishSession`, `ReadMeta`, `ListSessions`, `ReadEvents`.
- Non-fatal errors: logging only, never terminates the session.

#### `internal/server/server.go` — Session store integration + new endpoints
- Added `store *sessionstore.Store` field to `Server`.
- `New()` now accepts store as a parameter.
- On `user.message`: calls `store.StartSession` and subscribes to the event bus to `AppendEvent` for every event and `FinishSession` on `session.finished`/`error`.
- New endpoint `GET /sessions` — returns list of session metas.
- New endpoint `GET /sessions/{id}/events` — returns events array from JSONL.

#### `cmd/ai-engine/main.go` — Store instantiation
- Instantiates `sessionstore.New(workspacePath)` and passes it to `server.New`.

### Changed

#### `frontend/src/hooks/useSessionHistory.ts` — Complete rewrite
- Replaced `localStorage` with server-fetching via `GET /sessions` and `GET /sessions/{id}/events`.
- New interface: `{ sessions, activeSessionId, setActiveSessionId, loadSessionEvents, refreshSessions }`.

#### `frontend/src/App.tsx` — Updated session history usage
- `handleLoadSession` is now async, calls `loadSessionEvents(id)`.
- `useEffect` on `wsEvents` calls `refreshSessions` on `session.started`, `session.finished`, and `error` events.
- Removed `saveEvent` call (server handles persistence).

#### `frontend/src/components/layout/Sidebar.tsx` — Updated session type
- Updated `sessions` prop type from `SessionRecord` to `SessionMeta` (imported from `useSessionHistory`).

**Impact:** session history is now isolated per project (each workspace has its own `.ai-engine/sessions/`). Sessions persist across browser tabs, ports, and browser restarts.

---

## [0.0.6] — 2026-05-15

### Changed

#### `frontend/src/components/graph/LeaderNode.tsx` — Full redesign
- Glassmorphism card with purple gradient (`rgba(188,140,255,0.08)`).
- Hexagonal "L" badge with `linear-gradient(135deg, #bc8cff, #7c3aed)`.
- Dynamic border color by status; top accent line.
- `pulse-glow-purple` animation when running.
- Status labels in monospace uppercase.

#### `frontend/src/components/graph/ExecutorNode.tsx` — Full redesign
- Pill shape (`border-radius: 40px`), blue gradient (`rgba(88,166,255,0.06)`).
- Circular "E" badge with `linear-gradient(135deg, #58a6ff, #1d4ed8)`.
- Bottom status bar colored by state; glow by status.

#### `frontend/src/components/graph/AnimatedEdge.tsx` — Full redesign
- Base path + animated dashes (`stroke-dasharray: 8 12`).
- SVG `<animateMotion>` particle traversing the bezier path when agent is running.

#### `frontend/src/index.css` — New keyframe
- Added `@keyframes pulse-glow-purple` for leader node running animation.

### Added

#### `projects/link-vault/` — New test workspace
- **Link Vault** — personal bookmark manager.
- Binary: `ai-engine.exe` v0.0.6.
- Config: `claude-sonnet-4-6`, port `8085`, `max_tool_calls: 60`.
- Agent tree: `swarmito` → `backend-leader` → `backend-executor` + `frontend-leader` → `frontend-executor`.
- Backend spec: Go REST API on port `8085` — save links (title, URL, tags, notes), list with tag filter, delete, list unique tags.
- Frontend spec: single-file HTML, dark UI, link cards with colored tag badges, tag filter, open link in new tab.

---

## [0.0.5] — 2026-05-15

### Fixed

#### `frontend/src/components/graph/AgentGraph.tsx` — React Flow CSS override
- Replaced `@xyflow/react/dist/style.css` with `@xyflow/react/dist/base.css` to prevent React Flow from injecting white node backgrounds that overrode custom dark styles.

#### `frontend/src/index.css` — Residual node style removal
- Added `.react-flow__node { background: transparent; border: none; padding: 0; border-radius: 0 }` override block to remove residual React Flow default node styles.

#### `frontend/src/App.tsx` — Session history save fix
- Added `saveEvent` to `useSessionHistory` destructuring and called it in `useEffect`.
- **Root cause:** events were never being saved to session history, making session replay show empty results.

#### `internal/server/server.go` — Ping no-op
- Added `case "ping":` no-op in the WebSocket event switch to silence the `unknown event type "ping"` log noise from the frontend heartbeat.

### Added

#### `projects/habit-tracker/` — New test workspace
- **Habit Tracker** workspace.
- Binary: `ai-engine.exe` v0.0.5.
- Config: `claude-sonnet-4-6`, port `8084`, `max_tool_calls: 60`.
- Agent tree: `swarmito` → `backend-leader` → `backend-executor` + `frontend-leader` → `frontend-executor`.
- Backend spec: Go REST API on port `8084` — create habits, mark as done for today, list with streak (consecutive days completed).
- Frontend spec: single-file HTML, dark UI, habit list with checkboxes, streak indicator.

---

## [0.0.4] — 2026-05-15

### Added

#### `frontend/` — Frontend v2 "Mission Control" (branch `frontend-v2-test2`)
Complete rewrite of the frontend. The previous single-panel layout (graph + event feed + prompt input) was replaced by a full "Mission Control" cockpit interface designed for developers.

**Layout — 3 columns:**
- **Sidebar (240px, collapsible)** — session history persisted in `localStorage` (max 20 sessions), relative timestamps, status badges, replay mode, "New Mission" button. Collapse/expand via `Ctrl+B`.
- **Cockpit Area (flex: 1)** — vertically split between Agent Graph (top) and Live Terminal (bottom). Split ratio is drag-resizable via a `ResizeHandle` component.
- **Mission Panel (360px)** — all mission controls: connection status + latency, prompt editor, launch button, task progress, agent roster, quick stats.

**Agent Graph — redesigned:**
- `LeaderNode.tsx` — glassmorphism card, hexagonal "L" badge, pulsing glow when running, `--purple` accent border.
- `ExecutorNode.tsx` — pill shape, circular "E" badge, status-colored border.
- `AnimatedEdge.tsx` — SVG `stroke-dashoffset` particle animation while agent is active; static line when idle.
- `AgentGraph.tsx` — dot grid background, `<MiniMap>`, `<Controls>`, empty state message, click-to-open drawer.
- Node type inference updated: `swarmito` added to leader detection pattern.

**Live Terminal — replaces EventFeed:**
- `LiveTerminal.tsx` — real terminal aesthetic, `JetBrains Mono` font, scanline CRT overlay, auto-scroll, Clear button, Export `.jsonl` download.
- `TerminalLine.tsx` — color-coded lines by event type: `[SESSION]` cyan, `[AGENT]` violet, `[TOOL]` yellow, `[RESULT]` green, `[ERROR]` red.

**Mission Panel components:**
- `PromptEditor.tsx` — auto-resize textarea, `Ctrl+Enter` to submit, focus glow.
- `LaunchButton.tsx` — countdown animation 3→2→1→🚀 before firing.
- `TaskProgress.tsx` — parses Markdown checklist from `tasks.updated` events, renders progress bar with `[x]`/`[-]`/`[ ]` items.
- `AgentRoster.tsx` — live list of all agents with status dot, type badge (L/E), tool call count.
- `QuickStats.tsx` — total tool calls, unique agents, live session duration counter.

**Agent Detail Drawer:**
- `AgentDetailDrawer.tsx` — slides in from the right when a graph node is clicked. Shows per-agent event timeline (started, tool calls, results, finished). Close via `×` button or `Escape`.

**New hooks:**
- `useSessionHistory.ts` — `localStorage` persistence, `startSession`, `saveEvent`, `finishSession`, `loadSession`, max 20 sessions.
- `useResizable.ts` — drag-to-resize vertical split ratio via `mousemove`/`mouseup` on `document`.
- `useKeyboardShortcuts.ts` — global `keydown` handler, `Ctrl+B` (sidebar), `Escape` (drawer).
- `useWebSocket.ts` — extended with auto-reconnect (max 10 attempts, 3s interval), `latency: number | null`, `sessionStartTime: Date | null`, `connectionStatus` string.

**New types:**
- `types/session.ts` — `SessionRecord` interface for localStorage persistence.
- `types/graph.ts` — `AgentType` updated to `'leaderNode' | 'executorNode'` (React Flow custom node type names).

**Design system:**
- `index.css` — full CSS variable palette (`--bg-base: #080c14`, `--accent: #58a6ff`, `--purple: #bc8cff`, etc.), global reset, scrollbar styling, keyframe animations (`pulse-glow`, `pulse-border`, `dash-flow`, `slide-in-right`).
- Typography: `Inter` (UI) + `JetBrains Mono` (terminal), loaded from Google Fonts in `index.html`.
- Build: 194 modules, 441KB bundle, 0 TypeScript errors.

#### `projects/expense-tracker/` — New example workspace
- New example workspace: **Personal Expense Tracker** — expense management with category filtering and summary.
- Binary: `ai-engine.exe` v0.0.4 (with Frontend v2 embedded).
- Config: `claude-sonnet-4-6`, port `8080`, `max_tool_calls: 60`.
- Agent tree: `swarmito` → `backend-leader` → `backend-executor` + `frontend-leader` → `frontend-executor`.
- Backend spec: Go REST API on port `8083`, endpoints `POST /expenses`, `GET /expenses`, `DELETE /expenses/{id}`, `GET /summary`, `GET /categories`. Standard library only, no external deps.
- Frontend spec: single-file `frontend/index.html`, dark theme, expense form with category selector, filterable list, CSS-only horizontal bar chart for spending by category.
- Workspace initialised via `ai-engine.exe init`.

---

## [0.0.3] — 2026-05-15

### Fixed

#### `frontend/src/components/AgentGraphNode.tsx` — Missing `Handle` components
- Added `<Handle type="target" position={Position.Top}>` and `<Handle type="source" position={Position.Bottom}>` to the custom node component.
- **Root cause:** React Flow v12 requires custom nodes to explicitly declare `Handle` components for edges to be rendered. Without handles, React Flow has no connection points and silently discards all edges — the edge container existed in the DOM but was always empty (confirmed via browser inspection: 0 `.react-flow__edge` elements).
- Handles are styled with `width: 8`, `height: 8`, `border: none`, and dynamic color matching the node status.
- **Result:** 4 edges now render correctly in the agent graph (swarmito → frontend-leader, swarmito → backend-leader, frontend-leader → frontend-executor, backend-leader → backend-executor).

---

## [0.0.2] — 2026-05-15

### Fixed

#### `frontend/src/App.tsx` + `frontend/src/components/AgentGraph.tsx` — `ReactFlowProvider` placement
- Moved `ReactFlowProvider` from `AgentGraph.tsx` (same component as `ReactFlow`) to `App.tsx` (parent component).
- **Root cause:** React Flow v12 requires `ReactFlowProvider` to be in a **parent component** in the tree — having it wrap `ReactFlow` in the same component creates an isolated context that prevents edges from rendering.
- `AgentGraph.tsx` now renders `<ReactFlow>` directly inside `<div className="graph-panel">` without a provider wrapper.

---

## [0.0.1] — 2026-05-15

### Added

#### `projects/quiz-app/` — New example workspace
- New example workspace: **Quiz App** — a multiple-choice quiz with session management, scoring, and leaderboard.
- Backend: Go REST API on port `8082` with 10+ hardcoded questions across Science and History categories.
  - Endpoints: `GET /questions`, `POST /session/start`, `POST /session/answer`, `GET /session/{id}/result`, `GET /leaderboard`
- Frontend: single-file `frontend/index.html` with dark theme, start screen, quiz flow with answer feedback, result screen, and leaderboard view.
- Agent tree: `swarmito` → `backend-leader` → `backend-executor` + `frontend-leader` → `frontend-executor`
- Config: `claude-sonnet-4-6`, port `8080`, `max_tool_calls: 60`
- Workspace initialised via `ai-engine.exe init`

#### `build.cmd` — Semver versioning system
- `build.cmd` now accepts a bump type argument: `patch` (default), `minor`, or `major`.
- Reads and writes current version from/to [`version.txt`](../version.txt) (semver format `MAJOR.MINOR.PATCH`).
- Compiles binary to `bin\{version}\ai-engine.exe`.
- Always copies the latest build to `bin\latest\ai-engine.exe`.
- No longer accepts a destination folder as argument.

#### `version.txt`
- New file at repository root tracking the current semver version.
- Initial value: `0.0.0`.

### Fixed

#### `frontend/src/components/AgentGraph.tsx` — Missing `ReactFlowProvider`
- Added `ReactFlowProvider` wrapper around `<ReactFlow>` to fix missing edges in the agent graph view.
- **Root cause:** `@xyflow/react` v12 requires a `ReactFlowProvider` in the component tree to render edges. Without it, nodes appeared but connections were invisible.

---

## Usage

### Build

```cmd
build.cmd          # patch bump (default): 0.0.0 → 0.0.1
build.cmd patch    # same as above
build.cmd minor    # minor bump: 0.0.1 → 0.1.0
build.cmd major    # major bump: 0.1.0 → 1.0.0
```

Output:
- `bin\{version}\ai-engine.exe` — versioned binary
- `bin\latest\ai-engine.exe` — always the latest build

### Running the Quiz App workspace

```cmd
cd projects\quiz-app
:: Add ANTHROPIC_API_KEY to .ai-engine\.env first
ai-engine.exe
```

Then open `http://localhost:8080` in your browser and send the prompt:

> Build a Quiz App with a Go REST API backend (port 8082) and a single-file HTML/CSS/JS frontend. The backend should have at least 10 multiple-choice questions across Science and History categories, session management with scoring, and a leaderboard. The frontend should have a start screen, quiz flow with answer feedback, result screen, and leaderboard view with a dark modern UI.

### Running the Expense Tracker workspace (Frontend v2 "Mission Control")

```cmd
cd projects\expense-tracker
:: Add ANTHROPIC_API_KEY to .ai-engine\.env first
ai-engine.exe
```

Then open `http://localhost:8080` in your browser (Frontend v2 "Mission Control") and send the prompt:

> Build a Personal Expense Tracker app. The backend should be a Go REST API on port 8083 with endpoints to create, list, and delete expenses, plus a summary endpoint grouped by category. The frontend should be a single HTML file with a form to add expenses (description, amount, category, date), a filterable expense list with delete buttons, and a summary panel with a CSS-only bar chart showing spending by category. Use a clean dark UI.
