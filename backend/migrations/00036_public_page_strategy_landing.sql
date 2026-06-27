-- +goose Up
INSERT INTO public_pages (slug, title, content_html, meta_description, template, sort_order)
VALUES (
    'ai-strategy',
    'AI Strategy Generator',
    '',
    'Персональная стратегия внедрения AI для вашего бизнеса: roadmap, KPI, риски и план на 30 дней. Связка с калькулятором ROI. Генерация за один сеанс.',
    'strategy-landing',
    6
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    meta_description = EXCLUDED.meta_description,
    template = EXCLUDED.template,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

-- +goose Down
DELETE FROM public_pages WHERE slug = 'ai-strategy' AND template = 'strategy-landing';
