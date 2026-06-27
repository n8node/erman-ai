-- +goose Up
INSERT INTO public_pages (slug, title, content_html, meta_description, template, sort_order)
VALUES (
    'ai-audit',
    'Automation Audit Report',
    '',
    'Аудит автоматизации Erman AI: приоритизация 2–7 процессов, automation score, quick wins и roadmap. Оценка экономии за один сеанс.',
    'audit-landing',
    8
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    meta_description = EXCLUDED.meta_description,
    template = EXCLUDED.template,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

-- +goose Down
DELETE FROM public_pages WHERE slug = 'ai-audit' AND template = 'audit-landing';
