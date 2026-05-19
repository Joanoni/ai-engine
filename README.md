# AI Engine

AI Engine is a Go binary that orchestrates a hierarchical tree of AI agents. A root agent (**Swarmito**) receives a user prompt via WebSocket, decomposes the objective, and delegates work down a tree of specialized sub-agents. Leaders coordinate teams; executors perform concrete tasks using tools. The engine serves an embedded React frontend and exposes a WebSocket API for real-time observability. All configuration lives in a `.ai-engine/` folder inside the user's project directory (the workspace), which is hot-reloaded on every session start.

## Tech Stack

| Layer | Technology |
|---|---|
| Runtime | Go (`github.com/swarmit/ai-engine`) |
| LLM Provider | Anthropic (Claude) — via `LLMProvider` interface |
| Transport | WebSocket (bidirectional, default port `8080`) |
| Frontend | React 19 + TypeScript, Vite |
| Agent Prompts | Markdown |
| Config | JSON + `.env` |

## Build

```cmd
build.cmd [patch|minor|major]
```

Reads the current version from [`version.txt`](./version.txt), increments it (semver), builds the React frontend (`npm run build` inside `src/frontend/`), compiles the Go binary with the frontend embedded via [`src/backend/embed.go`](./src/backend/embed.go), and outputs:

- `bin\{version}\ai-engine.exe` — versioned binary
- `bin\latest\ai-engine.exe` — always the latest build

| Command | Example |
|---|---|
| `build.cmd` or `build.cmd patch` | `0.0.25 → 0.0.26` |
| `build.cmd minor` | `0.0.26 → 0.1.0` |
| `build.cmd major` | `0.1.0 → 1.0.0` |

## Run

Place the binary in your project directory (the workspace) and run:

```cmd
ai-engine init    # scaffold .ai-engine/ workspace structure
ai-engine         # start the server (no arguments)
```

Then open `http://localhost:{port}` in your browser.

## Configuration

Set your API key in `.ai-engine/.env`:

```
ANTHROPIC_API_KEY=your-key-here
```

The engine also reads `ANTHROPIC_API_KEY` from the system environment. All other configuration is in `.ai-engine/config.json`.

## Documentation

| Document | Description |
|---|---|
| [`docs/getting-started.md`](./docs/getting-started.md) | Step-by-step setup guide |
| [`docs/architecture.md`](./docs/architecture.md) | System architecture, Go packages, design decisions |
| [`docs/agents.md`](./docs/agents.md) | Agent model: types, lifecycle, tools, 5-layer system prompt |
| [`docs/workspace.md`](./docs/workspace.md) | `.ai-engine/` folder structure and configuration reference |
| [`docs/api.md`](./docs/api.md) | WebSocket and HTTP API reference |
| [`docs/frontend.md`](./docs/frontend.md) | Frontend components, hooks, design system |
| [`docs/chat-log-format.md`](./docs/chat-log-format.md) | JSONL chat log format reference |
| [`changelog/releases.md`](./changelog/releases.md) | Full version history |
