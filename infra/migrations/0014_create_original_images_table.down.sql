-- Reverse migration for original_images table

-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_update_original_images_updated_at ON original_images;
DROP FUNCTION IF EXISTS update_original_images_updated_at();

-- Drop foreign key and index from images table
DROP INDEX IF EXISTS idx_images_original_image_id;
ALTER TABLE images DROP COLUMN IF EXISTS original_image_id;

-- Restore original_url as NOT NULL (data must exist before running this)
ALTER TABLE images ALTER COLUMN original_url SET NOT NULL;

-- Drop original_images table and its indexes
DROP INDEX IF EXISTS idx_original_images_ref_count;
DROP INDEX IF EXISTS idx_original_images_hash;
DROP TABLE IF EXISTS original_images;
