-- +goose Up
ALTER TABLE users
    ADD COLUMN email_verified_at TIMESTAMPTZ;

UPDATE users SET email_verified_at = COALESCE(created_at, NOW());

CREATE TABLE smtp_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    config JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO smtp_settings (id, config) VALUES (1, '{
  "enabled": false,
  "from_email": "",
  "from_name": "Erman AI",
  "force_from_email": true,
  "force_from_name": true,
  "reply_to_from_email": false,
  "host": "",
  "port": 465,
  "encryption": "ssl",
  "auto_tls": true,
  "auth": true,
  "username": "",
  "password": ""
}'::jsonb);

CREATE TABLE email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);
CREATE INDEX idx_email_verification_tokens_hash ON email_verification_tokens(token_hash);

-- +goose Down
DROP TABLE IF EXISTS email_verification_tokens;
DROP TABLE IF EXISTS smtp_settings;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
