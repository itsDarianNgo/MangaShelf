package library

import "errors"

var (
	// ErrMangaExists is returned when trying to add a manga that already exists.
	ErrMangaExists = errors.New("manga already exists in library")

	// ErrMangaNotFound is returned when a manga is not found.
	ErrMangaNotFound = errors.New("manga not found")
)
