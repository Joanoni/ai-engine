package sessionstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// SessionMeta holds the metadata for a persisted session.
type SessionMeta struct {
	ID         string `json:"id"`
	Prompt     string `json:"prompt"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt,omitempty"`
	Status     string `json:"status"` // "running" | "done" | "error"
}

// Store persists session metadata and events to .ai-engine/sessions/.
type Store struct {
	baseDir string // absolute path to .ai-engine/sessions/
}

// New creates a Store rooted at workspacePath/.ai-engine/sessions/.
func New(workspacePath string) *Store {
	return &Store{
		baseDir: filepath.Join(workspacePath, ".ai-engine", "sessions"),
	}
}

func (s *Store) sessionDir(id string) string {
	return filepath.Join(s.baseDir, id)
}

func (s *Store) metaPath(id string) string {
	return filepath.Join(s.sessionDir(id), "meta.json")
}

func (s *Store) eventsPath(id string) string {
	return filepath.Join(s.sessionDir(id), "events.jsonl")
}

// StartSession creates the session directory and writes initial meta.json.
func (s *Store) StartSession(id, prompt string) error {
	dir := s.sessionDir(id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	meta := SessionMeta{
		ID:        id,
		Prompt:    prompt,
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		Status:    "running",
	}
	return s.writeMeta(id, meta)
}

// AppendEvent appends a JSON-encoded event line to events.jsonl.
func (s *Store) AppendEvent(id string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.eventsPath(id), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(data, '\n'))
	return err
}

// FinishSession updates meta.json with final status and finishedAt.
func (s *Store) FinishSession(id, status string) error {
	meta, err := s.ReadMeta(id)
	if err != nil {
		return err
	}
	meta.Status = status
	meta.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	return s.writeMeta(id, *meta)
}

// ReadMeta reads and returns the meta.json for a session.
func (s *Store) ReadMeta(id string) (*SessionMeta, error) {
	data, err := os.ReadFile(s.metaPath(id))
	if err != nil {
		return nil, err
	}
	var meta SessionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// ListSessions returns all session metas sorted by startedAt descending.
func (s *Store) ListSessions() ([]SessionMeta, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SessionMeta{}, nil
		}
		return nil, err
	}

	var sessions []SessionMeta
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		meta, err := s.ReadMeta(entry.Name())
		if err != nil {
			continue // skip corrupted sessions
		}
		sessions = append(sessions, *meta)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartedAt > sessions[j].StartedAt
	})

	return sessions, nil
}

// ReadEvents reads all events from events.jsonl and returns them as raw JSON lines.
func (s *Store) ReadEvents(id string) ([]json.RawMessage, error) {
	data, err := os.ReadFile(s.eventsPath(id))
	if err != nil {
		if os.IsNotExist(err) {
			return []json.RawMessage{}, nil
		}
		return nil, err
	}

	var events []json.RawMessage
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		events = append(events, json.RawMessage(line))
	}
	return events, nil
}

func (s *Store) writeMeta(id string, meta SessionMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.metaPath(id), data, 0644)
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
