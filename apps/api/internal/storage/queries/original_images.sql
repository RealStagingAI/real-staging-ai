-- name: CreateOriginalImage :one
INSERT INTO original_images (
  content_hash, s3_key, file_size, mime_type, width, height, reference_count
) VALUES (
  $1, $2, $3, $4, $5, $6, 1
) RETURNING *;

-- name: GetOriginalImageByID :one
SELECT * FROM original_images
WHERE id = $1;

-- name: GetOriginalImageByHash :one
SELECT * FROM original_images
WHERE content_hash = $1;

-- name: IncrementReferenceCount :exec
UPDATE original_images
SET reference_count = reference_count + 1,
    updated_at = now()
WHERE id = $1;

-- name: DecrementReferenceCount :exec
UPDATE original_images
SET reference_count = GREATEST(reference_count - 1, 0),
    updated_at = now()
WHERE id = $1;

-- name: ListOrphanedOriginalImages :many
SELECT * FROM original_images
WHERE reference_count = 0
  AND updated_at < (NOW() - $1::interval)
ORDER BY updated_at ASC
LIMIT $2;

-- name: DeleteOriginalImage :exec
DELETE FROM original_images
WHERE id = $1;

-- name: GetOriginalImageStats :one
SELECT 
  COUNT(*) as total_count,
  SUM(file_size) as total_size,
  SUM(CASE WHEN reference_count = 0 THEN 1 ELSE 0 END) as orphaned_count,
  SUM(CASE WHEN reference_count = 0 THEN file_size ELSE 0 END) as orphaned_size,
  AVG(reference_count) as avg_references
FROM original_images;
