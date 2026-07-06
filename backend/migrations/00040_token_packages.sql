-- +goose Up
CREATE TABLE token_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    tokens BIGINT NOT NULL CHECK (tokens > 0),
    price_rub INTEGER NOT NULL CHECK (price_rub > 0),
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_public BOOLEAN NOT NULL DEFAULT true,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_token_balances (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE token_package_checkouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_package_id UUID NOT NULL REFERENCES token_packages(id),
    provider VARCHAR(50) NOT NULL,
    amount_rub INTEGER NOT NULL,
    tokens_amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    external_id VARCHAR(255),
    inv_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ
);

CREATE INDEX idx_token_package_checkouts_user_id ON token_package_checkouts(user_id);
CREATE INDEX idx_token_package_checkouts_status ON token_package_checkouts(status);
CREATE UNIQUE INDEX idx_token_package_checkouts_provider_external ON token_package_checkouts(provider, external_id)
    WHERE external_id IS NOT NULL;
CREATE UNIQUE INDEX idx_token_package_checkouts_robokassa_inv ON token_package_checkouts(inv_id)
    WHERE inv_id IS NOT NULL;

INSERT INTO token_packages (slug, name, tokens, price_rub, sort_order) VALUES
    ('starter-100k', '100 000 токенов', 100000, 490, 10),
    ('pro-500k', '500 000 токенов', 500000, 1990, 20),
    ('business-2m', '2 000 000 токенов', 2000000, 6990, 30);

-- +goose Down
DROP TABLE IF EXISTS token_package_checkouts;
DROP TABLE IF EXISTS user_token_balances;
DROP TABLE IF EXISTS token_packages;
