-- +goose Up
UPDATE plans SET features = features || '{"legal_scan_max_pages": 5}'::jsonb WHERE slug = 'free';
UPDATE plans SET features = features || '{"legal_scan_max_pages": 30}'::jsonb WHERE slug = 'pro';
UPDATE plans SET features = features || '{"legal_scan_max_pages": 50}'::jsonb WHERE slug = 'business';

-- +goose Down
UPDATE plans SET features = features - 'legal_scan_max_pages' WHERE slug IN ('free', 'pro', 'business');
