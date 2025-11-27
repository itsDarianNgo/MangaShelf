-- name: ListQueue :many
SELECT * FROM download_queue
ORDER BY priority DESC, created_at ASC;

-- name: EnqueueDownload :one
INSERT INTO download_queue (
    chapter_id, priority
) VALUES (?, ?)
ON CONFLICT (chapter_id) DO UPDATE SET
    priority = excluded.priority,
    status = 'queued',
    attempts = 0,
    last_error = NULL,
    started_at = NULL,
    completed_at = NULL
RETURNING *;

-- name: UpdateDownloadStatus :one
UPDATE download_queue SET
    status = ?,
    attempts = ?,
    last_error = ?,
    started_at = ?,
    completed_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteDownload :exec
DELETE FROM download_queue WHERE id = ?;
