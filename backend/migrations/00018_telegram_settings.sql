-- +goose Up
CREATE TABLE telegram_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    config JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO telegram_settings (id, config) VALUES (1, '{
  "enabled": false,
  "chat_id": "",
  "bot_token": "",
  "notify_registration": true,
  "registration_template": "🆕 Новый пользователь\nEmail: {email}\nСегмент: {accountSegment}\nРеферал: {referral}",
  "notify_email_verified": true,
  "email_verified_template": "✅ Email подтверждён\nEmail: {email}",
  "notify_payment": true,
  "payment_template": "💰 Оплата тарифа\nПользователь: {userEmail}\nТариф: {planName}\nСумма: {amount} {currency}"
}'::jsonb);

-- +goose Down
DROP TABLE IF EXISTS telegram_settings;
