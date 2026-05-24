-- +goose Up
INSERT INTO ui_tooltips (key, label, text_ru, text_en, sort_order) VALUES
(
    'calculator.result.net_benefit',
    'Чистая выгода / мес',
    'B_m — чистая ежемесячная выгода: Hm × Ch + Sm − Om. Основной показатель экономии после вычета поддержки решения.',
    'B_m — net monthly benefit: Hm × Ch + Sm − Om. Main savings metric after solution support costs.',
    210
),
(
    'calculator.result.payback',
    'Окупаемость',
    'Срок окупаемости внедрения: I₀ ÷ B_m (месяцы). Меньше 12 мес — рекомендуется автоматизировать.',
    'Implementation payback: I₀ ÷ B_m (months). Under 12 months — automation is recommended.',
    220
),
(
    'calculator.result.roi',
    'ROI за горизонт',
    'Return on Investment за выбранный горизонт: (B_m × n − I₀) ÷ I₀ × 100%. Показывает доходность вложений.',
    'Return on Investment over the selected horizon: (B_m × n − I₀) ÷ I₀ × 100%. Shows investment return.',
    230
),
(
    'calculator.result.tco',
    'TCO',
    'Total Cost of Ownership — совокупная стоимость владения за горизонт: I₀ + Om × n (внедрение + поддержка).',
    'Total Cost of Ownership over the horizon: I₀ + Om × n (implementation + support).',
    240
),
(
    'calculator.result.npv',
    'NPV',
    'Net Present Value — чистая приведённая стоимость. Дисконтирует B_m по ставке r и сравнивает с I₀. NPV > 0 — проект выгоден.',
    'Net Present Value — discounts B_m at rate r and compares to I₀. NPV > 0 means the project is worthwhile.',
    250
),
(
    'calculator.kpi.processing_time',
    'Время обработки единицы',
    'Суммарное время всех операций на одну единицу до и после автоматизации (мин/ед).',
    'Total time of all operations per unit before and after automation (min/unit).',
    310
),
(
    'calculator.kpi.throughput',
    'Объём на сотрудника',
    'Сколько единиц обрабатывает один сотрудник за день. После автоматизации растёт пропорционально сокращению времени.',
    'Units one employee processes per day. After automation it grows in proportion to time saved.',
    320
),
(
    'calculator.kpi.errors',
    'Ошибки при обработке',
    'Доля ошибок до автоматизации и прогноз после — ручной ввод устранён, ошибки стремятся к 0%.',
    'Error rate before automation and forecast after — manual entry removed, errors approach 0%.',
    330
),
(
    'calculator.kpi.monthly_hours',
    'Ежемесячно затраченное время',
    'Человеко-часы в месяц на процесс до и после. «После» = Hm × (1 − % автоматизации/100) от исходного объёма.',
    'Person-hours per month on the process before and after. "After" reflects hours remaining post-automation.',
    340
),
(
    'calculator.kpi.payroll_savings',
    'Экономия ФОТ',
    'Экономия на фонде оплаты труда: Hm × Ch — сколько рублей в месяц высвобождается с ФОТ.',
    'Payroll savings: Hm × Ch — monthly ruble amount freed from payroll.',
    350
),
(
    'calculator.kpi.payback',
    'Окупаемость проекта (KPI)',
    'Прогнозный срок окупаемости I₀ при текущей чистой выгоде B_m. Дублирует ключевую метрику в таблице KPI.',
    'Forecast payback of I₀ at current net benefit B_m. Mirrors the key metric in the KPI table.',
    360
);

-- +goose Down
DELETE FROM ui_tooltips WHERE key IN (
    'calculator.result.net_benefit',
    'calculator.result.payback',
    'calculator.result.roi',
    'calculator.result.tco',
    'calculator.result.npv',
    'calculator.kpi.processing_time',
    'calculator.kpi.throughput',
    'calculator.kpi.errors',
    'calculator.kpi.monthly_hours',
    'calculator.kpi.payroll_savings',
    'calculator.kpi.payback'
);
