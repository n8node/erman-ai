-- +goose Up
CREATE TABLE geological_journal_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    config JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO geological_journal_settings (id, config) VALUES (1, '{
  "provider": "openrouter",
  "openrouter_model": "google/gemini-2.5-flash",
  "yandex_model": "yandexgpt/latest",
  "system_prompt": "",
  "temperature": 0.1,
  "max_tokens": 8192
}'::jsonb);

CREATE TABLE geological_journal_access (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE geological_journal_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_name TEXT NOT NULL,
    asset_path TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    latest_result JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_geological_journal_pages_user_created
    ON geological_journal_pages(user_id, created_at DESC);

CREATE TABLE geological_journal_page_runs (
    page_id UUID NOT NULL REFERENCES geological_journal_pages(id) ON DELETE CASCADE,
    run_id UUID NOT NULL UNIQUE REFERENCES tool_runs(id) ON DELETE CASCADE,
    PRIMARY KEY (page_id, run_id)
);

CREATE TABLE geological_journal_result_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL REFERENCES geological_journal_pages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    result JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_geological_journal_versions_page_created
    ON geological_journal_result_versions(page_id, created_at DESC);

CREATE TABLE geological_journal_examples (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    asset_path TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO tools (slug, name, description, enabled) VALUES
('geological-journal', 'Geological Journal', 'Vision recognition of geological drilling journal pages', true);

-- +goose Down
DELETE FROM tools WHERE slug = 'geological-journal';
DROP TABLE IF EXISTS geological_journal_examples;
DROP TABLE IF EXISTS geological_journal_result_versions;
DROP TABLE IF EXISTS geological_journal_page_runs;
DROP TABLE IF EXISTS geological_journal_pages;
DROP TABLE IF EXISTS geological_journal_access;
DROP TABLE IF EXISTS geological_journal_settings;
