-- +goose Down
DELETE FROM tools WHERE slug IN ('calculator', 'strategy', 'proposal');
DELETE FROM plans WHERE slug IN ('free', 'pro', 'business');
