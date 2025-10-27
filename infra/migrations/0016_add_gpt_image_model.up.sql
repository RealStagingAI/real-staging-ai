-- Seed default configuration for GPT Image 1 model
INSERT INTO settings (key, value, description, model_settings)
VALUES (
    'model_config_gpt_image_1',
    'openai/gpt-image-1',
    'Configuration for OpenAI GPT Image 1 model',
    '{
        "openai_api_key": "",
        "prompt": "",
        "aspect_ratio": "1:1",
        "input_fidelity": "low",
        "input_images": [],
        "number_of_images": 1,
        "quality": "auto",
        "background": "auto",
        "output_compression": 90,
        "output_format": "webp",
        "moderation": "auto",
        "user_id": null
    }'::jsonb
) ON CONFLICT (key) DO UPDATE SET
    model_settings = EXCLUDED.model_settings,
    value = EXCLUDED.value;
