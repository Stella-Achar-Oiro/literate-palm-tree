// internal/service/search.go
package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"groupie-tracker/internal/models"
)

type SearchService struct {
	cache *CacheService
}

func NewSearchService(cache *CacheService) *SearchService {
	return &SearchService{cache: cache}
}

func (s *SearchService) GetSuggestions(query string) ([]models.Suggestion, error) {
	if query == "" {
		return []models.Suggestion{}, nil
	}

	data, err := s.cache.GetCachedData()
	if err != nil {
		return nil, fmt.Errorf("failed to get cached data: %w", err)
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))
	suggestions := make(map[string]models.Suggestion)

	// Create locations lookup map
	locationMap := make(map[int][]string)
	for _, loc := range data.LocationsData.Index {
		locationMap[loc.ID] = loc.Locations
	}

	for _, artist := range data.ArtistsData {
		// Artist name suggestions (case insensitive)
		if strings.Contains(strings.ToLower(artist.Name), queryLower) {
			suggestions[artist.Name] = models.Suggestion{
				Text: artist.Name,
				Type: "artist",
			}
		}

		// Members suggestions
		for _, member := range artist.Members {
			if strings.Contains(strings.ToLower(member), queryLower) {
				suggestions[member] = models.Suggestion{
					Text: member,
					Type: "member",
				}
			}
		}

		// Creation date suggestions
		creationDate := strconv.Itoa(artist.CreationDate)
		if strings.Contains(creationDate, query) {
			suggestions[creationDate] = models.Suggestion{
				Text: creationDate,
				Type: "creation date",
			}
		}

		// First album suggestions
		if strings.Contains(strings.ToLower(artist.FirstAlbum), queryLower) {
			suggestions[artist.FirstAlbum] = models.Suggestion{
				Text: artist.FirstAlbum,
				Type: "first album",
			}
		}

		// Location suggestions
		if locations, ok := locationMap[artist.ID]; ok {
			for _, location := range locations {
				location = strings.TrimSpace(location)
				if strings.Contains(strings.ToLower(location), queryLower) {
					suggestions[location] = models.Suggestion{
						Text: location,
						Type: "location",
					}
				}
			}
		}
	}

	// Convert map to slice and sort results
	result := make([]models.Suggestion, 0, len(suggestions))
	for _, suggestion := range suggestions {
		result = append(result, suggestion)
	}

	// Sort by type and then text
	sort.Slice(result, func(i, j int) bool {
		if result[i].Type == result[j].Type {
			return result[i].Text < result[j].Text
		}
		return result[i].Type < result[j].Type
	})

	return result, nil
}

// SearchArtists searches for artists based on query and filters
func (s *SearchService) SearchArtists(query string, filters models.FilterParams) ([]models.Artist, error) {
	data, err := s.cache.GetCachedData()
	if err != nil {
		return nil, fmt.Errorf("failed to get cached data: %w", err)
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))
	var results []models.Artist

	// Create locations lookup map
	locationMap := make(map[int][]string)
	for _, loc := range data.LocationsData.Index {
		locationMap[loc.ID] = loc.Locations
	}

	for _, artist := range data.ArtistsData {
		// Skip if doesn't match filters
		if !s.matchesFilters(artist, filters) {
			continue
		}

		// If no query, include all filtered artists
		if query == "" {
			results = append(results, artist)
			continue
		}

		// Check various fields
		if s.matchesArtist(artist, queryLower, locationMap[artist.ID]) {
			results = append(results, artist)
		}
	}

	return results, nil
}

func (s *SearchService) matchesFilters(artist models.Artist, filters models.FilterParams) bool {
	// Creation year filter
	if artist.CreationDate < filters.CreationYearMin ||
		artist.CreationDate > filters.CreationYearMax {
		return false
	}

	// First album year filter
	firstAlbumYear, err := parseFirstAlbumYear(artist.FirstAlbum)
	if err != nil || firstAlbumYear < filters.FirstAlbumYearMin ||
		firstAlbumYear > filters.FirstAlbumYearMax {
		return false
	}

	// Members filter
	if len(filters.Members) > 0 {
		memberCount := len(artist.Members)
		found := false
		for _, count := range filters.Members {
			if count == memberCount {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Locations filter
	if len(filters.Locations) > 0 {
		found := false
		for _, filterLoc := range filters.Locations {
			for _, loc := range artist.Locations {
				if strings.Contains(strings.ToLower(string(loc)), strings.ToLower(filterLoc)) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (s *SearchService) matchesArtist(artist models.Artist, query string, locations []string) bool {
	// Check name
	if strings.Contains(strings.ToLower(artist.Name), query) {
		return true
	}

	// Check members
	for _, member := range artist.Members {
		if strings.Contains(strings.ToLower(member), query) {
			return true
		}
	}

	// Check creation date
	if query == strconv.Itoa(artist.CreationDate) {
		return true
	}

	// Check first album
	if strings.Contains(strings.ToLower(artist.FirstAlbum), query) {
		return true
	}

	// Check locations
	for _, location := range locations {
		if strings.Contains(strings.ToLower(location), query) {
			return true
		}
	}

	return false
}

func parseFirstAlbumYear(firstAlbum string) (int, error) {
	parts := strings.Split(firstAlbum, "-")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid first album date format")
	}
	return strconv.Atoi(parts[2])
}

// GetLocations returns the geocoded locations for an artist
func (s *SearchService) GetLocations(id int, locationsData models.Location) []models.GeoLocation {
	var locations []models.GeoLocation

	for _, loc := range locationsData.Index {
		if loc.ID == id {
			for _, location := range loc.Locations {
				geoLoc, err := s.geocode(location)
				if err != nil {
					continue
				}
				locations = append(locations, geoLoc)
			}
			break
		}
	}

	return locations
}

// GetDates returns the concert dates for an artist
func (s *SearchService) GetDates(id int, datesData models.Date) []string {
	for _, date := range datesData.Index {
		if date.ID == id {
			return date.Dates
		}
	}
	return nil
}

// GetRelations returns the relations between dates and locations for an artist
func (s *SearchService) GetRelations(id int, relationsData models.Relation) map[string][]string {
	for _, rel := range relationsData.Index {
		if rel.ID == id {
			return rel.DatesLocations
		}
	}
	return nil
}

// geocode converts an address to coordinates using Mapbox API
func (s *SearchService) geocode(address string) (models.GeoLocation, error) {
	mapboxGeocodingAPI := models.GetMapboxGeocodingAPI()
	mapboxAccessToken := models.GetMapboxAccessToken()

	geocodingURL := fmt.Sprintf("%s/%s.json?access_token=%s",
		mapboxGeocodingAPI,
		url.QueryEscape(address),
		mapboxAccessToken,
	)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(geocodingURL)
	if err != nil {
		return models.GeoLocation{}, fmt.Errorf("failed to geocode address: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.GeoLocation{}, fmt.Errorf("geocoding API returned status %d", resp.StatusCode)
	}

	var result struct {
		Features []struct {
			Center [2]float64 `json:"center"`
		} `json:"features"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.GeoLocation{}, fmt.Errorf("failed to decode geocoding response: %w", err)
	}

	if len(result.Features) == 0 {
		return models.GeoLocation{}, fmt.Errorf("no results found for address: %s", address)
	}

	return models.GeoLocation{
		Address: address,
		Lon:     result.Features[0].Center[0],
		Lat:     result.Features[0].Center[1],
	}, nil
}
