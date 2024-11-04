// internal/models/models.go
package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Constants for external APIs
const (
	ArtistsAPI   = "https://groupietrackers.herokuapp.com/api/artists"
	LocationsAPI = "https://groupietrackers.herokuapp.com/api/locations"
	DatesAPI     = "https://groupietrackers.herokuapp.com/api/dates"
	RelationsAPI = "https://groupietrackers.herokuapp.com/api/relation"
)

// Artist represents a musical artist or band
type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	Relations    string   `json:"relations"`
}

// Location represents the concert locations for artists
type Location struct {
	Index []struct {
		ID        int      `json:"id"`
		Locations []string `json:"locations"`
		Dates     string   `json:"dates"`
	} `json:"index"`
}

// Date represents the concert dates for artists
type Date struct {
	Index []struct {
		ID    int      `json:"id"`
		Dates []string `json:"dates"`
	} `json:"index"`
}

// Relation represents the relationship between dates and locations
type Relation struct {
	Index []struct {
		ID             int                 `json:"id"`
		DatesLocations map[string][]string `json:"datesLocations"`
	} `json:"index"`
}

// Datas represents all the data combined
type Datas struct {
	ArtistsData   []Artist `json:"artists"`
	LocationsData Location `json:"locations"`
	DatesData     Date     `json:"dates"`
	RelationsData Relation `json:"relations"`
}

// SearchResult represents the result of a search query
type SearchResult struct {
	Artists []Artist `json:"artists"`
	Total   int      `json:"total,omitempty"`
	Page    int      `json:"page,omitempty"`
}

// Suggestion represents a search suggestion
type Suggestion struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

// GeoLocation represents a geographical location with coordinates
type GeoLocation struct {
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// Event represents a concert event
type Event struct {
	Location string    `json:"location"`
	Date     time.Time `json:"date"`
}

// ArtistDetail represents detailed information about an artist
type ArtistDetail struct {
	Artist    Artist              `json:"artist"`
	Locations []GeoLocation       `json:"locations"`
	Dates     []string            `json:"dates"`
	Relations map[string][]string `json:"relations"`
	Events    []Event             `json:"events,omitempty"`
}

// FilterParams represents the parameters for filtering artists
type FilterParams struct {
	CreationYearMin   int      `json:"creationYearMin"`
	CreationYearMax   int      `json:"creationYearMax"`
	FirstAlbumYearMin int      `json:"firstAlbumYearMin"`
	FirstAlbumYearMax int      `json:"firstAlbumYearMax"`
	Members           []int    `json:"members"`
	Locations         []string `json:"locations"`
}

// Error represents an API error response
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Mapbox configuration variables
var (
	MapboxAccessToken  string
	MapboxGeocodingAPI string
)

// InitConstants initializes the Mapbox configuration
func InitConstants(accessToken, geocodingAPI string) {
	MapboxAccessToken = accessToken
	MapboxGeocodingAPI = geocodingAPI
}

// GetMapboxAccessToken returns the Mapbox access token
func GetMapboxAccessToken() string {
	return MapboxAccessToken
}

// GetMapboxGeocodingAPI returns the Mapbox Geocoding API URL
func GetMapboxGeocodingAPI() string {
	return MapboxGeocodingAPI
}

// Validate validates the filter parameters
func (f *FilterParams) Validate() error {
	if f.CreationYearMin > f.CreationYearMax {
		return fmt.Errorf("creation year minimum cannot be greater than maximum")
	}
	if f.FirstAlbumYearMin > f.FirstAlbumYearMax {
		return fmt.Errorf("first album year minimum cannot be greater than maximum")
	}
	return nil
}

// ToJSON converts a struct to JSON string
func ToJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON converts JSON string to a struct
func FromJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

// ParseFirstAlbumYear extracts the year from the first album date string
func ParseFirstAlbumYear(firstAlbum string) (int, error) {
	parts := strings.Split(firstAlbum, "-")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid first album date format")
	}
	return strconv.Atoi(parts[2])
}

// ContainsInt checks if a slice contains an integer
func ContainsInt(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// ContainsString checks if a slice contains a string (case insensitive)
func ContainsString(slice []string, val string) bool {
	valLower := strings.ToLower(val)
	for _, item := range slice {
		if strings.ToLower(item) == valLower {
			return true
		}
	}
	return false
}

// FormatDate formats a date string to a consistent format
func FormatDate(date string) (string, error) {
	t, err := time.Parse("02-01-2006", date)
	if err != nil {
		return "", err
	}
	return t.Format("January 2, 2006"), nil
}
