/**
 * Adds proposal tooltip fallback strings to all locale message files.
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const messagesDir = path.join(__dirname, "../frontend/src/messages");

const tooltipsFallback = {
  en: {
    scenario:
      "Choose the situation: prior client contact, first outreach, or a proactive offer after ROI analysis. The tone of the entire document depends on the scenario.",
    scenario_after_contact:
      "The client already spoke with you: call, email, brief, or quote request. The document references the discussion and leads to project kickoff.",
    scenario_cold_outreach:
      "First touch with no prior contact. No “thanks for reaching out” or “as we discussed”. CTA is a short meeting, not contract signing.",
    scenario_proactive_offer:
      "You propose automation after analysis (often from the calculator). Focus on ROI and the process — the client did not request a quote.",
    prior_contact_summary:
      "Briefly: when and what you discussed, what the client asked for, agreements. The model uses this in the greeting and task understanding block.",
    problem_source:
      "How you know about the problem without contact: process observation, industry pattern, referral, public data. Do not invent facts — describe the basis.",
    include_pricing:
      "On: the proposal includes a specific price and payment terms. Off (soft CP): price after diagnostics, CTA is a meeting. Useful for cold first emails.",
    client_company:
      "Legal or trade name of the company the proposal is addressed to. In cold outreach you can address the company without a contact name.",
    client_contact:
      "Recipient name or title. Optional for cold outreach. If provided — used in the greeting without faking prior dialogue (unless “after contact”).",
    client_industry: "Helps the model choose industry wording and typical pain points.",
    client_problem:
      "Describe the client pain in your words: what fails, how much time or money is lost, consequences. Goes into the “Understanding the task” block.",
    calculator_run:
      "Link a ready ROI run from Automation Calculator — the model inserts savings, payback, and NPV figures into the proposal text.",
    solution_name: 'Short project title for the proposal header, e.g. “Lead intake automation”.',
    solution_description:
      "What you propose to do: high-level scope, key technologies or approach. Do not duplicate deliverables — this is the why and how.",
    deliverables:
      "Concrete artifacts and outcomes: documents, integrations, training, support. One item per line.",
    project_cost: 'Used only when “Include price” is on. Amount in ₽ for the cost block and PDF.',
    timeline_weeks: "Total project duration. Phases in the document should fit this timeline.",
    payment_schedule: "Payment scheme for a full proposal. Ignored in soft mode without price.",
    sender_company: "Integrator or consultant company sending the proposal.",
    sender_contact: "Sender name for signature and next-step block.",
    sender_phone: "Contact phone in the document signature.",
    sender_email: "Email for the client reply.",
  },
  ru: {
    scenario:
      "Выберите ситуацию: был ли уже контакт с клиентом, это первое письмо или вы сами предлагаете автоматизацию после расчёта ROI. Тон всего документа зависит от сценария.",
    scenario_after_contact:
      "Клиент уже общался с вами: созвон, письмо, бриф или запрос КП. Документ ссылается на обсуждение и ведёт к старту проекта.",
    scenario_cold_outreach:
      "Первое касание: контакта не было. Без «спасибо за обращение» и «как мы обсудили». CTA — короткая встреча, не подписание договора.",
    scenario_proactive_offer:
      "Вы сами предлагаете автоматизацию после анализа (часто из калькулятора). Акцент на ROI и процессе, клиент не запрашивал КП.",
    prior_contact_summary:
      "Кратко: когда и о чём общались, что просил клиент, договорённости. Модель использует это в обращении и блоке «понимание задачи».",
    problem_source:
      "Откуда вы знаете о проблеме без контакта: наблюдение процесса, отраслевой паттерн, рекомендация, публичные данные. Не выдумывайте факты — опишите основание.",
    include_pricing:
      "Включено: в КП будут конкретная сумма и условия оплаты. Выключено («мягкое КП»): цена после диагностики, CTA — встреча. Удобно для холодного первого письма.",
    client_company:
      "Юридическое или торговое название компании, которой адресовано КП. В холодном сценарии можно обращаться к компании без ФИО контакта.",
    client_contact:
      "ФИО или должность получателя. Необязательно для холодного сценария. Если указано — используется в обращении без имитации прошлого диалога (если не «после контакта»).",
    client_industry: "Помогает модели подобрать отраслевые формулировки и типовые боли.",
    client_problem:
      "Опишите боль клиента своими словами: что не работает, сколько времени или денег теряется, какие последствия. Это попадёт в блок «Понимание задачи» в КП.",
    calculator_run:
      "Привяжите готовый ROI-расчёт из Automation Calculator — модель подставит цифры экономии, окупаемости и NPV в текст КП.",
    solution_name: "Короткое название проекта для заголовка КП, например «Автоматизация обработки заявок».",
    solution_description:
      "Что именно вы предлагаете сделать: scope на высоком уровне, ключевые технологии или подход. Не дублируйте deliverables — здесь «зачем и как».",
    deliverables:
      "Конкретные артефакты и результаты: документы, интеграции, обучение, поддержка. Каждый пункт — отдельная строка.",
    project_cost: "Используется только если включено «Указать цену». Сумма в ₽ для блока стоимости и PDF.",
    timeline_weeks: "Общая длительность проекта. Этапы в документе должны укладываться в этот срок.",
    payment_schedule: "Схема оплаты для полного КП. Игнорируется в «мягком» режиме без цены.",
    sender_company: "Компания интегратора или консультанта, от имени которого отправляется КП.",
    sender_contact: "ФИО отправителя для подписи и блока «следующий шаг».",
    sender_phone: "Контактный телефон в подписи документа.",
    sender_email: "Email для ответа клиента.",
  },
  de: {
    scenario:
      "Wählen Sie die Situation: bestand bereits Kontakt, erstes Anschreiben oder proaktives Angebot nach ROI-Analyse. Der Ton des gesamten Dokuments hängt vom Szenario ab.",
    scenario_after_contact:
      "Der Kunde hat bereits mit Ihnen gesprochen: Anruf, E-Mail, Briefing oder Angebotsanfrage. Das Dokument bezieht sich auf das Gespräch und führt zum Projektstart.",
    scenario_cold_outreach:
      "Erstkontakt ohne vorherige Kommunikation. Kein „Danke für Ihre Anfrage“ oder „wie besprochen“. CTA ist ein kurzes Meeting, kein Vertragsabschluss.",
    scenario_proactive_offer:
      "Sie schlagen Automatisierung nach einer Analyse vor (oft aus dem Rechner). Fokus auf ROI und Prozess — der Kunde hat kein Angebot angefordert.",
    prior_contact_summary:
      "Kurz: wann und worüber gesprochen, was der Kunde wollte, Vereinbarungen. Das Modell nutzt dies in Anrede und Aufgabenverständnis.",
    problem_source:
      "Woher Sie das Problem ohne Kontakt kennen: Prozessbeobachtung, Branchenmuster, Empfehlung, öffentliche Daten. Keine erfundenen Fakten — Grundlage beschreiben.",
    include_pricing:
      "Ein: konkreter Preis und Zahlungsbedingungen im Angebot. Aus (weiches Angebot): Preis nach Diagnose, CTA ist ein Meeting. Nützlich für kalte Erstansprache.",
    client_company:
      "Rechtlicher oder Handelsname des Unternehmens. Bei Kaltakquise kann die Firma ohne Ansprechpartner angesprochen werden.",
    client_contact:
      "Name oder Rolle des Empfängers. Optional bei Kaltakquise. Falls angegeben — in der Anrede ohne fiktiven Dialog (außer „nach Kontakt“).",
    client_industry: "Hilft dem Modell bei branchenspezifischen Formulierungen und typischen Schmerzpunkten.",
    client_problem:
      "Beschreiben Sie das Problem: was nicht funktioniert, Zeit- oder Geldverlust, Folgen. Landet im Block „Aufgabenverständnis“.",
    calculator_run:
      "Verknüpfen Sie einen ROI-Lauf aus dem Automation Calculator — das Modell fügt Einsparungen, Amortisation und NPV ein.",
    solution_name: "Kurzer Projekttitel für die Überschrift, z. B. „Automatisierung der Lead-Bearbeitung“.",
    solution_description:
      "Was Sie vorschlagen: Scope auf hoher Ebene, Technologien oder Ansatz. Deliverables nicht wiederholen — hier das Warum und Wie.",
    deliverables:
      "Konkrete Ergebnisse: Dokumente, Integrationen, Schulung, Support. Ein Punkt pro Zeile.",
    project_cost: "Nur wenn „Preis angeben“ aktiv ist. Betrag in ₽ für Kostenblock und PDF.",
    timeline_weeks: "Gesamtdauer des Projekts. Phasen müssen in diese Frist passen.",
    payment_schedule: "Zahlungsschema für ein vollständiges Angebot. Ignoriert im weichen Modus ohne Preis.",
    sender_company: "Unternehmen des Integrators oder Beraters, das das Angebot sendet.",
    sender_contact: "Name des Absenders für Signatur und nächsten Schritt.",
    sender_phone: "Telefon in der Dokumentsignatur.",
    sender_email: "E-Mail für die Antwort des Kunden.",
  },
  es: {
    scenario:
      "Elija la situación: contacto previo, primer acercamiento o oferta proactiva tras el análisis ROI. El tono del documento depende del escenario.",
    scenario_after_contact:
      "El cliente ya habló con usted: llamada, email, briefing o solicitud de propuesta. El documento referencia la conversación y conduce al inicio del proyecto.",
    scenario_cold_outreach:
      "Primer contacto sin comunicación previa. Sin «gracias por contactarnos» ni «como acordamos». CTA: reunión breve, no firma de contrato.",
    scenario_proactive_offer:
      "Usted propone automatización tras un análisis (a menudo desde la calculadora). Enfoque en ROI y proceso — el cliente no pidió propuesta.",
    prior_contact_summary:
      "Brevemente: cuándo y de qué hablaron, qué pidió el cliente, acuerdos. El modelo lo usa en el saludo y comprensión de la tarea.",
    problem_source:
      "Cómo conoce el problema sin contacto: observación del proceso, patrón sectorial, recomendación, datos públicos. No invente hechos — describa la base.",
    include_pricing:
      "Activado: precio concreto y condiciones de pago. Desactivado (propuesta suave): precio tras diagnóstico, CTA es reunión. Útil para primer email en frío.",
    client_company:
      "Nombre legal o comercial de la empresa destinataria. En contacto en frío puede dirigirse a la empresa sin persona de contacto.",
    client_contact:
      "Nombre o cargo del destinatario. Opcional en contacto en frío. Si se indica — se usa en el saludo sin simular diálogo previo (salvo «tras contacto»).",
    client_industry: "Ayuda al modelo con formulaciones sectoriales y dolores típicos.",
    client_problem:
      "Describa el dolor del cliente: qué falla, tiempo o dinero perdido, consecuencias. Va al bloque «Comprensión de la tarea».",
    calculator_run:
      "Vincule un cálculo ROI del Automation Calculator — el modelo inserta ahorro, recuperación y NPV.",
    solution_name: "Título corto del proyecto, p. ej. «Automatización de leads».",
    solution_description:
      "Qué propone hacer: alcance general, tecnologías o enfoque. No duplique entregables — aquí el porqué y el cómo.",
    deliverables:
      "Entregables concretos: documentos, integraciones, formación, soporte. Un ítem por línea.",
    project_cost: "Solo si «Incluir precio» está activo. Importe en ₽ para coste y PDF.",
    timeline_weeks: "Duración total del proyecto. Las fases deben caber en este plazo.",
    payment_schedule: "Esquema de pago para propuesta completa. Ignorado en modo suave sin precio.",
    sender_company: "Empresa integradora o consultora que envía la propuesta.",
    sender_contact: "Nombre del remitente para firma y siguiente paso.",
    sender_phone: "Teléfono en la firma del documento.",
    sender_email: "Email para respuesta del cliente.",
  },
  fr: {
    scenario:
      "Choisissez la situation : contact antérieur, première prise de contact ou offre proactive après analyse ROI. Le ton du document dépend du scénario.",
    scenario_after_contact:
      "Le client a déjà échangé avec vous : appel, e-mail, brief ou demande de devis. Le document reprend la discussion et mène au lancement.",
    scenario_cold_outreach:
      "Premier contact sans échange préalable. Pas de « merci pour votre demande » ni « comme convenu ». CTA : courte réunion, pas signature de contrat.",
    scenario_proactive_offer:
      "Vous proposez l'automatisation après analyse (souvent via le calculateur). Accent sur le ROI et le processus — le client n'a pas demandé de devis.",
    prior_contact_summary:
      "Brièvement : quand et de quoi vous avez parlé, demandes et accords. Le modèle l'utilise dans la salutation et la compréhension de la tâche.",
    problem_source:
      "Comment vous connaissez le problème sans contact : observation, pattern sectoriel, recommandation, données publiques. Pas de faits inventés — décrivez la base.",
    include_pricing:
      "Activé : prix et conditions de paiement. Désactivé (devis soft) : prix après diagnostic, CTA = réunion. Utile pour premier e-mail à froid.",
    client_company:
      "Raison sociale ou nom commercial du destinataire. En prospection à froid, adressez l'entreprise sans contact nommé.",
    client_contact:
      "Nom ou fonction du destinataire. Optionnel en prospection à froid. Si renseigné — utilisé sans simuler un dialogue passé (sauf « après contact »).",
    client_industry: "Aide le modèle avec formulations sectorielles et pains typiques.",
    client_problem:
      "Décrivez la douleur client : dysfonctionnements, pertes temps/argent, conséquences. Va dans « Compréhension de la tâche ».",
    calculator_run:
      "Liez un calcul ROI du Automation Calculator — le modèle insère économies, retour sur investissement et NPV.",
    solution_name: "Titre court du projet, ex. « Automatisation du traitement des leads ».",
    solution_description:
      "Ce que vous proposez : périmètre, technologies ou approche. Ne dupliquez pas les livrables — ici le pourquoi et le comment.",
    deliverables:
      "Livrables concrets : documents, intégrations, formation, support. Un point par ligne.",
    project_cost: "Uniquement si « Indiquer le prix » est activé. Montant en ₽ pour coût et PDF.",
    timeline_weeks: "Durée totale du projet. Les phases doivent tenir dans ce délai.",
    payment_schedule: "Schéma de paiement pour devis complet. Ignoré en mode soft sans prix.",
    sender_company: "Entreprise intégratrice ou consultante émettrice du devis.",
    sender_contact: "Nom de l'expéditeur pour signature et prochaine étape.",
    sender_phone: "Téléphone dans la signature du document.",
    sender_email: "E-mail pour la réponse du client.",
  },
  zh: {
    scenario:
      "选择场景：是否已有客户联系、首次触达，或在 ROI 分析后主动报价。整份文档的语气取决于场景。",
    scenario_after_contact:
      "客户已与您沟通：电话、邮件、简报或报价请求。文档会引用讨论并引导项目启动。",
    scenario_cold_outreach:
      "首次触达，此前无联系。避免「感谢来信」「如我们讨论」。CTA 是简短会议，而非签约。",
    scenario_proactive_offer:
      "您在分析后主动提出自动化（常来自计算器）。侧重 ROI 与流程——客户未请求报价。",
    prior_contact_summary:
      "简要说明：何时谈了什么、客户需求与约定。模型用于称呼与任务理解部分。",
    problem_source:
      "无联系时如何得知问题：流程观察、行业模式、推荐、公开数据。勿虚构——说明依据。",
    include_pricing:
      "开启：报价含具体金额与付款条件。关闭（软性报价）：诊断后定价，CTA 为会议。适合冷启动首封邮件。",
    client_company: "报价对象公司的法定或商业名称。冷触达时可只写公司不写联系人。",
    client_contact:
      "收件人姓名或职位。冷触达可选。若填写——用于称呼，不模拟过往对话（「已有联系」除外）。",
    client_industry: "帮助模型选择行业表述与典型痛点。",
    client_problem: "描述客户痛点：问题、时间或金钱损失、后果。进入「任务理解」板块。",
    calculator_run: "关联 Automation Calculator 的 ROI 计算——模型会插入节省、回本与 NPV 数据。",
    solution_name: "项目短标题，如「线索处理自动化」。",
    solution_description: "您提议做什么：高层范围、关键技术或方法。勿重复交付物——此处写原因与方式。",
    deliverables: "具体交付物：文档、集成、培训、支持。每项一行。",
    project_cost: "仅当开启「标明价格」时使用。₽ 金额用于成本块与 PDF。",
    timeline_weeks: "项目总时长。文档中的阶段应在此期限内。",
    payment_schedule: "完整报价的付款方案。无价格的软性模式下忽略。",
    sender_company: "发送报价的集成商或咨询公司。",
    sender_contact: "发件人姓名，用于签名与下一步。",
    sender_phone: "文档签名中的联系电话。",
    sender_email: "客户回复邮箱。",
  },
};

for (const [locale, strings] of Object.entries(tooltipsFallback)) {
  const filePath = path.join(messagesDir, `${locale}.json`);
  const data = JSON.parse(fs.readFileSync(filePath, "utf8"));
  data.proposal = data.proposal || {};
  data.proposal.tooltipsFallback = strings;
  fs.writeFileSync(filePath, JSON.stringify(data, null, 2) + "\n", "utf8");
  console.log(`Patched tooltips ${locale}.json`);
}

console.log("Done.");
