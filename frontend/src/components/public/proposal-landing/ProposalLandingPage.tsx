import Link from "next/link";
import type { PublicPage } from "@/lib/api-public-pages";
import { ProposalLandingSteps } from "@/components/public/proposal-landing/ProposalLandingSteps";

const PROPOSAL_HREF = "/tools/proposal";
const ACCENT = "#3B6D11";
const ACCENT_SOFT = "#EAF3DE";
const EXAMPLE_COST = "600\u00a0000\u00a0₽";

type Props = {
  page: PublicPage;
};

const outcomes = [
  {
    icon: "01 — INTRO",
    name: "Обращение к клиенту",
    desc: "Персонализированное intro под сценарий: после звонка, холодное письмо или проактивная инициатива.",
    value: "§1",
  },
  {
    icon: "02 — SCOPE",
    name: "Scope of Work",
    desc: "Что входит и что не входит — без размытых формулировок. Клиент видит границы проекта до подписания.",
    value: "In/Out",
  },
  {
    icon: "03 — TIME",
    name: "Этапы и сроки",
    desc: "Фазы с длительностью в неделях и описанием deliverables на каждом этапе.",
    value: "3 фазы",
  },
  {
    icon: "04 — ₽",
    name: "Стоимость и ROI",
    desc: "Сумма проекта, условия оплаты и финансовое обоснование из калькулятора — если привязан расчёт.",
    value: EXAMPLE_COST,
  },
];

const scenarios = [
  {
    industry: "После контакта",
    name: "Классическое КП после созвона",
    stats: ["С ценой", "ROI из calc", "8 разделов"],
  },
  {
    industry: "Cold outreach",
    name: "Первое письмо без прайса",
    stats: ["Без цены", "Фокус на боли", "CTA на встречу"],
  },
  {
    industry: "Проактивное",
    name: "Инициатива с обоснованием",
    stats: ["Ценность", "Кейс + ROI", "С ценой"],
  },
  {
    industry: "Из калькулятора",
    name: "КП сразу после расчёта ROI",
    stats: ["1 клик", "NPV + payback", "Process name"],
  },
  {
    industry: "Из Legal Scan",
    name: "КП после аудита документов",
    stats: ["Problem auto", "Compliance", "Custom scope"],
  },
  {
    industry: "Агентство",
    name: "White-label под клиента",
    stats: ["Sender brand", "PDF export", "DOCX edit"],
  },
];

const documentSections = [
  { key: "01", title: "Greeting", text: "Персонализированное обращение с учётом сценария и контекста контакта." },
  { key: "02", title: "Понимание задачи", text: "Переформулировка проблемы клиента — показывает, что вы слушали." },
  { key: "03", title: "Предлагаемое решение", text: "Описание подхода, технологий и ожидаемого результата." },
  { key: "04", title: "Scope included / excluded", text: "Два списка — что входит и что явно не входит в проект." },
  { key: "05", title: "Timeline", text: "Таблица фаз: название, недели, описание deliverables." },
  { key: "06", title: "Стоимость и оплата", text: "Сумма, ROI из калькулятора, условия оплаты по выбранной схеме." },
];

const faqItems = [
  {
    q: "Сколько времени занимает генерация?",
    a: "Обычно 30–90 секунд. КП генерируется асинхронно — статус обновляется автоматически.",
  },
  {
    q: "Можно без цены?",
    a: "Да. В сценарии cold outreach цена скрывается — документ фокусируется на проблеме и приглашении к диалогу.",
  },
  {
    q: "Как связать с калькулятором ROI?",
    a: "На шаге «Решение» выберите расчёт из истории калькулятора — в КП попадут окупаемость, NPV и название процесса.",
  },
  {
    q: "PDF и DOCX — когда доступны?",
    a: "Экспорт PDF и DOCX доступен на тарифе Pro+. Сгенерированный текст можно скопировать из интерфейса на любом тарифе.",
  },
  {
    q: "На каком языке генерируется КП?",
    a: "На языке профиля — русский или английский. Язык интерфейса определяет язык документа.",
  },
  {
    q: "Можно редактировать результат?",
    a: "DOCX-экспорт (Pro+) даёт файл для правок в Word. PDF — для отправки клиенту в финальном виде.",
  },
];

function PrimaryButton({
  href,
  children,
  className = "",
  accent = false,
}: {
  href: string;
  children: React.ReactNode;
  className?: string;
  accent?: boolean;
}) {
  return (
    <Link
      href={href}
      className={`inline-flex items-center gap-2 rounded-[10px] px-[22px] py-3.5 text-[15px] font-medium text-white transition hover:-translate-y-px ${
        accent ? "bg-[#3B6D11] hover:bg-[#2f5710]" : "bg-[#0B0D0E] hover:bg-[#1a1d20]"
      } ${className}`}
    >
      {children}
    </Link>
  );
}

function GhostButton({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      className="inline-flex items-center gap-2 rounded-[10px] border border-[#E5E3DC] px-[22px] py-3.5 text-[15px] font-medium text-[#0B0D0E] transition hover:-translate-y-px hover:bg-[#F5F4EF]"
    >
      {children}
    </Link>
  );
}

function SectionEyebrow({ children }: { children: React.ReactNode }) {
  return (
    <p className="font-mono text-xs uppercase tracking-[0.12em] text-[#5A5D62]">
      <span style={{ color: ACCENT }}>●</span> {children}
    </p>
  );
}

function SectionHead({
  eyebrow,
  title,
  lead,
}: {
  eyebrow: string;
  title: string;
  lead: string;
}) {
  return (
    <div className="mb-14 grid gap-5 lg:grid-cols-[0.9fr_1fr] lg:items-end lg:gap-12">
      <div>
        <SectionEyebrow>{eyebrow}</SectionEyebrow>
        <h2 className="mt-5 text-[clamp(28px,3vw,44px)] font-semibold leading-[1.05] tracking-[-0.025em] text-[#0B0D0E]">
          {title}
        </h2>
      </div>
      <p className="max-w-[540px] text-[17px] leading-relaxed text-[#2A2D30]">{lead}</p>
    </div>
  );
}

function MidPageCta({
  title,
  text,
  buttonLabel,
}: {
  title: string;
  text: string;
  buttonLabel: string;
}) {
  return (
    <div
      className="rounded-[20px] border border-[#E5E3DC] px-8 py-10 sm:flex sm:items-center sm:justify-between sm:gap-8"
      style={{ backgroundColor: ACCENT_SOFT }}
    >
      <div className="max-w-xl">
        <h3 className="text-xl font-semibold tracking-[-0.02em] text-[#0B0D0E]">{title}</h3>
        <p className="mt-2 text-[15px] leading-relaxed text-[#2A2D30]">{text}</p>
      </div>
      <PrimaryButton href={PROPOSAL_HREF} accent className="mt-6 shrink-0 sm:mt-0">
        {buttonLabel} <span aria-hidden>→</span>
      </PrimaryButton>
    </div>
  );
}

export function ProposalLandingPage({ page }: Props) {
  return (
    <div className="tool-landing proposal-landing bg-white text-[#0B0D0E]">
      <header className="sticky top-0 z-50 border-b border-[#1A1D20] bg-[#0B0D0E] text-white">
        <div className="mx-auto flex h-16 max-w-[1200px] items-center justify-between gap-6 px-7">
          <Link href="/" className="flex items-center gap-2.5 font-semibold">
            <span className="grid h-[30px] w-[30px] place-items-center rounded-full bg-white text-sm font-bold text-[#0B0D0E]">
              E
            </span>
            <span className="text-[15px]">Erman AI</span>
          </Link>
          <nav className="hidden items-center gap-7 text-sm text-[#C9CCD1] md:flex">
            <a href="#what" className="hover:text-white">Что получите</a>
            <a href="#steps" className="hover:text-white">Как работает</a>
            <a href="#scenarios" className="hover:text-white">Сценарии</a>
            <a href="#document" className="hover:text-white">Структура</a>
            <a href="#faq" className="hover:text-white">FAQ</a>
            <Link
              href={PROPOSAL_HREF}
              className="rounded-lg bg-[#3B6D11] px-3.5 py-2 font-medium text-white hover:bg-[#2f5710]"
            >
              Сгенерировать КП →
            </Link>
          </nav>
        </div>
      </header>

      <section className="border-b border-[#E5E3DC] bg-white pb-16 pt-[88px]">
        <div className="mx-auto grid max-w-[1200px] gap-16 px-7 lg:grid-cols-[1.05fr_1fr] lg:items-center">
          <div>
            <p className="mb-7 flex items-center gap-2 font-mono text-xs uppercase tracking-[0.1em] text-[#5A5D62]">
              <span>Инструменты</span>
              <span className="text-[#8A8E94]">/</span>
              <span className="font-medium text-[#0B0D0E]">{page.title}</span>
            </p>
            <h1 className="text-[clamp(40px,6vw,84px)] font-semibold leading-[0.98] tracking-[-0.035em]">
              Коммерческое предложение{" "}
              <em className="font-mono not-italic" style={{ color: ACCENT }}>
                за 5 минут
              </em>
            </h1>
            <p className="mt-6 max-w-[540px] text-[17px] leading-relaxed text-[#2A2D30]">
              Опишите клиента и решение — получите готовое КП: обращение, scope, этапы, стоимость и
              ROI из калькулятора. PDF и DOCX для отправки.
            </p>
            <div className="mt-9 flex flex-wrap gap-3">
              <PrimaryButton href={PROPOSAL_HREF} accent>
                Сгенерировать КП <span aria-hidden>→</span>
              </PrimaryButton>
              <GhostButton href="#steps">Как это работает</GhostButton>
            </div>
            <div className="mt-7 flex flex-wrap gap-5 font-mono text-[12.5px] uppercase tracking-[0.08em] text-[#5A5D62]">
              <span><b className="font-medium text-[#0B0D0E]">3 шага</b> · клиент → решение → КП</span>
              <span><b className="font-medium text-[#0B0D0E]">3 сценария</b> · cold / contact / proactive</span>
              <span><b className="font-medium text-[#0B0D0E]">PDF + DOCX</b> · Pro+</span>
            </div>
          </div>

          <aside className="rounded-[20px] border border-[#E5E3DC] bg-white p-7 shadow-[0_30px_60px_-30px_rgba(11,13,14,0.18)]">
            <div className="mb-5 flex items-center justify-between border-b border-[#E5E3DC] pb-4 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">
              <span className="inline-flex items-center gap-1.5 font-medium text-[#0B0D0E]">
                <span className="grid h-5 w-5 place-items-center rounded-full bg-[#3B6D11] text-[11px] text-white">3</span>
                КП
              </span>
              <span>Документ · пример</span>
            </div>
            <div className="grid grid-cols-2 gap-px overflow-hidden rounded-[14px] bg-[#E5E3DC]">
              {[
                { label: "Разделов", value: "8", big: true },
                { label: "Срок проекта", value: "8 нед", big: true },
                { label: "Стоимость", value: EXAMPLE_COST, hint: "с ROI" },
                { label: "Оплата", value: "50 / 50" },
              ].map((m) => (
                <div key={m.label} className="flex flex-col gap-2 bg-white p-[18px]">
                  <div className="font-mono text-[11px] uppercase tracking-[0.1em] text-[#5A5D62]">{m.label}</div>
                  <div className={`whitespace-nowrap font-mono font-medium tracking-[-0.02em] text-[#0B0D0E] ${m.big ? "text-[30px]" : "text-[26px]"}`}>
                    {m.value}
                  </div>
                  {m.hint && <div className="text-xs text-[#5A5D62]">{m.hint}</div>}
                </div>
              ))}
            </div>
            <div className="mt-[18px] flex items-center gap-2.5 rounded-[10px] bg-[#EAF3DE] px-3.5 py-3 text-[13.5px] text-[#3B6D11] before:h-[7px] before:w-[7px] before:shrink-0 before:rounded-full before:bg-[#3B6D11] before:content-['']">
              Готово к отправке клиенту — scope + timeline + ROI
            </div>
          </aside>
        </div>
      </section>

      <section className="bg-[#0B0D0E] px-7 py-[22px] text-white">
        <div className="mx-auto flex max-w-[1200px] flex-wrap items-center justify-between gap-7">
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">Структура КП</span>
          <span className="font-mono text-[clamp(15px,1.6vw,19px)] tracking-[-0.01em]">
            <span className="text-[#9FD89F]">КП</span>
            <span className="px-1 text-[#7E8189]">=</span>
            <span className="text-[#9FD89F]">Клиент</span>
            <span className="px-1 text-[#7E8189]">+</span>
            <span className="text-[#9FD89F]">Решение</span>
            <span className="px-1 text-[#7E8189]">+</span>
            <span className="text-[#9FD89F]">Scope</span>
            <span className="px-1 text-[#7E8189]">+</span>
            <span className="text-[#9FD89F]">ROI</span>
          </span>
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">LLM · Claude Sonnet</span>
        </div>
      </section>

      <section id="what" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Что вы получите"
            title="Документ, который закрывает сделку — не шаблон из Word."
            lead="Каждый раздел генерируется под конкретного клиента и задачу. Scope in/out, timeline и ROI — не placeholder-текст."
          />
          <div className="grid gap-px overflow-hidden rounded-[20px] border border-[#E5E3DC] bg-[#E5E3DC] sm:grid-cols-2 xl:grid-cols-4">
            {outcomes.map((item) => (
              <div key={item.icon} className="flex min-h-[220px] flex-col gap-3.5 bg-white p-7">
                <div className="font-mono text-xs tracking-[0.08em]" style={{ color: ACCENT }}>{item.icon}</div>
                <h3 className="text-[15px] font-medium">{item.name}</h3>
                <p className="text-[13.5px] leading-relaxed text-[#5A5D62]">{item.desc}</p>
                <div className="mt-auto whitespace-nowrap font-mono text-[38px] font-medium leading-none tracking-[-0.03em]">{item.value}</div>
              </div>
            ))}
          </div>
          <div className="mt-10">
            <MidPageCta
              title="Соберите КП под своего клиента"
              text="Три шага в форме — и через минуту у вас документ для отправки. Можно привязать расчёт ROI."
              buttonLabel="Сгенерировать"
            />
          </div>
        </div>
      </section>

      <section id="steps" className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Как работает"
            title="Три шага — потому что КП без контекста клиента не продаёт."
            lead="Сначала клиент и сценарий, потом решение и финансы. На выходе — структурированный документ с scope и timeline. ROI подтягивается из калькулятора."
          />
          <ProposalLandingSteps />
          <div className="mt-10">
            <MidPageCta
              title="Готовы описать клиента?"
              text="Три шага — и КП с финансовым обоснованием готово к отправке."
              buttonLabel="Открыть генератор"
            />
          </div>
        </div>
      </section>

      <section id="scenarios" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Сценарии"
            title="Три сценария — под разные этапы воронки."
            lead="После созвона, холодный outreach или проактивное предложение. Плюс быстрый старт из калькулятора или Legal Scan."
          />
          <div className="grid gap-4 md:grid-cols-2">
            {scenarios.map((item) => (
              <div key={item.name} className="grid grid-cols-[1fr_auto] items-center gap-4 rounded-[14px] border border-[#E5E3DC] bg-white px-6 py-5">
                <div>
                  <div className="mb-1.5 font-mono text-[11px] uppercase tracking-[0.08em] text-[#5A5D62]">{item.industry}</div>
                  <div className="mb-3 text-base font-medium">{item.name}</div>
                  <div className="flex flex-wrap gap-3.5 font-mono text-xs text-[#2A2D30]">
                    {item.stats.map((s) => (
                      <span key={s} className="inline-flex items-center gap-1.5 before:h-[5px] before:w-[5px] before:rounded-full before:bg-[#3B6D11] before:content-['']">{s}</span>
                    ))}
                  </div>
                </div>
                <div className="grid h-[38px] w-[38px] place-items-center rounded-full border border-[#E5E3DC] font-mono text-[#5A5D62]">→</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Для кого"
            title="Один инструмент — две роли."
            lead="Продажник закрывает сделку быстрее. Агентство масштабирует подготовку КП."
          />
          <div className="grid overflow-hidden rounded-[20px] border border-[#E5E3DC] md:grid-cols-2">
            <div className="bg-white p-9">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">01 — Продажи</div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">КП после созвона — пока клиент ещё «горячий»</h3>
              <p className="mt-3.5 text-[15px] text-[#2A2D30]">Не три часа в Word — пять минут в форме. С ROI из калькулятора аргумент «окупается за 0.2 месяца» звучит убедительнее скидки.</p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#2A2D30]">
                {["3 сценария под воронку", "Scope in/out из коробки", "PDF для отправки"].map((item) => (
                  <li key={item} className="flex gap-2.5 before:text-[#5A5D62] before:content-['—']">{item}</li>
                ))}
              </ul>
            </div>
            <div className="bg-[#0B0D0E] p-9 text-white">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#9CA0A8]">02 — Агентство</div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">КП под каждого клиента — без копипаста</h3>
              <p className="mt-3.5 text-[15px] text-[#C9CCD1]">Бренд отправителя, deliverables, график оплаты — всё настраивается. DOCX для правок перед финальной отправкой.</p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#C9CCD1]">
                {["Данные отправителя в шапке", "Привязка к calc / Legal Scan", "История всех КП"].map((item) => (
                  <li key={item} className="flex gap-2.5 before:text-[#7E8189] before:content-['—']">{item}</li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </section>

      <section id="document" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Структура документа"
            title="8 разделов — готовый каркас коммерческого предложения."
            lead="Каждый блок валидируется по JSON-схheme. Можно вернуться к истории и скачать PDF повторно."
          />
          <div className="grid gap-10 lg:grid-cols-2">
            <div className="rounded-[20px] bg-[#0B0D0E] p-9 font-mono text-white">
              <h4 className="mb-5 font-sans text-[13px] font-medium uppercase tracking-[0.08em] text-[#9CA0A8]">Разделы КП</h4>
              {[
                "Greeting + понимание задачи",
                "Предлагаемое решение",
                "Scope included / excluded",
                "Timeline по фазам",
                "Стоимость + payment terms",
                "Почему мы + next step",
              ].map((eq) => (
                <div key={eq} className="border-t border-[#1A1D20] py-4 first:border-0 first:pt-0 text-[15px] leading-[2.2] text-[#C9CCD1]">{eq}</div>
              ))}
            </div>
            <div className="divide-y divide-[#E5E3DC] border-t border-[#E5E3DC]">
              {documentSections.map((row) => (
                <div key={row.key} className="grid grid-cols-[64px_1fr] gap-4 py-[18px]">
                  <div className="font-mono text-lg" style={{ color: ACCENT }}>{row.key}</div>
                  <div className="text-[14.5px] text-[#2A2D30]">
                    <b className="mb-1 block text-[13px] font-medium text-[#0B0D0E]">{row.title}</b>
                    {row.text}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      <section id="faq" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Частые вопросы"
            title="Что важно знать до генерации."
            lead="Если вопроса нет в списке — напишите в телеграм-бот, мы добавим его сюда."
          />
          <div className="divide-y divide-[#E5E3DC] border-b border-[#E5E3DC]">
            {faqItems.map((item) => (
              <details key={item.q} className="group py-6">
                <summary className="flex cursor-pointer list-none items-center justify-between text-[17px] font-medium marker:content-none">
                  {item.q}
                  <span className="font-mono text-xl text-[#5A5D62] transition group-open:rotate-45">+</span>
                </summary>
                <p className="mt-3 max-w-3xl text-[15px] leading-relaxed text-[#2A2D30]">{item.a}</p>
              </details>
            ))}
          </div>
        </div>
      </section>

      <section className="border-t border-[#0B0D0E] bg-[#0B0D0E] px-7 py-24 text-white">
        <div className="mx-auto grid max-w-[1200px] gap-12 lg:grid-cols-[1.2fr_1fr] lg:items-center">
          <div>
            <h2 className="text-[clamp(34px,4vw,56px)] font-semibold leading-[1.02] tracking-[-0.03em]">
              Одно КП — и разговор про цену становится разговором про ценность.
            </h2>
            <p className="mt-4 max-w-[520px] text-[17px] text-[#C9CCD1]">
              Три шага, один документ, ROI из калькулятора. Экспортируете PDF — и отправляете клиенту.
            </p>
            <div className="mt-7 flex flex-wrap gap-3">
              <Link href={PROPOSAL_HREF} className="inline-flex items-center gap-2 rounded-[10px] bg-[#3B6D11] px-[22px] py-3.5 text-[15px] font-medium text-white hover:bg-[#2f5710]">
                Сгенерировать КП <span aria-hidden>→</span>
              </Link>
              <Link href="/tools/calculator" className="inline-flex items-center gap-2 rounded-[10px] border border-[#2A2D30] px-[22px] py-3.5 text-[15px] font-medium text-white hover:border-[#4E5158]">
                Сначала ROI
              </Link>
            </div>
          </div>
          <aside className="rounded-[20px] border border-[#1A1D20] bg-[#111417] p-7">
            <div className="mb-4 font-mono text-[11px] uppercase tracking-[0.1em] text-[#7E8189]">Связка инструментов</div>
            <div className="font-mono text-base leading-[1.9] text-white">
              <span className="text-[#9FD89F]">Калькулятор</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#9FD89F]">ROI</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#B4A8FF]">Стратегия</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#9FD89F]">КП</span>
            </div>
          </aside>
        </div>
      </section>

      <footer className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-12">
        <div className="mx-auto max-w-[1200px]">
          <div className="grid gap-9 md:grid-cols-[2fr_1fr_1fr_1fr]">
            <div>
              <Link href="/" className="inline-flex items-center gap-2.5 font-semibold">
                <span className="grid h-[30px] w-[30px] place-items-center rounded-full bg-[#0B0D0E] text-sm font-bold text-white">E</span>
                Erman AI
              </Link>
              <p className="mt-3 max-w-[340px] text-sm text-[#5A5D62]">
                ROI, стратегия и КП — в одном кабинете. От расчёта до подписания сделки.
              </p>
            </div>
            <div>
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">Инструменты</div>
              <ul className="space-y-2 text-sm text-[#2A2D30]">
                <li><Link href="/tools/calculator" className="hover:text-[#0B0D0E]">Калькулятор ROI</Link></li>
                <li><Link href="/tools/strategy" className="hover:text-[#0B0D0E]">AI-стратегия</Link></li>
                <li><Link href={PROPOSAL_HREF} className="hover:text-[#0B0D0E]">Генерация КП</Link></li>
              </ul>
            </div>
            <div>
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">Компания</div>
              <ul className="space-y-2 text-sm text-[#2A2D30]">
                <li><Link href="/about" className="hover:text-[#0B0D0E]">О нас</Link></li>
                <li><Link href="/discuss" className="hover:text-[#0B0D0E]">Обсудить проект</Link></li>
              </ul>
            </div>
            <div>
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">Связь</div>
              <ul className="space-y-2 text-sm text-[#2A2D30]">
                <li><a href="https://t.me/ermanai" className="hover:text-[#0B0D0E]">@ermanai</a></li>
                <li><a href="https://t.me/ermanai_bot" className="hover:text-[#0B0D0E]">@ermanai_bot</a></li>
              </ul>
            </div>
          </div>
          <div className="mt-8 flex flex-wrap justify-between gap-3 border-t border-[#E5E3DC] pt-8 font-mono text-xs tracking-[0.06em] text-[#5A5D62]">
            <span>1982 — 2026 · ERMAN AI</span>
            <span>Proposal Generator v.1.0</span>
          </div>
        </div>
      </footer>
    </div>
  );
}
