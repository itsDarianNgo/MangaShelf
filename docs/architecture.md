# MangaShelf Architecture

This document provides a comprehensive technical overview of MangaShelf's architecture, design decisions, and implementation details.  It serves as the authoritative reference for developers contributing to the project.

## Table of Contents

1.  [System Overview](#system-overview)
2. [Design Principles](#design-principles)
3.  [Component Architecture](#component-architecture)
4. [Database Design](#database-design)
5. [API Design](#api-design)
6.  [Scraper System](#scraper-system)
7. [Download Engine](#download-engine)
8. [Reader System](#reader-system)
9. [Frontend Architecture](#frontend-architecture)
10. [Configuration System](#configuration-system)
11. [Error Handling](#error-handling)
12.  [Security Considerations](#security-considerations)
13.  [Performance Considerations](#performance-considerations)
14. [Deployment Architecture](#deployment-architecture)

---

## System Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        MangaShelf Binary                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐         │
│  │    Web UI      │  │   REST API     │  │   OPDS Feed    │         │
│  │   (Embedded)   │◄─┤   (Echo/Chi)   │◄─┤   Endpoint     │         │
│  └───────┬────────┘  └───────┬────────┘  └────────────────┘         │
│          │                   │                                       │
│          └─────────┬─────────┘                                       │
│                    │                                                 │
│  ┌─────────────────▼─────────────────────────────────────────────┐  │
│  │                      Service Layer                             │  │
│  ├────────────┬────────────┬────────────┬────────────┬───────────┤  │
│  │  Library   │  Download  │  Scraper   │  Reader    │ Scheduler │  │
│  │  Service   │  Service   │  Manager   │  Service   │  Service  │  │
│  └─────┬──────┴─────┬──────┴─────┬──────┴─────┬──────┴─────┬─────┘  │
│        │            │            │            │            │         │
│  ┌─────▼──────┐ ┌───▼────┐ ┌─────▼─────┐ ┌────▼─────┐ ┌────▼─────┐  │
│  │  SQLite    │ │ Worker │ │  Lua VM   │ │  Image   │ │  Cron    │  │
│  │  Database  │ │  Pool  │ │ (gopher)  │ │  Server  │ │  Jobs    │  │
│  └────────────┘ └────────┘ └───────────┘ └──────────┘ └──────────┘  │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌──────────────────────────────────────────────────────────────────────┐
│                           File System                                 │
│  /data/                                                               │
│  ├── mangashelf.db           (SQLite database)                       │
│  ├── config.yaml             (User configuration)                    │
│  ├── manga/                  (Downloaded manga library)              │
│  │   ├── One Piece/                                                  │
│  │   │   ├── cover.jpg                                               │
│  │   │   ├── series.json                                             │
│  │   │   ├── Chapter 0001.cbz                                        │
│  │   │   └── Chapter 0002.cbz                                        │
│  │   └── . ../                                                        │
│  ├── cache/                  (Temporary/cached files)                │
│  │   ├── covers/             (Cached cover thumbnails)               │
│  │   └── temp/               (In-progress downloads)                 │
│  └── scrapers/               (Custom Lua scrapers)                   │
│      ├── custom-source.lua                                           │
│      └── . ../                                                        │
└──────────────────────────────────────────────────────────────────────┘
```

### Request Flow

```
User Browser                MangaShelf                    External
     │                          │                            │
     │  GET /api/manga          │                            │
     │─────────────────────────►│                            │
     │                          │                            │
     │                    ┌─────┴─────┐                      │
     │                    │  Router   │                      │
     │                    └─────┬─────┘                      │
     │                          │                            │
     │                    ┌─────▼─────┐                      │
     │                    │ Middleware│ (logging, auth)      │
     │                    └─────┬─────┘                      │
     │                          │                            │
     │                    ┌─────▼─────┐                      │
     │                    │  Handler  │                      │
     │                    └─────┬─────┘                      │
     │                          │                            │
     │                    ┌─────▼─────┐                      │
     │                    │  Service  │                      │
     │                    └─────┬─────┘                      │
     │                          │                            │
     │                    ┌─────▼─────┐                      │
     │                    │ Database  │                      │
     │                    └─────┬─────┘                      │
     │                          │                            │
     │  JSON Response           │                            │
     │◄─────────────────────────│                            │
     │                          │                            │
```

---

## Design Principles

### 1.  Single Binary Distribution

Everything is embedded into one executable:

```go
//go:embed all:dist
var webAssets embed.FS

//go:embed schema.sql
var schemaSQL string
```

**Benefits:**
- No installation steps beyond downloading
- No runtime dependency management
- Trivial updates (replace binary)
- Works on any system without setup

### 2.  Zero External Dependencies

| Component | Traditional | MangaShelf |
|-----------|-------------|------------|
| Database | PostgreSQL/MySQL | Embedded SQLite |
| Cache/Queue | Redis | In-memory Go structures |
| Task Runner | Celery/Sidekiq | Goroutines + channels |
| Downloader | External binary | Native Go HTTP client |

### 3. Convention Over Configuration

Sensible defaults that work for 90% of users:

```go
var DefaultConfig = Config{
    Server: ServerConfig{
        Host: "0.0.0. 0",
        Port: 8080,
    },
    Library: LibraryConfig{
        Path: "./data/manga",
    },
    Downloader: DownloaderConfig{
        Workers:   3,
        Format:    "cbz",
        RateLimit: "2/s",
    },
}
```

### 4. Graceful Degradation

Features fail independently without crashing the application:

```go
// Metadata fetch failure doesn't prevent manga addition
func (s *LibraryService) AddManga(ctx context.Context, manga *Manga) error {
    if err := s.db.InsertManga(ctx, manga); err != nil {
        return fmt.Errorf("insert manga: %w", err)
    }
    
    // Metadata is best-effort
    go func() {
        if err := s.fetchMetadata(context.Background(), manga); err != nil {
            s.log. Warn(). Err(err). Str("manga", manga.Title). Msg("metadata fetch failed")
        }
    }()
    
    return nil
}
```

---

## Component Architecture

### Package Dependency Graph

```
cmd/mangashelf/main.go
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│                    internal/api                              │
│  ┌──────────┐  ┌──────────────┐  ┌────────────────────────┐ │
│  │  router  │──│  middleware  │──│       handlers         │ │
│  └──────────┘  └──────────────┘  └───────────┬────────────┘ │
└──────────────────────────────────────────────┼──────────────┘
                                               │
         ┌─────────────────┬───────────────────┼───────────────┬─────────────────┐
         │                 │                   │               │                 │
         ▼                 ▼                   ▼               ▼                 ▼
┌─────────────────┐ ┌─────────────┐ ┌─────────────────┐ ┌───────────┐ ┌─────────────────┐
│internal/library │ │internal/    │ │internal/scraper │ │internal/  │ │internal/        │
│                 │ │downloader   │ │                 │ │reader     │ │scheduler        │
│ - library. go    │ │             │ │ - manager.go    │ │           │ │                 │
│ - scanner.go    │ │ - queue.go  │ │ - provider.go   │ │ - cbz.go  │ │ - scheduler.go  │
│ - metadata. go   │ │ - worker.go │ │ - lua/runtime   │ │ - reader  │ │ - jobs.go       │
└────────┬────────┘ └──────┬──────┘ │ - builtin/*     │ └─────┬─────┘ └────────┬────────┘
         │                 │        └────────┬────────┘       │                │
         │                 │                 │                │                │
         └─────────────────┴─────────────────┼────────────────┴────────────────┘
                                             │
                                             ▼
                              ┌──────────────────────────┐
                              │    internal/database     │
                              │                          │
                              │  - db.go (sqlc generated)│
                              │  - schema.sql            │
                              │  - queries/*. sql         │
                              └─────────────┬────────────┘
                                            │
                                            ▼
                              ┌──────────────────────────┐
                              │    internal/config       │
                              │                          │
                              │  - config.go             │
                              │  - defaults.go           │
                              └──────────────────────────┘
```

### Service Layer Pattern

Each domain has a service that encapsulates business logic:

```go
// internal/library/library. go
type Service struct {
    db        *database. Queries
    log       zerolog.Logger
    scrapers  *scraper.Manager
    downloads *downloader.Queue
}

func NewService(db *database. Queries, scrapers *scraper. Manager, downloads *downloader.Queue) *Service {
    return &Service{
        db:        db,
        log:       log.With().Str("component", "library"). Logger(),
        scrapers:  scrapers,
        downloads: downloads,
    }
}

func (s *Service) AddManga(ctx context.Context, req AddMangaRequest) (*Manga, error) {
    // Validation
    if req. SourceID == "" {
        return nil, ErrMissingSourceID
    }
    
    // Business logic
    manga, err := s.scrapers.GetManga(ctx, req. Source, req.SourceID)
    if err != nil {
        return nil, fmt. Errorf("fetch manga: %w", err)
    }
    
    // Persistence
    dbManga, err := s.db.InsertManga(ctx, toDBManga(manga))
    if err != nil {
        return nil, fmt. Errorf("insert manga: %w", err)
    }
    
    return fromDBManga(dbManga), nil
}
```

---

## Database Design

### Schema

```sql
-- internal/database/schema. sql

-- Enable WAL mode for better concurrent performance
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-------------------------------------------------------------------------------
-- MANGA TABLE
-------------------------------------------------------------------------------
CREATE TABLE manga (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Identity
    title           TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    
    -- Source information
    source          TEXT NOT NULL,                -- e.g., "mangadex"
    source_id       TEXT NOT NULL,                -- ID on the source
    url             TEXT NOT NULL,                -- URL on the source
    
    -- Cover images
    cover_url       TEXT,                         -- Remote cover URL
    cover_path      TEXT,                         -- Local cover path
    
    -- Metadata
    description     TEXT,
    status          TEXT CHECK(status IN ('ongoing', 'completed', 'hiatus', 'cancelled', 'unknown')),
    author          TEXT,
    artist          TEXT,
    genres          TEXT,                         -- JSON array: ["Action", "Adventure"]
    tags            TEXT,                         -- JSON array: ["Shounen", "Supernatural"]
    
    -- External IDs
    anilist_id      INTEGER,
    mal_id          INTEGER,
    
    -- Settings
    update_interval TEXT DEFAULT '0 */6 * * *',   -- Cron expression
    auto_download   INTEGER DEFAULT 1,            -- Boolean: auto-download new chapters
    
    -- Timestamps
    created_at      TEXT DEFAULT (datetime('now')),
    updated_at      TEXT DEFAULT (datetime('now')),
    last_checked_at TEXT,
    
    -- Constraints
    UNIQUE(source, source_id)
);

CREATE INDEX idx_manga_slug ON manga(slug);
CREATE INDEX idx_manga_source ON manga(source);
CREATE INDEX idx_manga_last_checked ON manga(last_checked_at);

-------------------------------------------------------------------------------
-- CHAPTER TABLE
-------------------------------------------------------------------------------
CREATE TABLE chapter (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    manga_id        INTEGER NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    
    -- Chapter information
    title           TEXT NOT NULL,
    number          REAL NOT NULL,                -- Supports decimals: 10.5
    volume          TEXT,                         -- Optional volume: "Vol.  1"
    
    -- Source information
    source_id       TEXT NOT NULL,
    url             TEXT NOT NULL,
    
    -- Download status
    status          TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'queued', 'downloading', 'completed', 'failed')),
    file_path       TEXT,                         -- Path to CBZ/PDF file
    file_size       INTEGER,                      -- File size in bytes
    page_count      INTEGER,                      -- Number of pages
    
    -- Reading progress
    is_read         INTEGER DEFAULT 0,            -- Boolean
    current_page    INTEGER DEFAULT 0,            -- Last read page
    read_at         TEXT,                         -- When marked as read
    
    -- Timestamps
    published_at    TEXT,                         -- When published on source
    downloaded_at   TEXT,
    created_at      TEXT DEFAULT (datetime('now')),
    
    -- Constraints
    UNIQUE(manga_id, source_id)
);

CREATE INDEX idx_chapter_manga_id ON chapter(manga_id);
CREATE INDEX idx_chapter_status ON chapter(status);
CREATE INDEX idx_chapter_number ON chapter(manga_id, number);

-------------------------------------------------------------------------------
-- DOWNLOAD QUEUE TABLE
-------------------------------------------------------------------------------
CREATE TABLE download_queue (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    chapter_id      INTEGER NOT NULL REFERENCES chapter(id) ON DELETE CASCADE,
    
    -- Queue status
    priority        INTEGER DEFAULT 0,            -- Higher = more urgent
    attempts        INTEGER DEFAULT 0,
    max_attempts    INTEGER DEFAULT 3,
    last_error      TEXT,
    
    status          TEXT DEFAULT 'queued' CHECK(status IN ('queued', 'downloading', 'completed', 'failed', 'cancelled')),
    
    -- Timestamps
    created_at      TEXT DEFAULT (datetime('now')),
    started_at      TEXT,
    completed_at    TEXT,
    
    UNIQUE(chapter_id)
);

CREATE INDEX idx_download_queue_status ON download_queue(status);
CREATE INDEX idx_download_queue_priority ON download_queue(priority DESC, created_at ASC);

-------------------------------------------------------------------------------
-- SETTINGS TABLE
-------------------------------------------------------------------------------
CREATE TABLE settings (
    key             TEXT PRIMARY KEY,
    value           TEXT NOT NULL,
    updated_at      TEXT DEFAULT (datetime('now'))
);

-------------------------------------------------------------------------------
-- SCRAPER TABLE
-------------------------------------------------------------------------------
CREATE TABLE scraper (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    
    name            TEXT NOT NULL UNIQUE,
    type            TEXT NOT NULL CHECK(type IN ('builtin', 'lua')),
    path            TEXT,                         -- File path for Lua scrapers
    enabled         INTEGER DEFAULT 1,
    config          TEXT,                         -- JSON configuration
    
    created_at      TEXT DEFAULT (datetime('now')),
    updated_at      TEXT DEFAULT (datetime('now'))
);

-------------------------------------------------------------------------------
-- TRIGGERS FOR UPDATED_AT
-------------------------------------------------------------------------------
CREATE TRIGGER update_manga_timestamp 
AFTER UPDATE ON manga
BEGIN
    UPDATE manga SET updated_at = datetime('now') WHERE id = NEW.id;
END;

CREATE TRIGGER update_scraper_timestamp 
AFTER UPDATE ON scraper
BEGIN
    UPDATE scraper SET updated_at = datetime('now') WHERE id = NEW.id;
END;
```

### sqlc Configuration

```yaml
# sqlc. yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "internal/database/queries"
    schema: "internal/database/schema. sql"
    gen:
      go:
        package: "database"
        out: "internal/database"
        emit_json_tags: true
        emit_empty_slices: true
        emit_result_struct_pointers: true
```

### Example Queries

```sql
-- internal/database/queries/manga.sql

-- name: GetManga :one
SELECT * FROM manga WHERE id = ?  LIMIT 1;

-- name: GetMangaBySlug :one
SELECT * FROM manga WHERE slug = ? LIMIT 1;

-- name: ListManga :many
SELECT * FROM manga ORDER BY title ASC;

-- name: ListMangaWithUnread :many
SELECT 
    m.*,
    COUNT(c.id) FILTER (WHERE c.is_read = 0 AND c.status = 'completed') as unread_count
FROM manga m
LEFT JOIN chapter c ON c.manga_id = m.id
GROUP BY m.id
ORDER BY m.title ASC;

-- name: InsertManga :one
INSERT INTO manga (
    title, slug, source, source_id, url, cover_url, description,
    status, author, artist, genres, tags
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateManga :one
UPDATE manga SET
    title = ?,
    cover_url = ?,
    cover_path = ?,
    description = ?,
    status = ?,
    author = ?,
    artist = ?,
    genres = ?,
    tags = ?,
    anilist_id = ?,
    last_checked_at = datetime('now')
WHERE id = ? 
RETURNING *;

-- name: DeleteManga :exec
DELETE FROM manga WHERE id = ?;

-- name: GetMangaForUpdate :many
SELECT * FROM manga 
WHERE auto_download = 1 
  AND (last_checked_at IS NULL OR last_checked_at < datetime('now', '-1 hour'))
ORDER BY last_checked_at ASC NULLS FIRST
LIMIT ? ;
```

```sql
-- internal/database/queries/chapter.sql

-- name: GetChapter :one
SELECT * FROM chapter WHERE id = ? LIMIT 1;

-- name: ListChaptersByManga :many
SELECT * FROM chapter WHERE manga_id = ?  ORDER BY number DESC;

-- name: InsertChapter :one
INSERT INTO chapter (
    manga_id, title, number, volume, source_id, url, published_at
) VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (manga_id, source_id) DO UPDATE SET
    title = excluded.title,
    number = excluded.number,
    volume = excluded.volume,
    url = excluded.url
RETURNING *;

-- name: UpdateChapterStatus :one
UPDATE chapter SET
    status = ?,
    file_path = ?,
    file_size = ?,
    page_count = ?,
    downloaded_at = CASE WHEN ?  = 'completed' THEN datetime('now') ELSE downloaded_at END
WHERE id = ? 
RETURNING *;

-- name: MarkChapterRead :exec
UPDATE chapter SET is_read = 1, read_at = datetime('now') WHERE id = ?;

-- name: UpdateReadingProgress :exec
UPDATE chapter SET current_page = ?  WHERE id = ?;
```

---

## API Design

### RESTful Conventions

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/resource | List all resources |
| POST | /api/resource | Create a resource |
| GET | /api/resource/:id | Get a specific resource |
| PATCH | /api/resource/:id | Update a resource |
| DELETE | /api/resource/:id | Delete a resource |
| POST | /api/resource/:id/action | Perform an action |

### Response Format

**Success Response:**
```json
{
    "data": { ...  },
    "meta": {
        "page": 1,
        "perPage": 20,
        "total": 150
    }
}
```

**Error Response:**
```json
{
    "error": {
        "code": "MANGA_NOT_FOUND",
        "message": "The requested manga was not found",
        "details": {
            "id": 123
        }
    }
}
```

### Handler Structure

```go
// internal/api/handler/manga.go
package handler

type MangaHandler struct {
    library *library.Service
    log     zerolog.Logger
}

func NewMangaHandler(library *library.Service) *MangaHandler {
    return &MangaHandler{
        library: library,
        log:     log.With(). Str("handler", "manga").Logger(),
    }
}

// GET /api/manga
func (h *MangaHandler) List(c echo.Context) error {
    ctx := c.Request(). Context()
    
    manga, err := h. library.ListManga(ctx)
    if err != nil {
        h.log.Error(). Err(err). Msg("failed to list manga")
        return echo.NewHTTPError(http.StatusInternalServerError, "failed to list manga")
    }
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "data": manga,
    })
}

// POST /api/manga
func (h *MangaHandler) Create(c echo. Context) error {
    ctx := c. Request().Context()
    
    var req CreateMangaRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
    }
    
    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    
    manga, err := h.library.AddManga(ctx, req)
    if err != nil {
        if errors.Is(err, library.ErrMangaExists) {
            return echo.NewHTTPError(http.StatusConflict, "manga already exists")
        }
        h.log.Error(). Err(err). Msg("failed to create manga")
        return echo.NewHTTPError(http. StatusInternalServerError, "failed to create manga")
    }
    
    return c.JSON(http.StatusCreated, map[string]interface{}{
        "data": manga,
    })
}

// Request/Response types
type CreateMangaRequest struct {
    Source   string `json:"source" validate:"required"`
    SourceID string `json:"sourceId" validate:"required"`
}
```

### Router Setup

```go
// internal/api/router. go
package api

func NewRouter(
    library *library. Service,
    downloads *downloader.Queue,
    reader *reader.Service,
) *echo.Echo {
    e := echo.New()
    
    // Middleware
    e.Use(middleware. Logger())
    e.Use(middleware. Recover())
    e.Use(middleware. CORS())
    e.Use(middleware. RequestID())
    
    // Health check
    e. GET("/api/health", func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
    })
    
    // API routes
    api := e.Group("/api")
    
    // Manga
    mangaHandler := handler.NewMangaHandler(library)
    api.GET("/manga", mangaHandler.List)
    api.POST("/manga", mangaHandler.Create)
    api.GET("/manga/:id", mangaHandler.Get)
    api.PATCH("/manga/:id", mangaHandler.Update)
    api.DELETE("/manga/:id", mangaHandler.Delete)
    api.POST("/manga/:id/refresh", mangaHandler. Refresh)
    
    // Chapters
    chapterHandler := handler.NewChapterHandler(library, downloads)
    api. GET("/manga/:id/chapters", chapterHandler.List)
    api.GET("/chapters/:id", chapterHandler.Get)
    api.POST("/chapters/:id/download", chapterHandler. Download)
    api. PATCH("/chapters/:id", chapterHandler.Update)
    
    // Reader
    readerHandler := handler.NewReaderHandler(reader)
    api.GET("/read/:chapterId", readerHandler.GetChapter)
    api.GET("/read/:chapterId/page/:page", readerHandler. GetPage)
    api. PATCH("/read/:chapterId/progress", readerHandler. UpdateProgress)
    
    // Search
    searchHandler := handler.NewSearchHandler(library)
    api.GET("/search", searchHandler.Search)
    
    // Sources
    sourceHandler := handler.NewSourceHandler(library)
    api.GET("/sources", sourceHandler.List)
    
    // Downloads
    downloadHandler := handler.NewDownloadHandler(downloads)
    api. GET("/downloads", downloadHandler.List)
    api.DELETE("/downloads/:id", downloadHandler.Cancel)
    
    // Settings
    settingsHandler := handler.NewSettingsHandler()
    api.GET("/settings", settingsHandler.Get)
    api. PATCH("/settings", settingsHandler.Update)
    
    // Static files (embedded frontend)
    e.GET("/*", echo.WrapHandler(http.FileServer(http. FS(webAssets))))
    
    return e
}
```

---

## Scraper System

### Provider Interface

```go
// internal/scraper/provider. go
package scraper

import (
    "context"
    "time"
)

// Provider defines the interface all manga sources must implement
type Provider interface {
    // Info returns metadata about this provider
    Info() ProviderInfo
    
    // Search finds manga matching the query
    Search(ctx context.Context, query string) ([]MangaResult, error)
    
    // GetManga fetches full details for a manga
    GetManga(ctx context.Context, id string) (*Manga, error)
    
    // GetChapters fetches all chapters for a manga
    GetChapters(ctx context.Context, mangaID string) ([]Chapter, error)
    
    // GetPages fetches all page URLs for a chapter
    GetPages(ctx context.Context, chapterID string) ([]Page, error)
}

type ProviderInfo struct {
    ID           string   `json:"id"`
    Name         string   `json:"name"`
    BaseURL      string   `json:"baseUrl"`
    Languages    []string `json:"languages"`
    IsNSFW       bool     `json:"isNsfw"`
    SupportsLatest bool   `json:"supportsLatest"`
}

type MangaResult struct {
    ID       string `json:"id"`
    Title    string `json:"title"`
    CoverURL string `json:"coverUrl"`
    URL      string `json:"url"`
}

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

type Chapter struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Number      float64   `json:"number"`
    Volume      string    `json:"volume"`
    URL         string    `json:"url"`
    PublishedAt time.Time `json:"publishedAt"`
}

type Page struct {
    Index    int    `json:"index"`
    URL      string `json:"url"`
    Filename string `json:"filename"`
}
```

### Built-in Provider Example (MangaDex)

```go
// internal/scraper/builtin/mangadex.go
package builtin

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "time"
    
    "github. com/username/mangashelf/internal/scraper"
)

const (
    mangadexBaseURL = "https://api.mangadex. org"
    mangadexCDN     = "https://uploads.mangadex. org"
)

type MangaDex struct {
    client   *http.Client
    language string
}

func NewMangaDex(language string) *MangaDex {
    return &MangaDex{
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
        language: language,
    }
}

func (m *MangaDex) Info() scraper.ProviderInfo {
    return scraper.ProviderInfo{
        ID:             "mangadex",
        Name:           "MangaDex",
        BaseURL:        "https://mangadex.org",
        Languages:      []string{"en", "ja", "ko", "zh", /* ... */},
        IsNSFW:         false,
        SupportsLatest: true,
    }
}

func (m *MangaDex) Search(ctx context.Context, query string) ([]scraper.MangaResult, error) {
    endpoint := fmt.Sprintf("%s/manga? title=%s&limit=20&includes[]=cover_art",
        mangadexBaseURL, url. QueryEscape(query))
    
    req, err := http. NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    
    resp, err := m. client.Do(req)
    if err != nil {
        return nil, fmt. Errorf("execute request: %w", err)
    }
    defer resp. Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status: %d", resp. StatusCode)
    }
    
    var result mangadexSearchResponse
    if err := json.NewDecoder(resp.Body). Decode(&result); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }
    
    return m.convertSearchResults(result), nil
}

func (m *MangaDex) GetManga(ctx context. Context, id string) (*scraper. Manga, error) {
    // Implementation... 
}

func (m *MangaDex) GetChapters(ctx context. Context, mangaID string) ([]scraper.Chapter, error) {
    // Implementation with pagination handling... 
}

func (m *MangaDex) GetPages(ctx context.Context, chapterID string) ([]scraper. Page, error) {
    // Implementation using MangaDex@Home API...
}

// MangaDex API response types
type mangadexSearchResponse struct {
    Result   string `json:"result"`
    Response string `json:"response"`
    Data     []struct {
        ID         string `json:"id"`
        Type       string `json:"type"`
        Attributes struct {
            Title       map[string]string `json:"title"`
            Description map[string]string `json:"description"`
            Status      string            `json:"status"`
        } `json:"attributes"`
        Relationships []struct {
            ID         string `json:"id"`
            Type       string `json:"type"`
            Attributes struct {
                FileName string `json:"fileName"`
            } `json:"attributes,omitempty"`
        } `json:"relationships"`
    } `json:"data"`
}
```

### Scraper Manager

```go
// internal/scraper/manager.go
package scraper

import (
    "context"
    "fmt"
    "sync"
)

type Manager struct {
    providers map[string]Provider
    mu        sync.RWMutex
    luaVM     *LuaRuntime
}

func NewManager() *Manager {
    return &Manager{
        providers: make(map[string]Provider),
        luaVM:     NewLuaRuntime(),
    }
}

func (m *Manager) Register(provider Provider) {
    m.mu.Lock()
    defer m.mu. Unlock()
    m.providers[provider. Info().ID] = provider
}

func (m *Manager) Get(id string) (Provider, error) {
    m.mu.RLock()
    defer m.mu. RUnlock()
    
    provider, ok := m.providers[id]
    if !ok {
        return nil, fmt. Errorf("provider not found: %s", id)
    }
    return provider, nil
}

func (m *Manager) List() []ProviderInfo {
    m.mu. RLock()
    defer m.mu.RUnlock()
    
    infos := make([]ProviderInfo, 0, len(m.providers))
    for _, p := range m.providers {
        infos = append(infos, p.Info())
    }
    return infos
}

func (m *Manager) LoadLuaScrapers(dir string) error {
    // Load . lua files from directory and register as providers
    // Implementation uses gopher-lua
}

func (m *Manager) Search(ctx context.Context, source, query string) ([]MangaResult, error) {
    provider, err := m. Get(source)
    if err != nil {
        return nil, err
    }
    return provider.Search(ctx, query)
}

func (m *Manager) GetManga(ctx context.Context, source, id string) (*Manga, error) {
    provider, err := m.Get(source)
    if err != nil {
        return nil, err
    }
    return provider.GetManga(ctx, id)
}
```

---

## Download Engine

### Queue Design

```go
// internal/downloader/queue.go
package downloader

import (
    "context"
    "sync"
    "time"
)

type Job struct {
    ID        int64
    ChapterID int64
    MangaID   int64
    Priority  int
    Attempts  int
    Status    JobStatus
    Error     error
    CreatedAt time.Time
    StartedAt time.Time
}

type JobStatus string

const (
    StatusQueued      JobStatus = "queued"
    StatusDownloading JobStatus = "downloading"
    StatusCompleted   JobStatus = "completed"
    StatusFailed      JobStatus = "failed"
    StatusCancelled   JobStatus = "cancelled"
)

type Queue struct {
    jobs       chan *Job
    workers    int
    wg         sync.WaitGroup
    
    mu         sync.RWMutex
    active     map[int64]*Job   // Currently downloading
    pending    []*Job           // Waiting in queue
    
    ctx        context.Context
    cancel     context. CancelFunc
    
    db         *database. Queries
    scrapers   *scraper.Manager
    library    string           // Library path
    
    onProgress func(job *Job, progress Progress)
    onComplete func(job *Job)
    onError    func(job *Job, err error)
}

type Progress struct {
    CurrentPage int
    TotalPages  int
    BytesDown   int64
    Speed       float64 // bytes per second
}

func NewQueue(db *database. Queries, scrapers *scraper.Manager, workers int, libraryPath string) *Queue {
    ctx, cancel := context. WithCancel(context.Background())
    
    return &Queue{
        jobs:     make(chan *Job, 1000),
        workers:  workers,
        active:   make(map[int64]*Job),
        ctx:      ctx,
        cancel:   cancel,
        db:       db,
        scrapers: scrapers,
        library:  libraryPath,
    }
}

func (q *Queue) Start() {
    for i := 0; i < q.workers; i++ {
        q.wg.Add(1)
        go q. worker(i)
    }
}

func (q *Queue) Stop() {
    q.cancel()
    close(q.jobs)
    q.wg.Wait()
}

func (q *Queue) Enqueue(ctx context.Context, chapterID int64, priority int) error {
    // Create job in database
    dbJob, err := q.db.CreateDownloadJob(ctx, database.CreateDownloadJobParams{
        ChapterID: chapterID,
        Priority:  int64(priority),
    })
    if err != nil {
        return fmt.Errorf("create job: %w", err)
    }
    
    job := &Job{
        ID:        dbJob.ID,
        ChapterID: chapterID,
        Priority:  priority,
        Status:    StatusQueued,
        CreatedAt: time.Now(),
    }
    
    select {
    case q.jobs <- job:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (q *Queue) Status() QueueStatus {
    q.mu.RLock()
    defer q.mu. RUnlock()
    
    return QueueStatus{
        Active:    len(q.active),
        Pending:   len(q.pending),
        Workers:   q.workers,
    }
}

type QueueStatus struct {
    Active  int `json:"active"`
    Pending int `json:"pending"`
    Workers int `json:"workers"`
}
```

### Worker Implementation

```go
// internal/downloader/worker.go
package downloader

import (
    "archive/zip"
    "context"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"
)

func (q *Queue) worker(id int) {
    defer q.wg.Done()
    
    log := log.With().Int("worker", id).Logger()
    log.Info(). Msg("download worker started")
    
    for job := range q.jobs {
        select {
        case <-q.ctx. Done():
            return
        default:
        }
        
        q.processJob(job)
    }
}

func (q *Queue) processJob(job *Job) {
    log := log. With().Int64("job", job.ID).Int64("chapter", job.ChapterID).Logger()
    
    // Mark as downloading
    q.mu.Lock()
    q.active[job.ID] = job
    job.Status = StatusDownloading
    job.StartedAt = time.Now()
    q.mu. Unlock()
    
    defer func() {
        q.mu.Lock()
        delete(q.active, job.ID)
        q.mu. Unlock()
    }()
    
    // Get chapter info
    chapter, err := q. db.GetChapter(q.ctx, job.ChapterID)
    if err != nil {
        q.failJob(job, fmt.Errorf("get chapter: %w", err))
        return
    }
    
    manga, err := q. db.GetManga(q.ctx, chapter.MangaID)
    if err != nil {
        q.failJob(job, fmt. Errorf("get manga: %w", err))
        return
    }
    
    // Get page URLs
    provider, err := q. scrapers.Get(manga.Source)
    if err != nil {
        q.failJob(job, fmt. Errorf("get provider: %w", err))
        return
    }
    
    pages, err := provider.GetPages(q.ctx, chapter. SourceID)
    if err != nil {
        q.failJob(job, fmt.Errorf("get pages: %w", err))
        return
    }
    
    // Create CBZ
    cbzPath, err := q.downloadToCBZ(job, manga, chapter, pages)
    if err != nil {
        q.failJob(job, err)
        return
    }
    
    // Update database
    _, err = q.db.UpdateChapterStatus(q.ctx, database.UpdateChapterStatusParams{
        ID:        chapter.ID,
        Status:    "completed",
        FilePath:  sql.NullString{String: cbzPath, Valid: true},
        FileSize:  sql. NullInt64{Int64: getFileSize(cbzPath), Valid: true},
        PageCount: sql.NullInt64{Int64: int64(len(pages)), Valid: true},
    })
    if err != nil {
        log.Error(). Err(err). Msg("failed to update chapter status")
    }
    
    job.Status = StatusCompleted
    if q.onComplete != nil {
        q.onComplete(job)
    }
}

func (q *Queue) downloadToCBZ(job *Job, manga *database.Manga, chapter *database.Chapter, pages []scraper.Page) (string, error) {
    // Create manga directory
    mangaDir := filepath.Join(q.library, sanitizeFilename(manga.Title))
    if err := os.MkdirAll(mangaDir, 0755); err != nil {
        return "", fmt.Errorf("create manga dir: %w", err)
    }
    
    // Create CBZ file
    cbzName := fmt.Sprintf("Chapter %04. 1f.cbz", chapter.Number)
    cbzPath := filepath.Join(mangaDir, cbzName)
    
    cbzFile, err := os.Create(cbzPath)
    if err != nil {
        return "", fmt.Errorf("create cbz file: %w", err)
    }
    defer cbzFile.Close()
    
    zipWriter := zip.NewWriter(cbzFile)
    defer zipWriter.Close()
    
    // Download each page
    client := &http.Client{Timeout: 30 * time. Second}
    
    for i, page := range pages {
        if err := q.downloadPage(client, zipWriter, page, chapter. Url); err != nil {
            return "", fmt. Errorf("download page %d: %w", i+1, err)
        }
        
        if q.onProgress != nil {
            q. onProgress(job, Progress{
                CurrentPage: i + 1,
                TotalPages:  len(pages),
            })
        }
    }
    
    return cbzPath, nil
}

func (q *Queue) downloadPage(client *http.Client, zw *zip.Writer, page scraper. Page, referer string) error {
    req, err := http.NewRequest(http.MethodGet, page.URL, nil)
    if err != nil {
        return err
    }
    req.Header.Set("Referer", referer)
    req.Header.Set("User-Agent", "MangaShelf/1.0")
    
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp. StatusCode != http. StatusOK {
        return fmt. Errorf("HTTP %d", resp.StatusCode)
    }
    
    filename := fmt.Sprintf("%03d%s", page.Index, filepath.Ext(page.Filename))
    w, err := zw.Create(filename)
    if err != nil {
        return err
    }
    
    _, err = io.Copy(w, resp. Body)
    return err
}

func (q *Queue) failJob(job *Job, err error) {
    job.Status = StatusFailed
    job.Error = err
    job. Attempts++
    
    // Retry logic
    if job. Attempts < 3 {
        time.AfterFunc(time. Duration(job. Attempts)*time. Minute, func() {
            q.jobs <- job
        })
    } else if q.onError != nil {
        q. onError(job, err)
    }
}
```

---

## Reader System

### CBZ Handling

```go
// internal/reader/cbz. go
package reader

import (
    "archive/zip"
    "fmt"
    "io"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
)

// CBZReader handles reading pages from CBZ (Comic Book ZIP) archives
type CBZReader struct {
    path   string
    reader *zip.ReadCloser
    pages  []PageInfo
}

type PageInfo struct {
    Index    int
    Filename string
    Size     int64
}

// OpenCBZ opens a CBZ file and indexes its pages
func OpenCBZ(path string) (*CBZReader, error) {
    r, err := zip. OpenReader(path)
    if err != nil {
        return nil, fmt. Errorf("open cbz: %w", err)
    }
    
    cbz := &CBZReader{
        path:   path,
        reader: r,
        pages:  make([]PageInfo, 0),
    }
    
    // Index image files
    imageExts := map[string]bool{
        ".jpg": true, ".jpeg": true, ".png": true,
        ".gif": true, ".webp": true, ". avif": true,
    }
    
    for _, file := range r.File {
        ext := strings.ToLower(filepath.Ext(file.Name))
        if ! imageExts[ext] {
            continue
        }
        
        // Skip macOS metadata files
        if strings.HasPrefix(file.Name, "__MACOSX") || strings.HasPrefix(file.Name, ". ") {
            continue
        }
        
        cbz.pages = append(cbz.pages, PageInfo{
            Filename: file.Name,
            Size:     int64(file.UncompressedSize64),
        })
    }
    
    // Sort pages by filename (assumes zero-padded numbers)
    sort. Slice(cbz.pages, func(i, j int) bool {
        return naturalSort(cbz. pages[i]. Filename, cbz. pages[j]. Filename)
    })
    
    // Assign indices after sorting
    for i := range cbz.pages {
        cbz. pages[i].Index = i + 1
    }
    
    return cbz, nil
}

// PageCount returns the number of pages in the CBZ
func (c *CBZReader) PageCount() int {
    return len(c.pages)
}

// Pages returns metadata for all pages
func (c *CBZReader) Pages() []PageInfo {
    return c.pages
}

// GetPage returns a reader for a specific page (1-indexed)
func (c *CBZReader) GetPage(pageNum int) (io.ReadCloser, string, error) {
    if pageNum < 1 || pageNum > len(c.pages) {
        return nil, "", fmt. Errorf("page %d out of range (1-%d)", pageNum, len(c.pages))
    }
    
    page := c.pages[pageNum-1]
    
    for _, file := range c.reader.File {
        if file.Name == page.Filename {
            rc, err := file. Open()
            if err != nil {
                return nil, "", fmt.Errorf("open page: %w", err)
            }
            
            contentType := getContentType(page.Filename)
            return rc, contentType, nil
        }
    }
    
    return nil, "", fmt.Errorf("page file not found: %s", page.Filename)
}

// Close releases resources
func (c *CBZReader) Close() error {
    if c.reader != nil {
        return c.reader. Close()
    }
    return nil
}

// naturalSort compares strings with natural number ordering
// e.g., "page2. jpg" < "page10.jpg"
func naturalSort(a, b string) bool {
    aNum := extractNumber(a)
    bNum := extractNumber(b)
    
    if aNum != -1 && bNum != -1 {
        return aNum < bNum
    }
    return a < b
}

func extractNumber(s string) int {
    base := filepath.Base(s)
    name := strings.TrimSuffix(base, filepath.Ext(base))
    
    // Extract digits from the filename
    var numStr strings.Builder
    for _, r := range name {
        if r >= '0' && r <= '9' {
            numStr. WriteRune(r)
        }
    }
    
    if numStr.Len() == 0 {
        return -1
    }
    
    num, err := strconv.Atoi(numStr. String())
    if err != nil {
        return -1
    }
    return num
}

func getContentType(filename string) string {
    ext := strings.ToLower(filepath. Ext(filename))
    switch ext {
    case ".jpg", ".jpeg":
        return "image/jpeg"
    case ".png":
        return "image/png"
    case ".gif":
        return "image/gif"
    case ".webp":
        return "image/webp"
    case ".avif":
        return "image/avif"
    default:
        return "application/octet-stream"
    }
}
```

### Reader Service

```go
// internal/reader/reader.go
package reader

import (
    "context"
    "fmt"
    "io"
    "sync"
    
    "github.com/username/mangashelf/internal/database"
)

// Service handles reading operations
type Service struct {
    db    *database. Queries
    cache *CBZCache
}

// CBZCache maintains a pool of open CBZ readers for performance
type CBZCache struct {
    mu      sync.RWMutex
    readers map[int64]*cachedReader
    maxSize int
}

type cachedReader struct {
    reader   *CBZReader
    lastUsed time.Time
}

func NewService(db *database. Queries) *Service {
    s := &Service{
        db: db,
        cache: &CBZCache{
            readers: make(map[int64]*cachedReader),
            maxSize: 10, // Keep up to 10 CBZ files open
        },
    }
    
    // Start cache cleanup goroutine
    go s.cache.cleanup()
    
    return s
}

// GetChapterForReading returns chapter info and page count for the reader
func (s *Service) GetChapterForReading(ctx context.Context, chapterID int64) (*ReadingSession, error) {
    chapter, err := s. db.GetChapter(ctx, chapterID)
    if err != nil {
        return nil, fmt.Errorf("get chapter: %w", err)
    }
    
    if chapter. Status != "completed" || ! chapter.FilePath.Valid {
        return nil, ErrChapterNotDownloaded
    }
    
    cbz, err := s. cache.Get(chapterID, chapter.FilePath. String)
    if err != nil {
        return nil, fmt. Errorf("open cbz: %w", err)
    }
    
    manga, err := s. db.GetManga(ctx, chapter. MangaID)
    if err != nil {
        return nil, fmt.Errorf("get manga: %w", err)
    }
    
    // Get adjacent chapters for navigation
    prevChapter, _ := s.db. GetPreviousChapter(ctx, database.GetPreviousChapterParams{
        MangaID: chapter.MangaID,
        Number:  chapter.Number,
    })
    nextChapter, _ := s.db.GetNextChapter(ctx, database.GetNextChapterParams{
        MangaID: chapter.MangaID,
        Number:  chapter. Number,
    })
    
    return &ReadingSession{
        Chapter: ChapterInfo{
            ID:          chapter.ID,
            Title:       chapter.Title,
            Number:      chapter.Number,
            Volume:      chapter.Volume. String,
            CurrentPage: int(chapter.CurrentPage),
            PageCount:   cbz.PageCount(),
        },
        Manga: MangaInfo{
            ID:    manga.ID,
            Title: manga. Title,
        },
        PreviousChapter: toOptionalChapter(prevChapter),
        NextChapter:     toOptionalChapter(nextChapter),
        Pages:           cbz.Pages(),
    }, nil
}

// GetPage returns a specific page image
func (s *Service) GetPage(ctx context.Context, chapterID int64, pageNum int) (io.ReadCloser, string, error) {
    chapter, err := s. db.GetChapter(ctx, chapterID)
    if err != nil {
        return nil, "", fmt.Errorf("get chapter: %w", err)
    }
    
    if ! chapter.FilePath. Valid {
        return nil, "", ErrChapterNotDownloaded
    }
    
    cbz, err := s. cache.Get(chapterID, chapter. FilePath.String)
    if err != nil {
        return nil, "", fmt. Errorf("open cbz: %w", err)
    }
    
    return cbz.GetPage(pageNum)
}

// UpdateProgress saves the current reading position
func (s *Service) UpdateProgress(ctx context.Context, chapterID int64, page int) error {
    err := s.db. UpdateReadingProgress(ctx, database.UpdateReadingProgressParams{
        ID:          chapterID,
        CurrentPage: int64(page),
    })
    if err != nil {
        return fmt.Errorf("update progress: %w", err)
    }
    
    return nil
}

// MarkAsRead marks a chapter as read
func (s *Service) MarkAsRead(ctx context.Context, chapterID int64) error {
    return s.db. MarkChapterRead(ctx, chapterID)
}

// Response types
type ReadingSession struct {
    Chapter         ChapterInfo       `json:"chapter"`
    Manga           MangaInfo         `json:"manga"`
    PreviousChapter *ChapterNavInfo   `json:"previousChapter,omitempty"`
    NextChapter     *ChapterNavInfo   `json:"nextChapter,omitempty"`
    Pages           []PageInfo        `json:"pages"`
}

type ChapterInfo struct {
    ID          int64   `json:"id"`
    Title       string  `json:"title"`
    Number      float64 `json:"number"`
    Volume      string  `json:"volume,omitempty"`
    CurrentPage int     `json:"currentPage"`
    PageCount   int     `json:"pageCount"`
}

type MangaInfo struct {
    ID    int64  `json:"id"`
    Title string `json:"title"`
}

type ChapterNavInfo struct {
    ID     int64   `json:"id"`
    Number float64 `json:"number"`
    Title  string  `json:"title"`
}

// Errors
var (
    ErrChapterNotDownloaded = fmt.Errorf("chapter not downloaded")
    ErrPageNotFound         = fmt.Errorf("page not found")
)
```

### CBZ Cache Management

```go
// internal/reader/cache.go
package reader

import (
    "sync"
    "time"
)

// Get returns a cached CBZ reader or opens a new one
func (c *CBZCache) Get(chapterID int64, path string) (*CBZReader, error) {
    c.mu.RLock()
    if cached, ok := c.readers[chapterID]; ok {
        cached.lastUsed = time.Now()
        c.mu. RUnlock()
        return cached.reader, nil
    }
    c.mu.RUnlock()
    
    // Open new reader
    reader, err := OpenCBZ(path)
    if err != nil {
        return nil, err
    }
    
    c.mu.Lock()
    defer c. mu.Unlock()
    
    // Evict old entries if at capacity
    if len(c.readers) >= c.maxSize {
        c.evictOldest()
    }
    
    c.readers[chapterID] = &cachedReader{
        reader:   reader,
        lastUsed: time.Now(),
    }
    
    return reader, nil
}

// evictOldest removes the least recently used entry
func (c *CBZCache) evictOldest() {
    var oldestID int64
    var oldestTime time. Time
    
    for id, cached := range c.readers {
        if oldestTime.IsZero() || cached.lastUsed. Before(oldestTime) {
            oldestID = id
            oldestTime = cached.lastUsed
        }
    }
    
    if oldestID != 0 {
        c.readers[oldestID]. reader.Close()
        delete(c. readers, oldestID)
    }
}

// cleanup periodically removes stale cache entries
func (c *CBZCache) cleanup() {
    ticker := time.NewTicker(5 * time. Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        c.mu. Lock()
        threshold := time.Now(). Add(-10 * time.Minute)
        
        for id, cached := range c.readers {
            if cached.lastUsed. Before(threshold) {
                cached.reader.Close()
                delete(c.readers, id)
            }
        }
        c.mu.Unlock()
    }
}

// Close closes all cached readers
func (c *CBZCache) Close() {
    c.mu.Lock()
    defer c.mu. Unlock()
    
    for _, cached := range c.readers {
        cached. reader.Close()
    }
    c.readers = make(map[int64]*cachedReader)
}
```

---

## Frontend Architecture

### Technology Stack

| Technology | Purpose |
|------------|---------|
| **Svelte** | UI framework (reactive, small bundle) |
| **SvelteKit** | Routing and SSR capabilities |
| **Vite** | Build tool and dev server |
| **Tailwind CSS** | Utility-first styling |
| **TypeScript** | Type safety |

### Project Structure

```
web/
├── src/
│   ├── lib/
│   │   ├── components/           # Reusable components
│   │   │   ├── MangaCard.svelte
│   │   │   ├── MangaGrid.svelte
│   │   │   ├── ChapterList.svelte
│   │   │   ├── Reader.svelte
│   │   │   ├── ReaderControls.svelte
│   │   │   ├── SearchModal.svelte
│   │   │   ├── Navbar.svelte
│   │   │   ├── Sidebar.svelte
│   │   │   ├── Toast.svelte
│   │   │   └── ui/              # Base UI components
│   │   │       ├── Button.svelte
│   │   │       ├── Input. svelte
│   │   │       ├── Modal.svelte
│   │   │       ├── Dropdown.svelte
│   │   │       └── Spinner.svelte
│   │   ├── api/
│   │   │   ├── client.ts        # API client
│   │   │   ├── types.ts         # TypeScript types
│   │   │   └── endpoints.ts     # API endpoint definitions
│   │   ├── stores/
│   │   │   ├── library.ts       # Library state
│   │   │   ├── reader.ts        # Reader state
│   │   │   ├── downloads.ts     # Download queue state
│   │   │   └── settings.ts      # User preferences
│   │   └── utils/
│   │       ├── format.ts        # Formatting helpers
│   │       └── keyboard.ts      # Keyboard handling
│   ├── routes/
│   │   ├── +page.svelte         # Home / Library
│   │   ├── +layout.svelte       # App layout
│   │   ├── manga/
│   │   │   └── [id]/
│   │   │       └── +page.svelte # Manga detail
│   │   ├── read/
│   │   │   └── [chapterId]/
│   │   │       └── +page.svelte # Reader
│   │   ├── search/
│   │   │   └── +page.svelte     # Search results
│   │   ├── downloads/
│   │   │   └── +page.svelte     # Download queue
│   │   └── settings/
│   │       └── +page.svelte     # Settings
│   ├── app.html                 # HTML template
│   ├── app.css                  # Global styles (Tailwind)
│   └── app. d.ts                 # TypeScript declarations
├── static/
│   ├── favicon.ico
│   └── icons/
├── package.json
├── svelte.config.js
├── tailwind.config.js
├── tsconfig.json
└── vite.config.ts
```

### API Client

```typescript
// web/src/lib/api/client.ts

const BASE_URL = '/api';

interface ApiError {
    code: string;
    message: string;
    details?: Record<string, unknown>;
}

interface ApiResponse<T> {
    data: T;
    meta?: {
        page: number;
        perPage: number;
        total: number;
    };
}

class ApiClient {
    private baseUrl: string;

    constructor(baseUrl: string = BASE_URL) {
        this.baseUrl = baseUrl;
    }

    private async request<T>(
        method: string,
        path: string,
        options?: {
            body?: unknown;
            params?: Record<string, string>;
        }
    ): Promise<T> {
        let url = `${this.baseUrl}${path}`;
        
        if (options?.params) {
            const searchParams = new URLSearchParams(options.params);
            url += `?${searchParams.toString()}`;
        }

        const response = await fetch(url, {
            method,
            headers: {
                'Content-Type': 'application/json',
            },
            body: options?.body ? JSON.stringify(options.body) : undefined,
        });

        if (!response.ok) {
            const error: { error: ApiError } = await response.json();
            throw new ApiClientError(error.error);
        }

        return response.json();
    }

    // Library
    async getLibrary(): Promise<ApiResponse<Manga[]>> {
        return this.request('GET', '/manga');
    }

    async getManga(id: number): Promise<ApiResponse<MangaDetail>> {
        return this.request('GET', `/manga/${id}`);
    }

    async addManga(source: string, sourceId: string): Promise<ApiResponse<Manga>> {
        return this. request('POST', '/manga', {
            body: { source, sourceId },
        });
    }

    async deleteManga(id: number): Promise<void> {
        await this.request('DELETE', `/manga/${id}`);
    }

    async refreshManga(id: number): Promise<ApiResponse<Manga>> {
        return this.request('POST', `/manga/${id}/refresh`);
    }

    // Chapters
    async getChapters(mangaId: number): Promise<ApiResponse<Chapter[]>> {
        return this.request('GET', `/manga/${mangaId}/chapters`);
    }

    async downloadChapter(chapterId: number): Promise<void> {
        await this.request('POST', `/chapters/${chapterId}/download`);
    }

    async downloadChapters(chapterIds: number[]): Promise<void> {
        await this.request('POST', '/chapters/bulk-download', {
            body: { chapterIds },
        });
    }

    async markChapterRead(chapterId: number): Promise<void> {
        await this.request('PATCH', `/chapters/${chapterId}`, {
            body: { isRead: true },
        });
    }

    // Reader
    async getReadingSession(chapterId: number): Promise<ApiResponse<ReadingSession>> {
        return this. request('GET', `/read/${chapterId}`);
    }

    getPageUrl(chapterId: number, page: number): string {
        return `${this.baseUrl}/read/${chapterId}/page/${page}`;
    }

    async updateProgress(chapterId: number, page: number): Promise<void> {
        await this.request('PATCH', `/read/${chapterId}/progress`, {
            body: { page },
        });
    }

    // Search
    async search(query: string, source?: string): Promise<ApiResponse<SearchResult[]>> {
        const params: Record<string, string> = { q: query };
        if (source) params. source = source;
        return this.request('GET', '/search', { params });
    }

    // Sources
    async getSources(): Promise<ApiResponse<Source[]>> {
        return this. request('GET', '/sources');
    }

    // Downloads
    async getDownloads(): Promise<ApiResponse<DownloadJob[]>> {
        return this.request('GET', '/downloads');
    }

    async cancelDownload(id: number): Promise<void> {
        await this.request('DELETE', `/downloads/${id}`);
    }

    // Settings
    async getSettings(): Promise<ApiResponse<Settings>> {
        return this.request('GET', '/settings');
    }

    async updateSettings(settings: Partial<Settings>): Promise<ApiResponse<Settings>> {
        return this.request('PATCH', '/settings', { body: settings });
    }
}

class ApiClientError extends Error {
    code: string;
    details?: Record<string, unknown>;

    constructor(error: ApiError) {
        super(error.message);
        this.code = error.code;
        this.details = error.details;
    }
}

export const api = new ApiClient();
export { ApiClientError };
```

### TypeScript Types

```typescript
// web/src/lib/api/types.ts

export interface Manga {
    id: number;
    title: string;
    slug: string;
    source: string;
    coverUrl: string | null;
    coverPath: string | null;
    status: 'ongoing' | 'completed' | 'hiatus' | 'cancelled' | 'unknown';
    unreadCount: number;
    chapterCount: number;
    lastUpdated: string;
}

export interface MangaDetail extends Manga {
    description: string | null;
    author: string | null;
    artist: string | null;
    genres: string[];
    tags: string[];
    anilistId: number | null;
    chapters: Chapter[];
}

export interface Chapter {
    id: number;
    title: string;
    number: number;
    volume: string | null;
    status: 'pending' | 'queued' | 'downloading' | 'completed' | 'failed';
    isRead: boolean;
    currentPage: number;
    pageCount: number | null;
    publishedAt: string | null;
    downloadedAt: string | null;
}

export interface ReadingSession {
    chapter: {
        id: number;
        title: string;
        number: number;
        volume: string | null;
        currentPage: number;
        pageCount: number;
    };
    manga: {
        id: number;
        title: string;
    };
    previousChapter: ChapterNav | null;
    nextChapter: ChapterNav | null;
    pages: PageInfo[];
}

export interface ChapterNav {
    id: number;
    number: number;
    title: string;
}

export interface PageInfo {
    index: number;
    filename: string;
    size: number;
}

export interface SearchResult {
    id: string;
    title: string;
    coverUrl: string;
    source: string;
    status: string;
}

export interface Source {
    id: string;
    name: string;
    baseUrl: string;
    languages: string[];
    isNsfw: boolean;
}

export interface DownloadJob {
    id: number;
    chapterId: number;
    mangaTitle: string;
    chapterTitle: string;
    status: 'queued' | 'downloading' | 'completed' | 'failed';
    progress: number;
    error: string | null;
}

export interface Settings {
    library: {
        path: string;
    };
    downloader: {
        workers: number;
        format: 'cbz' | 'pdf' | 'raw';
    };
    reader: {
        defaultMode: 'single' | 'double' | 'vertical';
        defaultDirection: 'rtl' | 'ltr';
    };
    updates: {
        enabled: boolean;
        interval: string;
    };
}
```

### Reader Component

```svelte
<!-- web/src/lib/components/Reader.svelte -->
<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { api } from '$lib/api/client';
    import type { ReadingSession } from '$lib/api/types';
    import { readerSettings } from '$lib/stores/settings';
    import ReaderControls from './ReaderControls. svelte';

    export let chapterId: number;

    let session: ReadingSession | null = null;
    let currentPage = 1;
    let loading = true;
    let error: string | null = null;
    let imageUrl = '';
    let showControls = true;
    let controlsTimeout: ReturnType<typeof setTimeout>;

    // Preload adjacent pages
    let preloadedImages: Map<number, HTMLImageElement> = new Map();

    $: if (session) {
        imageUrl = api.getPageUrl(chapterId, currentPage);
        preloadAdjacentPages();
        saveProgress();
    }

    onMount(async () => {
        await loadSession();
        document.addEventListener('keydown', handleKeydown);
    });

    onDestroy(() => {
        document.removeEventListener('keydown', handleKeydown);
        if (controlsTimeout) clearTimeout(controlsTimeout);
    });

    async function loadSession() {
        try {
            loading = true;
            const response = await api. getReadingSession(chapterId);
            session = response. data;
            currentPage = session.chapter.currentPage || 1;
        } catch (e) {
            error = e instanceof Error ? e. message : 'Failed to load chapter';
        } finally {
            loading = false;
        }
    }

    function handleKeydown(e: KeyboardEvent) {
        if (! session) return;

        switch (e.key) {
            case 'ArrowRight':
            case ' ':
                if ($readerSettings.direction === 'ltr') {
                    nextPage();
                } else {
                    prevPage();
                }
                break;
            case 'ArrowLeft':
                if ($readerSettings.direction === 'ltr') {
                    prevPage();
                } else {
                    nextPage();
                }
                break;
            case 'ArrowUp':
                prevPage();
                break;
            case 'ArrowDown':
                nextPage();
                break;
            case 'f':
                toggleFullscreen();
                break;
            case 'Escape':
                showControls = ! showControls;
                break;
        }
    }

    function nextPage() {
        if (! session) return;
        
        if (currentPage < session.chapter.pageCount) {
            currentPage++;
        } else if (session.nextChapter) {
            navigateToChapter(session.nextChapter.id);
        }
    }

    function prevPage() {
        if (! session) return;
        
        if (currentPage > 1) {
            currentPage--;
        } else if (session. previousChapter) {
            navigateToChapter(session. previousChapter.id);
        }
    }

    function navigateToChapter(id: number) {
        window.location.href = `/read/${id}`;
    }

    async function saveProgress() {
        if (!session) return;
        try {
            await api.updateProgress(chapterId, currentPage);
            
            // Mark as read when reaching last page
            if (currentPage === session. chapter.pageCount) {
                await api.markChapterRead(chapterId);
            }
        } catch (e) {
            console.error('Failed to save progress:', e);
        }
    }

    function preloadAdjacentPages() {
        if (! session) return;
        
        const pagesToPreload = [currentPage - 1, currentPage + 1, currentPage + 2];
        
        for (const page of pagesToPreload) {
            if (page >= 1 && page <= session.chapter.pageCount && !preloadedImages.has(page)) {
                const img = new Image();
                img.src = api.getPageUrl(chapterId, page);
                preloadedImages.set(page, img);
            }
        }
    }

    function toggleFullscreen() {
        if (document.fullscreenElement) {
            document.exitFullscreen();
        } else {
            document.documentElement.requestFullscreen();
        }
    }

    function handleImageClick(e: MouseEvent) {
        const rect = (e.target as HTMLElement).getBoundingClientRect();
        const x = e.clientX - rect.left;
        const width = rect.width;
        
        // Click on left 1/3 = previous, right 2/3 = next
        if (x < width / 3) {
            if ($readerSettings.direction === 'rtl') {
                nextPage();
            } else {
                prevPage();
            }
        } else {
            if ($readerSettings. direction === 'rtl') {
                prevPage();
            } else {
                nextPage();
            }
        }
        
        // Show controls briefly
        showControls = true;
        if (controlsTimeout) clearTimeout(controlsTimeout);
        controlsTimeout = setTimeout(() => {
            showControls = false;
        }, 2000);
    }
</script>

<div class="reader bg-black min-h-screen flex flex-col">
    {#if loading}
        <div class="flex-1 flex items-center justify-center">
            <div class="animate-spin rounded-full h-12 w-12 border-4 border-white border-t-transparent"></div>
        </div>
    {:else if error}
        <div class="flex-1 flex items-center justify-center text-red-500">
            <p>{error}</p>
        </div>
    {:else if session}
        <!-- Controls overlay -->
        <ReaderControls
            {session}
            {currentPage}
            visible={showControls}
            on:prev={prevPage}
            on:next={nextPage}
            on:goToPage={(e) => currentPage = e.detail}
            on:close={() => window.history.back()}
        />

        <!-- Main image -->
        <div 
            class="flex-1 flex items-center justify-center cursor-pointer select-none"
            on:click={handleImageClick}
            role="button"
            tabindex="0"
        >
            <img
                src={imageUrl}
                alt="Page {currentPage}"
                class="max-h-screen max-w-full object-contain"
                class:w-full={$readerSettings.mode === 'vertical'}
                draggable="false"
            />
        </div>

        <!-- Progress bar -->
        <div class="h-1 bg-gray-800">
            <div 
                class="h-full bg-blue-500 transition-all duration-200"
                style="width: {(currentPage / session. chapter.pageCount) * 100}%"
            ></div>
        </div>
    {/if}
</div>

<style>
    .reader {
        user-select: none;
        -webkit-user-select: none;
    }
</style>
```

### Embedding Frontend in Go

```go
// internal/api/static. go
package api

import (
    "embed"
    "io/fs"
    "net/http"
)

//go:embed all:dist
var webDist embed.FS

// GetFileSystem returns the embedded frontend filesystem
func GetFileSystem() http.FileSystem {
    // Strip the "dist" prefix
    stripped, err := fs.Sub(webDist, "dist")
    if err != nil {
        panic(err)
    }
    return http. FS(stripped)
}

// SetupStaticRoutes configures serving of embedded frontend files
func SetupStaticRoutes(e *echo.Echo) {
    // Serve static files
    assetHandler := http. FileServer(GetFileSystem())
    
    // Serve index.html for SPA routes
    e.GET("/*", echo.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http. Request) {
        // Try to serve the file directly
        path := r.URL.Path
        
        // Check if file exists
        f, err := webDist.Open("dist" + path)
        if err == nil {
            f.Close()
            assetHandler.ServeHTTP(w, r)
            return
        }
        
        // Fallback to index. html for SPA routing
        r.URL.Path = "/"
        assetHandler.ServeHTTP(w, r)
    })))
}
```

---

## Configuration System

### Configuration Structure

```go
// internal/config/config.go
package config

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/spf13/viper"
)

type Config struct {
    Server       ServerConfig       `mapstructure:"server"`
    Library      LibraryConfig      `mapstructure:"library"`
    Downloader   DownloaderConfig   `mapstructure:"downloader"`
    Formats      FormatsConfig      `mapstructure:"formats"`
    Updates      UpdatesConfig      `mapstructure:"updates"`
    Metadata     MetadataConfig     `mapstructure:"metadata"`
    Notifications NotificationsConfig `mapstructure:"notifications"`
    Reader       ReaderConfig       `mapstructure:"reader"`
    Sources      SourcesConfig      `mapstructure:"sources"`
    Logging      LoggingConfig      `mapstructure:"logging"`
    Database     DatabaseConfig     `mapstructure:"database"`
}

type ServerConfig struct {
    Host    string `mapstructure:"host"`
    Port    int    `mapstructure:"port"`
    BaseURL string `mapstructure:"baseUrl"`
}

type LibraryConfig struct {
    Path            string `mapstructure:"path"`
    ScanOnStartup   bool   `mapstructure:"scanOnStartup"`
    WatchForChanges bool   `mapstructure:"watchForChanges"`
}

type DownloaderConfig struct {
    Workers       int           `mapstructure:"workers"`
    RetryAttempts int           `mapstructure:"retryAttempts"`
    RetryDelay    time.Duration `mapstructure:"retryDelay"`
    Timeout       time.Duration `mapstructure:"timeout"`
    RateLimit     string        `mapstructure:"rateLimit"`
    UserAgent     string        `mapstructure:"userAgent"`
}

type FormatsConfig struct {
    Default           string `mapstructure:"default"`
    CompressImages    bool   `mapstructure:"compressImages"`
    JpegQuality       int    `mapstructure:"jpegQuality"`
    MaxImageWidth     int    `mapstructure:"maxImageWidth"`
    GenerateComicInfo bool   `mapstructure:"generateComicInfo"`
}

type UpdatesConfig struct {
    Enabled         bool   `mapstructure:"enabled"`
    DefaultInterval string `mapstructure:"defaultInterval"`
    CheckOnStartup  bool   `mapstructure:"checkOnStartup"`
    AutoDownload    bool   `mapstructure:"autoDownload"`
}

type MetadataConfig struct {
    FetchAnilist   bool   `mapstructure:"fetchAnilist"`
    DownloadCovers bool   `mapstructure:"downloadCovers"`
    CoverSize      string `mapstructure:"coverSize"`
}

type NotificationsConfig struct {
    Enabled bool     `mapstructure:"enabled"`
    URLs    []string `mapstructure:"urls"`
}

type ReaderConfig struct {
    DefaultMode      string `mapstructure:"defaultMode"`
    DefaultDirection string `mapstructure:"defaultDirection"`
    PreloadPages     int    `mapstructure:"preloadPages"`
    SaveProgress     bool   `mapstructure:"saveProgress"`
}

type SourcesConfig struct {
    CustomPath string                 `mapstructure:"customPath"`
    Default    string                 `mapstructure:"default"`
    MangaDex   MangaDexSourceConfig   `mapstructure:"mangadex"`
}

type MangaDexSourceConfig struct {
    Language        string `mapstructure:"language"`
    NSFW            bool   `mapstructure:"nsfw"`
    ShowUnavailable bool   `mapstructure:"showUnavailable"`
}

type LoggingConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"`
}

type DatabaseConfig struct {
    Path    string `mapstructure:"path"`
    WALMode bool   `mapstructure:"walMode"`
}
```

### Default Configuration

```go
// internal/config/defaults.go
package config

func SetDefaults() {
    // Server
    viper.SetDefault("server.host", "0.0.0. 0")
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.baseUrl", "")

    // Library
    viper.SetDefault("library.path", "./data/manga")
    viper.SetDefault("library.scanOnStartup", true)
    viper.SetDefault("library.watchForChanges", false)

    // Downloader
    viper.SetDefault("downloader.workers", 3)
    viper.SetDefault("downloader.retryAttempts", 3)
    viper.SetDefault("downloader.retryDelay", "5s")
    viper.SetDefault("downloader.timeout", "30s")
    viper.SetDefault("downloader.rateLimit", "2/s")
    viper.SetDefault("downloader.userAgent", "MangaShelf/1.0")

    // Formats
    viper. SetDefault("formats. default", "cbz")
    viper.SetDefault("formats.compressImages", false)
    viper.SetDefault("formats.jpegQuality", 85)
    viper.SetDefault("formats.maxImageWidth", 0)
    viper.SetDefault("formats.generateComicInfo", true)

    // Updates
    viper. SetDefault("updates. enabled", true)
    viper.SetDefault("updates.defaultInterval", "0 */6 * * *")
    viper.SetDefault("updates.checkOnStartup", true)
    viper.SetDefault("updates.autoDownload", true)

    // Metadata
    viper.SetDefault("metadata.fetchAnilist", true)
    viper.SetDefault("metadata.downloadCovers", true)
    viper.SetDefault("metadata.coverSize", "large")

    // Notifications
    viper.SetDefault("notifications.enabled", false)
    viper.SetDefault("notifications.urls", []string{})

    // Reader
    viper.SetDefault("reader.defaultMode", "single")
    viper.SetDefault("reader.defaultDirection", "rtl")
    viper.SetDefault("reader.preloadPages", 2)
    viper.SetDefault("reader.saveProgress", true)

    // Sources
    viper. SetDefault("sources. customPath", "./data/scrapers")
    viper.SetDefault("sources.default", "mangadex")
    viper.SetDefault("sources.mangadex.language", "en")
    viper.SetDefault("sources.mangadex.nsfw", false)
    viper.SetDefault("sources.mangadex. showUnavailable", false)

    // Logging
    viper.SetDefault("logging.level", "info")
    viper.SetDefault("logging.format", "text")

    // Database
    viper.SetDefault("database.path", "./data/mangashelf.db")
    viper.SetDefault("database.walMode", true)
}
```

### Configuration Loading

```go
// internal/config/loader.go
package config

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/spf13/viper"
)

// Load reads configuration from file, environment, and flags
func Load(configPath string) (*Config, error) {
    SetDefaults()

    // Config file
    if configPath != "" {
        viper.SetConfigFile(configPath)
    } else {
        // Search in standard locations
        viper.SetConfigName("config")
        viper.SetConfigType("yaml")
        viper.AddConfigPath(".")
        viper.AddConfigPath("./data")
        viper.AddConfigPath(getConfigDir())
    }

    // Environment variables
    viper. SetEnvPrefix("MANGASHELF")
    viper.SetEnvKeyReplacer(strings. NewReplacer(".", "_"))
    viper.AutomaticEnv()

    // Read config file (optional)
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ! ok {
            return nil, fmt. Errorf("read config: %w", err)
        }
        // Config file not found is OK - use defaults
    }

    var cfg Config
    if err := viper. Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }

    // Ensure directories exist
    if err := ensureDirectories(&cfg); err != nil {
        return nil, fmt.Errorf("create directories: %w", err)
    }

    return &cfg, nil
}

func getConfigDir() string {
    if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
        return filepath. Join(xdgConfig, "mangashelf")
    }
    
    home, err := os.UserHomeDir()
    if err != nil {
        return "."
    }
    
    return filepath.Join(home, ".config", "mangashelf")
}

func ensureDirectories(cfg *Config) error {
    dirs := []string{
        cfg.Library.Path,
        cfg.Sources.CustomPath,
        filepath.Dir(cfg.Database. Path),
    }
    
    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("create %s: %w", dir, err)
        }
    }
    
    return nil
}
```

---

## Error Handling

### Error Types

```go
// internal/errors/errors.go
package errors

import (
    "errors"
    "fmt"
)

// Sentinel errors for common cases
var (
    ErrNotFound       = errors.New("not found")
    ErrAlreadyExists  = errors.New("already exists")
    ErrInvalidInput   = errors.New("invalid input")
    ErrUnauthorized   = errors.New("unauthorized")
    ErrRateLimited    = errors.New("rate limited")
    ErrSourceError    = errors.New("source error")
    ErrDownloadFailed = errors.New("download failed")
)

// DomainError wraps errors with domain context
type DomainError struct {
    Domain  string // e.g., "manga", "chapter", "scraper"
    Op      string // e.g., "get", "create", "download"
    Err     error
    Details map[string]interface{}
}

func (e *DomainError) Error() string {
    if e.Details != nil {
        return fmt.Sprintf("%s. %s: %v (%v)", e. Domain, e.Op, e.Err, e.Details)
    }
    return fmt.Sprintf("%s.%s: %v", e.Domain, e. Op, e. Err)
}

func (e *DomainError) Unwrap() error {
    return e. Err
}

// Helper constructors
func NewMangaError(op string, err error, details map[string]interface{}) error {
    return &DomainError{Domain: "manga", Op: op, Err: err, Details: details}
}

func NewChapterError(op string, err error, details map[string]interface{}) error {
    return &DomainError{Domain: "chapter", Op: op, Err: err, Details: details}
}

func NewScraperError(op string, err error, details map[string]interface{}) error {
    return &DomainError{Domain: "scraper", Op: op, Err: err, Details: details}
}

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
    return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if error is a duplicate error
func IsAlreadyExists(err error) bool {
    return errors.Is(err, ErrAlreadyExists)
}
```

### HTTP Error Handling

```go
// internal/api/middleware/error.go
package middleware

import (
    "errors"
    "net/http"

    "github.com/labstack/echo/v4"
    apperrors "github.com/username/mangashelf/internal/errors"
)

type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

func ErrorHandler(err error, c echo.Context) {
    if c.Response(). Committed {
        return
    }

    var (
        code     = http.StatusInternalServerError
        errCode  = "INTERNAL_ERROR"
        message  = "An internal error occurred"
        details  map[string]interface{}
    )

    // Handle echo HTTP errors
    var he *echo.HTTPError
    if errors.As(err, &he) {
        code = he. Code
        if m, ok := he. Message.(string); ok {
            message = m
        }
        errCode = httpCodeToErrorCode(code)
    }

    // Handle domain errors
    var de *apperrors.DomainError
    if errors.As(err, &de) {
        details = de.Details
        
        switch {
        case errors. Is(de.Err, apperrors.ErrNotFound):
            code = http.StatusNotFound
            errCode = "NOT_FOUND"
            message = fmt.Sprintf("%s not found", de.Domain)
        case errors.Is(de. Err, apperrors.ErrAlreadyExists):
            code = http. StatusConflict
            errCode = "ALREADY_EXISTS"
            message = fmt. Sprintf("%s already exists", de.Domain)
        case errors. Is(de. Err, apperrors.ErrInvalidInput):
            code = http.StatusBadRequest
            errCode = "INVALID_INPUT"
            message = de.Err.Error()
        case errors.Is(de. Err, apperrors.ErrRateLimited):
            code = http. StatusTooManyRequests
            errCode = "RATE_LIMITED"
            message = "Too many requests, please try again later"
        }
    }

    c.JSON(code, ErrorResponse{
        Error: ErrorDetail{
            Code:    errCode,
            Message: message,
            Details: details,
        },
    })
}

func httpCodeToErrorCode(code int) string {
    switch code {
    case http.StatusBadRequest:
        return "BAD_REQUEST"
    case http.StatusUnauthorized:
        return "UNAUTHORIZED"
    case http.StatusForbidden:
        return "FORBIDDEN"
    case http.StatusNotFound:
        return "NOT_FOUND"
    case http.StatusConflict:
        return "CONFLICT"
    case http. StatusTooManyRequests:
        return "RATE_LIMITED"
    default:
        return "INTERNAL_ERROR"
    }
}
```

---

## Security Considerations

### Input Validation

```go
// internal/api/validation/validation.go
package validation

import (
    "regexp"
    "strings"

    "github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
    validate = validator. New()
    
    // Register custom validations
    validate. RegisterValidation("slug", validateSlug)
    validate.RegisterValidation("cron", validateCron)
    validate.RegisterValidation("source", validateSource)
}

func Validate(s interface{}) error {
    return validate. Struct(s)
}

var slugRegex = regexp. MustCompile(`^[a-z0-9]+(? :-[a-z0-9]+)*$`)

func validateSlug(fl validator.FieldLevel) bool {
    return slugRegex. MatchString(fl. Field().String())
}

func validateCron(fl validator. FieldLevel) bool {
    expr := fl.Field().String()
    if expr == "never" {
        return true
    }
    // Validate cron expression format
    parts := strings.Fields(expr)
    return len(parts) == 5
}

var validSources = map[string]bool{
    "mangadex":  true,
    "mangasee":  true,
    "manganato": true,
}

func validateSource(fl validator.FieldLevel) bool {
    source := fl. Field().String()
    return validSources[source]
}
```

### Path Traversal Prevention

```go
// internal/library/sanitize. go
package library

import (
    "path/filepath"
    "regexp"
    "strings"
)

var unsafeChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

// sanitizeFilename removes unsafe characters from filenames
func sanitizeFilename(name string) string {
    // Replace unsafe characters
    safe := unsafeChars.ReplaceAllString(name, "_")
    
    // Remove leading/trailing spaces and dots
    safe = strings. Trim(safe, " .")
    
    // Limit length
    if len(safe) > 200 {
        safe = safe[:200]
    }
    
    // Ensure not empty
    if safe == "" {
        safe = "unnamed"
    }
    
    return safe
}

// safePath ensures a path is within the library directory
func (s *Service) safePath(subpath string) (string, error) {
    // Clean and join paths
    full := filepath.Join(s.libraryPath, filepath.Clean(subpath))
    
    // Ensure the result is still within library
    if ! strings.HasPrefix(full, s.libraryPath) {
        return "", fmt.Errorf("path traversal detected")
    }
    
    return full, nil
}
```

### Rate Limiting

```go
// internal/api/middleware/ratelimit.go
package middleware

import (
    "net/http"
    "sync"
    "time"

    "github.com/labstack/echo/v4"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    visitors map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
    rl := &RateLimiter{
        visitors: make(map[string]*rate.Limiter),
        rate:     rate.Limit(requestsPerSecond),
        burst:    burst,
    }
    
    // Cleanup old entries periodically
    go rl.cleanup()
    
    return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate. Limiter {
    rl.mu.Lock()
    defer rl.mu. Unlock()
    
    limiter, exists := rl.visitors[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.visitors[ip] = limiter
    }
    
    return limiter
}

func (rl *RateLimiter) Middleware() echo. MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo. Context) error {
            ip := c. RealIP()
            limiter := rl. getLimiter(ip)
            
            if !limiter. Allow() {
                return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
            }
            
            return next(c)
        }
    }
}

func (rl *RateLimiter) cleanup() {
    ticker := time.NewTicker(time. Minute)
    for range ticker.C {
        rl. mu.Lock()
        // Simple cleanup: clear all (rate limiters are cheap to recreate)
        rl.visitors = make(map[string]*rate.Limiter)
        rl.mu.Unlock()
    }
}
```

---

## Performance Considerations

### Image Optimization

```go
// internal/library/images.go
package library

import (
    "bytes"
    "image"
    "image/jpeg"
    _ "image/png"
    
    "github.com/disintegration/imaging"
)

// OptimizeImage resizes and compresses an image
func OptimizeImage(data []byte, maxWidth int, quality int) ([]byte, error) {
    img, _, err := image.Decode(bytes.NewReader(data))
    if err != nil {
        return nil, err
    }
    
    bounds := img.Bounds()
    width := bounds.Dx()
    
    // Only resize if larger than max width
    if maxWidth > 0 && width > maxWidth {
        img = imaging.Resize(img, maxWidth, 0, imaging. Lanczos)
    }
    
    // Encode as JPEG with specified quality
    var buf bytes.Buffer
    err = jpeg. Encode(&buf, img, &jpeg.Options{Quality: quality})
    if err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

// GenerateThumbnail creates a small thumbnail for covers
func GenerateThumbnail(data []byte, width, height int) ([]byte, error) {
    img, _, err := image. Decode(bytes. NewReader(data))
    if err != nil {
        return nil, err
    }
    
    thumb := imaging.Thumbnail(img, width, height, imaging. Lanczos)
    
    var buf bytes.Buffer
    err = jpeg. Encode(&buf, thumb, &jpeg.Options{Quality: 80})
    if err != nil {
        return nil, err
    }
    
    return buf. Bytes(), nil
}
```

### Database Optimization

```go
// internal/database/optimize.go
package database

import (
    "context"
    "database/sql"
)

// OptimizeDatabase runs maintenance operations
func OptimizeDatabase(db *sql.DB) error {
    ctx := context.Background()
    
    // Analyze tables for query optimization
    if _, err := db.ExecContext(ctx, "ANALYZE"); err != nil {
        return err
    }
    
    // Vacuum to reclaim space (run periodically, not on every start)
    // if _, err := db. ExecContext(ctx, "VACUUM"); err != nil {
    //     return err
    // }
    
    return nil
}

// SetPragmas configures SQLite for optimal performance
func SetPragmas(db *sql.DB) error {
    pragmas := []string{
        "PRAGMA journal_mode = WAL",
        "PRAGMA synchronous = NORMAL",
        "PRAGMA cache_size = -64000", // 64MB cache
        "PRAGMA temp_store = MEMORY",
        "PRAGMA mmap_size = 268435456", // 256MB mmap
        "PRAGMA foreign_keys = ON",
    }
    
    for _, pragma := range pragmas {
        if _, err := db. Exec(pragma); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Connection Pooling

```go
// internal/database/pool.go
package database

import (
    "database/sql"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
)

func OpenDatabase(path string) (*sql.DB, error) {
    // SQLite connection string with optimizations
    dsn := path + "?_journal=WAL&_timeout=5000&_fk=1"
    
    db, err := sql.Open("sqlite3", dsn)
    if err != nil {
        return nil, err
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(1)           // SQLite supports one writer
    db.SetMaxIdleConns(1)
    db.SetConnMaxLifetime(time.Hour)
    
    // Apply pragmas
    if err := SetPragmas(db); err != nil {
        db.Close()
        return nil, err
    }
    
    return db, nil
}
```

---

## Deployment Architecture

### Single Binary Deployment

```
┌─────────────────────────────────────────┐
│              User's Server              │
├─────────────────────────────────────────┤
│                                         │
│   ┌─────────────────────────────────┐   │
│   │         mangashelf              │   │
│   │       (single binary)           │   │
│   │                                 │   │
│   │  ┌───────────────────────────┐  │   │
│   │  │   Embedded SQLite DB      │  │   │
│   │  └───────────────────────────┘  │   │
│   │                                 │   │
│   │  ┌───────────────────────────┐  │   │
│   │  │   Embedded Web Assets     │  │   │
│   │  └───────────────────────────┘  │   │
│   │                                 │   │
│   │  ┌───────────────────────────┐  │   │
│   │  │   Goroutine Workers       │  │   │
│   │  └───────────────────────────┘  │   │
│   └─────────────────────────────────┘   │
│                   │                     │
│                   ▼                     │
│   ┌─────────────────────────────────┐   │
│   │       /data directory            │   │
│   │                                 │   │
│   │   ├── mangashelf.db            │   │
│   │   ├── config.yaml              │   │
│   │   ├── manga/                   │   │
│   │   │   └── <downloaded manga>   │   │
│   │   └── scrapers/                │   │
│   │       └── <custom scrapers>    │   │
│   └─────────────────────────────────┘   │
│                                         │
└─────────────────────────────────────────┘
```

### Docker Deployment

```dockerfile
# Dockerfile

# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make nodejs npm

# Copy go mod files
COPY go.mod go. sum ./
RUN go mod download

# Copy source
COPY . . 

# Build frontend
WORKDIR /build/web
RUN npm ci && npm run build

# Build backend
WORKDIR /build
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o mangashelf \
    ./cmd/mangashelf

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /build/mangashelf /usr/local/bin/

# Create non-root user
RUN adduser -D -h /data mangashelf
USER mangashelf

VOLUME /data
WORKDIR /data

EXPOSE 8080

ENTRYPOINT ["mangashelf"]
CMD ["serve"]
```

```yaml
# docker-compose.yml

version: '3.8'

services:
  mangashelf:
    image: ghcr.io/username/mangashelf:latest
    container_name: mangashelf
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - TZ=UTC
      - MANGASHELF_SERVER_PORT=8080
    # Optional: resource limits for shared hosting
    # deploy:
    #   resources:
    #     limits:
    #       memory: 512M
    #       cpus: '1.0'
```

### Reverse Proxy Configuration

```nginx
# nginx. conf

upstream mangashelf {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name manga.example.com;

    ssl_certificate /etc/letsencrypt/live/manga.example. com/fullchain. pem;
    ssl_certificate_key /etc/letsencrypt/live/manga.example.com/privkey.pem;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Increase body size for potential cover uploads
    client_max_body_size 10M;

    location / {
        proxy_pass http://mangashelf;
        proxy_http_version 1. 1;
        
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support (for future real-time features)
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Timeouts
        proxy_connect_timeout 