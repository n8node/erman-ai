-- +goose Up
ALTER TABLE project_inquiries
    ADD COLUMN calculator_snapshot JSONB;

-- +goose Down
ALTER TABLE project_inquiries
    DROP COLUMN IF EXISTS calculator_snapshot;
