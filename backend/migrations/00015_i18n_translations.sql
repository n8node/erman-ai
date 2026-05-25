-- +goose Up
CREATE TABLE ui_translations (
    key TEXT NOT NULL,
    locale VARCHAR(10) NOT NULL,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (key, locale)
);

CREATE INDEX idx_ui_translations_locale ON ui_translations (locale);
CREATE INDEX idx_ui_translations_key ON ui_translations (key);
CREATE INDEX idx_ui_translations_search ON ui_translations USING gin (to_tsvector('simple', key || ' ' || value));

ALTER TABLE ui_tooltips ADD COLUMN IF NOT EXISTS locale_texts JSONB NOT NULL DEFAULT '{}';

UPDATE ui_tooltips
SET locale_texts = jsonb_build_object('ru', text_ru, 'en', text_en)
WHERE locale_texts = '{}'::jsonb OR locale_texts IS NULL;

-- +goose Down
ALTER TABLE ui_tooltips DROP COLUMN IF EXISTS locale_texts;
DROP TABLE IF EXISTS ui_translations;
