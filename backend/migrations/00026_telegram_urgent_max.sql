-- +goose Up
CREATE TABLE max_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    config JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO max_settings (id, config) VALUES (1, '{
  "enabled": false,
  "bot_token": "",
  "bot_username": "",
  "notify_user_id": "",
  "notify_chat_id": "",
  "urgent_alerts_enabled": true
}'::jsonb);

CREATE TABLE telegram_bot_user_state (
    user_chat_id VARCHAR(32) PRIMARY KEY,
    mode VARCHAR(32) NOT NULL DEFAULT 'idle',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE telegram_urgent_sends (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_chat_id VARCHAR(32) NOT NULL,
    message_preview VARCHAR(500) NOT NULL DEFAULT '',
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_telegram_urgent_sends_user_day
    ON telegram_urgent_sends (user_chat_id, sent_at DESC);

UPDATE telegram_settings
SET config = config || '{
  "urgent_enabled": true,
  "urgent_email": "erman.ai@yandex.ru"
}'::jsonb
WHERE id = 1;

-- +goose Down
DROP TABLE IF EXISTS telegram_urgent_sends;
DROP TABLE IF EXISTS telegram_bot_user_state;
DROP TABLE IF EXISTS max_settings;
UPDATE telegram_settings
SET config = config - 'urgent_enabled' - 'urgent_email' - 'urgent_instruction'
WHERE id = 1;
