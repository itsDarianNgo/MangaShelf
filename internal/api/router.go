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

	r.Get("/api/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			writeError(w, http.StatusBadRequest, "MISSING_QUERY", "query parameter 'q' is required")
			return
		}

		source := r.URL.Query().Get("source")
		if source == "" {
			source = "mangadex"
		}

		results, err := scrapers.Search(r.Context(), source, query)
		if err != nil {
			log.Error().Err(err).Str("source", source).Str("query", query).Msg("search failed")

			if err.Error() == "provider not found: "+source {
				writeError(w, http.StatusBadRequest, "UNKNOWN_SOURCE", "source '"+source+"' not found")
				return
			}

			writeError(w, http.StatusInternalServerError, "SEARCH_FAILED", "failed to search manga")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": results,
		})
	})

	return r
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
