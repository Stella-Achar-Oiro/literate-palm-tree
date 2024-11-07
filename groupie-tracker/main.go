package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"groupie-tracker/internal/handlers"
	"groupie-tracker/internal/models"
	"groupie-tracker/internal/service"
)

var (
	indexTpl         *template.Template
	artistDetailsTpl *template.Template
	logger           *log.Logger
)

const cacheDuration = 1 * time.Hour

func main() {
	// Initialize logger with file name and line number
	logger = log.New(os.Stdout, "GROUPIE-TRACKER: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialize models package with required constants
	models.InitConstants(service.GetMapboxAccessToken(), service.GetMapboxGeocodingAPI())
	logger.Println("Models initialized with Mapbox constants")

	// Initialize services
	cacheService := service.NewCacheService(cacheDuration)
	filterService := service.NewFilterService(cacheService)
	searchService := service.NewSearchService(cacheService)

	// Initialize cache with initial data
	if err := cacheService.RefreshCache(); err != nil {
		logger.Fatalf("Failed to fetch initial data: %v", err)
	}
	logger.Println("Initial data fetched successfully")

	// Parse HTML templates
	var err error
	indexTpl, err = template.ParseFiles("web/templates/index.html")
	if err != nil {
		logger.Fatalf("Failed to parse index template: %v", err)
	}

	artistDetailsTpl, err = template.ParseFiles("web/templates/artist-details.html")
	if err != nil {
		logger.Fatalf("Failed to parse artist details template: %v", err)
	}

	// Initialize handlers
	handlerConfig := handlers.Config{
		IndexTemplate:  indexTpl,
		ArtistTemplate: artistDetailsTpl,
		CacheService:   cacheService,
		FilterService:  filterService,
		SearchService:  searchService,
		Logger:         logger,
	}

	h := handlers.NewHandler(handlerConfig)

	// Set up routes
	http.HandleFunc("/", h.HandleIndex())
	http.HandleFunc("/artist/", h.HandleArtistDetails())
	http.HandleFunc("/api/search", h.HandleSearch)
	http.HandleFunc("/api/artist/", h.HandleArtist)
	http.HandleFunc("/api/suggestions", h.HandleSuggestions)

	// Serve static files
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	logger.Println("Routes and static file server set up")

	// Start server
	port := ":8000"
	logger.Printf("Server starting on %s", port)
	server := &http.Server{
		Addr:         port,
		Handler:      nil, // Use default ServeMux
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Fatalf("Server failed to start: %v", err)
	}
}
