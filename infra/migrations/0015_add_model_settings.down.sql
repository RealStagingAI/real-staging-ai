-- Remove model configuration entries
DELETE FROM settings WHERE key LIKE 'model_config_%';

-- Drop index
DROP INDEX IF EXISTS idx_settings_model_settings;

-- Remove model_settings column
ALTER TABLE settings 
DROP COLUMN IF EXISTS model_settings;
