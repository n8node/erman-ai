-- +goose Up
INSERT INTO public_pages (slug, title, content_html, meta_description, template, sort_order)
VALUES (
    'ai-proposal',
    'Proposal Generator',
    '',
    'Генератор коммерческих предложений Erman AI: готовое КП за 5 минут — scope, timeline, стоимость и ROI из калькулятора. PDF и DOCX.',
    'proposal-landing',
    7
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    meta_description = EXCLUDED.meta_description,
    template = EXCLUDED.template,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

-- +goose Down
DELETE FROM public_pages WHERE slug = 'ai-proposal' AND template = 'proposal-landing';
