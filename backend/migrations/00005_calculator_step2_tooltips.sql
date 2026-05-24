-- +goose Up
INSERT INTO ui_tooltips (key, label, text_ru, text_en, sort_order) VALUES
(
    'calculator.wizard.hm_label',
    'Hm — экономия часов/мес',
    'Человеко-часы в месяц, которые снимает автоматизация. Подтягивается с шага 1 — можно скорректировать вручную перед расчётом.',
    'Person-hours per month saved by automation. Carried over from step 1 — you can adjust before calculating.',
    120
),
(
    'calculator.wizard.ch',
    'Ch — полная стоимость часа',
    'Полная стоимость часа сотрудника (Ch): зарплата + налоги + бонусы + накладные, делённые на рабочие часы. Используется для расчёта экономии на ФОТ.',
    'Fully loaded hourly cost (Ch): salary + taxes + bonuses + overhead divided by working hours. Used to calculate payroll savings.',
    130
),
(
    'calculator.wizard.om',
    'Om — поддержка решения / мес',
    'Ежемесячные расходы на поддержку решения: VPS, домен, лицензии, техподдержка. Вычитается из чистой выгоды B_m.',
    'Monthly solution support costs: VPS, domain, licenses, maintenance. Subtracted from net benefit B_m.',
    140
),
(
    'calculator.wizard.capex',
    'I₀ — бюджет внедрения',
    'Разовые затраты на внедрение: аудит, разработка, обучение, интеграции. Используется для окупаемости, ROI и NPV.',
    'One-time implementation costs: audit, development, training, integrations. Used for payback, ROI and NPV.',
    150
),
(
    'calculator.wizard.horizon',
    'Горизонт оценки, мес',
    'Период, за который считаются ROI, TCO и NPV. Обычно 12, 24 или 36 месяцев — в зависимости от горизонта планирования.',
    'Period over which ROI, TCO and NPV are calculated. Typically 12, 24 or 36 months depending on planning horizon.',
    160
),
(
    'calculator.wizard.sm',
    'Sm — прочая выгода / мес',
    'Дополнительная ежемесячная выгода помимо экономии на ФОТ: снижение штрафов, рост качества, ускорение оборота.',
    'Additional monthly benefit beyond payroll savings: fewer penalties, quality gains, faster turnaround.',
    170
),
(
    'calculator.wizard.utilization',
    'L — коэфф. загрузки',
    'Коэффициент загрузки L (0–1). В модели по умолчанию 0,8 — сотрудник занят 80% рабочего времени. Нужен для расчёта FTE.',
    'Utilization factor L (0–1). Default 0.8 means 80% of working time is productive. Used to calculate FTE.',
    180
),
(
    'calculator.wizard.discount',
    'r — ставка дисконтирования',
    'Годовая ставка дисконтирования r для расчёта NPV. По умолчанию 20% — типичная ставка для оценки инвестиций в автоматизацию.',
    'Annual discount rate r for NPV calculation. Default 20% — typical rate for automation investment appraisal.',
    190
),
(
    'calculator.wizard.fte',
    'FTE',
    'Full-Time Equivalent — эквивалент полных ставок, высвобождаемых автоматизацией. Формула: FTE = Hm ÷ (168 × L), где 168 — рабочие часы в месяц.',
    'Full-Time Equivalent — full-time positions freed by automation. Formula: FTE = Hm ÷ (168 × L), where 168 is working hours per month.',
    200
);

-- +goose Down
DELETE FROM ui_tooltips WHERE key IN (
    'calculator.wizard.hm_label',
    'calculator.wizard.ch',
    'calculator.wizard.om',
    'calculator.wizard.capex',
    'calculator.wizard.horizon',
    'calculator.wizard.sm',
    'calculator.wizard.utilization',
    'calculator.wizard.discount',
    'calculator.wizard.fte'
);
