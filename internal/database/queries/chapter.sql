-- name: GetChapter :one
SELECT * FROM chapter WHERE id = ? LIMIT 1;

-- name: ListChaptersByManga :many
SELECT * FROM chapter WHERE manga_id = ? ORDER BY number DESC;

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
    downloaded_at = CASE WHEN ? = 'completed' THEN datetime('now') ELSE downloaded_at END
WHERE id = ?
RETURNING *;

-- name: MarkChapterRead :exec
UPDATE chapter SET is_read = 1, read_at = datetime('now') WHERE id = ?;

-- name: UpdateReadingProgress :exec
UPDATE chapter SET current_page = ? WHERE id = ?;
