# Agents

## Status

🟢 Defined — agent model finalized. Ready for implementation.

---

## Agent Types

### Leader
A **leader** is a non-leaf node in the agent tree. It has a team of agents directly below it. Its responsibility is to decompose the received objective into tasks, delegate them to team members, review results, and report back.

**Swarmito** is the root leader — the only agent that communicates directly with the user.

### Executor
An **executor** is a leaf node in the agent tree. It has no team. Its responsibility is to perform concrete tasks using tools (terminal commands, file operations) and report the result back to its leader.

---

## System Prompt Structure

Every agent's system prompt is composed of **3 layers** assembled by the engine at runtime:

```
Layer 1 — ENGINE CONTEXT (optional)
  Content of .ai-engine/engine_context.md — injected as a prefix into every agent.
  Describes the engine, tools, and conventions. Separated from Layer 2 by "---".

Layer 2 — AGENT ROLE (always present)
  Content of .ai-engine/agents/{name}/system_prompt.md — the agent's identity and skills.
  Structure:
    [Leader-only] ## Team — list of sub-agents with brief role descriptions
    # Agent Description — who this agent is and what it does
    # Skills — following Anthropic's skill definition pattern
    # Task — what this agent is expected to accomplish (generic; specific task comes from Layer 3)

Layer 3 — CURRENT TASK (optional)
  Content of .ai-engine/chats/{session-id}/{agent-name}/task_context.md
  Written by the leader via the set_task_context tool before calling create_chat.
  Injected as a "# Current Task" section appended to the system prompt.
  Archived to .ai-engine/history/{session-id}/{agent-name}/task_context.md after the agent finishes.
```

---

## Tools

### Leader Tools
| Tool | Description |
|---|---|
| `create_chat` | Opens a new chat session with an agent from the team. Loads the agent's `task_context.md` (Layer 3) before running. |
| `set_task_context` | Writes a task context file for a sub-agent at `.ai-engine/chats/{session-id}/{agent}/task_context.md`. This content is injected as Layer 3 of the sub-agent's system prompt when `create_chat` is called. |
| `create_task_file` | Creates a Markdown checklist task file for a sub-agent at `.ai-engine/chats/{session-id}/{agent}/tasks.md`. |
| `update_task_file` | Overwrites the task file for a sub-agent (mark tasks done, add new tasks). |
| `list_files` | Lists files and directories at a given path. Available to leaders for workspace inspection. |
| `read_file` | Reads the content of a file. Available to leaders for workspace inspection. |
| `finish_work` | Signals the end of this agent's execution. Must be called exactly once, as the last action. Contains the response message. |

### Executor Tools
| Tool | Description |
|---|---|
| `run_terminal_command` | Executes a shell command inside the workspace sandbox. Agents send commands as if at the filesystem root; the engine resolves paths to the workspace. |
| `list_files` | Lists files and directories at a given path. Supports recursive listing. |
| `read_file` | Reads the content of a file, returning it with line numbers. Supports offset and limit for partial reads. |
| `write_file` | Creates a new file or fully overwrites an existing one. Use for new files or intentional full rewrites. |
| `apply_diff` | Applies precise, targeted edits to an existing file using one or more search/replace blocks. Preferred over `write_file` for modifications to existing files — avoids full overwrite and handles large files efficiently. |
| `search_files` | Performs a regex search across files in a directory, returning matches with surrounding context. |
| `delete_file` | Deletes a file in the workspace. |
| `finish_work` | Signals the end of this agent's execution. Must be called exactly once, as the last action. Contains the response message. |

> **Note:** All file paths and terminal commands are relative to the workspace root. The engine resolves absolute paths transparently — agents never see or use the real workspace path.
>
> **Guideline:** Prefer `apply_diff` over `write_file` when modifying existing files. Use `write_file` only for creating new files or when a complete rewrite is intentionally required.

---

## Agent Lifecycle

### Leader Lifecycle

```
┌──────────────────────────────────────────────────────────┐
│                     Leader Lifecycle                     │
│                                                          │
│  1. Receive objective (initial prompt: "start")          │
│  2. Analyze team (from system prompt)                    │
│  3. Create task file                                     │
│  4. For each task:                                       │
│     a. Open chat with the appropriate agent              │
│     b. Send initial prompt                               │
│     c. Wait for agent's finish_work response             │
│     d. Review the result                                 │
│     e. Update task file                                  │
│  5. When all tasks are done:                             │
│     a. General review vs. original objective             │
│     b. Call finish_work with final response              │
└──────────────────────────────────────────────────────────┘
```

### Executor Lifecycle

```
┌──────────────────────────────────────────────────────────┐
│                    Executor Lifecycle                    │
│                                                          │
│  1. Receive task (from leader's chat)                    │
│  2. Execute task using tools (terminal, files)           │
│  3. Call finish_work with result                         │
└──────────────────────────────────────────────────────────┘
```

---

## Execution Flow (Full Tree)

```
User
 └─▶ Swarmito (root leader)
       ├─▶ [create_task_file]
       ├─▶ [create_chat → Leader A]
       │     ├─▶ [create_task_file]
       │     ├─▶ [create_chat → Executor D] → [finish_work]
       │     ├─▶ [create_chat → Executor E] → [finish_work]
       │     ├─▶ [create_chat → Executor F] → [finish_work]
       │     └─▶ [finish_work]
       ├─▶ [create_chat → Leader B]
       │     └─▶ ... (same pattern)
       ├─▶ [create_chat → Leader C]
       │     └─▶ ... (same pattern)
       └─▶ [finish_work] ──▶ User receives final response
```

---

## Decisions

| Question | Decision |
|---|---|
| Sequential or parallel chats? | **Sequential in V1.** Parallel is a future enhancement. |
| What happens when an executor fails? | **V1: fail-fast** — error propagates to user immediately. Future: configurable retry + leader escalation. See [`docs/error-handling.md`](./error-handling.md). |
| Maximum tree depth? | **No hard limit in V1.** Practically bounded by LLM context and recursion. |
| Task file format? | **Markdown checklist.** See [`docs/workspace-structure.md`](./workspace-structure.md). |
| Multi-turn conversation with user? | **V1: multi-turn supported.** The user can send follow-up messages to Swarmito via WebSocket after a session finishes. Each message starts a new session. |
