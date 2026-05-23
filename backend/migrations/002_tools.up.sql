-- +goose Up
INSERT INTO plans (slug, name, price_monthly_rub, price_yearly_rub, tool_limits, features, support_level) VALUES
(
    'free',
    'Free',
    0,
    0,
    '{"calculator": -1, "strategy": 3, "proposal": 1}'::jsonb,
    '{"export_pdf": false, "export_docx": false, "api_access": false, "priority_queue": false, "white_label": false}'::jsonb,
    'community'
),
(
    'pro',
    'Pro',
    2900,
    29000,
    '{"calculator": -1, "strategy": 30, "proposal": 10}'::jsonb,
    '{"export_pdf": true, "export_docx": true, "api_access": true, "priority_queue": false, "white_label": false}'::jsonb,
    'email'
),
(
    'business',
    'Business',
    9900,
    99000,
    '{"calculator": -1, "strategy": -1, "proposal": -1}'::jsonb,
    '{"export_pdf": true, "export_docx": true, "api_access": true, "priority_queue": true, "white_label": true}'::jsonb,
    'priority'
);

INSERT INTO tools (slug, name, description, enabled) VALUES
('calculator', 'Automation Calculator', 'ROI calculator for business process automation', true),
('strategy', 'AI Strategy Generator', 'Personalized AI adoption strategy', true),
('proposal', 'Proposal Generator', 'Commercial proposal generator', true);

-- +goose Down
DELETE FROM tools WHERE slug IN ('calculator', 'strategy', 'proposal');
DELETE FROM plans WHERE slug IN ('free', 'pro', 'business');
