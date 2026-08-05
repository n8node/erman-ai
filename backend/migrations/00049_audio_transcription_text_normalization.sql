-- +goose Up
UPDATE audio_transcription_settings
SET config = config || '{
  "text_normalization_enabled": true,
  "literature_text": true,
  "profanity_filter": false
}'::jsonb,
    updated_at = NOW()
WHERE id = 1;

-- +goose Down
UPDATE audio_transcription_settings
SET config = config - 'text_normalization_enabled' - 'literature_text' - 'profanity_filter',
    updated_at = NOW()
WHERE id = 1;
