// internal/handlers/search.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"groupie-tracker/internal/models"
)

// HandleSearch handles the search API endpoint
func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	// Set CORS and content type headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		return
	}

	// Only allow POST method
	if r.Method != http.MethodPost {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get and validate query parameter
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	// Get cached data
	cachedData, err := h.cache.GetCachedData()
	if err != nil {
		h.logger.Printf("Error getting cached data: %v", err)
		h.sendError(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	// Initialize filter parameters with data-driven defaults
	filters := h.getDefaultFilters(cachedData)

	// Parse filter parameters from request body if present
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
			h.logger.Printf("Warning: Error decoding filter parameters: %v", err)
			// Continue with default filters
		}
	}

	// Validate and normalize filter parameters
	if err := h.validateFilters(&filters, cachedData); err != nil {
		h.logger.Printf("Invalid filter parameters: %v", err)
		h.sendError(w, "Invalid filter parameters: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Search for artists
	results, err := h.search.SearchArtists(query, filters)
	if err != nil {
		h.logger.Printf("Error searching artists: %v", err)
		h.sendError(w, "Failed to search artists", http.StatusInternalServerError)
		return
	}

	// Send response
	response := models.SearchResult{
		Artists: results,
		Total:   len(results),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Error encoding search results: %v", err)
		h.sendError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getDefaultFilters returns data-driven default filter values
func (h *Handler) getDefaultFilters(data models.Datas) models.FilterParams {
	currentYear := time.Now().Year()

	// Find the earliest creation year from actual data
	minCreationYear := currentYear
	for _, artist := range data.ArtistsData {
		if artist.CreationDate < minCreationYear {
			minCreationYear = artist.CreationDate
		}
	}

	// Find the earliest first album year
	minAlbumYear := currentYear
	for _, artist := range data.ArtistsData {
		if year, err := parseFirstAlbumYear(artist.FirstAlbum); err == nil && year < minAlbumYear {
			minAlbumYear = year
		}
	}

	return models.FilterParams{
		CreationYearMin:   minCreationYear,
		CreationYearMax:   currentYear,
		FirstAlbumYearMin: minAlbumYear,
		FirstAlbumYearMax: currentYear,
	}
}

// validateFilters validates and normalizes filter parameters
func (h *Handler) validateFilters(filters *models.FilterParams, data models.Datas) error {
	currentYear := time.Now().Year()

	// Validate year ranges
	defaults := h.getDefaultFilters(data)

	if filters.CreationYearMin < defaults.CreationYearMin {
		filters.CreationYearMin = defaults.CreationYearMin
	}
	if filters.CreationYearMax > currentYear {
		filters.CreationYearMax = currentYear
	}
	if filters.CreationYearMin > filters.CreationYearMax {
		return fmt.Errorf("invalid creation year range: min (%d) > max (%d)",
			filters.CreationYearMin, filters.CreationYearMax)
	}

	if filters.FirstAlbumYearMin < defaults.FirstAlbumYearMin {
		filters.FirstAlbumYearMin = defaults.FirstAlbumYearMin
	}
	if filters.FirstAlbumYearMax > currentYear {
		filters.FirstAlbumYearMax = currentYear
	}
	if filters.FirstAlbumYearMin > filters.FirstAlbumYearMax {
		return fmt.Errorf("invalid first album year range: min (%d) > max (%d)",
			filters.FirstAlbumYearMin, filters.FirstAlbumYearMax)
	}

	// Validate member counts
	for _, count := range filters.Members {
		if count < 1 {
			return fmt.Errorf("invalid member count: %d (must be positive)", count)
		}
	}

	// Validate locations (trim and normalize)
	for i, loc := range filters.Locations {
		filters.Locations[i] = strings.TrimSpace(loc)
		if filters.Locations[i] == "" {
			return fmt.Errorf("empty location not allowed")
		}
	}

	return nil
}

// Helper function to parse first album year
func parseFirstAlbumYear(firstAlbum string) (int, error) {
	parts := strings.Split(firstAlbum, "-")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid first album date format")
	}
	return strconv.Atoi(parts[2])
}

// sendError sends a JSON error response
func (h *Handler) sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := models.Error{
		Code:    code,
		Message: message,
	}

	if encodeErr := json.NewEncoder(w).Encode(err); encodeErr != nil {
		h.logger.Printf("Error encoding error response: %v", encodeErr)
	}
}
