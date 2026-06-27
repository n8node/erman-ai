import Link from "next/link";
import type { PublicPage } from "@/lib/api-public-pages";
import { AuditLandingSteps } from "@/components/public/audit-landing/AuditLandingSteps";

const AUDIT_HREF = "/tools/audit";
const ACCENT = "#BA7517";
const ACCENT_SOFT = "#FAEEDA";
const SAVINGS_VALUE = "890\u00a0000\u00a0₽";

type Props = {
  page: PublicPage;
};

const outcomes = [
  {
    icon: "01 — RANK",
    name: "Приоритетный рейтинг",
    desc: "2–7 процессов ранжированы по automation score — с обоснованием для каждого.",
    value: "Top-7",
  },
  {
    icon: "02 — ₽",
    name: "Оценка экономии",
    desc: "Суммарная потенциальная экономия в месяц по всем процессам — на основе ваших диапазонов.",
    value: SAVINGS_VALUE,
  },
  {
    icon: "03 — WIN",
    name: "Quick wins",
    desc: "Процессы с высоким score и низкой сложностью — старт автоматизации без долгого внедрения.",
    value: "3 шт.",
  },
  {
    icon: "04 — MAP",
    name: "Roadmap",
    desc: "Фазы внедрения с процессами и deliverables — от быстрых побед к масштабированию.",
    value: "3 фазы",
  },
];

const processTemplates = [
  { industry: "Продажи", name: "Обработка входящих лидов", stats: ["CRM", "Quick win", "Score 7.9+"] },
  { industry: "Операции", name: "Обработка заказов", stats: ["ERP + 1C", "High impact", "2–3 FTE"] },
  { industry: "Поддержка", name: "Клиентская поддержка", stats: ["Helpdesk", "Volume", "4–10 FTE"] },
  { industry: "Финансы", name: "Выставление счетов", stats: ["1C", "Errors", "Compliance"] },
  { industry: "HR", name: "Онбординг сотрудников", stats: ["Docs", "Medium", "Standardize"] },
  { industry: "Логистика", name: "Управление складом", stats: ["WMS", "Integrations", "Bottleneck"] },
  { industry: "Финансы", name: "Ежемесячная отчётность", stats: ["Excel", "Manual", "Reporting"] },
  { industry: "Свой", name: "Любой процесс вручную", stats: ["2–7 шт.", "Custom", "Full score"] },
];

const documentSections = [
  { key: "01", title: "Executive Summary", text: "Главный вывод аудита — куда вкладываться первым и почему." },
  { key: "02", title: "Контекст компании", text: "IT-ландшафт, цели и ограничения, учтённые при ранжировании." },
  { key: "03", title: "Process scores", text: "Детализация по каждому процессу: часы, ФОТ, ошибки, automation score." },
  { key: "04", title: "Priority ranking", text: "Таблица приоритетов с экономией, окупаемостью и quick win." },
  { key: "05", title: "Roadmap", text: "Фазы внедрения: период, процессы, deliverables." },
  { key: "06", title: "Риски и метрики", text: "Митигация рисков, next steps и KPI для отслеживания прогресса." },
];

const faqItems = [
  {
    q: "Сколько процессов можно добавить?",
    a: "От 2 до 7. Меньше двух — аудит не запустится: нужен минимум для сравнительного ранжирования.",
  },
  {
    q: "Как считается automation score?",
    a: "Формула на backend — не LLM. Учитываются частота, FTE, ошибки, стандартизация, готовность и сложность интеграций.",
  },
  {
    q: "Сколько времени занимает генерация?",
    a: "Обычно 30–90 секунд. Отчёт генерируется асинхронно — статус обновляется автоматически.",
  },
  {
    q: "Чем аудит отличается от калькулятора?",
    a: "Калькулятор — один процесс, точные цифры ROI. Аудит — портфель из 2–7 процессов с приоритизацией и roadmap.",
  },
  {
    q: "Можно перейти к калькулятору после аудита?",
    a: "Да. Quick wins и топ-процессы — кандидаты для детального ROI-расчёта в Automation Calculator.",
  },
  {
    q: "На каком языке генерируется отчёт?",
    a: "На языке профиля — русский или английский.",
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
        accent ? "bg-[#BA7517] hover:bg-[#9a6314]" : "bg-[#0B0D0E] hover:bg-[#1a1d20]"
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
      <PrimaryButton href={AUDIT_HREF} accent className="mt-6 shrink-0 sm:mt-0">
        {buttonLabel} <span aria-hidden>→</span>
      </PrimaryButton>
    </div>
  );
}

export function AuditLandingPage({ page }: Props) {
  return (
    <div className="tool-landing audit-landing bg-white text-[#0B0D0E]">
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
            <a href="#templates" className="hover:text-white">Шаблоны</a>
            <a href="#document" className="hover:text-white">Структура</a>
            <a href="#faq" className="hover:text-white">FAQ</a>
            <Link
              href={AUDIT_HREF}
              className="rounded-lg bg-[#BA7517] px-3.5 py-2 font-medium text-white hover:bg-[#9a6314]"
            >
              Запустить аудит →
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
              Аудит автоматизации{" "}
              <em className="font-mono not-italic" style={{ color: ACCENT }}>
                за один сеанс
              </em>
            </h1>
            <p className="mt-6 max-w-[540px] text-[17px] leading-relaxed text-[#2A2D30]">
              Опишите 2–7 процессов — получите приоритетный рейтинг, оценку экономии, quick wins и
              roadmap внедрения. Score считается формулой, отчёт дополняется AI.
            </p>
            <div className="mt-9 flex flex-wrap gap-3">
              <PrimaryButton href={AUDIT_HREF} accent>
                Запустить аудит <span aria-hidden>→</span>
              </PrimaryButton>
              <GhostButton href="#steps">Как это работает</GhostButton>
            </div>
            <div className="mt-7 flex flex-wrap gap-5 font-mono text-[12.5px] uppercase tracking-[0.08em] text-[#5A5D62]">
              <span><b className="font-medium text-[#0B0D0E]">2–7 процессов</b> · портфель</span>
              <span><b className="font-medium text-[#0B0D0E]">8 шаблонов</b> · готовые профили</span>
              <span><b className="font-medium text-[#0B0D0E]">Score + AI</b> · приоритеты</span>
            </div>
          </div>

          <aside className="rounded-[20px] border border-[#E5E3DC] bg-white p-7 shadow-[0_30px_60px_-30px_rgba(11,13,14,0.18)]">
            <div className="mb-5 flex items-center justify-between border-b border-[#E5E3DC] pb-4 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">
              <span className="inline-flex items-center gap-1.5 font-medium text-[#0B0D0E]">
                <span className="grid h-5 w-5 place-items-center rounded-full bg-[#BA7517] text-[11px] text-white">3</span>
                Аудит
              </span>
              <span>Отчёт · пример</span>
            </div>
            <div className="grid grid-cols-2 gap-px overflow-hidden rounded-[14px] bg-[#E5E3DC]">
              {[
                { label: "Процессов", value: "4", big: true },
                { label: "Экономия / мес", value: SAVINGS_VALUE, big: true },
                { label: "Quick wins", value: "3", hint: "старт сразу" },
                { label: "Top score", value: "8.7" },
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
            <div className="mt-[18px] flex items-center gap-2.5 rounded-[10px] bg-[#FAEEDA] px-3.5 py-3 text-[13.5px] text-[#633806] before:h-[7px] before:w-[7px] before:shrink-0 before:rounded-full before:bg-[#BA7517] before:content-['']">
              Начните с quick wins — окупаемость &lt; 6 мес
            </div>
          </aside>
        </div>
      </section>

      <section className="bg-[#0B0D0E] px-7 py-[22px] text-white">
        <div className="mx-auto flex max-w-[1200px] flex-wrap items-center justify-between gap-7">
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">Модель приоритизации</span>
          <span className="font-mono text-[clamp(15px,1.6vw,19px)] tracking-[-0.01em]">
            <span className="text-[#E8C078]">Score</span>
            <span className="px-1 text-[#7E8189]">=</span>
            <span className="text-[#E8C078]">f(объём)</span>
            <span className="px-1 text-[#7E8189]">×</span>
            <span className="text-[#E8C078]">готовность</span>
            <span className="px-1 text-[#7E8189]">−</span>
            <span className="text-[#E8C078]">сложность</span>
          </span>
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">Формула + LLM</span>
        </div>
      </section>

      <section id="what" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Что вы получите"
            title="Отчёт, который отвечает на вопрос «с чего начать автоматизацию»."
            lead="Не список идей — ранжированный портфель процессов с экономией, quick wins и roadmap. Score прозрачен, LLM дополняет контекст."
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
              title="Узнайте приоритеты своих процессов"
              text="Опишите компанию и 2–7 процессов — через минуту получите ранжированный отчёт с roadmap."
              buttonLabel="Запустить аудит"
            />
          </div>
        </div>
      </section>

      <section id="steps" className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Как работает"
            title="Три шага — потому что приоритизация без портфеля процессов невозможна."
            lead="Сначала контекст компании, потом описание каждого процесса. На выходе — score, ranking, roadmap и quick wins."
          />
          <AuditLandingSteps />
          <div className="mt-10">
            <MidPageCta
              title="Готовы описать процессы?"
              text="Три шага — и у вас будет приоритетный план автоматизации с оценкой экономии."
              buttonLabel="Открыть аудит"
            />
          </div>
        </div>
      </section>

      <section id="templates" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Шаблоны процессов"
            title="8 типовых процессов — не нужно описывать с нуля."
            lead="Каждый шаблон подставляет отдел, типовые диапазоны и системы. Можно править вручную или добавить свой процесс."
          />
          <div className="grid gap-4 md:grid-cols-2">
            {processTemplates.map((tpl) => (
              <div key={tpl.name} className="grid grid-cols-[1fr_auto] items-center gap-4 rounded-[14px] border border-[#E5E3DC] bg-white px-6 py-5">
                <div>
                  <div className="mb-1.5 font-mono text-[11px] uppercase tracking-[0.08em] text-[#5A5D62]">{tpl.industry}</div>
                  <div className="mb-3 text-base font-medium">{tpl.name}</div>
                  <div className="flex flex-wrap gap-3.5 font-mono text-xs text-[#2A2D30]">
                    {tpl.stats.map((s) => (
                      <span key={s} className="inline-flex items-center gap-1.5 before:h-[5px] before:w-[5px] before:rounded-full before:bg-[#BA7517] before:content-['']">{s}</span>
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
            lead="Операционный директор выбирает приоритеты. Интегратор — обосновывает портфель проектов."
          />
          <div className="grid overflow-hidden rounded-[20px] border border-[#E5E3DC] md:grid-cols-2">
            <div className="bg-white p-9">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">01 — Бизнес</div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">С чего начать автоматизацию — не гадать, а ранжировать</h3>
              <p className="mt-3.5 text-[15px] text-[#2A2D30]">Портфель из 7 процессов — score показывает, где максимальный эффект при минимальной сложности.</p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#2A2D30]">
                {["Quick wins для быстрого старта", "Roadmap на 6–12 месяцев", "Переход в калькулятор ROI"].map((item) => (
                  <li key={item} className="flex gap-2.5 before:text-[#5A5D62] before:content-['—']">{item}</li>
                ))}
              </ul>
            </div>
            <div className="bg-[#0B0D0E] p-9 text-white">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#9CA0A8]">02 — Интегратор</div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">Обоснование портфеля проектов для клиента</h3>
              <p className="mt-3.5 text-[15px] text-[#C9CCD1]">Приоритетная таблица с экономией — аргумент для бюджета. Дальше — калькулятор, стратегия, КП.</p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#C9CCD1]">
                {["2–7 процессов в одном отчёте", "Automation score по формуле", "Риски и next steps"].map((item) => (
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
            eyebrow="Структура отчёта"
            title="Полный аудит — не только таблица score."
            lead="LLM дополняет формульный score контекстом компании, roadmap, рисками и метриками. Всё сохраняется в истории."
          />
          <div className="grid gap-10 lg:grid-cols-2">
            <div className="rounded-[20px] bg-[#0B0D0E] p-9 font-mono text-white">
              <h4 className="mb-5 font-sans text-[13px] font-medium uppercase tracking-[0.08em] text-[#9CA0A8]">Разделы отчёта</h4>
              {[
                "Executive Summary + контекст",
                "Process scores по каждому процессу",
                "Priority ranking + quick wins",
                "Roadmap по фазам",
                "Риски + next steps",
                "Метрики для отслеживания",
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
            title="Что важно знать до аудита."
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
              Один аудит — и портфель автоматизации становится планом, а не хаосом.
            </h2>
            <p className="mt-4 max-w-[520px] text-[17px] text-[#C9CCD1]">
              2–7 процессов, прозрачный score, roadmap и quick wins. Дальше — калькулятор ROI по топ-процессам.
            </p>
            <div className="mt-7 flex flex-wrap gap-3">
              <Link href={AUDIT_HREF} className="inline-flex items-center gap-2 rounded-[10px] bg-[#BA7517] px-[22px] py-3.5 text-[15px] font-medium text-white hover:bg-[#9a6314]">
                Запустить аудит <span aria-hidden>→</span>
              </Link>
              <Link href="/tools/calculator" className="inline-flex items-center gap-2 rounded-[10px] border border-[#2A2D30] px-[22px] py-3.5 text-[15px] font-medium text-white hover:border-[#4E5158]">
                Детальный ROI
              </Link>
            </div>
          </div>
          <aside className="rounded-[20px] border border-[#1A1D20] bg-[#111417] p-7">
            <div className="mb-4 font-mono text-[11px] uppercase tracking-[0.1em] text-[#7E8189]">Связка инструментов</div>
            <div className="font-mono text-base leading-[1.9] text-white">
              <span className="text-[#E8C078]">Аудит</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#8FA1FF]">Калькулятор</span>
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
                Аудит, ROI, стратегия и КП — полный цикл от приоритизации до сделки.
              </p>
            </div>
            <div>
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">Инструменты</div>
              <ul className="space-y-2 text-sm text-[#2A2D30]">
                <li><Link href={AUDIT_HREF} className="hover:text-[#0B0D0E]">Аудит автоматизации</Link></li>
                <li><Link href="/tools/calculator" className="hover:text-[#0B0D0E]">Калькулятор ROI</Link></li>
                <li><Link href="/tools/strategy" className="hover:text-[#0B0D0E]">AI-стратегия</Link></li>
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
            <span>Automation Audit v.1.0</span>
          </div>
        </div>
      </footer>
    </div>
  );
}
