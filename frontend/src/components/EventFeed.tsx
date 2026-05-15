import { useEffect, useRef } from 'react';
import type { EngineEvent, EventType } from '../types/events';

interface EventFeedProps {
  events: EngineEvent[];
}

function getEventColor(type: EventType): string {
  switch (type) {
    case 'session.started':
    case 'session.finished':
      return 'event-blue';
    case 'agent.started':
    case 'agent.finished':
      return 'event-purple';
    case 'tool.called':
    case 'tool.result':
      return 'event-grey';
    case 'tasks.updated':
      return 'event-teal';
    case 'error':
      return 'event-red';
    default:
      return '';
  }
}

function formatPayload(payload?: Record<string, unknown>): string {
  if (!payload) return '';
  return Object.entries(payload)
    .map(([k, v]) => `${k}: ${String(v)}`)
    .join(' | ');
}

export function EventFeed({ events }: EventFeedProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [events]);

  return (
    <div className="event-feed">
      {events.map((event, index) => (
        <div key={index} className={`event-row ${getEventColor(event.type)}`}>
          <span className="event-time">
            {new Date(event.receivedAt).toLocaleTimeString()}
          </span>
          <span className="event-type">{event.type}</span>
          {event.agent_name && (
            <span className="event-agent">{event.agent_name}</span>
          )}
          {event.payload && (
            <span className="event-payload">{formatPayload(event.payload)}</span>
          )}
        </div>
      ))}
      <div ref={bottomRef} />
    </div>
  );
}
