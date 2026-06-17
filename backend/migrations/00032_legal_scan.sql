-- +goose Up
CREATE TABLE legal_risks (
    risk_id           TEXT PRIMARY KEY,
    title_ru          TEXT NOT NULL,
    title_en          TEXT NOT NULL DEFAULT '',
    what_ru           TEXT NOT NULL,
    article           TEXT NOT NULL,
    fine_text_ru      TEXT NOT NULL,
    fine_min          INTEGER,
    fine_max          INTEGER,
    severity          TEXT NOT NULL CHECK (severity IN ('high', 'medium', 'low')),
    how_to_fix_ru     TEXT NOT NULL,
    how_to_fix_en     TEXT NOT NULL DEFAULT '',
    trigger_findings  JSONB NOT NULL DEFAULT '{}',
    trigger_flags     JSONB NOT NULL DEFAULT '{}',
    is_turnover_fine  BOOLEAN NOT NULL DEFAULT false,
    is_context_only   BOOLEAN NOT NULL DEFAULT false,
    is_active         BOOLEAN NOT NULL DEFAULT true,
    sort_order        INTEGER NOT NULL DEFAULT 0,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO legal_risks (
    risk_id, title_ru, what_ru, article, fine_text_ru, fine_min, fine_max, severity,
    how_to_fix_ru, trigger_findings, trigger_flags, is_turnover_fine, is_context_only, sort_order
) VALUES
(
    'no_privacy_policy',
    'Нет политики обработки персональных данных',
    'Нет общедоступной политики обработки ПД',
    'ч.10 ст. 13.11 КоАП / ст. 18.1 152-ФЗ',
    'для юрлиц 30 000–60 000 ₽, для ИП 10 000–20 000 ₽',
    30000, 60000, 'medium',
    'Опубликовать политику обработки ПД и поставить ссылку в подвал сайта.',
    '{"privacy_policy": false}',
    '{}',
    false, false, 20
),
(
    'no_consent',
    'Обработка персональных данных без согласия',
    'Обработка ПД без согласия',
    'ч.1/ч.2 ст. 13.11 КоАП',
    'до 300 000 ₽ (первое), до 1 500 000 ₽ (повторное)',
    0, 300000, 'high',
    'Добавить чекбокс согласия со ссылкой на политику обработки ПД ко всем формам.',
    '{"form_consent": false, "forms_collect_pd": true}',
    '{"has_forms": true}',
    false, false, 10
),
(
    'no_rkn_notice',
    'Не уведомлён Роскомнадзор об обработке ПД',
    'Не уведомлён Роскомнадзор о намерении обрабатывать ПД',
    'ч.10 ст. 13.11 КоАП',
    'для организаций 100 000–300 000 ₽',
    100000, 300000, 'medium',
    'Подать уведомление оператора персональных данных в Роскомнадзор.',
    '{"forms_collect_pd": true}',
    '{"has_forms": true}',
    false, false, 15
),
(
    'data_leak_exposure',
    'Риск ответственности за утечку персональных данных',
    'Риск ответственности за утечку ПД',
    'ст. 13.11 КоАП (новые части)',
    'до 15 000 000 ₽; при повторной утечке — оборотный штраф 1–3% выручки (от 20 до 500 млн ₽)',
    NULL, NULL, 'high',
    'Усилить меры защиты ПД, провести аудит безопасности и обновить политики.',
    '{"forms_collect_pd": true}',
    '{"has_forms": true}',
    true, true, 5
),
(
    'no_cookie_banner',
    'Нет cookie-баннера',
    'Нет информирования/согласия на cookie',
    'производно от 152-ФЗ (cookie с ПД)',
    'в рамках штрафов за обработку ПД без согласия',
    0, 0, 'medium',
    'Подключить cookie-баннер с возможностью отказа и ссылкой на политику cookie.',
    '{"cookie_banner": false, "has_trackers": true}',
    '{}',
    false, false, 25
),
(
    'no_ad_marking',
    'Реклама без маркировки (токен erid)',
    'Реклама без токена erid / пометки «Реклама»',
    'ст. 14.3 КоАП (РКН + ФАС)',
    'для юрлиц — крупные, накопительно по числу публикаций',
    0, 500000, 'high',
    'Регистрировать креативы в ОРД, получать erid и ставить пометку «Реклама».',
    '{"ad_marking": false}',
    '{"has_traffic_from_ads": true}',
    false, false, 12
),
(
    'no_offer',
    'Нет публичной оферты',
    'Нет публичной оферты / условий продажи',
    'ГК РФ, ЗоЗПП',
    'фикс. штрафа нет — риск споров и претензий потребителей',
    0, 0, 'medium',
    'Опубликовать оферту с условиями оплаты, доставки и возврата.',
    '{"offer": false}',
    '{"has_online_sales": true}',
    false, false, 30
),
(
    'no_ssl',
    'Передача данных без шифрования (нет HTTPS)',
    'Передача ПД без шифрования',
    'ст. 19 152-ФЗ (меры защиты)',
    'усиливает риск утечки и претензий',
    0, 0, 'medium',
    'Установить SSL-сертификат и перевести сайт на HTTPS.',
    '{"ssl": false}',
    '{}',
    false, false, 35
),
(
    'no_requisites',
    'Не указаны реквизиты продавца',
    'Нет реквизитов продавца',
    'ст. 9 ЗоЗПП',
    'риск претензий, снижает доверие',
    0, 0, 'low',
    'Добавить в подвал наименование, ИНН/ОГРН и контактные данные.',
    '{"requisites_inn": false}',
    '{}',
    false, false, 40
),
(
    'no_foreign_loc',
    'Иностранные сервисы и локализация данных',
    'Возможное нарушение требований о локализации ПД в РФ',
    'ч.5 ст. 18 152-ФЗ',
    'ограничение доступа + штрафы за нарушение локализации',
    0, 0, 'medium',
    'Хранить ПД россиян на серверах в РФ; проверить используемые сервисы.',
    '{"foreign_trackers": true}',
    '{"has_foreign_services": true}',
    false, false, 45
);

CREATE TABLE legal_scan_llm_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    config JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO legal_scan_llm_settings (id, config) VALUES (1, '{
  "provider": "openrouter",
  "openrouter_model": "google/gemini-flash-1.5-8b",
  "deepseek_model": "deepseek-chat",
  "system_prompt": "",
  "temperature": 0.3,
  "max_tokens": 4096
}'::jsonb);

INSERT INTO tools (slug, name, description, enabled) VALUES
('legal-scan', 'Legal Site Scan', 'RF legal risk scan for client websites', true);

UPDATE plans SET tool_limits = tool_limits || '{"legal-scan": 3}'::jsonb WHERE slug = 'free';
UPDATE plans SET tool_limits = tool_limits || '{"legal-scan": 20}'::jsonb WHERE slug = 'pro';
UPDATE plans SET tool_limits = tool_limits || '{"legal-scan": -1}'::jsonb WHERE slug = 'business';

-- +goose Down
DELETE FROM tools WHERE slug = 'legal-scan';
UPDATE plans SET tool_limits = tool_limits - 'legal-scan' WHERE slug IN ('free', 'pro', 'business');
DROP TABLE IF EXISTS legal_scan_llm_settings;
DROP TABLE IF EXISTS legal_risks;
