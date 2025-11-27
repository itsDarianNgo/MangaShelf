package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"github.com/mangashelf/mangashelf/internal/scraper"
)

// NewRouter configures the HTTP routes for the API.
func NewRouter(log zerolog.Logger, scrapers *scraper.Manager) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		log.Debug().Msg("health check")
	})

	r.Get("/api/sources", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		sources := scrapers.List()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": sources,
		})
	})

	return r
}
