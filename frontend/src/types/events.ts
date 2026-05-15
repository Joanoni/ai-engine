export type EventType =
  | 'session.started'
  | 'session.finished'
  | 'agent.started'
  | 'agent.finished'
  | 'tool.called'
  | 'tool.result'
  | 'tasks.updated'
  | 'error';

export interface EngineEvent {
  type: EventType;
  session_id?: string;
  agent_name?: string;
  payload?: Record<string, unknown>;
  receivedAt: string; // added client-side on receipt
}
