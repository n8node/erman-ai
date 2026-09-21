-- +goose Up
ALTER TABLE geological_journal_document_pages
    ADD COLUMN table_result JSONB NOT NULL DEFAULT '{"rows": []}'::jsonb;

-- +goose Down
ALTER TABLE geological_journal_document_pages
    DROP COLUMN table_result;
