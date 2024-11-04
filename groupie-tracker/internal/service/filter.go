// internal/service/filter.go
package service

import (
	"fmt"
	"strconv"
	"strings"

	"groupie-tracker/internal/models"
)

type FilterService struct {
	cache *CacheService
}

func NewFilterService(cache *CacheService) *FilterService {
	return &FilterService{
		cache: cache,
	}
}

func (s *FilterService) FilterArtists(query string, filters models.FilterParams) ([]models.Artist, error) {
	// Validate filters
	if err := filters.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filters: %w", err)
	}

	// Get cached data
	data, err := s.cache.GetCachedData()
	if err != nil {
		return nil, fmt.Errorf("failed to get cached data: %w", err)
	}

	var results []models.Artist
	queryLower := strings.ToLower(query)

	// Create a map for quick location lookup
	artistLocations := make(map[int][]string)
	for _, loc := range data.LocationsData.Index {
		artistLocations[loc.ID] = loc.Locations
	}

	// Filter artists
	for _, artist := range data.ArtistsData {
		// Get locations for this artist
		locations := artistLocations[artist.ID]

		// Check if artist matches filters
		if !s.matchesFilters(artist, locations, filters) {
			continue
		}

		// If no query, include all artists that match filters
		if query == "" {
			results = append(results, artist)
			continue
		}

		// If it's a year query, only check the creation date
		if isYearQuery(query) {
			queryYear, _ := strconv.Atoi(query)
			if queryYear == artist.CreationDate {
				results = append(results, artist)
			}
			continue
		}

		// Check if artist matches query
		if s.matchesQuery(artist, locations, queryLower) {
			results = append(results, artist)
		}
	}

	return results, nil
}

func (s *FilterService) matchesFilters(artist models.Artist, locations []string, filters models.FilterParams) bool {
	// Check creation year
	if artist.CreationDate < filters.CreationYearMin || artist.CreationDate > filters.CreationYearMax {
		return false
	}

	// Check first album year
	firstAlbumYear, err := models.ParseFirstAlbumYear(artist.FirstAlbum)
	if err != nil || firstAlbumYear < filters.FirstAlbumYearMin || firstAlbumYear > filters.FirstAlbumYearMax {
		return false
	}

	// Check number of members
	if len(filters.Members) > 0 {
		memberCount := len(artist.Members)
		if !models.ContainsInt(filters.Members, memberCount) {
			return false
		}
	}

	// Check locations
	if len(filters.Locations) > 0 {
		matched := false
		for _, loc := range locations {
			locLower := strings.ToLower(strings.TrimSpace(loc))
			for _, filterLoc := range filters.Locations {
				if strings.Contains(locLower, strings.ToLower(filterLoc)) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func (s *FilterService) matchesQuery(artist models.Artist, locations []string, query string) bool {
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

func isYearQuery(query string) bool {
	_, err := strconv.Atoi(query)
	return err == nil && len(query) == 4
}
