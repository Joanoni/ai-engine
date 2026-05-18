# AI Engine

> A modular engine for running AI agents — built in Go.

## Overview

**AI Engine** is a platform designed to orchestrate and execute a hierarchical tree of AI agents. A root agent (**Swarmito**) serves as the interface with the user, decomposing objectives and delegating work down a tree of specialized agents. Leaders coordinate teams; executors perform concrete tasks using tools.

The engine exposes a **WebSocket API** (port `8080`) so a frontend can observe the full execution flow in real time — and send user prompts bidirectionally through the same connection.

All engine configuration lives inside a `.ai-engine/` folder in the user's workspace — hot-reloaded on every interaction, without polluting the user's project.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Runtime | Go |
| LLM Provider (V1) | Anthropic (Claude) — provider-agnostic interface for future expansion |
| Frontend Interface | WebSocket (bidirectional, port `8080`) |
| Config Format | JSON |
| Agent Prompts | Markdown |

---

## Running

### Build (Windows)

```cmd
build.cmd [patch|minor|major]
```

This script:
1. Reads the current version from [`version.txt`](./version.txt) and increments it (semver).
2. Builds the React frontend (`npm run build` inside `frontend/`).
3. Compiles the Go binary with the frontend embedded (`embed.go`).
4. Outputs the versioned binary to `bin\{version}\ai-engine.exe`.
5. Copies the binary to `bin\latest\ai-engine.exe`.

The bump type argument controls which semver component is incremented:

| Command | Example |
|---|---|
| `build.cmd` or `build.cmd patch` | `0.0.0 → 0.0.1` |
| `build.cmd minor` | `0.0.1 → 0.1.0` |
| `build.cmd major` | `0.1.0 → 1.0.0` |

### Configuration

Place a `.env` file at `.ai-engine/.env` in your workspace with:

```
ANTHROPIC_API_KEY=your-key-here
```

The engine also accepts `ANTHROPIC_API_KEY` as a system environment variable. Run `ai-engine init` to scaffold the workspace skeleton (including a blank `.ai-engine/.env`) automatically.

---

## Documentation

Detailed documentation is organized by topic inside the [`docs/`](./docs/) folder:

| Document | Description |
|---|---|
| [`docs/architecture.md`](./docs/architecture.md) | System architecture, components, and design decisions |
| [`docs/agents.md`](./docs/agents.md) | Agent model: types, lifecycle, tools, 3-layer system prompt, and execution flow |
| [`docs/workspace-structure.md`](./docs/workspace-structure.md) | `.ai-engine/` folder structure, config and agent file specs |
| [`docs/events.md`](./docs/events.md) | WebSocket event system for frontend observability |
| [`docs/error-handling.md`](./docs/error-handling.md) | Error handling: retry limits, tool call cap, terminal command behaviour |
| [`docs/project-structure.md`](./docs/project-structure.md) | Go project layout, package responsibilities, interfaces, env vars |
| [`docs/frontend.md`](./docs/frontend.md) | Frontend spec: React + TypeScript SPA, components, WebSocket protocol |
| [`docs/roadmap.md`](./docs/roadmap.md) | Full development roadmap — completed phases, current status, remaining work |
| [`docs/changelog.md`](./docs/changelog.md) | Release history — changes per version |
| [`docs/analytics.md`](./docs/analytics.md) | Analytics BI Panel — architecture, components, endpoints, types, bug history |

---

## Status

🟢 **Phase 4 (Partial) Complete** — Engine operational with logging, retry, 3-layer system prompts, and graceful shutdown. Remaining: session persistence, leader escalation.
