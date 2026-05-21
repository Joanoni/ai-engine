package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"lorekeeper/models"
	"lorekeeper/store"
)

const eventsFile = "events.json"

// RegisterEvents registers the /events and /events/{id} routes.
func RegisterEvents(mux *http.ServeMux) {
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listEvents(w, r)
		case http.MethodPost:
			createEvent(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/events/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/events/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			getEvent(w, r, id)
		case http.MethodPut:
			updateEvent(w, r, id)
		case http.MethodDelete:
			deleteEvent(w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func listEvents(w http.ResponseWriter, _ *http.Request) {
	items, err := store.Load[models.Event](eventsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, items)
}

func getEvent(w http.ResponseWriter, _ *http.Request, id string) {
	items, err := store.Load[models.Event](eventsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, e := range items {
		if e.ID == id {
			jsonResponse(w, http.StatusOK, e)
			return
		}
	}
	http.Error(w, "event not found", http.StatusNotFound)
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	var e models.Event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	e.ID = models.NewUUID()

	items, err := store.Load[models.Event](eventsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	items = append(items, e)
	if err := store.Save(eventsFile, items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, e)
}

func updateEvent(w http.ResponseWriter, r *http.Request, id string) {
	items, err := store.Load[models.Event](eventsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var updated models.Event
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	updated.ID = id

	found := false
	for i, e := range items {
		if e.ID == id {
			items[i] = updated
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	if err := store.Save(eventsFile, items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, updated)
}

func deleteEvent(w http.ResponseWriter, _ *http.Request, id string) {
	items, err := store.Load[models.Event](eventsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filtered := items[:0]
	found := false
	for _, e := range items {
		if e.ID == id {
			found = true
		} else {
			filtered = append(filtered, e)
		}
	}
	if !found {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	if err := store.Save(eventsFile, filtered); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
