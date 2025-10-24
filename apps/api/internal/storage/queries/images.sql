-- name: CreateImage :one
INSERT INTO images (project_id, original_url, room_type, style, seed, prompt)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, project_id, original_url, staged_url, room_type, style, seed, prompt, status, error, created_at, updated_at, deleted_at;

-- name: GetImageByID :one
SELECT id, project_id, original_url, staged_url, room_type, style, seed, prompt, status, error, created_at, updated_at, deleted_at
FROM images
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetImagesByProjectID :many
SELECT id, project_id, original_url, staged_url, room_type, style, seed, prompt, status, error, created_at, updated_at, deleted_at
FROM images
WHERE project_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateImageStatus :one
UPDATE images
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING id, project_id, original_url, staged_url, room_type, style, seed, prompt, status, error, created_at, updated_at, deleted_at;

-- name: UpdateImageWithStagedURL :one
UPDATE images
SET staged_url = $2, status = $3, updated_at = now()
WHERE id = $1
RETURNING id, project_id, original_url, staged_url, room_type, style, seed, prompt, status, error, created_at, updated_at, deleted_at;

-- name: UpdateImageWithError :one
UPDATE images
SET status = 'error', error = $2, updated_at = now()
WHERE id = $1
RETURNING id, project_id, original_url, staged_url, room_type, style, seed, prompt, status, error, created_at, updated_at, deleted_at;

-- name: SoftDeleteImage :exec
-- Soft delete an image - marks it as deleted but keeps it in DB for usage tracking
UPDATE images
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: DeleteImage :exec
-- Hard delete an image - only use for cleanup operations
DELETE FROM images
WHERE id = $1;

-- name: DeleteImagesByProjectID :exec
-- Hard delete all images in a project - used when cascading project deletion
DELETE FROM images
WHERE project_id = $1;

-- name: DeleteStuckQueuedImages :many
-- Hard delete stuck queued images - cleanup operation for failed uploads
DELETE FROM images
WHERE status = 'queued'
  AND created_at < NOW() - $1::interval
RETURNING id, project_id, created_at;

-- name: ListImagesForReconcile :many
-- List images for reconciliation - only non-deleted images
SELECT id, project_id, original_url, staged_url, room_type, style, seed, prompt, status, error, created_at, updated_at, deleted_at
FROM images
WHERE ($1::uuid IS NULL OR project_id = $1::uuid)
  AND ($2::text IS NULL OR $2::text = '' OR status = $2::image_status)
  AND ($3::uuid IS NULL OR id > $3::uuid)
  AND deleted_at IS NULL
ORDER BY id ASC
LIMIT $4;

-- name: GetOriginalImageIDForImage :one
-- Get the original_image_id for an image before deletion
SELECT original_image_id
FROM images
WHERE id = $1;

-- name: GetOriginalImageIDsForProject :many
-- Get all unique original_image_ids for a project (for bulk deletion)
SELECT DISTINCT original_image_id
FROM images
WHERE project_id = $1
  AND original_image_id IS NOT NULL;
