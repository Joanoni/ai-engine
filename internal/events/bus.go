package events

import "sync"

// EventType identifies the kind of event.
type EventType string

const (
	EventTypeAgentStarted   EventType = "agent.started"
	EventTypeAgentFinished  EventType = "agent.finished"
	EventTypeToolCalled     EventType = "tool.called"
	EventTypeToolResult     EventType = "tool.result"
	EventTypeSessionStarted EventType = "session.started"
	EventTypeSessionFinished EventType = "session.finished"
	EventTypeError          EventType = "error"
	EventTypeTasksUpdated   EventType = "tasks.updated"
)

// Event is a single event published on the Bus.
type Event struct {
	Type      EventType   `json:"type"`
	SessionID string      `json:"session_id,omitempty"`
	AgentName string      `json:"agent_name,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
}

// Handler is a function that receives events.
type Handler func(event Event)

// Bus broadcasts events to all registered handlers. Thread-safe.
type Bus struct {
	mu       sync.RWMutex
	handlers []Handler
}

// NewBus creates a new event Bus.
func NewBus() *Bus {
	return &Bus{}
}

// Subscribe registers a handler to receive all future events.
func (b *Bus) Subscribe(h Handler) {
	b.mu.Lock()
	b.handlers = append(b.handlers, h)
	b.mu.Unlock()
}

// Publish sends an event to all registered handlers.
func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	handlers := make([]Handler, len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.RUnlock()

	for _, h := range handlers {
		h(e)
	}
}
