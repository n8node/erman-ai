import Link from "next/link";
import type { PublicPage } from "@/lib/api-public-pages";
import { StrategyLandingSteps } from "@/components/public/strategy-landing/StrategyLandingSteps";

const STRATEGY_HREF = "/tools/strategy";
const ACCENT = "#534AB7";
const ACCENT_SOFT = "#EEEDFE";

type Props = {
  page: PublicPage;
};

const outcomes = [
  {
    icon: "01 — DOC",
    name: "Executive Summary",
    desc: "Диагностика текущей ситуации и главный вывод — зачем внедрять AI именно сейчас и с каким фокусом.",
    value: "2–3 стр.",
  },
  {
    icon: "02 — ТОП-3",
    name: "Рекомендуемые решения",
    desc: "Три приоритетных AI-инициативы с обоснованием, impact/effort и привязкой к вашим процессам.",
    value: "3 кейса",
  },
  {
    icon: "03 — Q1–Q4",
    name: "Roadmap внедрения",
    desc: "Поквартальный план инициатив — от быстрых побед до масштабирования. Не абстрактный «план трансформации».",
    value: "4 фазы",
  },
  {
    icon: "04 — KPI",
    name: "Метрики успеха",
    desc: "Измеримые цели с горизонтом: что считать успехом через 3, 6 и 12 месяцев после старта.",
    value: "6+ KPI",
  },
];

const useCases = [
  { industry: "Розничная торговля", name: "Персонализация и прогноз спроса", stats: ["CRM + ML", "Q1–Q2", "Impact: высокий"] },
  { industry: "Финансы", name: "Автоматизация комплаенса", stats: ["Документы", "Q1", "Impact: высокий"] },
  { industry: "Производство", name: "Predictive maintenance", stats: ["IoT + AI", "Q2–Q3", "Effort: высокий"] },
  { industry: "B2B-услуги", name: "AI-квалификация лидов", stats: ["CRM + LLM", "Q1", "ROI из калькулятора"] },
  { industry: "IT и digital", name: "Генерация КП и ТЗ", stats: ["LLM", "Q1", "Effort: низкий"] },
  { industry: "Логистика", name: "Маршрутизация и SLA", stats: ["Оптимизация", "Q2", "Impact: средний"] },
  { industry: "HR", name: "Скрининг и онбординг", stats: ["ATS + AI", "Q1–Q2", "FTE-эффект"] },
  { industry: "Маркетинг", name: "Контент и аналитика", stats: ["GenAI + BI", "Q1", "Speed ×3"] },
];

const documentSections = [
  { key: "01", title: "Диагностика и цели", text: "Текущая ситуация, зрелость данных, обоснование выбранных целей." },
  { key: "02", title: "Анализ процессов", text: "Каждый ключевой процесс: as-is, боли, потенциал AI." },
  { key: "03", title: "AI use cases", text: "Приоритизированный список с impact/effort и матрицей." },
  { key: "04", title: "Архитектура и стек", text: "Рекомендуемые инструменты, интеграции, data governance." },
  { key: "05", title: "Бюджет и ROI", text: "Capex/Opex по фазам, связь с расчётами калькулятора." },
  { key: "06", title: "Риски и 30 дней", text: "Митигация, stakeholder plan, конкретные шаги на месяц." },
];

const faqItems = [
  {
    q: "Сколько времени занимает генерация?",
    a: "Обычно 30–90 секунд. Стратегия генерируется асинхронно — можно следить за прогрессом в реальном времени.",
  },
  {
    q: "Нужна ли регистрация?",
    a: "Можно заполнить форму без аккаунта. Для сохранения истории, PDF-экспорта и привязки расчётов калькулятора — нужен вход.",
  },
  {
    q: "Можно связать с калькулятором ROI?",
    a: "Да. До 10 расчётов из Automation Calculator — стратегия получит финансовый контекст: NPV, окупаемость, рекомендации по процессам.",
  },
  {
    q: "Чем standard отличается от consulting?",
    a: "Consulting — расширенный отчёт: maturity matrix, stakeholder plan, budget phases, Mermaid-диаграммы. Standard — компактная версия для быстрого старта.",
  },
  {
    q: "На каком языке генерируется документ?",
    a: "На языке вашего профиля — русский или английский. Язык интерфейса определяет язык стратегии.",
  },
  {
    q: "Можно экспортировать в PDF?",
    a: "Да, на тарифе Pro+. PDF включает все разделы, таблицы и диаграммы — готов для презентации совету.",
  },
];

function PrimaryButton({
  href,
  children,
  className = "",
  ai = false,
}: {
  href: string;
  children: React.ReactNode;
  className?: string;
  ai?: boolean;
}) {
  return (
    <Link
      href={href}
      className={`inline-flex items-center gap-2 rounded-[10px] px-[22px] py-3.5 text-[15px] font-medium text-white transition hover:-translate-y-px ${
        ai ? "bg-[#534AB7] hover:bg-[#4338a8]" : "bg-[#0B0D0E] hover:bg-[#1a1d20]"
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
      <PrimaryButton href={STRATEGY_HREF} ai className="mt-6 shrink-0 sm:mt-0">
        {buttonLabel} <span aria-hidden>→</span>
      </PrimaryButton>
    </div>
  );
}

export function StrategyLandingPage({ page }: Props) {
  return (
    <div className="tool-landing strategy-landing bg-white text-[#0B0D0E]">
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
            <a href="#use-cases" className="hover:text-white">Use cases</a>
            <a href="#document" className="hover:text-white">Структура</a>
            <a href="#faq" className="hover:text-white">FAQ</a>
            <Link
              href={STRATEGY_HREF}
              className="rounded-lg bg-[#534AB7] px-3.5 py-2 font-medium text-white hover:bg-[#4338a8]"
            >
              Сгенерировать →
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
              Персональная стратегия внедрения AI{" "}
              <em className="font-mono not-italic" style={{ color: ACCENT }}>
                за один сеанс
              </em>
            </h1>
            <p className="mt-6 max-w-[540px] text-[17px] leading-relaxed text-[#2A2D30]">
              Опишите бизнес, цели и ключевые процессы — получите структурированный документ:
              roadmap, KPI, риски и план на 30 дней. С финансовым обоснованием из калькулятора ROI.
            </p>
            <div className="mt-9 flex flex-wrap gap-3">
              <PrimaryButton href={STRATEGY_HREF} ai>
                Сгенерировать стратегию <span aria-hidden>→</span>
              </PrimaryButton>
              <GhostButton href="#steps">Как это работает</GhostButton>
            </div>
            <div className="mt-7 flex flex-wrap gap-5 font-mono text-[12.5px] uppercase tracking-[0.08em] text-[#5A5D62]">
              <span><b className="font-medium text-[#0B0D0E]">2 шага</b> · контекст → цели</span>
              <span><b className="font-medium text-[#0B0D0E]">15+ разделов</b> · готовый документ</span>
              <span><b className="font-medium text-[#0B0D0E]">PDF</b> · Pro+</span>
            </div>
          </div>

          <aside className="rounded-[20px] border border-[#E5E3DC] bg-white p-7 shadow-[0_30px_60px_-30px_rgba(11,13,14,0.18)]">
            <div className="mb-5 flex items-center justify-between border-b border-[#E5E3DC] pb-4 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">
              <span className="inline-flex items-center gap-1.5 font-medium text-[#0B0D0E]">
                <span className="grid h-5 w-5 place-items-center rounded-full bg-[#534AB7] text-[11px] text-white">3</span>
                Стратегия
              </span>
              <span>Документ · пример</span>
            </div>
            <div className="grid grid-cols-2 gap-px overflow-hidden rounded-[14px] bg-[#E5E3DC]">
              {[
                { label: "Решений / топ", value: "3 кейса", big: true },
                { label: "Roadmap", value: "4 кварт.", big: true },
                { label: "KPI", value: "6 метрик", hint: "с горизонтом" },
                { label: "Срок плана", value: "12 мес" },
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
            <div className="mt-[18px] flex items-center gap-2.5 rounded-[10px] bg-[#EEEDFE] px-3.5 py-3 text-[13.5px] text-[#534AB7] before:h-[7px] before:w-[7px] before:shrink-0 before:rounded-full before:bg-[#534AB7] before:content-['']">
              Готово к презентации совету — roadmap + KPI + ROI
            </div>
          </aside>
        </div>
      </section>

      <section className="bg-[#0B0D0E] px-7 py-[22px] text-white">
        <div className="mx-auto flex max-w-[1200px] flex-wrap items-center justify-between gap-7">
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">Структура документа</span>
          <span className="font-mono text-[clamp(15px,1.6vw,19px)] tracking-[-0.01em]">
            <span className="text-[#B4A8FF]">Стратегия</span>
            <span className="px-1 text-[#7E8189]">=</span>
            <span className="text-[#B4A8FF]">Диагностика</span>
            <span className="px-1 text-[#7E8189]">+</span>
            <span className="text-[#B4A8FF]">Roadmap</span>
            <span className="px-1 text-[#7E8189]">+</span>
            <span className="text-[#B4A8FF]">KPI</span>
            <span className="px-1 text-[#7E8189]">+</span>
            <span className="text-[#B4A8FF]">ROI</span>
          </span>
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">LLM · Claude Sonnet</span>
        </div>
      </section>

      <section id="what" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Что вы получите"
            title="Документ, который отвечает на вопрос «с чего начать AI»."
            lead="Не общие тренды — конкретный план под ваш бизнес: приоритеты, сроки, бюджет, метрики и риски. Каждый раздел можно открыть и экспортировать."
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
              title="Получите стратегию под свой бизнес"
              text="Два шага в форме — и через минуту у вас будет документ для разговора с командой, советом или инвесторами."
              buttonLabel="Сгенерировать"
            />
          </div>
        </div>
      </section>

      <section id="steps" className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Как работает"
            title="Три шага — потому что стратегия без контекста бизнеса бесполезна."
            lead="Сначала контекст компании, потом цели и процессы. Только после этого — генерация документа с roadmap и KPI. Можно привязать расчёты ROI из калькулятора."
          />
          <StrategyLandingSteps />
          <div className="mt-10">
            <MidPageCta
              title="Готовы описать свой бизнес?"
              text="Три шага в форме — и у вас будет стратегия с финансовым обоснованием и планом на 30 дней."
              buttonLabel="Открыть генератор"
            />
          </div>
        </div>
      </section>

      <section id="use-cases" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Use cases"
            title="8 отраслей — типовые сценарии как отправная точка."
            lead="Стратегия адаптируется под ваш ввод. Отраслевые паттерны помогают LLM точнее приоритизировать инициативы и оценить effort."
          />
          <div className="grid gap-4 md:grid-cols-2">
            {useCases.map((item) => (
              <div key={item.name} className="grid grid-cols-[1fr_auto] items-center gap-4 rounded-[14px] border border-[#E5E3DC] bg-white px-6 py-5">
                <div>
                  <div className="mb-1.5 font-mono text-[11px] uppercase tracking-[0.08em] text-[#5A5D62]">{item.industry}</div>
                  <div className="mb-3 text-base font-medium">{item.name}</div>
                  <div className="flex flex-wrap gap-3.5 font-mono text-xs text-[#2A2D30]">
                    {item.stats.map((s) => (
                      <span key={s} className="inline-flex items-center gap-1.5 before:h-[5px] before:w-[5px] before:rounded-full before:bg-[#534AB7] before:content-['']">{s}</span>
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
            lead="Собственник получает план для решения. Консультант — готовый каркас под клиента."
          />
          <div className="grid overflow-hidden rounded-[20px] border border-[#E5E3DC] md:grid-cols-2">
            <div className="bg-white p-9">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">01 — Бизнес</div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">План AI-внедрения для совета — без месяца консалтинга</h3>
              <p className="mt-3.5 text-[15px] text-[#2A2D30]">Стратегия отвечает: что внедрять первым, сколько это стоит и как измерить успех через 12 месяцев.</p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#2A2D30]">
                {["Roadmap по кварталам", "KPI с целевыми значениями", "Связь с ROI калькулятора"].map((item) => (
                  <li key={item} className="flex gap-2.5 before:text-[#5A5D62] before:content-['—']">{item}</li>
                ))}
              </ul>
            </div>
            <div className="bg-[#0B0D0E] p-9 text-white">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#9CA0A8]">02 — Консультант</div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">Каркас стратегии под клиента за 5 минут</h3>
              <p className="mt-3.5 text-[15px] text-[#C9CCD1]">Consulting-режим даёт maturity matrix, stakeholder plan и budget phases — основа для доработки под проект.</p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#C9CCD1]">
                {["15+ разделов out of the box", "Mermaid-диаграммы", "PDF под бренд клиента"].map((item) => (
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
            title="15+ разделов — не один абзац от ChatGPT."
            lead="Стратегия генерируется по JSON-схеме: каждый блок валидируется и сохраняется в истории. Можно вернуться и сравнить версии."
          />
          <div className="grid gap-10 lg:grid-cols-2">
            <div className="rounded-[20px] bg-[#0B0D0E] p-9 font-mono text-white">
              <h4 className="mb-5 font-sans text-[13px] font-medium uppercase tracking-[0.08em] text-[#9CA0A8]">Consulting-режим</h4>
              {[
                "Executive Summary + диагностика",
                "Process analysis + AI use cases",
                "Priority matrix + roadmap Q1–Q4",
                "Tech stack + data governance",
                "Budget phases + ROI summary",
                "Risks + next 30 days",
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
              Одна стратегия — и разговор про AI становится разговором про план.
            </h2>
            <p className="mt-4 max-w-[520px] text-[17px] text-[#C9CCD1]">
              Два шага, один документ, roadmap на год. Если результат подходит — экспортируете PDF и делитесь с командой.
            </p>
            <div className="mt-7 flex flex-wrap gap-3">
              <Link href={STRATEGY_HREF} className="inline-flex items-center gap-2 rounded-[10px] bg-[#534AB7] px-[22px] py-3.5 text-[15px] font-medium text-white hover:bg-[#4338a8]">
                Сгенерировать стратегию <span aria-hidden>→</span>
              </Link>
              <Link href="/tools/calculator" className="inline-flex items-center gap-2 rounded-[10px] border border-[#2A2D30] px-[22px] py-3.5 text-[15px] font-medium text-white hover:border-[#4E5158]">
                Сначала ROI
              </Link>
            </div>
          </div>
          <aside className="rounded-[20px] border border-[#1A1D20] bg-[#111417] p-7">
            <div className="mb-4 font-mono text-[11px] uppercase tracking-[0.1em] text-[#7E8189]">Связка инструментов</div>
            <div className="font-mono text-base leading-[1.9] text-white">
              <span className="text-[#B4A8FF]">Калькулятор</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#B4A8FF]">ROI</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#B4A8FF]">Стратегия</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#B4A8FF]">КП</span>
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
                Внедрение AI и автоматизация бизнес-процессов. Стратегия, ROI и КП — в одном кабинете.
              </p>
            </div>
            <div>
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">Инструменты</div>
              <ul className="space-y-2 text-sm text-[#2A2D30]">
                <li><Link href="/tools/calculator" className="hover:text-[#0B0D0E]">Калькулятор ROI</Link></li>
                <li><Link href={STRATEGY_HREF} className="hover:text-[#0B0D0E]">AI-стратегия</Link></li>
                <li><Link href="/tools/proposal" className="hover:text-[#0B0D0E]">Генерация КП</Link></li>
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
            <span>AI Strategy Generator v.1.0</span>
          </div>
        </div>
      </footer>
    </div>
  );
}
