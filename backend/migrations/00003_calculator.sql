-- +goose Up
ALTER TABLE users
    ADD COLUMN account_segment VARCHAR(20) NOT NULL DEFAULT 'partner',
    ADD COLUMN onboarding_completed BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE shared_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES tool_runs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    view_count INTEGER NOT NULL DEFAULT 0,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shared_reports_token ON shared_reports(token);
CREATE INDEX idx_shared_reports_run_id ON shared_reports(run_id);

CREATE TABLE leads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    run_id UUID NOT NULL REFERENCES tool_runs(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    company VARCHAR(200),
    phone VARCHAR(50),
    email VARCHAR(255) NOT NULL,
    message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_leads_created_at ON leads(created_at DESC);

UPDATE plans SET features = features || '{"share_report": false, "share_report_limit": 0}'::jsonb
WHERE slug = 'free';

UPDATE plans SET features = features || '{"share_report": true, "share_report_limit": 10}'::jsonb
WHERE slug = 'pro';

UPDATE plans SET features = features || '{"share_report": true, "share_report_limit": -1}'::jsonb
WHERE slug = 'business';

UPDATE users SET onboarding_completed = true;

-- +goose Down
DROP TABLE IF EXISTS leads;
DROP TABLE IF EXISTS shared_reports;
ALTER TABLE users DROP COLUMN IF EXISTS onboarding_completed;
ALTER TABLE users DROP COLUMN IF EXISTS account_segment;
