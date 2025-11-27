package mangadex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mangashelf/mangashelf/internal/scraper"
)

const (
	baseURL   = "https://api.mangadex.org"
	coversURL = "https://uploads.mangadex.org/covers"
	userAgent = "MangaShelf/1.0"
)

// MangaDex implements the scraper.Provider interface.
type MangaDex struct {
	client   *http.Client
	language string
}

// New creates a new MangaDex provider.
func New(language string) *MangaDex {
	return &MangaDex{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		language: language,
	}
}

// Info returns provider metadata.
func (m *MangaDex) Info() scraper.ProviderInfo {
	return scraper.ProviderInfo{
		ID:        "mangadex",
		Name:      "MangaDex",
		BaseURL:   "https://mangadex.org",
		Languages: []string{"en", "ja", "ko", "zh", "es", "fr", "de", "it", "pt-br", "ru"},
		IsNSFW:    false,
	}
}

// Search finds manga matching the query.
func (m *MangaDex) Search(ctx context.Context, query string) ([]scraper.MangaResult, error) {
	endpoint := fmt.Sprintf("%s/manga?title=%s&limit=20&includes[]=cover_art", baseURL, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, scraper.ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var response searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return m.convertSearchResults(response), nil
}

// GetManga fetches full details for a manga.
func (m *MangaDex) GetManga(ctx context.Context, id string) (*scraper.Manga, error) {
	// TODO: Implement in next task
	return nil, fmt.Errorf("not implemented")
}

// GetChapters fetches all chapters for a manga.
func (m *MangaDex) GetChapters(ctx context.Context, mangaID string) ([]scraper.Chapter, error) {
	// TODO: Implement in next task
	return nil, fmt.Errorf("not implemented")
}

// GetPages fetches all page URLs for a chapter.
func (m *MangaDex) GetPages(ctx context.Context, chapterID string) ([]scraper.Page, error) {
	// TODO: Implement in next task
	return nil, fmt.Errorf("not implemented")
}

// convertSearchResults transforms API response to scraper types.
func (m *MangaDex) convertSearchResults(resp searchResponse) []scraper.MangaResult {
	results := make([]scraper.MangaResult, 0, len(resp.Data))

	for _, manga := range resp.Data {
		title := m.getTitle(manga.Attributes.Title)
		coverURL := m.getCoverURL(manga.ID, manga.Relationships)

		results = append(results, scraper.MangaResult{
			ID:       manga.ID,
			Title:    title,
			CoverURL: coverURL,
			URL:      fmt.Sprintf("https://mangadex.org/title/%s", manga.ID),
		})
	}

	return results
}

// getTitle extracts the best title based on language preference.
func (m *MangaDex) getTitle(titles map[string]string) string {
	if title, ok := titles[m.language]; ok && title != "" {
		return title
	}

	if title, ok := titles["en"]; ok && title != "" {
		return title
	}

	if title, ok := titles["ja-ro"]; ok && title != "" {
		return title
	}

	for _, title := range titles {
		if title != "" {
			return title
		}
	}

	return "Unknown Title"
}

// getCoverURL extracts the cover image URL from relationships.
func (m *MangaDex) getCoverURL(mangaID string, relationships []relationship) string {
	for _, rel := range relationships {
		if rel.Type == "cover_art" && rel.Attributes != nil {
			if filename, ok := rel.Attributes["fileName"].(string); ok {
				return fmt.Sprintf("%s/%s/%s.256.jpg", coversURL, mangaID, filename)
			}
		}
	}

	return ""
}
