package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"lorekeeper/models"
	"lorekeeper/store"
)

const charactersFile = "characters.json"

// RegisterCharacters registers the /characters and /characters/{id} routes.
func RegisterCharacters(mux *http.ServeMux) {
	mux.HandleFunc("/characters", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listCharacters(w, r)
		case http.MethodPost:
			createCharacter(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/characters/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/characters/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			getCharacter(w, r, id)
		case http.MethodPut:
			updateCharacter(w, r, id)
		case http.MethodDelete:
			deleteCharacter(w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func listCharacters(w http.ResponseWriter, _ *http.Request) {
	items, err := store.Load[models.Character](charactersFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, items)
}

func getCharacter(w http.ResponseWriter, _ *http.Request, id string) {
	items, err := store.Load[models.Character](charactersFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, c := range items {
		if c.ID == id {
			jsonResponse(w, http.StatusOK, c)
			return
		}
	}
	http.Error(w, "character not found", http.StatusNotFound)
}

func createCharacter(w http.ResponseWriter, r *http.Request) {
	var c models.Character
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	c.ID = models.NewUUID()

	items, err := store.Load[models.Character](charactersFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	items = append(items, c)
	if err := store.Save(charactersFile, items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, c)
}

func updateCharacter(w http.ResponseWriter, r *http.Request, id string) {
	items, err := store.Load[models.Character](charactersFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var updated models.Character
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	updated.ID = id

	found := false
	for i, c := range items {
		if c.ID == id {
			items[i] = updated
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "character not found", http.StatusNotFound)
		return
	}
	if err := store.Save(charactersFile, items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, updated)
}

func deleteCharacter(w http.ResponseWriter, _ *http.Request, id string) {
	items, err := store.Load[models.Character](charactersFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filtered := items[:0]
	found := false
	for _, c := range items {
		if c.ID == id {
			found = true
		} else {
			filtered = append(filtered, c)
		}
	}
	if !found {
		http.Error(w, "character not found", http.StatusNotFound)
		return
	}
	if err := store.Save(charactersFile, filtered); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
