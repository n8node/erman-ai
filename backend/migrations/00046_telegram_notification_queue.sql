-- +goose Up
CREATE TABLE telegram_notification_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    message_text TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempt_count INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error TEXT,
    last_attempt_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    locked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT telegram_notification_queue_status_check
        CHECK (status IN ('pending', 'processing', 'sent', 'failed'))
);

CREATE INDEX idx_telegram_notification_queue_pick
    ON telegram_notification_queue (status, next_attempt_at, created_at);

CREATE INDEX idx_telegram_notification_queue_created
    ON telegram_notification_queue (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS telegram_notification_queue;
