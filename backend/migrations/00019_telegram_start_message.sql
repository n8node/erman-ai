-- +goose Up
UPDATE telegram_settings
SET config = config || '{
  "start_enabled": false,
  "start_text": "Добро пожаловать в Erman AI!\n\nЗдесь вы получите уведомления и сможете связаться с командой.",
  "start_image_filename": ""
}'::jsonb
WHERE id = 1;

-- +goose Down
UPDATE telegram_settings
SET config = config - 'start_enabled' - 'start_text' - 'start_image_filename'
WHERE id = 1;
