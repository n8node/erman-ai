-- +goose Up
UPDATE strategy_llm_settings
SET config = config || '{
  "openrouter_api_key": "",
  "deepseek_api_key": "",
  "openrouter_models": [],
  "deepseek_models": []
}'::jsonb
WHERE id = 1;

-- +goose Down
UPDATE strategy_llm_settings
SET config = config - 'openrouter_api_key' - 'deepseek_api_key' - 'openrouter_models' - 'deepseek_models'
WHERE id = 1;
