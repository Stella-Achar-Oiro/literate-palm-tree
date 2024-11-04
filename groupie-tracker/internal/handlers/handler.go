// internal/handlers/handler.go
package handlers

import (
	"html/template"
	"log"

	"groupie-tracker/internal/service"
)

type Config struct {
	IndexTemplate  *template.Template
	ArtistTemplate *template.Template
	CacheService   *service.CacheService
	FilterService  *service.FilterService
	SearchService  *service.SearchService
	Logger         *log.Logger
}

type Handler struct {
	indexTpl  *template.Template
	artistTpl *template.Template
	cache     *service.CacheService
	filter    *service.FilterService
	search    *service.SearchService
	logger    *log.Logger
}

func NewHandler(config Config) *Handler {
	return &Handler{
		indexTpl:  config.IndexTemplate,
		artistTpl: config.ArtistTemplate,
		cache:     config.CacheService,
		filter:    config.FilterService,
		search:    config.SearchService,
		logger:    config.Logger,
	}
}
