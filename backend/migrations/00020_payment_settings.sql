-- +goose Up
CREATE TABLE payment_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    config JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO payment_settings (id, config) VALUES (1, '{
  "active_provider": "yookassa",
  "yookassa": {
    "shop_id": "",
    "secret_key": "",
    "return_url": "",
    "enabled": false
  },
  "robokassa": {
    "merchant_login": "",
    "password1": "",
    "password2": "",
    "test_mode": false,
    "enabled": false
  }
}'::jsonb);

-- +goose Down
DROP TABLE IF EXISTS payment_settings;
