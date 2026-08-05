-- +goose Up
CREATE TABLE audio_transcription_access (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE audio_transcription_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    config JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO audio_transcription_settings (id, config) VALUES (1, '{
  "model": "general",
  "language_code": "ru-RU",
  "price_rub_per_minute": 0.16
}'::jsonb);

CREATE TABLE audio_transcription_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_name TEXT NOT NULL,
    asset_path TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    duration_sec NUMERIC(10, 2),
    transcript_path TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audio_transcription_files_user_created
    ON audio_transcription_files(user_id, created_at DESC);

CREATE TABLE audio_transcription_file_runs (
    file_id UUID NOT NULL REFERENCES audio_transcription_files(id) ON DELETE CASCADE,
    run_id UUID NOT NULL UNIQUE REFERENCES tool_runs(id) ON DELETE CASCADE,
    PRIMARY KEY (file_id, run_id)
);

INSERT INTO tools (slug, name, description, enabled) VALUES
('audio-transcription', 'Audio Transcription', 'Speech-to-text transcription via Yandex SpeechKit', true);

UPDATE plans SET tool_limits = tool_limits || '{"audio-transcription": 5}'::jsonb WHERE slug = 'free';
UPDATE plans SET tool_limits = tool_limits || '{"audio-transcription": 50}'::jsonb WHERE slug = 'pro';
UPDATE plans SET tool_limits = tool_limits || '{"audio-transcription": -1}'::jsonb WHERE slug = 'business';

-- +goose Down
DELETE FROM tools WHERE slug = 'audio-transcription';
UPDATE plans SET tool_limits = tool_limits - 'audio-transcription' WHERE slug IN ('free', 'pro', 'business');
DROP TABLE IF EXISTS audio_transcription_file_runs;
DROP TABLE IF EXISTS audio_transcription_files;
DROP TABLE IF EXISTS audio_transcription_settings;
DROP TABLE IF EXISTS audio_transcription_access;
