package library

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog"

	"github.com/mangashelf/mangashelf/internal/database"
	"github.com/mangashelf/mangashelf/internal/scraper"
)

// Service manages the manga library.
type Service struct {
	db       *database.Queries
	scrapers *scraper.Manager
	log      zerolog.Logger
}

// NewService creates a new library service.
func NewService(db *database.Queries, scrapers *scraper.Manager, log zerolog.Logger) *Service {
	return &Service{
		db:       db,
		scrapers: scrapers,
		log:      log.With().Str("component", "library").Logger(),
	}
}

// AddMangaRequest contains parameters for adding manga to the library.
type AddMangaRequest struct {
	Source   string `json:"source"`
	SourceID string `json:"sourceId"`
}

// AddManga fetches manga from a source and adds it to the library.
func (s *Service) AddManga(ctx context.Context, req AddMangaRequest) (*database.Manga, error) {
	manga, err := s.scrapers.GetManga(ctx, req.Source, req.SourceID)
	if err != nil {
		return nil, fmt.Errorf("fetch manga: %w", err)
	}

	slug := generateSlug(manga.Title)

	genresJSON, _ := json.Marshal(manga.Genres)
	tagsJSON, _ := json.Marshal(manga.Tags)

	dbManga, err := s.db.InsertManga(ctx, database.InsertMangaParams{
		Title:       manga.Title,
		Slug:        slug,
		Source:      req.Source,
		SourceID:    req.SourceID,
		Url:         manga.URL,
		CoverUrl:    toNullString(manga.CoverURL),
		Description: toNullString(manga.Description),
		Status:      toNullString(manga.Status),
		Author:      toNullString(manga.Author),
		Artist:      toNullString(manga.Artist),
		Genres:      toNullString(string(genresJSON)),
		Tags:        toNullString(string(tagsJSON)),
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrMangaExists
		}
		return nil, fmt.Errorf("insert manga: %w", err)
	}

	s.log.Info().
		Str("title", manga.Title).
		Str("source", req.Source).
		Int64("id", dbManga.ID).
		Msg("manga added to library")

	return dbManga, nil
}

// GetManga retrieves a manga by ID.
func (s *Service) GetManga(ctx context.Context, id int64) (*database.Manga, error) {
	manga, err := s.db.GetManga(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrMangaNotFound
		}
		return nil, fmt.Errorf("get manga: %w", err)
	}
	return manga, nil
}

// ListManga returns all manga in the library.
func (s *Service) ListManga(ctx context.Context) ([]*database.Manga, error) {
	manga, err := s.db.ListManga(ctx)
	if err != nil {
		return nil, fmt.Errorf("list manga: %w", err)
	}
	return manga, nil
}

// DeleteManga removes a manga from the library.
func (s *Service) DeleteManga(ctx context.Context, id int64) error {
	err := s.db.DeleteManga(ctx, id)
	if err != nil {
		return fmt.Errorf("delete manga: %w", err)
	}
	s.log.Info().Int64("id", id).Msg("manga deleted from library")
	return nil
}

// generateSlug creates a URL-friendly slug from a title.
func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")

	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// toNullString converts a string to sql.NullString.
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
