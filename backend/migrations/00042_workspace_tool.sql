-- +goose Up
INSERT INTO tools (slug, name, description, enabled) VALUES
('workspace', 'Workspace', 'Collaborative documents and edgeless whiteboards powered by AFFiNE', true);

UPDATE plans
SET tool_limits = tool_limits || '{"workspace": -1}'::jsonb
WHERE slug IN ('free', 'pro', 'business');

-- +goose Down
DELETE FROM tools WHERE slug = 'workspace';
UPDATE plans
SET tool_limits = tool_limits - 'workspace'
WHERE slug IN ('free', 'pro', 'business');
