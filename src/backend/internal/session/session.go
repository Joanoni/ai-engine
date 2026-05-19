package session

import (
	"sync"

	"github.com/google/uuid"
)

// Session represents an active user interaction session.
type Session struct {
	ID string
}

// New creates a new Session with a unique UUID v4 ID.
func New() *Session {
	return &Session{ID: uuid.New().String()}
}

// Manager tracks active sessions. Thread-safe.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewManager creates a new session Manager.
func NewManager() *Manager {
	return &Manager{sessions: make(map[string]*Session)}
}

// Create creates a new session, registers it, and returns it.
func (m *Manager) Create() *Session {
	s := New()
	m.mu.Lock()
	m.sessions[s.ID] = s
	m.mu.Unlock()
	return s
}

// Get returns the session with the given ID, or nil if not found.
func (m *Manager) Get(id string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

// Delete removes a session from the manager.
func (m *Manager) Delete(id string) {
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()
}
