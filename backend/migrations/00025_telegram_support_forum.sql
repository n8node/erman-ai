-- +goose Up
CREATE TABLE telegram_support_threads (
    user_chat_id VARCHAR(32) PRIMARY KEY,
    forum_chat_id VARCHAR(32) NOT NULL,
    topic_id INTEGER NOT NULL,
    display_name VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_telegram_support_threads_forum_topic
    ON telegram_support_threads (forum_chat_id, topic_id);

UPDATE telegram_settings
SET config = config || '{
  "support_enabled": false,
  "support_forum_chat_id": "",
  "dashboard_url": "https://erman.ai/dashboard/"
}'::jsonb
WHERE id = 1;

-- +goose Down
DROP TABLE IF EXISTS telegram_support_threads;
UPDATE telegram_settings
SET config = config - 'support_enabled' - 'support_forum_chat_id' - 'dashboard_url'
WHERE id = 1;
