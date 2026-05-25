-- +goose Up
INSERT INTO ui_tooltips (key, label, text_ru, text_en, sort_order) VALUES
(
    'proposal.wizard.scenario',
    'Proposal — сценарий',
    'Выберите ситуацию: был ли уже контакт с клиентом, это первое письмо или вы сами предлагаете автоматизацию после расчёта ROI. Тон всего документа зависит от сценария.',
    'Choose the situation: prior client contact, first cold outreach, or proactive offer after ROI analysis. The entire document tone depends on the scenario.',
    195
),
(
    'proposal.wizard.scenario_after_contact',
    'Proposal — после контакта',
    'Клиент уже общался с вами: созвон, письмо, бриф или запрос КП. Документ ссылается на обсуждение и ведёт к старту проекта.',
    'You already spoke with the client: call, email, brief, or quote request. The document references the discussion and moves toward project start.',
    196
),
(
    'proposal.wizard.scenario_cold_outreach',
    'Proposal — холодное обращение',
    'Первое касание: контакта не было. Без «спасибо за обращение» и «как мы обсудили». CTA — короткая встреча, не подписание договора.',
    'First touch: no prior contact. No “thank you for reaching out” or “as we discussed”. CTA is a short meeting, not contract signing.',
    197
),
(
    'proposal.wizard.scenario_proactive_offer',
    'Proposal — проактивное предложение',
    'Вы сами предлагаете автоматизацию после анализа (часто из калькулятора). Акцент на ROI и процессе, клиент не запрашивал КП.',
    'You proactively offer automation after analysis (often from Calculator). Focus on ROI and process; the client did not request a quote.',
    198
),
(
    'proposal.wizard.prior_contact_summary',
    'Proposal — контекст контакта',
    'Кратко: когда и о чём общались, что просил клиент, договорённости. Модель использует это в обращении и блоке «понимание задачи».',
    'Briefly: when you spoke, what the client asked, any agreements. The model uses this in the opening and task understanding sections.',
    199
),
(
    'proposal.wizard.problem_source',
    'Proposal — источник гипотезы',
    'Откуда вы знаете о проблеме без контакта: наблюдение процесса, отраслевой паттерн, рекомендация, публичные данные. Не выдумывайте факты — опишите основание.',
    'How you know about the problem without contact: process observation, industry pattern, referral, public data. Do not invent facts — describe your basis.',
    200
),
(
    'proposal.wizard.include_pricing',
    'Proposal — указать цену',
    'Включено: в КП будут конкретная сумма и условия оплаты. Выключено («мягкое КП»): цена после диагностики, CTA — встреча. Удобно для холодного первого письма.',
    'On: exact price and payment terms in the document. Off (“soft CP”): price after discovery, CTA is a meeting. Useful for cold first touch.',
    201
),
(
    'proposal.wizard.client_company',
    'Proposal — компания клиента',
    'Юридическое или торговое название компании, которой адресовано КП. В холодном сценарии можно обращаться к компании без ФИО контакта.',
    'Legal or trade name of the company receiving the proposal. In cold outreach you may address the company without a contact name.',
    202
),
(
    'proposal.wizard.client_contact',
    'Proposal — контакт клиента',
    'ФИО или должность получателя. Необязательно для холодного сценария. Если указано — используется в обращении без имитации прошлого диалога (если не «после контакта»).',
    'Recipient name or title. Optional for cold outreach. If set, used in salutation without implying prior dialogue (unless after-contact scenario).',
    203
),
(
    'proposal.wizard.client_industry',
    'Proposal — отрасль',
    'Помогает модели подобрать отраслевые формулировки и типовые боли.',
    'Helps the model choose industry-appropriate language and typical pain points.',
    204
),
(
    'proposal.wizard.solution_name',
    'Proposal — название решения',
    'Короткое название проекта для заголовка КП, например «Автоматизация обработки заявок».',
    'Short project title for the proposal header, e.g. “Lead intake automation”.',
    205
),
(
    'proposal.wizard.project_cost',
    'Proposal — стоимость проекта',
    'Используется только если включено «Указать цену». Сумма в ₽ для блока стоимости и PDF.',
    'Used only when “Include pricing” is on. Amount in RUB for cost section and PDF.',
    206
),
(
    'proposal.wizard.timeline_weeks',
    'Proposal — срок в неделях',
    'Общая длительность проекта. Этапы в документе должны укладываться в этот срок.',
    'Total project duration. Timeline phases in the document should fit this duration.',
    207
),
(
    'proposal.wizard.payment_schedule',
    'Proposal — условия оплаты',
    'Схема оплаты для полного КП. Игнорируется в «мягком» режиме без цены.',
    'Payment scheme for full pricing mode. Ignored in soft mode without price.',
    208
),
(
    'proposal.wizard.sender_company',
    'Proposal — ваша компания',
    'Компания интегратора или консультанта, от имени которого отправляется КП.',
    'Integrator or consultant company sending the proposal.',
    209
),
(
    'proposal.wizard.sender_contact',
    'Proposal — ваш контакт',
    'ФИО отправителя для подписи и блока «следующий шаг».',
    'Sender name for signature and next-step block.',
    210
),
(
    'proposal.wizard.sender_phone',
    'Proposal — телефон отправителя',
    'Контактный телефон в подписи документа.',
    'Contact phone in document signature.',
    211
),
(
    'proposal.wizard.sender_email',
    'Proposal — email отправителя',
    'Email для ответа клиента.',
    'Email for client replies.',
    212
);

-- +goose Down
DELETE FROM ui_tooltips WHERE key LIKE 'proposal.wizard.%' AND key IN (
    'proposal.wizard.scenario',
    'proposal.wizard.scenario_after_contact',
    'proposal.wizard.scenario_cold_outreach',
    'proposal.wizard.scenario_proactive_offer',
    'proposal.wizard.prior_contact_summary',
    'proposal.wizard.problem_source',
    'proposal.wizard.include_pricing',
    'proposal.wizard.client_company',
    'proposal.wizard.client_contact',
    'proposal.wizard.client_industry',
    'proposal.wizard.solution_name',
    'proposal.wizard.project_cost',
    'proposal.wizard.timeline_weeks',
    'proposal.wizard.payment_schedule',
    'proposal.wizard.sender_company',
    'proposal.wizard.sender_contact',
    'proposal.wizard.sender_phone',
    'proposal.wizard.sender_email'
);
