-- name: GetManga :one
SELECT * FROM manga WHERE id = ? LIMIT 1;

-- name: GetMangaBySlug :one
SELECT * FROM manga WHERE slug = ? LIMIT 1;

-- name: ListManga :many
SELECT * FROM manga ORDER BY title ASC;

-- name: ListMangaWithUnread :many
SELECT
    m.*,
    SUM(CASE WHEN c.is_read = 0 AND c.status = 'completed' THEN 1 ELSE 0 END) AS unread_count
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
ORDER BY last_checked_at ASC
LIMIT ?;
