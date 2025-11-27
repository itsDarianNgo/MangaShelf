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
	endpoint := fmt.Sprintf("%s/manga/%s?includes[]=cover_art&includes[]=author&includes[]=artist", baseURL, url.PathEscape(id))

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

	if resp.StatusCode == http.StatusNotFound {
		return nil, scraper.ErrMangaNotFound
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, scraper.ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var response mangaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return m.convertManga(response.Data), nil
}

// GetChapters fetches all chapters for a manga.
func (m *MangaDex) GetChapters(ctx context.Context, mangaID string) ([]scraper.Chapter, error) {
	var allChapters []scraper.Chapter
	offset := 0
	limit := 100

	for {
		chapters, total, err := m.fetchChapterPage(ctx, mangaID, offset, limit)
		if err != nil {
			return nil, err
		}

		allChapters = append(allChapters, chapters...)
		offset += limit

		if offset >= total {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}

	return allChapters, nil
}

// GetPages fetches all page URLs for a chapter.
func (m *MangaDex) GetPages(ctx context.Context, chapterID string) ([]scraper.Page, error) {
	// TODO: Implement in next task
	return nil, fmt.Errorf("not implemented")
}

// fetchChapterPage fetches a single page of chapters.
func (m *MangaDex) fetchChapterPage(ctx context.Context, mangaID string, offset, limit int) ([]scraper.Chapter, int, error) {
	endpoint := fmt.Sprintf(
		"%s/manga/%s/feed?translatedLanguage[]=%s&order[chapter]=asc&limit=%d&offset=%d&includes[]=scanlation_group",
		baseURL, url.PathEscape(mangaID), m.language, limit, offset,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, 0, scraper.ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var response chapterFeedResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, 0, fmt.Errorf("decode response: %w", err)
	}

	chapters := m.convertChapters(response.Data)
	return chapters, response.Total, nil
}

// convertChapters transforms API chapter data to scraper.Chapter slice.
func (m *MangaDex) convertChapters(data []chapterData) []scraper.Chapter {
	chapters := make([]scraper.Chapter, 0, len(data))

	for _, ch := range data {
		number := parseChapterNumber(ch.Attributes.Chapter)

		title := ch.Attributes.Title
		if title == "" {
			title = fmt.Sprintf("Chapter %g", number)
		}

		var publishedAt time.Time
		if ch.Attributes.PublishAt != "" {
			publishedAt, _ = time.Parse(time.RFC3339, ch.Attributes.PublishAt)
		}

		chapters = append(chapters, scraper.Chapter{
			ID:          ch.ID,
			Title:       title,
			Number:      number,
			Volume:      ch.Attributes.Volume,
			URL:         fmt.Sprintf("https://mangadex.org/chapter/%s", ch.ID),
			PublishedAt: publishedAt,
			PageCount:   ch.Attributes.Pages,
		})
	}

	return chapters
}

// parseChapterNumber parses chapter number string to float64.
func parseChapterNumber(s string) float64 {
	if s == "" {
		return 0
	}
	var n float64
	fmt.Sscanf(s, "%f", &n)
	return n
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

// convertManga transforms API manga data to scraper.Manga.
func (m *MangaDex) convertManga(data mangaData) *scraper.Manga {
	title := m.getTitle(data.Attributes.Title)
	description := m.getDescription(data.Attributes.Description)
	coverURL := m.getCoverURL(data.ID, data.Relationships)
	author, artist := m.getStaff(data.Relationships)
	genres, tags := m.getTags(data.Attributes.Tags)

	return &scraper.Manga{
		ID:          data.ID,
		Title:       title,
		Description: description,
		CoverURL:    coverURL,
		Status:      data.Attributes.Status,
		Author:      author,
		Artist:      artist,
		Genres:      genres,
		Tags:        tags,
		URL:         fmt.Sprintf("https://mangadex.org/title/%s", data.ID),
	}
}

// getDescription extracts description based on language preference.
func (m *MangaDex) getDescription(descriptions map[string]string) string {
	if desc, ok := descriptions[m.language]; ok && desc != "" {
		return desc
	}

	if desc, ok := descriptions["en"]; ok && desc != "" {
		return desc
	}

	for _, desc := range descriptions {
		if desc != "" {
			return desc
		}
	}

	return ""
}

// getStaff extracts author and artist from relationships.
func (m *MangaDex) getStaff(relationships []relationship) (author, artist string) {
	for _, rel := range relationships {
		if rel.Attributes == nil {
			continue
		}

		name, _ := rel.Attributes["name"].(string)
		if name == "" {
			continue
		}

		switch rel.Type {
		case "author":
			if author == "" {
				author = name
			}
		case "artist":
			if artist == "" {
				artist = name
			}
		}
	}

	return author, artist
}

// getTags extracts genres and tags from tag data.
func (m *MangaDex) getTags(tagList []tagData) (genres, tags []string) {
	for _, tag := range tagList {
		name := ""
		if n, ok := tag.Attributes.Name["en"]; ok {
			name = n
		}

		if name == "" {
			continue
		}

		switch tag.Attributes.Group {
		case "genre":
			genres = append(genres, name)
		default:
			tags = append(tags, name)
		}
	}

	return genres, tags
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
