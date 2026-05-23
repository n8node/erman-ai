-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    price_monthly_rub INTEGER NOT NULL DEFAULT 0,
    price_yearly_rub INTEGER NOT NULL DEFAULT 0,
    tool_limits JSONB NOT NULL DEFAULT '{}',
    features JSONB NOT NULL DEFAULT '{}',
    support_level VARCHAR(50) NOT NULL DEFAULT 'community',
    is_public BOOLEAN NOT NULL DEFAULT true,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    plan_id UUID REFERENCES plans(id),
    locale VARCHAR(10) NOT NULL DEFAULT 'ru',
    is_blocked BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMPTZ
);

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    amount_rub INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tool_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tool_slug VARCHAR(50) NOT NULL,
    plan_tier VARCHAR(50) NOT NULL DEFAULT 'free',
    input JSONB NOT NULL DEFAULT '{}',
    output JSONB,
    artifact_url TEXT,
    tokens_used BIGINT NOT NULL DEFAULT 0,
    model_used VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_msg TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_tool_runs_user_id ON tool_runs(user_id);
CREATE INDEX idx_tool_runs_tool_slug ON tool_runs(tool_slug);
CREATE INDEX idx_tool_runs_created_at ON tool_runs(created_at DESC);

CREATE TABLE usage_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    run_id UUID REFERENCES tool_runs(id) ON DELETE SET NULL,
    model VARCHAR(100) NOT NULL,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    cost_usd NUMERIC(12, 6) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usage_log_user_id ON usage_log(user_id);
CREATE INDEX idx_usage_log_created_at ON usage_log(created_at DESC);

CREATE TABLE tool_usage_counters (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tool_slug VARCHAR(50) NOT NULL,
    period_month DATE NOT NULL,
    runs_count INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, tool_slug, period_month)
);

CREATE TABLE tools (
    slug VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS tool_usage_counters;
DROP TABLE IF EXISTS usage_log;
DROP TABLE IF EXISTS tool_runs;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tools;
DROP TABLE IF EXISTS plans;
