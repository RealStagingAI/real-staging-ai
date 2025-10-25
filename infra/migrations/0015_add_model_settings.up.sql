-- Add model_settings column to store model-specific JSON configurations
ALTER TABLE settings 
ADD COLUMN IF NOT EXISTS model_settings JSONB DEFAULT '{}'::jsonb;

-- Create index for faster JSON queries
CREATE INDEX IF NOT EXISTS idx_settings_model_settings ON settings USING gin(model_settings);

-- Insert default configurations for each model
INSERT INTO settings (key, value, description, model_settings) 
VALUES (
    'model_config_qwen',
    'qwen/qwen-image-edit',
    'Configuration for Qwen Image Edit model',
    '{
        "go_fast": true,
        "aspect_ratio": "match_input_image",
        "output_format": "webp",
        "output_quality": 80
    }'::jsonb
) ON CONFLICT (key) DO UPDATE SET
    model_settings = EXCLUDED.model_settings;

INSERT INTO settings (key, value, description, model_settings) 
VALUES (
    'model_config_flux_kontext_max',
    'black-forest-labs/flux-kontext-max',
    'Configuration for Flux Kontext Max model',
    '{
        "aspect_ratio": "match_input_image",
        "output_format": "png",
        "safety_tolerance": 4,
        "prompt_upsampling": false,
        "num_outputs": 1,
        "output_quality": 90
    }'::jsonb
) ON CONFLICT (key) DO UPDATE SET
    model_settings = EXCLUDED.model_settings;

INSERT INTO settings (key, value, description, model_settings) 
VALUES (
    'model_config_flux_kontext_pro',
    'black-forest-labs/flux-kontext-pro',
    'Configuration for Flux Kontext Pro model',
    '{
        "aspect_ratio": "match_input_image",
        "output_format": "png",
        "safety_tolerance": 4,
        "prompt_upsampling": false,
        "num_outputs": 1,
        "output_quality": 90
    }'::jsonb
) ON CONFLICT (key) DO UPDATE SET
    model_settings = EXCLUDED.model_settings;

INSERT INTO settings (key, value, description, model_settings) 
VALUES (
    'model_config_seedream_3',
    'bytedance/seedream-3',
    'Configuration for Seedream 3 model',
    '{
        "aspect_ratio": "1:1",
        "num_inference_steps": 50,
        "guidance_scale": 7.5,
        "output_quality": 95
    }'::jsonb
) ON CONFLICT (key) DO UPDATE SET
    model_settings = EXCLUDED.model_settings;

INSERT INTO settings (key, value, description, model_settings) 
VALUES (
    'model_config_seedream_4',
    'bytedance/seedream-4',
    'Configuration for Seedream 4 model',
    '{
        "aspect_ratio": "1:1",
        "num_inference_steps": 50,
        "guidance_scale": 7.5,
        "output_quality": 95
    }'::jsonb
) ON CONFLICT (key) DO UPDATE SET
    model_settings = EXCLUDED.model_settings;
