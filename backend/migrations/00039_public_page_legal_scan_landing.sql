-- +goose Up
INSERT INTO public_pages (slug, title, content_html, meta_description, template, sort_order)
VALUES (
    'legal-scan',
    'Legal-скан сайта',
    '',
    'Legal-скан сайта Erman AI: проверка по 152-ФЗ за 3 минуты — 12 автоматических проверок, риски, штрафы и рекомендации. PDF и переход в КП.',
    'legal-scan-landing',
    9
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    meta_description = EXCLUDED.meta_description,
    template = EXCLUDED.template,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

-- +goose Down
DELETE FROM public_pages WHERE slug = 'legal-scan' AND template = 'legal-scan-landing';
