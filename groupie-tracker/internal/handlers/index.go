// internal/handlers/index.go
package handlers

import (
	"net/http"
)

func (h *Handler) HandleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only handle the root path
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		// Set headers
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Execute template
		if err := h.indexTpl.Execute(w, nil); err != nil {
			h.logger.Printf("Error executing index template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
