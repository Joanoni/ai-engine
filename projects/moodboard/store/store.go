package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// ── Models ───────────────────────────────────────────────────────────────────

type Board struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
}

type Reference struct {
	ID        string `json:"id"`
	BoardID   string `json:"boardId"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	Label     string `json:"label"`
	CreatedAt string `json:"createdAt"`
}

// ── In-memory state ──────────────────────────────────────────────────────────

var (
	Boards     []Board
	References []Reference
	Mu         sync.RWMutex
)

var counter int64

// ── ID generation ────────────────────────────────────────────────────────────

// NewID returns a unique string ID. A global atomic counter is appended to the
// nanosecond timestamp to prevent collisions when IDs are generated in rapid
// succession (e.g. during seeding).
func NewID() string {
	n := atomic.AddInt64(&counter, 1)
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), n)
}

// ── Persistence ──────────────────────────────────────────────────────────────

const (
	dataDir        = "data"
	boardsFile     = "data/boards.json"
	referencesFile = "data/references.json"
)

func LoadData() {
	// Ensure data directory exists.
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		panic("store: cannot create data dir: " + err.Error())
	}

	// If boards file doesn't exist, seed and return.
	if _, err := os.Stat(boardsFile); os.IsNotExist(err) {
		Seed()
		return
	}

	// Load boards.
	if data, err := os.ReadFile(boardsFile); err == nil {
		_ = json.Unmarshal(data, &Boards)
	}
	if Boards == nil {
		Boards = []Board{}
	}

	// Load references.
	if data, err := os.ReadFile(referencesFile); err == nil {
		_ = json.Unmarshal(data, &References)
	}
	if References == nil {
		References = []Reference{}
	}
}

func SaveBoards() error {
	data, err := json.MarshalIndent(Boards, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(boardsFile, data, 0o644)
}

func SaveReferences() error {
	data, err := json.MarshalIndent(References, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(referencesFile, data, 0o644)
}

// ── Seed ─────────────────────────────────────────────────────────────────────

func Seed() {
	now := time.Now().UTC().Format(time.RFC3339)

	b1 := Board{ID: NewID(), Name: "Branding Palette", Description: "Color and image references for the brand identity.", CreatedAt: now}
	b2 := Board{ID: NewID(), Name: "UI Inspiration", Description: "Dark UI patterns and typographic notes.", CreatedAt: now}

	Boards = []Board{b1, b2}

	References = []Reference{
		// Branding Palette — image, color, note
		{
			ID: NewID(), BoardID: b1.ID, Type: "image",
			Content:   "https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=800",
			Label:     "Hero texture",
			CreatedAt: now,
		},
		{
			ID: NewID(), BoardID: b1.ID, Type: "color",
			Content:   "#F4A261",
			Label:     "Primary orange",
			CreatedAt: now,
		},
		{
			ID: NewID(), BoardID: b1.ID, Type: "note",
			Content:   "Keep the palette warm and earthy — avoid cold blues.",
			Label:     "Direction note",
			CreatedAt: now,
		},
		// UI Inspiration — image, color, note
		{
			ID: NewID(), BoardID: b2.ID, Type: "image",
			Content:   "https://images.unsplash.com/photo-1555066931-4365d14bab8c?w=800",
			Label:     "Dark code editor",
			CreatedAt: now,
		},
		{
			ID: NewID(), BoardID: b2.ID, Type: "color",
			Content:   "#1E1E2E",
			Label:     "Background dark",
			CreatedAt: now,
		},
		{
			ID: NewID(), BoardID: b2.ID, Type: "note",
			Content:   "Use subtle glows instead of hard borders for depth.",
			Label:     "UI principle",
			CreatedAt: now,
		},
	}

	_ = SaveBoards()
	_ = SaveReferences()
}
