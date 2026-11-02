-- +goose Up
INSERT INTO tools (slug, name, description, enabled) VALUES
('ai-assessment', 'Анализ готовности к AI', 'AI readiness, autonomy and financial assessment for a business process', true)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description;

-- New plans receive a conservative beta limit; existing plans keep their current JSON unchanged.
UPDATE plans SET tool_limits = tool_limits || '{"ai-assessment": 1}'::jsonb WHERE slug = 'free';
UPDATE plans SET tool_limits = tool_limits || '{"ai-assessment": 10}'::jsonb WHERE slug = 'pro';
UPDATE plans SET tool_limits = tool_limits || '{"ai-assessment": -1}'::jsonb WHERE slug = 'business';

-- +goose Down
DELETE FROM tools WHERE slug = 'ai-assessment';
UPDATE plans SET tool_limits = tool_limits - 'ai-assessment';