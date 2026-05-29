-- +goose Up
CREATE TABLE project_inquiries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    calculator_run_id UUID REFERENCES tool_runs(id) ON DELETE SET NULL,
    name VARCHAR(200) NOT NULL,
    email VARCHAR(255) NOT NULL,
    telegram VARCHAR(100) NOT NULL,
    project_title VARCHAR(300) NOT NULL DEFAULT '',
    project_description TEXT NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending_email',
    email_verified_at TIMESTAMPTZ,
    locale VARCHAR(10) NOT NULL DEFAULT 'ru',
    ip_hash VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_project_inquiries_created_at ON project_inquiries(created_at DESC);
CREATE INDEX idx_project_inquiries_status ON project_inquiries(status);
CREATE INDEX idx_project_inquiries_email ON project_inquiries(email);

CREATE TABLE inquiry_verification_tokens (
    inquiry_id UUID PRIMARY KEY REFERENCES project_inquiries(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS inquiry_verification_tokens;
DROP TABLE IF EXISTS project_inquiries;
