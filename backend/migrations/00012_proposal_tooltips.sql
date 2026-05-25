-- +goose Up
INSERT INTO ui_tooltips (key, label, text_ru, text_en, sort_order) VALUES
(
    'proposal.wizard.client_problem',
    'Proposal — описание проблемы',
    'Опишите боль клиента своими словами: что не работает, сколько времени или денег теряется, какие последствия. Это попадёт в блок «Понимание задачи» в КП.',
    'Describe the client pain in plain language: what is broken, how much time or money is lost, and the business impact. This feeds the «Understanding the task» section.',
    200
),
(
    'proposal.wizard.calculator_run',
    'Proposal — расчёт из калькулятора',
    'Привяжите готовый ROI-расчёт из Automation Calculator — модель подставит цифры экономии, окупаемости и NPV в текст КП.',
    'Link a completed Automation Calculator run — the model will weave savings, payback, and NPV figures into the proposal.',
    210
),
(
    'proposal.wizard.solution_description',
    'Proposal — описание решения',
    'Что именно вы предлагаете сделать: scope на высоком уровне, ключевые технологии или подход. Не дублируйте deliverables — здесь «зачем и как».',
    'What you propose to deliver at a high level: approach, stack, or methodology. Do not repeat deliverables — this is the «why and how».',
    220
),
(
    'proposal.wizard.deliverables',
    'Proposal — что получит клиент',
    'Конкретные артеfacts и результаты: документы, интеграции, обучение, поддержка. Каждый пункт — отдельная строка; их можно добавить кнопкой ниже.',
    'Concrete artifacts and outcomes: documents, integrations, training, support. One item per line; add more with the button below.',
    230
);

-- +goose Down
DELETE FROM ui_tooltips WHERE key IN (
    'proposal.wizard.client_problem',
    'proposal.wizard.calculator_run',
    'proposal.wizard.solution_description',
    'proposal.wizard.deliverables'
);
