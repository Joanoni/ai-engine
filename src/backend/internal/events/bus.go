package events

import (
	"sync"
	"sync/atomic"
	"time"
)

// EventType identifies the kind of event.
type EventType string

const (
	EventTypeAgentStarted    EventType = "agent.started"
	EventTypeAgentFinished   EventType = "agent.finished"
	EventTypeToolCalled      EventType = "tool.called"
	EventTypeToolResult      EventType = "tool.result"
	EventTypeSessionStarted  EventType = "session.started"
	EventTypeSessionFinished EventType = "session.finished"
	EventTypeError           EventType = "error"
	EventTypeTasksUpdated    EventType = "tasks.updated"
)

// Event is a single event published on the Bus.
type Event struct {
	Type      EventType   `json:"type"`
	SessionID string      `json:"session_id,omitempty"`
	AgentName string      `json:"agent_name,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp string      `json:"timestamp,omitempty"`
}

// Handler is a function that receives events.
type Handler func(event Event)

// SubscriptionID is a unique identifier for a subscription.
type SubscriptionID uint64

// nextID is an atomic counter used to generate unique subscription IDs.
var nextID uint64

// Bus broadcasts events to all registered handlers. Thread-safe.
type Bus struct {
	mu       sync.RWMutex
	handlers map[SubscriptionID]Handler
}

// NewBus creates a new event Bus.
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[SubscriptionID]Handler),
	}
}

// Subscribe registers a handler to receive all future events.
// It returns a SubscriptionID that can be used to unsubscribe later.
func (b *Bus) Subscribe(h Handler) SubscriptionID {
	id := SubscriptionID(atomic.AddUint64(&nextID, 1))
	b.mu.Lock()
	b.handlers[id] = h
	b.mu.Unlock()
	return id
}

// Unsubscribe removes the handler associated with the given SubscriptionID.
func (b *Bus) Unsubscribe(id SubscriptionID) {
	b.mu.Lock()
	delete(b.handlers, id)
	b.mu.Unlock()
}

// Publish sends an event to all registered handlers.
func (b *Bus) Publish(e Event) {
	if e.Timestamp == "" {
		e.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	b.mu.RLock()
	handlers := make([]Handler, 0, len(b.handlers))
	for _, h := range b.handlers {
		handlers = append(handlers, h)
	}
	b.mu.RUnlock()

	for _, h := range handlers {
		h(e)
	}
}
