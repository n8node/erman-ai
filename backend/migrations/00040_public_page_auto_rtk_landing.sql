-- +goose Up
INSERT INTO public_pages (slug, title, content_html, meta_description, template, sort_order)
VALUES (
    'auto-rtk',
    'Авто РТК',
    '',
    'Кейс Erman AI: модуль Авто РТК для автоматизированной буровой — иерархия скважин, план/факт, графики SyncDrill и согласование режимно-технологических карт.',
    'auto-rtk-landing',
    10
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    meta_description = EXCLUDED.meta_description,
    template = EXCLUDED.template,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

-- +goose Down
DELETE FROM public_pages WHERE slug = 'auto-rtk' AND template = 'auto-rtk-landing';
