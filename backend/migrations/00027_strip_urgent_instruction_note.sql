-- +goose Up
UPDATE telegram_settings
SET config = jsonb_set(
    config,
    '{urgent_instruction}',
    to_jsonb(
        trim(both from replace(
            replace(config->>'urgent_instruction', E'(Сейчас доставка настроена на email, Telegram и MAX; VK и Instagram подключим позже.)\n\n', ''),
            E'(Сейчас доставка настроена на email, Telegram и MAX; VK и Instagram подключим позже.)', ''
        ))
    )
)
WHERE id = 1
  AND config->>'urgent_instruction' LIKE '%Сейчас доставка настроена%';

-- +goose Down
-- no-op: text change not reverted
