package main

import (
	"log"
	"net/http"
	"strings"

	"moodboard/handlers"
	"moodboard/store"
)

// corsMiddleware adds CORS headers to every response and handles OPTIONS preflight.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	store.LoadData()

	mux := http.NewServeMux()

	// ── /boards (exact) ───────────────────────────────────────────────────────
	mux.HandleFunc("/boards", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.ListBoards(w, r)
		case http.MethodPost:
			handlers.CreateBoard(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	// ── /boards/{id}[/references] ─────────────────────────────────────────────
	mux.HandleFunc("/boards/", func(w http.ResponseWriter, r *http.Request) {
		// Trim the prefix and split into segments.
		// Path examples:
		//   /boards/123              → segments: ["123"]
		//   /boards/123/references   → segments: ["123", "references"]
		trimmed := strings.TrimPrefix(r.URL.Path, "/boards/")
		trimmed = strings.Trim(trimmed, "/")
		segments := strings.Split(trimmed, "/")

		switch len(segments) {
		case 1:
			// /boards/{id}
			if segments[0] == "" {
				http.NotFound(w, r)
				return
			}
			switch r.Method {
			case http.MethodGet:
				handlers.GetBoard(w, r)
			case http.MethodDelete:
				handlers.DeleteBoard(w, r)
			default:
				http.NotFound(w, r)
			}

		case 2:
			// /boards/{id}/references
			if segments[1] != "references" {
				http.NotFound(w, r)
				return
			}
			switch r.Method {
			case http.MethodGet:
				handlers.ListReferences(w, r)
			case http.MethodPost:
				handlers.CreateReference(w, r)
			default:
				http.NotFound(w, r)
			}

		default:
			http.NotFound(w, r)
		}
	})

	// ── /references/{id} ──────────────────────────────────────────────────────
	mux.HandleFunc("/references/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			handlers.DeleteReference(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	log.Println("Moodboard API listening on :3000")
	log.Fatal(http.ListenAndServe(":3000", corsMiddleware(mux)))
}
