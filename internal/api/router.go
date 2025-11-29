package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"github.com/mangashelf/mangashelf/internal/library"
	"github.com/mangashelf/mangashelf/internal/scraper"
)

// NewRouter configures the HTTP routes for the API.
func NewRouter(log zerolog.Logger, scrapers *scraper.Manager, lib *library.Service) *chi.Mux {
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

	r.Get("/api/manga", func(w http.ResponseWriter, req *http.Request) {
		manga, err := lib.ListManga(req.Context())
		if err != nil {
			log.Error().Err(err).Msg("failed to list manga")
			writeError(w, http.StatusInternalServerError, "LIST_FAILED", "failed to list manga")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": manga})
	})

	r.Post("/api/manga", func(w http.ResponseWriter, req *http.Request) {
		var body library.AddMangaRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
			return
		}

		if body.Source == "" || body.SourceID == "" {
			writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "source and sourceId are required")
			return
		}

		manga, err := lib.AddManga(req.Context(), body)
		if err != nil {
			log.Error().Err(err).Str("source", body.Source).Str("sourceId", body.SourceID).Msg("failed to add manga")
			if err == library.ErrMangaExists {
				writeError(w, http.StatusConflict, "MANGA_EXISTS", "manga already exists in library")
				return
			}
			writeError(w, http.StatusInternalServerError, "ADD_FAILED", "failed to add manga")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": manga})
	})

	r.Get("/api/manga/{id}", func(w http.ResponseWriter, req *http.Request) {
		idStr := chi.URLParam(req, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid manga ID")
			return
		}

		manga, err := lib.GetManga(req.Context(), id)
		if err != nil {
			if err == library.ErrMangaNotFound {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "manga not found")
				return
			}
			log.Error().Err(err).Int64("id", id).Msg("failed to get manga")
			writeError(w, http.StatusInternalServerError, "GET_FAILED", "failed to get manga")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": manga})
	})

	r.Delete("/api/manga/{id}", func(w http.ResponseWriter, req *http.Request) {
		idStr := chi.URLParam(req, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid manga ID")
			return
		}

		if err := lib.DeleteManga(req.Context(), id); err != nil {
			log.Error().Err(err).Int64("id", id).Msg("failed to delete manga")
			writeError(w, http.StatusInternalServerError, "DELETE_FAILED", "failed to delete manga")
			return
		}

		w.WriteHeader(http.StatusNoContent)
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
