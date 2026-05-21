package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"lorekeeper/models"
	"lorekeeper/store"
)

const factionsFile = "factions.json"

// RegisterFactions registers the /factions and /factions/{id} routes.
func RegisterFactions(mux *http.ServeMux) {
	mux.HandleFunc("/factions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listFactions(w, r)
		case http.MethodPost:
			createFaction(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/factions/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/factions/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			getFaction(w, r, id)
		case http.MethodPut:
			updateFaction(w, r, id)
		case http.MethodDelete:
			deleteFaction(w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func listFactions(w http.ResponseWriter, _ *http.Request) {
	items, err := store.Load[models.Faction](factionsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, items)
}

func getFaction(w http.ResponseWriter, _ *http.Request, id string) {
	items, err := store.Load[models.Faction](factionsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, f := range items {
		if f.ID == id {
			jsonResponse(w, http.StatusOK, f)
			return
		}
	}
	http.Error(w, "faction not found", http.StatusNotFound)
}

func createFaction(w http.ResponseWriter, r *http.Request) {
	var f models.Faction
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	f.ID = models.NewUUID()

	items, err := store.Load[models.Faction](factionsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	items = append(items, f)
	if err := store.Save(factionsFile, items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, f)
}

func updateFaction(w http.ResponseWriter, r *http.Request, id string) {
	items, err := store.Load[models.Faction](factionsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var updated models.Faction
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	updated.ID = id

	found := false
	for i, f := range items {
		if f.ID == id {
			items[i] = updated
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "faction not found", http.StatusNotFound)
		return
	}
	if err := store.Save(factionsFile, items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, updated)
}

func deleteFaction(w http.ResponseWriter, _ *http.Request, id string) {
	items, err := store.Load[models.Faction](factionsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filtered := items[:0]
	found := false
	for _, f := range items {
		if f.ID == id {
			found = true
		} else {
			filtered = append(filtered, f)
		}
	}
	if !found {
		http.Error(w, "faction not found", http.StatusNotFound)
		return
	}
	if err := store.Save(factionsFile, filtered); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
