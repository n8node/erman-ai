-- +goose Up
CREATE SEQUENCE plan_checkout_inv_id_seq START WITH 1000;

CREATE TABLE plan_checkouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id),
    provider VARCHAR(50) NOT NULL,
    amount_rub INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    external_id VARCHAR(255),
    inv_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ
);

CREATE INDEX idx_plan_checkouts_user_id ON plan_checkouts(user_id);
CREATE INDEX idx_plan_checkouts_status ON plan_checkouts(status);
CREATE UNIQUE INDEX idx_plan_checkouts_provider_external ON plan_checkouts(provider, external_id)
    WHERE external_id IS NOT NULL;
CREATE UNIQUE INDEX idx_plan_checkouts_robokassa_inv ON plan_checkouts(inv_id)
    WHERE inv_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS plan_checkouts;
DROP SEQUENCE IF EXISTS plan_checkout_inv_id_seq;
