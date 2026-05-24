-- +goose Up
ALTER TABLE proposal_requests
    ADD COLUMN requester_name VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN telegram VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN business_note TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE proposal_requests
    DROP COLUMN IF EXISTS business_note,
    DROP COLUMN IF EXISTS telegram,
    DROP COLUMN IF EXISTS requester_name;
