-- internal/database/schema.sql

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
    source          TEXT NOT NULL,
    source_id       TEXT NOT NULL,
    url             TEXT NOT NULL,

    -- Cover images
    cover_url       TEXT,
    cover_path      TEXT,

    -- Metadata
    description     TEXT,
    status          TEXT CHECK(status IN ('ongoing', 'completed', 'hiatus', 'cancelled', 'unknown')),
    author          TEXT,
    artist          TEXT,
    genres          TEXT,
    tags            TEXT,

    -- External IDs
    anilist_id      INTEGER,
    mal_id          INTEGER,

    -- Settings
    update_interval TEXT DEFAULT '0 */6 * * *',
    auto_download   INTEGER DEFAULT 1,

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
    number          REAL NOT NULL,
    volume          TEXT,

    -- Source information
    source_id       TEXT NOT NULL,
    url             TEXT NOT NULL,

    -- Download status
    status          TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'queued', 'downloading', 'completed', 'failed')),
    file_path       TEXT,
    file_size       INTEGER,
    page_count      INTEGER,

    -- Reading progress
    is_read         INTEGER DEFAULT 0,
    current_page    INTEGER DEFAULT 0,
    read_at         TEXT,

    -- Timestamps
    published_at    TEXT,
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
    priority        INTEGER DEFAULT 0,
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
    path            TEXT,
    enabled         INTEGER DEFAULT 1,
    config          TEXT,

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
