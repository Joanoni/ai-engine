# Error Handling

## Status

🟢 Implemented — retry + tool call limit are live. Future leader escalation documented for reference.

---

## Current Behaviour — Retry with Fail-Fast Escalation

Tool errors are fed back to the agent as tool results (with `is_error: true`). The agent can attempt a different approach. If the same tool fails **`max_tool_retries` consecutive times** (default: 3, configurable in `config.json`), the session terminates immediately.

Additionally, if an agent exceeds **`max_tool_calls` total tool calls** (default: 50, configurable in `config.json`), the session terminates to prevent runaway loops.

```
Tool fails
    └─▶ Error returned to agent as tool result (is_error: true)
    └─▶ Agent retries (up to max_tool_retries consecutive failures)
    └─▶ If limit exceeded → session terminates → error event sent to user

Total tool calls > max_tool_calls
    └─▶ Session terminates immediately → error event sent to user
```

LLM API errors (e.g., network failure, `max_tokens` truncation) terminate the session immediately — no retry.

---

## Terminal Command Errors

`run_terminal_command` uses a **persistent shell per agent** (see `internal/sandbox/Shell`). The shell process stays alive for the entire agent execution, so working directory and environment variables are preserved between calls.

| Outcome | Behavior |
|---|---|
| Command succeeds | Output returned to agent — no error. |
| Command exits with non-zero status (e.g., compilation error) | Output returned to agent — **not an error**. Agent decides how to handle. |
| Command times out (exceeds `timeout_seconds`, default 30s) | Shell process killed (and via Windows Job Object, all child processes); partial output returned with `[TIMEOUT after Ns]` prefix — **not an error**. A new shell is started for the next command. |
| Shell process could not be started | Engine error — counts toward `max_tool_retries`; session terminates if limit exceeded. |

**Windows Job Object:** On Windows, the shell process is associated with a Job Object configured with `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`. When the shell is killed (timeout or agent finish), the OS automatically terminates all child processes in the job tree — including background processes started with `&`. This prevents orphan processes from holding ports open after a timeout.

**Working directory persistence:** Because all commands run in the same shell process, `cd backend` in one call persists for subsequent calls. Agents no longer need to repeat `cd` prefixes on every command.

---

## Future — Leader Escalation

The current retry model feeds errors back to the **same agent**. A future enhancement would escalate exhausted errors to the **leader**, which could then reassign the task, retry with a different approach, or escalate further up the tree. This requires no engine changes — only a change in how `create_chat` reports sub-agent failures.

---

## Design Principles

1. **Transparency:** Tool errors are always returned to the agent as tool results — no information is lost.
2. **Agent-driven recovery:** The agent's LLM decides how to handle errors (retry, change approach, give up). The engine enforces limits, not strategies.
3. **Configurable limits:** `max_tool_retries` and `max_tool_calls` are set in `config.json` — not hardcoded.
4. **Runaway loop prevention:** `max_tool_calls` caps total tool calls per agent, preventing infinite loops regardless of error state.
