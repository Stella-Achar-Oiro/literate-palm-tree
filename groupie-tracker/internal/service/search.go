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

	// Helper function to generate unique keys
	makeKey := func(text, type_ string) string {
		return fmt.Sprintf("%s|%s", text, type_)
	}

	// Helper function to check if string contains parts of query
	containsQueryParts := func(text string) bool {
		text = strings.ToLower(text)
		queryParts := strings.Fields(queryLower)
		for _, part := range queryParts {
			if !strings.Contains(text, part) {
				return false
			}
		}
		return true
	}

	for _, artist := range data.ArtistsData {
		// Artist name suggestions (case insensitive)
		if containsQueryParts(artist.Name) {
			key := makeKey(artist.Name, "artist")
			suggestions[key] = models.Suggestion{
				Text: artist.Name,
				Type: "artist/band",
			}
		}

		// Members suggestions - check each part of member name
		for _, member := range artist.Members {
			if containsQueryParts(member) {
				key := makeKey(member, "member")
				suggestions[key] = models.Suggestion{
					Text: member,
					Type: "member",
				}
				// Also add the artist for member searches
				artistKey := makeKey(artist.Name, "artist")
				suggestions[artistKey] = models.Suggestion{
					Text: artist.Name,
					Type: "artist/band",
				}
			}
		}

		// Creation date suggestions
		creationDate := strconv.Itoa(artist.CreationDate)
		if strings.Contains(creationDate, queryLower) {
			key := makeKey(creationDate, "creation date")
			suggestions[key] = models.Suggestion{
				Text: creationDate,
				Type: "creation date",
			}
		}

		// First album suggestions
		if strings.Contains(strings.ToLower(artist.FirstAlbum), queryLower) {
			key := makeKey(artist.FirstAlbum, "first album")
			suggestions[key] = models.Suggestion{
				Text: artist.FirstAlbum,
				Type: "first album",
			}
		}

		// Location suggestions with improved matching
		locations := s.GetLocationsForArtist(artist.ID)
		for _, location := range locations {
			location = strings.TrimSpace(location)
			if containsQueryParts(location) {
				key := makeKey(location, "location")
				suggestions[key] = models.Suggestion{
					Text: location,
					Type: "location",
				}
				// Also add the artist for location searches
				artistKey := makeKey(artist.Name, "artist")
				suggestions[artistKey] = models.Suggestion{
					Text: artist.Name,
					Type: "artist/band",
				}
			}
		}
	}

	// Convert map to slice and sort results
	result := make([]models.Suggestion, 0, len(suggestions))
	for _, suggestion := range suggestions {
		result = append(result, suggestion)
	}

	// sorting with priorities
	sort.Slice(result, func(i, j int) bool {
		typePriority := map[string]int{
			"artist/band":   1,
			"member":        2,
			"location":      3,
			"creation date": 4,
			"first album":   5,
		}

		// First compare by exact match
		iExact := strings.ToLower(result[i].Text) == queryLower
		jExact := strings.ToLower(result[j].Text) == queryLower
		if iExact != jExact {
			return iExact
		}

		// Then by type priority
		if result[i].Type != result[j].Type {
			return typePriority[result[i].Type] < typePriority[result[j].Type]
		}

		// Finally by text length and alphabetically
		if len(result[i].Text) != len(result[j].Text) {
			return len(result[i].Text) < len(result[j].Text)
		}
		return result[i].Text < result[j].Text
	})

	return result, nil
}

func (s *SearchService) SearchArtists(query string, filters models.FilterParams) ([]models.Artist, error) {
	data, err := s.cache.GetCachedData()
	if err != nil {
		return nil, fmt.Errorf("failed to get cached data: %w", err)
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))
	queryParts := strings.Fields(queryLower)
	var results []models.Artist

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

		// Check if matches any search criteria
		if s.matchesArtist(artist, queryParts) {
			results = append(results, artist)
		}
	}

	return results, nil
}

func (s *SearchService) matchesArtist(artist models.Artist, queryParts []string) bool {
	// Helper function to check if text contains all query parts
	containsAllParts := func(text string) bool {
		text = strings.ToLower(text)
		for _, part := range queryParts {
			if !strings.Contains(text, part) {
				return false
			}
		}
		return true
	}

	// Check name (case insensitive)
	if containsAllParts(artist.Name) {
		return true
	}

	// Check members
	for _, member := range artist.Members {
		if containsAllParts(member) {
			return true
		}
	}

	// Check creation date
	creationDate := strconv.Itoa(artist.CreationDate)
	if len(queryParts) == 1 && strings.Contains(creationDate, queryParts[0]) {
		return true
	}

	// Check first album
	if containsAllParts(artist.FirstAlbum) {
		return true
	}

	// Check locations
	locations := s.GetLocationsForArtist(artist.ID)
	for _, location := range locations {
		if containsAllParts(location) {
			return true
		}
	}

	return false
}

func (s *SearchService) matchesFilters(artist models.Artist, filters models.FilterParams) bool {
	// Get locations data for the artist
	cachedData, err := s.cache.GetCachedData()
	if err != nil {
		return false
	}

	// Get locations data for the artist
	var artistLocations []string
	for _, loc := range cachedData.LocationsData.Index {
		if loc.ID == artist.ID {
			artistLocations = loc.Locations
			break
		}
	}

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
	if len(filters.Locations) > 0 && len(artistLocations) > 0 {
		found := false
		for _, filterLoc := range filters.Locations {
			for _, loc := range artistLocations {
				if strings.Contains(strings.ToLower(loc), strings.ToLower(filterLoc)) {
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

// Add this helper function
func (s *SearchService) GetLocationsForArtist(artistID int) []string {
	data, err := s.cache.GetCachedData()
	if err != nil {
		return nil
	}

	for _, loc := range data.LocationsData.Index {
		if loc.ID == artistID {
			return loc.Locations
		}
	}
	return nil
}
