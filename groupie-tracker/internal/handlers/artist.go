// internal/handlers/artist.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"groupie-tracker/internal/models"
)

func (h *Handler) HandleArtist(w http.ResponseWriter, r *http.Request) {
	// Extract artist ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/artist/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Printf("Invalid artist ID: %v", err)
		http.Error(w, "Invalid artist ID", http.StatusBadRequest)
		return
	}

	// Get cached data
	cachedData, err := h.cache.GetCachedData()
	if err != nil {
		h.logger.Printf("Error getting cached data: %v", err)
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	// Find artist
	var artist models.Artist
	for _, a := range cachedData.ArtistsData {
		if a.ID == id {
			artist = a
			break
		}
	}

	if artist.ID == 0 {
		http.Error(w, "Artist not found", http.StatusNotFound)
		return
	}

	// Get artist details
	details := models.ArtistDetail{
		Artist:    artist,
		Locations: h.search.GetLocations(id, cachedData.LocationsData),
		Dates:     h.search.GetDates(id, cachedData.DatesData),
		Relations: h.search.GetRelations(id, cachedData.RelationsData),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(details); err != nil {
		h.logger.Printf("Error encoding artist details: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleArtistDetails() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract artist ID from URL
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			http.Error(w, "Invalid artist ID", http.StatusBadRequest)
			return
		}

		// Get artist ID
		artistID := parts[len(parts)-1]

		// Create template data
		data := struct {
			ArtistID string
		}{
			ArtistID: artistID,
		}

		// Set headers
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Execute template
		if err := h.artistTpl.Execute(w, data); err != nil {
			h.logger.Printf("Error executing artist template: %v", err)
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			return
		}
	}
}
