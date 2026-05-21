package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"moodboard/store"
)

// ── Shared helper ─────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ── Board handlers ────────────────────────────────────────────────────────────

// ListBoards — GET /boards
func ListBoards(w http.ResponseWriter, r *http.Request) {
	store.Mu.RLock()
	boards := make([]store.Board, len(store.Boards))
	copy(boards, store.Boards)
	store.Mu.RUnlock()
	writeJSON(w, http.StatusOK, boards)
}

// CreateBoard — POST /boards
func CreateBoard(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	board := store.Board{
		ID:          store.NewID(),
		Name:        strings.TrimSpace(body.Name),
		Description: strings.TrimSpace(body.Description),
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	store.Mu.Lock()
	store.Boards = append(store.Boards, board)
	_ = store.SaveBoards()
	store.Mu.Unlock()

	writeJSON(w, http.StatusCreated, board)
}

// GetBoard — GET /boards/{id}
func GetBoard(w http.ResponseWriter, r *http.Request) {
	id := lastSegment(r.URL.Path)

	store.Mu.RLock()
	defer store.Mu.RUnlock()

	for _, b := range store.Boards {
		if b.ID == id {
			writeJSON(w, http.StatusOK, b)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "board not found"})
}

// DeleteBoard — DELETE /boards/{id}
func DeleteBoard(w http.ResponseWriter, r *http.Request) {
	id := lastSegment(r.URL.Path)

	store.Mu.Lock()
	defer store.Mu.Unlock()

	found := false
	newBoards := store.Boards[:0:0]
	for _, b := range store.Boards {
		if b.ID == id {
			found = true
			continue
		}
		newBoards = append(newBoards, b)
	}
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "board not found"})
		return
	}
	store.Boards = newBoards

	// Remove all references belonging to this board.
	newRefs := store.References[:0:0]
	for _, ref := range store.References {
		if ref.BoardID != id {
			newRefs = append(newRefs, ref)
		}
	}
	store.References = newRefs

	_ = store.SaveBoards()
	_ = store.SaveReferences()

	w.WriteHeader(http.StatusNoContent)
}

// ListReferences — GET /boards/{id}/references
func ListReferences(w http.ResponseWriter, r *http.Request) {
	// Path: /boards/{id}/references  — board id is the second segment.
	id := boardIDFromPath(r.URL.Path)

	store.Mu.RLock()
	defer store.Mu.RUnlock()

	// Verify board exists.
	boardFound := false
	for _, b := range store.Boards {
		if b.ID == id {
			boardFound = true
			break
		}
	}
	if !boardFound {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "board not found"})
		return
	}

	result := []store.Reference{}
	for _, ref := range store.References {
		if ref.BoardID == id {
			result = append(result, ref)
		}
	}
	writeJSON(w, http.StatusOK, result)
}

// CreateReference — POST /boards/{id}/references
func CreateReference(w http.ResponseWriter, r *http.Request) {
	boardID := boardIDFromPath(r.URL.Path)

	// Verify board exists first.
	store.Mu.RLock()
	boardFound := false
	for _, b := range store.Boards {
		if b.ID == boardID {
			boardFound = true
			break
		}
	}
	store.Mu.RUnlock()

	if !boardFound {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "board not found"})
		return
	}

	var body struct {
		Type    string `json:"type"`
		Content string `json:"content"`
		Label   string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(body.Type) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type is required"})
		return
	}
	if strings.TrimSpace(body.Content) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required"})
		return
	}

	ref := store.Reference{
		ID:        store.NewID(),
		BoardID:   boardID,
		Type:      strings.TrimSpace(body.Type),
		Content:   strings.TrimSpace(body.Content),
		Label:     strings.TrimSpace(body.Label),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	store.Mu.Lock()
	store.References = append(store.References, ref)
	_ = store.SaveReferences()
	store.Mu.Unlock()

	writeJSON(w, http.StatusCreated, ref)
}

// ── Path helpers ──────────────────────────────────────────────────────────────

// lastSegment returns the final non-empty path segment.
// e.g. "/boards/123" → "123"
func lastSegment(path string) string {
	path = strings.TrimRight(path, "/")
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return path
	}
	return path[idx+1:]
}

// boardIDFromPath extracts the board ID from paths like /boards/{id}/references.
func boardIDFromPath(path string) string {
	// Trim leading/trailing slashes, then split.
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// parts: ["boards", "{id}", "references"]
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
