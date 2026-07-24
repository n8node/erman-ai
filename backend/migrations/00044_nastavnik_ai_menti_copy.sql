-- +goose Up
UPDATE public_pages
SET
    meta_description = 'Обучение менторов и проверка навыков: тренажёр менторской беседы с ИИ-менти и автоматический разбор реальных видеосессий.',
    updated_at = NOW()
WHERE slug = 'nastavnik-ai' AND template = 'nastavnik-ai-landing';

-- +goose Down
UPDATE public_pages
SET
    meta_description = 'Обучение менторов и проверка навыков: тренажёр менторской беседы с ИИ-подопечным и автоматический разбор реальных видеосессий.',
    updated_at = NOW()
WHERE slug = 'nastavnik-ai' AND template = 'nastavnik-ai-landing';
