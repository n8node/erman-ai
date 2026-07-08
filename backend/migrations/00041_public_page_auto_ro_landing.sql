-- +goose Up
INSERT INTO public_pages (slug, title, content_html, meta_description, template, sort_order)
VALUES (
    'auto-ro',
    'Авто РО',
    '',
    'Кейс Erman AI: модуль Авто РО — автоклассификация 18 буровых операций с точностью 97.5%. Гибрид алгоритмов и ML в единой системе автоматизированной буровой.',
    'auto-ro-landing',
    11
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    meta_description = EXCLUDED.meta_description,
    template = EXCLUDED.template,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

-- +goose Down
DELETE FROM public_pages WHERE slug = 'auto-ro' AND template = 'auto-ro-landing';
