-- +goose Up
INSERT INTO public_pages (slug, title, content_html, meta_description, template, sort_order)
VALUES (
    'nastavnik-ai',
    'Наставник ИИ',
    '',
    'Обучение менторов и проверка навыков: тренажёр менторской беседы с ИИ-подопечным и автоматический разбор реальных видеосессий.',
    'nastavnik-ai-landing',
    12
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    meta_description = EXCLUDED.meta_description,
    template = EXCLUDED.template,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

-- +goose Down
DELETE FROM public_pages
WHERE slug = 'nastavnik-ai' AND template = 'nastavnik-ai-landing';
