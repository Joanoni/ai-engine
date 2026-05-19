package tokenstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/swarmit/ai-engine/internal/pricing"
)

// SessionTokens holds token usage for a single session.
type SessionTokens struct {
	SessionID        string  `json:"session_id"`
	InputTokens      int     `json:"input_tokens"`
	OutputTokens     int     `json:"output_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
	UpdatedAt        string  `json:"updated_at"`
}

// ProjectTokens holds cumulative token usage for the entire workspace.
type ProjectTokens struct {
	InputTokens      int     `json:"input_tokens"`
	OutputTokens     int     `json:"output_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
	SessionCount     int     `json:"session_count"`
	LastUpdatedAt    string  `json:"last_updated_at"`
}

// Store persists token usage to .ai-engine/tokens.json (project) and
// .ai-engine/sessions/{id}/tokens.json (per session).
type Store struct {
	workspacePath string
	mu            sync.Mutex
}

// New creates a Store for the given workspace.
func New(workspacePath string) *Store {
	return &Store{workspacePath: workspacePath}
}

// StartSession increments the SessionCount in the project tokens file.
// Errors are non-fatal — callers should log but not terminate on failure.
func (s *Store) StartSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC().Format(time.RFC3339)
	path := s.projectTokensPath()
	pt := s.readProjectTokens()
	pt.SessionCount++
	pt.LastUpdatedAt = now
	return writeJSON(path, pt)
}

// AddUsage records input+output tokens for a session and model.
// Errors are non-fatal — callers should log but not terminate on failure.
func (s *Store) AddUsage(sessionID, model string, inputTokens, outputTokens int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prices, _ := pricing.Load(s.workspacePath) // best-effort; empty map on error
	cost := pricing.CalcCost(prices, model, inputTokens, outputTokens)
	now := time.Now().UTC().Format(time.RFC3339)

	// Update session tokens.
	if err := s.updateSession(sessionID, inputTokens, outputTokens, cost, now); err != nil {
		return err
	}

	// Update project tokens.
	return s.updateProject(inputTokens, outputTokens, cost, now)
}

func (s *Store) sessionTokensPath(sessionID string) string {
	return filepath.Join(s.workspacePath, ".ai-engine", "sessions", sessionID, "tokens.json")
}

func (s *Store) projectTokensPath() string {
	return filepath.Join(s.workspacePath, ".ai-engine", "tokens.json")
}

func (s *Store) updateSession(sessionID string, input, output int, cost float64, now string) error {
	path := s.sessionTokensPath(sessionID)
	st := s.readSessionTokens(sessionID)
	st.SessionID = sessionID
	st.InputTokens += input
	st.OutputTokens += output
	st.TotalTokens = st.InputTokens + st.OutputTokens
	st.EstimatedCostUSD += cost
	st.UpdatedAt = now
	return writeJSON(path, st)
}

func (s *Store) updateProject(input, output int, cost float64, now string) error {
	path := s.projectTokensPath()
	pt := s.readProjectTokens()
	pt.InputTokens += input
	pt.OutputTokens += output
	pt.TotalTokens = pt.InputTokens + pt.OutputTokens
	pt.EstimatedCostUSD += cost
	pt.LastUpdatedAt = now
	return writeJSON(path, pt)
}

// ReadSession returns the token usage for a session. Returns zero-value if not found.
func (s *Store) ReadSession(sessionID string) SessionTokens {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readSessionTokens(sessionID)
}

// ReadProject returns the cumulative project token usage.
func (s *Store) ReadProject() ProjectTokens {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readProjectTokens()
}

func (s *Store) readSessionTokens(sessionID string) SessionTokens {
	var st SessionTokens
	data, err := os.ReadFile(s.sessionTokensPath(sessionID))
	if err != nil {
		return st
	}
	_ = json.Unmarshal(data, &st)
	return st
}

func (s *Store) readProjectTokens() ProjectTokens {
	var pt ProjectTokens
	data, err := os.ReadFile(s.projectTokensPath())
	if err != nil {
		return pt
	}
	_ = json.Unmarshal(data, &pt)
	return pt
}

func writeJSON(path string, v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
