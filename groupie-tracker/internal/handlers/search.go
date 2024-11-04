// internal/handlers/search.go
package handlers

import (
	"encoding/json"
	"net/http"

	"groupie-tracker/internal/models"
)

func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameter
	query := r.URL.Query().Get("q")

	// Parse filter parameters from request body
	var filters models.FilterParams
	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		h.logger.Printf("Error decoding filter parameters: %v", err)
		http.Error(w, "Invalid filter parameters", http.StatusBadRequest)
		return
	}

	// Search for artists
	results, err := h.filter.FilterArtists(query, filters)
	if err != nil {
		h.logger.Printf("Error searching artists: %v", err)
		http.Error(w, "Failed to search artists", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := models.SearchResult{
		Artists: results,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Error encoding search results: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
