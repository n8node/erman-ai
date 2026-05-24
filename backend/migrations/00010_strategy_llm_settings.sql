-- +goose Up
CREATE TABLE strategy_llm_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    config JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO strategy_llm_settings (id, config) VALUES (1, '{
  "provider": "openrouter",
  "openrouter_model": "anthropic/claude-sonnet-4-5",
  "deepseek_model": "deepseek-chat",
  "system_prompt": "",
  "temperature": 0.7,
  "max_tokens": 8192
}'::jsonb);

-- +goose Down
DROP TABLE IF EXISTS strategy_llm_settings;
