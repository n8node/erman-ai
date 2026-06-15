-- +goose Up
CREATE TABLE yandex_metrika_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    config JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO yandex_metrika_settings (id, config) VALUES (1, '{
  "enabled": false,
  "counter_code": ""
}'::jsonb);

-- +goose Down
DROP TABLE IF EXISTS yandex_metrika_settings;
