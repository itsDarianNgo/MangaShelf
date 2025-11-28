package scraper

import (
	"context"
	"time"
)

// Provider defines the interface all manga sources must implement.
type Provider interface {
	// Info returns metadata about this provider.
	Info() ProviderInfo

	// Search finds manga matching the query.
	Search(ctx context.Context, query string) ([]MangaResult, error)

	// GetManga fetches full details for a manga.
	GetManga(ctx context.Context, id string) (*Manga, error)

	// GetChapters fetches all chapters for a manga.
	GetChapters(ctx context.Context, mangaID string) ([]Chapter, error)

	// GetPages fetches all page URLs for a chapter.
	GetPages(ctx context.Context, chapterID string) ([]Page, error)
}

// ProviderInfo contains metadata about a provider.
type ProviderInfo struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	BaseURL   string   `json:"baseUrl"`
	Languages []string `json:"languages"`
	IsNSFW    bool     `json:"isNsfw"`
}

// MangaResult is a search result item.
type MangaResult struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	CoverURL string `json:"coverUrl"`
	URL      string `json:"url"`
}

// Manga contains full manga details.
type Manga struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CoverURL    string   `json:"coverUrl"`
	Status      string   `json:"status"`
	Author      string   `json:"author"`
	Artist      string   `json:"artist"`
	Genres      []string `json:"genres"`
	Tags        []string `json:"tags"`
	URL         string   `json:"url"`
}

// Chapter represents a manga chapter.
type Chapter struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Number      float64   `json:"number"`
	Volume      string    `json:"volume"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"publishedAt"`
	PageCount   int       `json:"pageCount"`
}

// Page represents a single page in a chapter.
type Page struct {
	Index    int    `json:"index"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
}
