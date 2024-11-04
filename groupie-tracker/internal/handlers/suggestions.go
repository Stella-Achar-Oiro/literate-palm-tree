// internal/handlers/suggestions.go
package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) HandleSuggestions(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	// Get suggestions
	suggestions, err := h.search.GetSuggestions(query)
	if err != nil {
		h.logger.Printf("Error getting suggestions: %v", err)
		http.Error(w, "Failed to get suggestions", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(suggestions); err != nil {
		h.logger.Printf("Error encoding suggestions: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
