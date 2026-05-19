# Changelog

## [0.0.27] — 2026-05-19

### Changed — Source code reorganized into `src/`

The entire source tree was reorganized from a flat layout into a structured `src/` directory, separating frontend and backend concerns.

#### Files moved

| Before | After |
|--------|-------|
| `frontend/` | `src/frontend/` |
| `cmd/` | `src/backend/cmd/` |
| `internal/` | `src/backend/internal/` |
| `embed.go` | `src/backend/embed.go` |
| `go.mod` | `src/backend/go.mod` |
| `go.sum` | `src/backend/go.sum` |

#### Files modified

- [`src/frontend/vite.config.ts`](../src/frontend/vite.config.ts) — `outDir` changed from `'dist'` to `'../backend/frontend-dist'`; added `emptyOutDir: true`. Vite now outputs directly into `src/backend/frontend-dist/` so the Go embed directive can reference it without `..` path restrictions.
- [`src/backend/embed.go`](../src/backend/embed.go) — `//go:embed frontend/dist` → `//go:embed frontend-dist`.
- [`src/backend/cmd/ai-engine/main.go`](../src/backend/cmd/ai-engine/main.go) — `fs.Sub(aiengine.Files, "frontend/dist")` → `fs.Sub(aiengine.Files, "frontend-dist")`.
- [`build.cmd`](../build.cmd) — frontend build step now `cd src\frontend`; Go build step executes from `src\backend\` with output path adjusted back to repo root.
- [`.gitignore`](../.gitignore) — added `src/backend/frontend-dist/` (generated at build time).
- [`README.md`](../README.md) — updated path references.
- [`docs/frontend.md`](../docs/frontend.md) — updated all component links (`../frontend/src/...` → `../src/frontend/src/...`), development section (`cd frontend` → `cd src/frontend`), and build output path.

#### Design rationale

- `go.mod` moved to `src/backend/` so the Go module root is co-located with all Go source files.
- Vite `outDir` set to `../backend/frontend-dist` (relative to `src/frontend/`) so the build output lands inside `src/backend/` — this avoids the `//go:embed` restriction that prohibits `..` paths.
- The Go module name `github.com/swarmit/ai-engine` is unchanged; no import paths changed.

### Added — New test workspace `projects/recipe-roulette/`

New example workspace: **Recipe Roulette** — a web app where users spin a wheel to get a random recipe suggestion.

- Binary: `ai-engine.exe` v0.0.27 (first binary built from the `src/` layout).
- Config: model `claude-sonnet-4-6`, port `8080`, `max_tool_calls: 80`.
- Agent tree: `swarmito` → `backend-leader` → `backend-executor` + `frontend-leader` → `frontend-executor`.
- Backend spec: Go REST API on port `8081` — recipe list, recipe detail, random recipe, category filter, favorites (add/remove/list). In-memory storage, 12+ seeded recipes across 4 categories (`breakfast`, `lunch`, `dinner`, `dessert`). Each recipe: `id`, `name`, `category`, `prep_time`, `ingredients` (array), `instructions`.
- Frontend spec: single-file `frontend/index.html` — spinning wheel animation for random recipe, recipe detail card, favorite toggle, favorites page, category filter tabs. Dark UI with orange accent (`#f97316`).

## [0.0.26] — 2026-05-19

### Changed — Filesystem-based agent hierarchy + in-memory AgentTree

Complete refactor of how agent hierarchy is defined and loaded. The `type` and `team` fields in `agent.json` are replaced by the filesystem folder structure. An in-memory `AgentTree` is built once per session and used by all components.

#### New agent folder structure

Agents are now nested: subdirectories with a `system_prompt.md` are direct children of the parent agent. A directory with children is automatically a `leader`; without children it is an `executor`. No explicit `type` or `team` declaration needed.

```
.ai-engine/agents/
└── swarmito/
    ├── agent.json          ← optional: only { "description", "model" }
    ├── system_prompt.md
    ├── backend-leader/
    │   ├── system_prompt.md
    │   └── backend-executor/
    │       └── system_prompt.md
    └── frontend-leader/
        ├── system_prompt.md
        └── frontend-executor/
            └── system_prompt.md
```

#### New `agent.json` format

```json
{
  "description": "Implements the Go REST API backend",
  "model": "claude-opus-4"
}
```

Both fields are optional. `name`, `type`, and `team` are ignored if present (backward compatibility).

#### Backend changes

- [`internal/registry/registry.go`](../internal/registry/registry.go) — Full rewrite: replaced `AgentDefinition` + `AgentTreeNode` with `AgentNode` struct; new `LoadTree(rootName)` builds the full in-memory tree recursively from the filesystem; new `FindNode(root, name)` performs BFS by agent name; `LoadAgentTree` updated to use `LoadTree` internally (frontend `/agents` endpoint unchanged); `agent.json` is optional; `LoadAgent` and `ListAgents` removed.
- [`internal/agent/agent.go`](../internal/agent/agent.go) — `Agent.Definition` changed from `*registry.AgentDefinition` to `*registry.AgentNode`.
- [`internal/agent/runner.go`](../internal/agent/runner.go) — New 5-layer `composeSystemPrompt(engineCtx, agentRole, teamCtx, dynamicCtx, taskCtx)`; added `buildTeamContext(children)` generating `## Team` markdown with `description` (falls back to name); `teamCtx` computed once per `Run()` for leaders; `SystemLayers` log entry updated with `TeamContext` field.
- [`internal/chatlog/logger.go`](../internal/chatlog/logger.go) — Added `TeamContext string` field to `SystemLayers`; fields reordered to match layer order.
- [`internal/tools/create_chat.go`](../internal/tools/create_chat.go) — Replaced `*registry.Registry` with `*registry.AgentNode` (`currentNode`); agent resolution now searches `currentNode.Children` in-memory (no disk I/O); returns clear error if target is not a direct child.
- [`internal/tools/registry.go`](../internal/tools/registry.go) — `NewRegistry` parameter `reg *registry.Registry` replaced with `currentNode *registry.AgentNode`.
- [`internal/server/server.go`](../internal/server/server.go) — `LoadAgent` replaced with `LoadTree`; `runAgent` closure receives `*registry.AgentNode`; sub-agent tool registry passes the node as `currentNode`.
- [`internal/scaffold/templates/swarmito_agent.json`](../internal/scaffold/templates/swarmito_agent.json) — Removed `name`, `type`, `team`; now only `description` and `model`.

#### New system prompt layer order (5 layers)

| Layer | Content | Source |
|---|---|---|
| 1 | Engine Context | `engine_context.md` (static) |
| 2 | Agent Role | `system_prompt.md` (static) |
| 3 | Team | Auto-generated for leaders from `Children` + `description` |
| 4 | Dynamic Context | Workspace tree (runtime) |
| 5 | Current Task | `task_context.md` (runtime, written by leader) |

### Added

#### `projects/pomodoro-timer/` — New test workspace (v0.0.26 validation)

- New example workspace: **Pomodoro Timer** — productivity timer with focus sessions, breaks, session history, and statistics.
- Binary: `ai-engine.exe` v0.0.26 (first binary with filesystem-based agent hierarchy).
- Config: `claude-sonnet-4-6`, port `8080`, `max_tool_calls: 60`.
- Agent tree (nested structure): `swarmito` → `backend-leader` → `backend-executor` + `frontend-leader` → `frontend-executor`.
- Backend spec: Go REST API on port `8081` — session CRUD (focus/short_break/long_break), active session endpoint, aggregate stats (total focus sessions, today's sessions, current streak), settings management.
- Frontend spec: single-file `frontend/index.html` — circular SVG countdown timer with animated progress ring, session type selector, Start/Pause/Stop controls, stats panel, session history list, collapsible settings panel. Dark UI with tomato red/green/blue accent per session type.

---

## [0.0.25] — 2026-05-18

### Fixed — Frontend Review bugs (full pass)

Complete frontend hardening based on a systematic Playwright review. All 5 identified bugs resolved.

#### Bug 1 — Timestamps `NaN:NaN:NaN` in Live Log and Agent Detail Drawer (historical replay)

- [`internal/events/bus.go`](../internal/events/bus.go) — Added `Timestamp string \`json:"timestamp,omitempty"\`` field to `Event` struct; `Publish()` now auto-fills `Timestamp` with `time.Now().UTC().Format(time.RFC3339)` if empty — all future persisted events carry a timestamp.
- [`frontend/src/hooks/useSessionHistory.ts`](../frontend/src/hooks/useSessionHistory.ts) — `loadSessionEvents` now maps `timestamp → receivedAt` as fallback for historical events: `receivedAt: e.receivedAt || e.timestamp || ''`.
- [`frontend/src/types/events.ts`](../frontend/src/types/events.ts) — Added `timestamp?: string` to `EngineEvent` interface.
- [`frontend/src/components/terminal/TerminalLine.tsx`](../frontend/src/components/terminal/TerminalLine.tsx) — `formatTime(event.receivedAt ?? event.timestamp ?? '')`.
- [`frontend/src/components/drawers/AgentDetailDrawer.tsx`](../frontend/src/components/drawers/AgentDetailDrawer.tsx) — same fallback applied.

#### Bug 2 — Clear button resets agent graph, Agent Roster, and Quick Stats

- [`frontend/src/App.tsx`](../frontend/src/App.tsx) — Added `terminalClearedAt: number | null` state; `handleClearEvents` now only calls `setTerminalClearedAt(Date.now())` (no longer calls `clearEvents()`); `terminalEvents` memo filters events after the cleared timestamp; `handleNewMission` and `handleLoadSession` reset `terminalClearedAt` to `null`.
- [`frontend/src/components/layout/CockpitArea.tsx`](../frontend/src/components/layout/CockpitArea.tsx) — Added `terminalEvents` prop; passes it to `LiveTerminal` while full `events` array continues to drive the graph, roster, and stats.

#### Bug 3 — Session status always `RUNNING` in Analytics (ProjectView and SessionView)

- [`internal/server/server.go`](../internal/server/server.go) — Both the success path (`EventTypeSessionFinished`) and error path (`EventTypeError`) in the agent goroutine now use `s.bus.Publish()` instead of direct `writeCh` writes. The persistence subscriber receives these events and calls `FinishSession("done")` / `FinishSession("error")` correctly.

#### Bug 4 — Agent graph does not re-fit after panel resize

- [`frontend/src/components/graph/AgentGraph.tsx`](../frontend/src/components/graph/AgentGraph.tsx) — Added `ResizeObserver` on `containerRef` that calls `fitView({ padding: 0.4, duration: 300 })` after 50ms debounce whenever the container size changes. Decoupled from parent resize mechanism — no additional props needed.

#### Bug 5 — Ghost drawer persisting off-screen from page load

- [`frontend/src/components/drawers/AgentDetailDrawer.tsx`](../frontend/src/components/drawers/AgentDetailDrawer.tsx) — Added `if (!agent) return null` early return; replaced `transform`/`transition` toggle with `animation: 'slide-in-right 250ms ease'` on mount.

### Tested

Playwright headed browser test on `projects/standup-tracker/` (v0.0.25):

| Bug | Result |
|-----|--------|
| Bug 5 — Ghost drawer | ✅ PASS |
| Bug 1 — Timestamps (new sessions) | ✅ PASS (legacy session data has no timestamp — expected) |
| Bug 2 — Clear does not reset graph | ✅ PASS |
| Bug 4 — Graph re-fits after resize | ✅ PASS |
| Bug 3 — Analytics status (new sessions) | ✅ PASS (legacy meta.json unaffected — expected) |

---

## [0.0.24] — 2026-05-18

### Fixed — Frontend Review bugs (partial — intermediate build)

- Bugs 2, 4, 5 fixed (frontend-only changes).
- Bug 3 fixed (backend `server.go`).
- Bug 1 partially fixed (frontend fallback added, but backend `Event` struct still lacked `Timestamp` field — root cause addressed in v0.0.25).

---

## [0.0.23] — 2026-05-18

### Fixed — Backend Code Review (full pass)

Complete backend hardening based on a systematic code review. All 33 identified issues were resolved across 4 priority levels.

#### 🔴 P1 — Critical fixes

**`internal/events/bus.go` + `internal/server/server.go` — Event bus memory leak + panic on closed channel (P1 #1, #2)**
- `Subscribe` now returns a `SubscriptionID` (atomic uint64); new `Unsubscribe(id)` method removes handlers from a `map[SubscriptionID]Handler`.
- Persistence handler self-unsubscribes after `session.finished`/`error` events — eliminates zombie handlers after N sessions.
- Forwarding handler uses `select { case writeCh <- ev: default: }` (non-blocking send) + `defer Unsubscribe` on connection close — eliminates panic on closed channel.

**`internal/sandbox/shell.go` — Race condition in timeout path + blocking cmd.Wait() (P1 #3, #6)**
- Replaced shared `strings.Builder buf` with a `chan string` (buffer 256) — goroutine sends lines to channel; `buf` is built only after `close(lineCh)`, eliminating the data race.
- Added `waitDone chan struct{}` field; background goroutine calls `cmd.Wait()` exactly once and closes the channel. `Close()` and the timeout path wait on `waitDone` with 3s/1s timeouts — `Close()` never blocks indefinitely on a dead process.

**`internal/agent/runner.go` — Batch tool call protocol violations (P1 #4, #5)**
- `consecutiveErrors >= maxToolRetries` check moved outside the tool call loop — all tool results are collected before terminating, satisfying the Anthropic protocol requirement that every `tool_use` has a corresponding `tool_result`.
- `finish_work` in the middle of a batch no longer breaks immediately — remaining calls receive a `"skipped: finish_work already called"` result and the loop continues.
- Post-loop ordering: `chat.AddMessage(toolResults)` always runs first, then retry limit check, then finish check.

#### 🟠 P2 — High priority fixes

**`internal/server/server.go` — Concurrent session guard (P2 #2)**
- Per-connection `sessionMu sync.Mutex` + `sessionActive bool` flag; second `user.message` while a session is running receives an error response instead of creating a concurrent session.

**`internal/sessionstore/store.go` — File handle kept open during session (P2 #3)**
- Added `mu sync.Mutex` + `openFiles map[string]*os.File`; `StartSession` opens `events.jsonl` and stores the handle; `AppendEvent` uses the open handle under lock; `FinishSession` closes and removes the handle — eliminates open/write/close per event.

**`internal/chatlog/logger.go` — Thread-safe WriteEntry (P2 #4)**
- Added `sync.Mutex` to `Logger`; `WriteEntry` acquires the lock before writing.

**`internal/sandbox/sandbox.go` — Path traversal check robustness (P2 #6)**
- Replaced `strings.HasPrefix` with `filepath.Rel` + `strings.HasPrefix(rel, "..")` — correctly handles Windows case-insensitive paths and edge cases like `C:\Projects\foo-evil` vs `C:\Projects\foo`.

**`internal/agent/runner.go` — Nudge message limit (P2 #7)**
- Added `consecutiveNudges` counter (limit: 5); session terminates with a clear error if the LLM returns 5 consecutive responses without tool calls.

**`internal/tools/create_chat.go` — Infinite recursion guard (P2 #9)**
- Agent call stack passed via `context.Value`; `Execute` checks if the target agent is already in the stack and returns an error if so — prevents `agent-a → agent-b → agent-a` infinite loops.

**`internal/llm/provider.go` + `internal/llm/anthropic/anthropic.go` — Configurable MaxTokens (P2 #10)**
- Added `MaxOutputTokens int` to `llm.Request`; Anthropic provider uses it with fallback to 16000 — allows per-request token limit configuration.

**`internal/tools/registry.go` — Tool access validation by agent type (P2 #14)**
- Added `agentType AgentToolSet` field to `Registry`; `Get()` validates the requested tool is in the allowed set for the agent type — executors cannot call leader tools even if the LLM requests them.

**Additional P2 fixes:**
- `internal/server/server.go`: CORS policy documented; session errors logged with `[session=ID]`; indentation fixed in `runAgent` block.
- `internal/agent/runner.go`: All `//nolint:errcheck` directives removed from `logger.WriteEntry` calls.
- `internal/tools/apply_diff.go`: `Description()` and `InputSchema()` updated to document first-occurrence-only behavior.
- `internal/scaffold/scaffold.go`: `config.json` template now includes `"max_tool_calls": 50`.

#### 🟡 P3 — Robustness improvements

- **`internal/config/config.go`**: `.env` parser strips surrounding single/double quotes from values; `Load` validates `provider`, `default_model`, and `root_agent` are non-empty with clear error messages.
- **`internal/tools/search_files.go`**: Files >10MB are skipped; binary files (null byte in first 512 bytes) are skipped — prevents memory exhaustion on large workspaces.
- **`internal/tools/read_file.go`**: Replaced `os.ReadFile` + `strings.Split` with `bufio.Scanner` (1MB buffer) — applies offset/limit without loading the entire file into memory.
- **`internal/tools/list_files.go`**: `.ai-engine` filtered at any depth in recursive mode, not just at the workspace root.
- **`internal/dyncontext/workspace_tree.go`**: Added `ignoredDirs` map (`.git`, `node_modules`, `vendor`, `dist`, `build`, `__pycache__`, `.venv`, `venv`, `.ai-engine`) and `maxTreeDepth = 6` — prevents enormous workspace trees from consuming LLM tokens.
- **`internal/scaffold/scaffold.go`**: `writeNew` is now idempotent — skips existing files with a log message instead of returning an error; `ai-engine init` is safe to run multiple times.
- **`internal/server/server.go`**: `handleSessionEvents` validates the session ID as a UUID v4 before use.
- **`internal/pricing/pricing.go`**: `CalcCost` logs a warning when the model is not found in the pricing map.
- **`internal/tools/update_task_file.go`**: Returns an error if the task file doesn't exist — enforces the `create_task_file` → `update_task_file` workflow.
- **`internal/llm/provider.go`**: `ToolDefinition.InputSchema` changed from `interface{}` to `json.RawMessage` for type safety.

#### 🟢 P4 — Code quality

- **`internal/agent/runner.go`**: Removed dead code `var schema interface{}; _ = schema`.
- **`internal/sessionstore/store.go`**: `splitLines` function removed; replaced with `bytes.Split(data, []byte("\n"))`.

### Added

#### `projects/standup-tracker/` — New test workspace (v0.0.23 validation)

- New example workspace: **Daily Standup Tracker** — team standup entry management with per-date grouping and author filtering.
- Binary: `ai-engine.exe` v0.0.23.
- Config: `claude-sonnet-4-6`, port `8080`, `max_tool_calls: 60`.
- Agent tree: `swarmito` → `backend-leader` → `backend-executor` + `frontend-leader` → `frontend-executor`.
- Backend spec: Go REST API on port `8081` — create standup entry (date, author, "what I did", "what I'll do", blockers), list all entries sorted by date descending, filter by author, delete by ID. In-memory storage, standard library only.
- Frontend spec: single-file `frontend/index.html`, dark modern UI, standup form with all fields, entries grouped by date, dynamic author filter dropdown, delete button per entry, Fetch API communication with backend on port 8081.

---

## [0.0.22] — 2026-05-18

### Fixed

#### Analytics Panel — crash on open (`TypeError: Cannot read properties of undefined (reading 'toFixed')`)

- [`frontend/src/components/analytics/ProjectView.tsx`](../frontend/src/components/analytics/ProjectView.tsx) — interface `TokenData.cost_usd` → `estimated_cost_usd`; guard `?.estimated_cost_usd != null` added in StatCard "Total Cost"; usage in session cost fetch corrected.
- [`frontend/src/components/analytics/SessionView.tsx`](../frontend/src/components/analytics/SessionView.tsx) — interface `TokenData.cost_usd` → `estimated_cost_usd`; render of total session cost corrected with guard `?.estimated_cost_usd != null`.

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
- Three helper functions added (unexported): `toMessageLog()`, `toToolLog()`, `toToolCallEntries()`.
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

#### `frontend/` — Frontend v2 "Mission Control"
Complete rewrite of the frontend. The previous single-panel layout was replaced by a full "Mission Control" cockpit interface.

**Layout — 3 columns:**
- **Sidebar (240px, collapsible)** — session history, relative timestamps, status badges, replay mode, "New Mission" button. Collapse/expand via `Ctrl+B`.
- **Cockpit Area (flex: 1)** — vertically split between Agent Graph (top) and Live Terminal (bottom). Split ratio is drag-resizable via a `ResizeHandle` component.
- **Mission Panel (360px)** — connection status + latency, prompt editor, launch button, task progress, agent roster, quick stats.

**Agent Graph — redesigned:**
- `LeaderNode.tsx` — glassmorphism card, hexagonal "L" badge, pulsing glow when running, `--purple` accent border.
- `ExecutorNode.tsx` — pill shape, circular "E" badge, status-colored border.
- `AnimatedEdge.tsx` — SVG `stroke-dashoffset` particle animation while agent is active; static line when idle.
- `AgentGraph.tsx` — dot grid background, `<MiniMap>`, `<Controls>`, empty state message, click-to-open drawer.

**Live Terminal — replaces EventFeed:**
- `LiveTerminal.tsx` — real terminal aesthetic, `JetBrains Mono` font, scanline CRT overlay, auto-scroll, Clear button, Export `.jsonl` download.
- `TerminalLine.tsx` — color-coded lines by event type: `[SESSION]` cyan, `[AGENT]` violet, `[TOOL]` yellow, `[RESULT]` green, `[ERROR]` red.

**Mission Panel components:**
- `PromptEditor.tsx` — auto-resize textarea, `Ctrl+Enter` to submit, focus glow.
- `LaunchButton.tsx` — countdown animation 3→2→1→🚀 before firing.
- `TaskProgress.tsx` — parses Markdown checklist from `tasks.updated` events, renders progress bar.
- `AgentRoster.tsx` — live list of all agents with status dot, type badge (L/E), tool call count.
- `QuickStats.tsx` — total tool calls, unique agents, live session duration counter.

**Agent Detail Drawer:**
- `AgentDetailDrawer.tsx` — slides in from the right when a graph node is clicked. Shows per-agent event timeline. Close via `×` button or `Escape`.

**New hooks:**
- `useSessionHistory.ts` — `localStorage` persistence, `startSession`, `saveEvent`, `finishSession`, `loadSession`, max 20 sessions.
- `useResizable.ts` — drag-to-resize vertical split ratio via `mousemove`/`mouseup` on `document`.
- `useKeyboardShortcuts.ts` — global `keydown` handler, `Ctrl+B` (sidebar), `Escape` (drawer).
- `useWebSocket.ts` — extended with auto-reconnect (max 10 attempts, 3s interval), `latency: number | null`, `sessionStartTime: Date | null`, `connectionStatus` string.

**Design system:**
- `index.css` — full CSS variable palette (`--bg-base: #080c14`, `--accent: #58a6ff`, `--purple: #bc8cff`, etc.), global reset, scrollbar styling, keyframe animations.
- Typography: `Inter` (UI) + `JetBrains Mono` (terminal), loaded from Google Fonts in `index.html`.

#### `projects/expense-tracker/` — New example workspace
- New example workspace: **Personal Expense Tracker** — expense management with category filtering and summary.
- Binary: `ai-engine.exe` v0.0.4 (with Frontend v2 embedded).
- Config: `claude-sonnet-4-6`, port `8080`, `max_tool_calls: 60`.
- Agent tree: `swarmito` → `backend-leader` → `backend-executor` + `frontend-leader` → `frontend-executor`.
- Backend spec: Go REST API on port `8083`, endpoints `POST /expenses`, `GET /expenses`, `DELETE /expenses/{id}`, `GET /summary`, `GET /categories`. Standard library only, no external deps.
- Frontend spec: single-file `frontend/index.html`, dark theme, expense form with category selector, filterable list, CSS-only horizontal bar chart for spending by category.

---

## [0.0.3] — 2026-05-15

### Fixed

#### `frontend/src/components/AgentGraphNode.tsx` — Missing `Handle` components
- Added `<Handle type="target" position={Position.Top}>` and `<Handle type="source" position={Position.Bottom}>` to the custom node component.
- Handles are styled with `width: 8`, `height: 8`, `border: none`, and dynamic color matching the node status.

---

## [0.0.2] — 2026-05-15

### Fixed

#### `frontend/src/App.tsx` + `frontend/src/components/AgentGraph.tsx` — `ReactFlowProvider` placement
- Moved `ReactFlowProvider` from `AgentGraph.tsx` (same component as `ReactFlow`) to `App.tsx` (parent component).
- React Flow v12 requires `ReactFlowProvider` to be in a **parent component** in the tree — having it wrap `ReactFlow` in the same component creates an isolated context that prevents edges from rendering.

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
- Reads and writes current version from/to `version.txt` (semver format `MAJOR.MINOR.PATCH`).
- Compiles binary to `bin\{version}\ai-engine.exe`.
- Always copies the latest build to `bin\latest\ai-engine.exe`.

#### `version.txt`
- New file at repository root tracking the current semver version.
- Initial value: `0.0.0`.

### Fixed

#### `frontend/src/components/AgentGraph.tsx` — Missing `ReactFlowProvider`
- Added `ReactFlowProvider` wrapper around `<ReactFlow>` to fix missing edges in the agent graph view.