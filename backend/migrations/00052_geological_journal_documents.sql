-- +goose Up
CREATE TABLE geological_journal_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_name TEXT NOT NULL,
    asset_path TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL DEFAULT 'application/pdf',
    size_bytes BIGINT NOT NULL,
    page_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'uploaded',
    is_shared BOOLEAN NOT NULL DEFAULT false,
    error_msg TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    analysis_started_at TIMESTAMPTZ,
    analysis_completed_at TIMESTAMPTZ
);
CREATE INDEX idx_geological_journal_documents_owner_created
    ON geological_journal_documents(owner_id, created_at DESC);
CREATE INDEX idx_geological_journal_documents_shared_created
    ON geological_journal_documents(is_shared, created_at DESC);

CREATE TABLE geological_journal_document_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES geological_journal_documents(id) ON DELETE CASCADE,
    page_number INTEGER NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    original_asset_path TEXT,
    oriented_asset_path TEXT,
    preprocessed_asset_path TEXT,
    ocr_text TEXT NOT NULL DEFAULT '',
    ocr_tsv TEXT NOT NULL DEFAULT '',
    analysis JSONB NOT NULL DEFAULT '{}',
    content_type VARCHAR(32),
    orientation_degrees INTEGER NOT NULL DEFAULT 0,
    orientation_confidence REAL NOT NULL DEFAULT 0,
    table_count INTEGER NOT NULL DEFAULT 0,
    text_char_count INTEGER NOT NULL DEFAULT 0,
    error_msg TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(document_id, page_number)
);
CREATE INDEX idx_geological_journal_document_pages_document
    ON geological_journal_document_pages(document_id, page_number);

CREATE TABLE geological_journal_document_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES geological_journal_documents(id) ON DELETE CASCADE,
    page_id UUID NOT NULL REFERENCES geological_journal_document_pages(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    phase VARCHAR(32) NOT NULL DEFAULT 'queued',
    attempts INTEGER NOT NULL DEFAULT 0,
    lease_until TIMESTAMPTZ,
    error_msg TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    UNIQUE(page_id)
);
CREATE INDEX idx_geological_journal_document_jobs_queue
    ON geological_journal_document_jobs(status, lease_until, created_at);

CREATE TABLE geological_journal_document_llm_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES geological_journal_documents(id) ON DELETE CASCADE,
    page_id UUID NOT NULL REFERENCES geological_journal_document_pages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mode VARCHAR(32) NOT NULL,
    result JSONB NOT NULL,
    model_used VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_geological_journal_document_llm_results_page
    ON geological_journal_document_llm_results(document_id, page_id, created_at DESC);

CREATE TABLE geological_journal_document_chat_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES geological_journal_documents(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    page_numbers JSONB NOT NULL DEFAULT '[]',
    include_neighbors BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE geological_journal_document_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES geological_journal_document_chat_sessions(id) ON DELETE CASCADE,
    role VARCHAR(16) NOT NULL,
    content TEXT NOT NULL,
    sources JSONB NOT NULL DEFAULT '[]',
    confidence VARCHAR(16),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_geological_journal_document_chat_messages_session
    ON geological_journal_document_chat_messages(session_id, created_at);

-- +goose Down
DROP TABLE IF EXISTS geological_journal_document_chat_messages;
DROP TABLE IF EXISTS geological_journal_document_chat_sessions;
DROP TABLE IF EXISTS geological_journal_document_llm_results;
DROP TABLE IF EXISTS geological_journal_document_jobs;
DROP TABLE IF EXISTS geological_journal_document_pages;
DROP TABLE IF EXISTS geological_journal_documents;