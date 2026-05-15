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

`run_terminal_command` distinguishes between three types of outcomes:

| Outcome | Behavior |
|---|---|
| Command succeeds (exit code 0) | Output returned to agent — no error. |
| Command exits with non-zero status (e.g., compilation error, test failure) | Output returned to agent with `Exit code N:` prefix — **not an error**. Agent decides how to handle. |
| Command times out (exceeds `timeout_seconds`, default 30s) | Process killed; partial output returned with `[TIMEOUT after Ns]` prefix — **not an error**. Agent can continue. |
| Command could not be started (e.g., executable not found in PATH) | Engine error — counts toward `max_tool_retries`; session terminates if limit exceeded. |

This allows agents to iteratively fix compilation errors, test failures, and other recoverable command failures without terminating the session.

---

## Future — Leader Escalation

The current retry model feeds errors back to the **same agent**. A future enhancement would escalate exhausted errors to the **leader**, which could then reassign the task, retry with a different approach, or escalate further up the tree. This requires no engine changes — only a change in how `create_chat` reports sub-agent failures.

---

## Design Principles

1. **Transparency:** Tool errors are always returned to the agent as tool results — no information is lost.
2. **Agent-driven recovery:** The agent's LLM decides how to handle errors (retry, change approach, give up). The engine enforces limits, not strategies.
3. **Configurable limits:** `max_tool_retries` and `max_tool_calls` are set in `config.json` — not hardcoded.
4. **Runaway loop prevention:** `max_tool_calls` caps total tool calls per agent, preventing infinite loops regardless of error state.
