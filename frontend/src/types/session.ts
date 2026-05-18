import type { EngineEvent } from './events';

export interface SessionRecord {
  id: string;
  prompt: string;
  startedAt: string;   // ISO timestamp
  finishedAt?: string;
  status: 'running' | 'done' | 'error';
  events: EngineEvent[];
}
