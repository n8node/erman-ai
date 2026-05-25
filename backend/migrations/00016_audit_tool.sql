-- +goose Up
INSERT INTO tools (slug, name, description, enabled) VALUES
('audit', 'Automation Audit Report', 'Multi-process automation audit with prioritization roadmap', true);

UPDATE plans SET tool_limits = tool_limits || '{"audit": 2}'::jsonb WHERE slug = 'free';
UPDATE plans SET tool_limits = tool_limits || '{"audit": 15}'::jsonb WHERE slug = 'pro';
UPDATE plans SET tool_limits = tool_limits || '{"audit": -1}'::jsonb WHERE slug = 'business';

-- +goose Down
DELETE FROM tools WHERE slug = 'audit';
UPDATE plans SET tool_limits = tool_limits - 'audit' WHERE slug IN ('free', 'pro', 'business');
