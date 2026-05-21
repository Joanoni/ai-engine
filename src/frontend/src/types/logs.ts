export interface SystemLayers {
  engine_context?: string;
  dynamic_context?: string;
  agent_role?: string;
  task_context?: string;
}

export interface ContentLog {
  type: string;
  text?: string;
  id?: string;
  name?: string;
  input?: unknown;
  tool_use_id?: string;
  content?: string;
  is_error?: boolean;
}

export interface MessageLog {
  role: string;
  content: ContentLog[];
}

export interface ToolCallEntry {
  id: string;
  name: string;
  input: unknown;
}

export interface ToolLog {
  name: string;
  description: string;
  input_schema: unknown;
}

export interface LogEntry {
  ts: string;
  turn: number;
  role: 'agent_init' | 'user' | 'llm_request' | 'llm_response' | 'tool_result' | 'warning' | 'error' | 'finish';

  // agent_init
  agent_name?: string;
  agent_type?: string;
  session_id?: string;
  model?: string;

  // user
  content?: string;

  // llm_request
  system_prompt?: string;
  system_layers?: SystemLayers;
  messages?: MessageLog[];
  tools?: ToolLog[];
  message_count?: number;
  total_tool_calls_so_far?: number;
  consecutive_errors?: number;

  // llm_response
  text?: string;
  tool_calls?: ToolCallEntry[];
  stop_reason?: string;
  input_tokens?: number;
  output_tokens?: number;

  // tool_result
  tool_use_id?: string;
  tool?: string;
  success?: boolean;
  output?: string;
  duration_ms?: number;

  // error
  message?: string;

  // finish
  result?: string;
}

export interface SessionLogs {
  agents: Record<string, LogEntry[]>;
}
