-- +goose Up
CREATE TABLE ui_tooltips (
    key VARCHAR(120) PRIMARY KEY,
    label VARCHAR(200) NOT NULL,
    text_ru TEXT NOT NULL DEFAULT '',
    text_en TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO ui_tooltips (key, label, text_ru, text_en, sort_order) VALUES
(
    'calculator.wizard.hm_mode_direct',
    'Hm уже известен',
    'Если вы уже знаете, сколько часов в месяц снимет автоматизация — включите этот режим и введите Hm напрямую, без разбора операций.',
    'If you already know how many hours per month automation will save — enable this mode and enter Hm directly, without breaking down operations.',
    10
),
(
    'calculator.wizard.col_operation',
    'Операция',
    'Название шага процесса: что делает сотрудник на одну единицу (например, «Ввод счёта в систему»).',
    'Process step name: what the employee does per unit (e.g. "Invoice entry").',
    20
),
(
    'calculator.wizard.col_executor',
    'Исполнитель',
    'Роль или должность сотрудника, выполняющего операцию. Нужна для понятности отчёта.',
    'Role or job title of the person performing the operation.',
    30
),
(
    'calculator.wizard.col_minutes',
    'Мин/ед',
    'Сколько минут занимает эта операция на одну единицу (один заказ, документ, транзакцию).',
    'Minutes this operation takes per unit (one order, document, transaction).',
    40
),
(
    'calculator.wizard.col_rate',
    '₽/мин',
    'Полная стоимость минуты работы исполнителя: зарплата + налоги + накладные, делённые на рабочие минуты.',
    'Fully loaded cost per minute: salary + taxes + overhead divided by working minutes.',
    50
),
(
    'calculator.wizard.col_cost',
    '₽/ед',
    'Стоимость операции на одну единицу. Считается автоматически: Мин/ед × ₽/мин.',
    'Cost per unit for this operation. Calculated as Min/unit × RUB/min.',
    60
),
(
    'calculator.wizard.units_per_month',
    'Единиц в месяц',
    'Сколько единиц (заказов, документов, операций) процесс обрабатывает за месяц.',
    'How many units (orders, documents, operations) the process handles per month.',
    70
),
(
    'calculator.wizard.automation_pct',
    '% автоматизации',
    'Доля ручного времени, которую уберёт автоматизация. 80% означает, что 80% текущих часов больше не нужны.',
    'Share of manual time removed by automation. 80% means 80% of current hours are no longer needed.',
    80
),
(
    'calculator.wizard.error_rate_before',
    '% ошибок до автоматизации',
    'Процент ошибок, переделок или брака до внедрения. Используется в KPI «до/после».',
    'Error or rework rate before automation. Used in before/after KPI rows.',
    90
),
(
    'calculator.wizard.throughput_before',
    'Объём на сотрудника / день',
    'Сколько единиц один сотрудник обрабатывает за рабочий день до автоматизации.',
    'How many units one employee processes per working day before automation.',
    100
),
(
    'calculator.wizard.hm_preview',
    'Hm — экономия часов',
    'Hm — человеко-часы в месяц, которые снимает автоматизация. Формула: (сумма мин/ед × единиц/мес × % автоматизации) ÷ 60, либо прямой ввод.',
    'Hm — person-hours per month saved by automation. Formula: (total min/unit × units/month × automation %) ÷ 60, or direct input.',
    110
);

-- +goose Down
DROP TABLE IF EXISTS ui_tooltips;
