-- Add soft delete support to images table
-- Once an image is created, it counts toward usage limit even if deleted
ALTER TABLE images ADD COLUMN deleted_at TIMESTAMPTZ;

-- Index for efficient filtering of non-deleted images
CREATE INDEX idx_images_deleted_at ON images (deleted_at) WHERE deleted_at IS NULL;

-- Index for efficient user+deleted queries
CREATE INDEX idx_images_project_deleted ON images (project_id, deleted_at);
