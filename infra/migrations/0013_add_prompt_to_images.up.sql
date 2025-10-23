-- Add prompt column to images table for custom staging prompts
ALTER TABLE images ADD COLUMN prompt TEXT;

COMMENT ON COLUMN images.prompt IS 'Custom prompt for AI staging. If null, uses default prompt from library based on room_type and style';
