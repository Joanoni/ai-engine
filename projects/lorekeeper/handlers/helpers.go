package handlers

import (
	"encoding/json"
	"net/http"
)

// jsonResponse writes v as a JSON body with the given HTTP status code.
func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
