# Changelog

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
