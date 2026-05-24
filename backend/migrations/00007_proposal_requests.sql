-- +goose Up
CREATE TABLE proposal_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    run_id UUID NOT NULL REFERENCES tool_runs(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_proposal_requests_created_at ON proposal_requests(created_at DESC);
CREATE INDEX idx_proposal_requests_status ON proposal_requests(status);
CREATE INDEX idx_proposal_requests_user_id ON proposal_requests(user_id);

-- +goose Down
DROP TABLE IF EXISTS proposal_requests;
