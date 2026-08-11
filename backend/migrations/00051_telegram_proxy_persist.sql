-- +goose Up
-- Keep proxy enabled when URLs are configured (fixes missing proxy_enabled flag after reboot).
UPDATE telegram_settings
SET config = jsonb_set(config, '{proxy_enabled}', 'true'::jsonb, true)
WHERE id = 1
  AND jsonb_array_length(COALESCE(config->'proxy_urls', '[]'::jsonb)) > 0
  AND COALESCE((config->>'proxy_enabled')::boolean, false) = false;

-- +goose Down
-- no-op: do not revert inferred proxy_enabled
