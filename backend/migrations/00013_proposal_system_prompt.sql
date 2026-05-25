-- +goose Up
UPDATE strategy_llm_settings
SET config = config || '{"proposal_system_prompt": ""}'::jsonb
WHERE id = 1 AND NOT (config ? 'proposal_system_prompt');

-- +goose Down
UPDATE strategy_llm_settings
SET config = config - 'proposal_system_prompt'
WHERE id = 1;
