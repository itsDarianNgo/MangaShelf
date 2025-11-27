package scraper

import "errors"

var (
	// ErrProviderNotFound is returned when a provider ID doesn't exist.
	ErrProviderNotFound = errors.New("provider not found")

	// ErrMangaNotFound is returned when a manga ID doesn't exist on the source.
	ErrMangaNotFound = errors.New("manga not found")

	// ErrChapterNotFound is returned when a chapter ID doesn't exist.
	ErrChapterNotFound = errors.New("chapter not found")

	// ErrRateLimited is returned when the source rate limits requests.
	ErrRateLimited = errors.New("rate limited")

	// ErrSourceUnavailable is returned when the source is down.
	ErrSourceUnavailable = errors.New("source unavailable")
)
