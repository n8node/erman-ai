-- +goose Up
ALTER TABLE usage_log ADD COLUMN IF NOT EXISTS cost_rub NUMERIC(12, 4) NOT NULL DEFAULT 0;
ALTER TABLE usage_log ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT '';

UPDATE strategy_llm_settings
SET config = config || '{
  "yandex_api_key": "",
  "yandex_folder_id": "",
  "yandex_model": "",
  "yandex_models": [],
  "pricing": {}
}'::jsonb
WHERE id = 1;

-- +goose Down
ALTER TABLE usage_log DROP COLUMN IF EXISTS provider;
ALTER TABLE usage_log DROP COLUMN IF EXISTS cost_rub;

UPDATE strategy_llm_settings
SET config = config
  - 'yandex_api_key'
  - 'yandex_folder_id'
  - 'yandex_model'
  - 'yandex_models'
  - 'pricing'
WHERE id = 1;
