-- +goose Up
ALTER TABLE public_pages
    ADD COLUMN IF NOT EXISTS template VARCHAR(64) NOT NULL DEFAULT '';

INSERT INTO public_pages (slug, title, content_html, meta_description, template, sort_order)
VALUES (
    'calculator-roi',
    'Калькулятор ROI автоматизации',
    '',
    'Финансовая модель Erman AI: посчитайте чистую выгоду, окупаемость, FTE и ROI от автоматизации бизнес-процессов за 3 минуты. 10 готовых шаблонов отраслей.',
    'calculator-landing',
    5
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    meta_description = EXCLUDED.meta_description,
    template = EXCLUDED.template,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

-- +goose Down
DELETE FROM public_pages WHERE slug = 'calculator-roi' AND template = 'calculator-landing';
ALTER TABLE public_pages DROP COLUMN IF EXISTS template;
