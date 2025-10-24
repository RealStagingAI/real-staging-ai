-- Create original_images table for content-addressable storage
-- This enables deduplication: store each unique original image only once
CREATE TABLE original_images (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  content_hash VARCHAR(64) NOT NULL UNIQUE,
  s3_key TEXT NOT NULL,
  file_size BIGINT NOT NULL,
  mime_type VARCHAR(50) NOT NULL,
  width INTEGER,
  height INTEGER,
  reference_count INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index for fast hash lookups during deduplication check
CREATE INDEX idx_original_images_hash ON original_images(content_hash);

-- Index for cleanup queries (finding unreferenced originals)
CREATE INDEX idx_original_images_ref_count ON original_images(reference_count);

-- Add foreign key to images table
-- Keep original_url nullable for backward compatibility during migration
ALTER TABLE images ADD COLUMN original_image_id UUID REFERENCES original_images(id) ON DELETE RESTRICT;
ALTER TABLE images ALTER COLUMN original_url DROP NOT NULL;

-- Index for foreign key lookups
CREATE INDEX idx_images_original_image_id ON images(original_image_id);

-- Add trigger to keep updated_at current
CREATE OR REPLACE FUNCTION update_original_images_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_original_images_updated_at
  BEFORE UPDATE ON original_images
  FOR EACH ROW
  EXECUTE FUNCTION update_original_images_updated_at();
