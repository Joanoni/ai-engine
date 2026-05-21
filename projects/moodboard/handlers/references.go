package handlers

import (
	"net/http"

	"moodboard/store"
)

// DeleteReference — DELETE /references/{id}
func DeleteReference(w http.ResponseWriter, r *http.Request) {
	id := lastSegment(r.URL.Path)

	store.Mu.Lock()
	defer store.Mu.Unlock()

	found := false
	newRefs := store.References[:0:0]
	for _, ref := range store.References {
		if ref.ID == id {
			found = true
			continue
		}
		newRefs = append(newRefs, ref)
	}
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "reference not found"})
		return
	}
	store.References = newRefs
	_ = store.SaveReferences()

	w.WriteHeader(http.StatusNoContent)
}
