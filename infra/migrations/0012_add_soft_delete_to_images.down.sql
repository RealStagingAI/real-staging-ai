-- Remove soft delete support from images table
DROP INDEX IF EXISTS idx_images_project_deleted;
DROP INDEX IF EXISTS idx_images_deleted_at;
ALTER TABLE images DROP COLUMN IF EXISTS deleted_at;
