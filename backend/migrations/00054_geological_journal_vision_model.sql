-- +goose Up
UPDATE geological_journal_settings
SET config = config || jsonb_build_object(
    'vision_model', COALESCE(config->>'vision_model', '')
), updated_at = NOW()
WHERE id = 1;

-- +goose Down
UPDATE geological_journal_settings
SET config = config - 'vision_model', updated_at = NOW()
WHERE id = 1;