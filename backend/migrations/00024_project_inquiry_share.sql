-- +goose Up
ALTER TABLE project_inquiries
    ADD COLUMN share_token VARCHAR(64);

CREATE UNIQUE INDEX idx_project_inquiries_share_token
    ON project_inquiries(share_token)
    WHERE share_token IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_project_inquiries_share_token;
ALTER TABLE project_inquiries
    DROP COLUMN IF EXISTS share_token;
