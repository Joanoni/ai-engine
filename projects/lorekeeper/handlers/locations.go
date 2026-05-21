package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"lorekeeper/models"
	"lorekeeper/store"
)

const locationsFile = "locations.json"

// RegisterLocations registers the /locations and /locations/{id} routes.
func RegisterLocations(mux *http.ServeMux) {
	mux.HandleFunc("/locations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listLocations(w, r)
		case http.MethodPost:
			createLocation(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/locations/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/locations/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			getLocation(w, r, id)
		case http.MethodPut:
			updateLocation(w, r, id)
		case http.MethodDelete:
			deleteLocation(w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func listLocations(w http.ResponseWriter, _ *http.Request) {
	items, err := store.Load[models.Location](locationsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, items)
}

func getLocation(w http.ResponseWriter, _ *http.Request, id string) {
	items, err := store.Load[models.Location](locationsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, l := range items {
		if l.ID == id {
			jsonResponse(w, http.StatusOK, l)
			return
		}
	}
	http.Error(w, "location not found", http.StatusNotFound)
}

func createLocation(w http.ResponseWriter, r *http.Request) {
	var l models.Location
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	l.ID = models.NewUUID()

	items, err := store.Load[models.Location](locationsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	items = append(items, l)
	if err := store.Save(locationsFile, items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, l)
}

func updateLocation(w http.ResponseWriter, r *http.Request, id string) {
	items, err := store.Load[models.Location](locationsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var updated models.Location
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	updated.ID = id

	found := false
	for i, l := range items {
		if l.ID == id {
			items[i] = updated
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "location not found", http.StatusNotFound)
		return
	}
	if err := store.Save(locationsFile, items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, updated)
}

func deleteLocation(w http.ResponseWriter, _ *http.Request, id string) {
	items, err := store.Load[models.Location](locationsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filtered := items[:0]
	found := false
	for _, l := range items {
		if l.ID == id {
			found = true
		} else {
			filtered = append(filtered, l)
		}
	}
	if !found {
		http.Error(w, "location not found", http.StatusNotFound)
		return
	}
	if err := store.Save(locationsFile, filtered); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
