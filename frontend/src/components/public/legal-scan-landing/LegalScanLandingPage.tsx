import Link from "next/link";
import type { PublicPage } from "@/lib/api-public-pages";
import { LandingBrandLink } from "@/components/public/LandingBrandLink";
import { LegalScanLandingSteps } from "@/components/public/legal-scan-landing/LegalScanLandingSteps";

const LEGAL_SCAN_HREF = "/tools/legal-scan";
const ACCENT = "#A32D2D";
const ACCENT_SOFT = "#FCEBEB";
const FINE_VALUE = "до\u00a06\u00a0М\u00a0₽";

type Props = {
  page: PublicPage;
};

const outcomes = [
  {
    icon: "01 — CHECK",
    name: "12 проверок",
    desc: "SSL, политика ПД, cookie, оферта, реквизиты, трекеры, erid и другие пункты чеклиста 152-ФЗ.",
    value: "12",
  },
  {
    icon: "02 — RISK",
    name: "Карточки рисков",
    desc: "Каждый риск — с severity, статьёй закона, текстом штрафа и пошаговым «как исправить».",
    value: "7+",
  },
  {
    icon: "03 — ₽",
    name: "Оценка штрафов",
    desc: "Суммарный диапазон фиксированных штрафов по найденным нарушениям — для аргументации клиенту.",
    value: FINE_VALUE,
  },
  {
    icon: "04 — PDF",
    name: "Отчёт PDF",
    desc: "Экспорт для клиента или внутреннего согласования. Доказательства — URL страниц и найденные данные.",
    value: "Pro+",
  },
];

const checklistItems = [
  { industry: "152-ФЗ", name: "Политика обработки ПД", stats: ["High", "Cookie", "Forms"] },
  { industry: "149-ФЗ", name: "Cookie и согласие", stats: ["Banner", "Policy", "Trackers"] },
  { industry: "ЗоЗПП", name: "Оферта и реквизиты", stats: ["ИНН/ОГРН", "Контакты", "Offer"] },
  { industry: "Реклама", name: "Маркировка erid", stats: ["Ads traffic", "Medium", "Evidence"] },
  { industry: "Техника", name: "SSL и формы", stats: ["HTTPS", "Consent", "Encryption"] },
  { industry: "ПД", name: "Отзыв согласия", stats: ["Withdraw", "Contacts", "Terms"] },
];

const documentSections = [
  { key: "01", title: "Layer 1 — crawl", text: "Обход сайта, извлечение ссылок, форм, текстов документов и метаданных." },
  { key: "02", title: "Checklist", text: "12 автоматических проверок с статусом ok / risk и evidence." },
  { key: "03", title: "Risk engine", text: "Сопоставление findings с базой legal_risks — статьи и штрафы." },
  { key: "04", title: "LLM enrich", text: "Контекст по отрасли и пояснения — без подмены фактов layer 1." },
  { key: "05", title: "Summary", text: "Количество рисков, суммарные штрафы, turnover note." },
  { key: "06", title: "Экспорт", text: "PDF-отчёт и переход в Proposal Generator с problem auto-fill." },
];

const faqItems = [
  {
    q: "Это юридическая консультация?",
    a: "Нет. Инструмент — автоматизированный скрининг по чеклисту и базе рисков. Disclaimer в каждом отчёте. Для судебных споров нужен юрист.",
  },
  {
    q: "Сколько страниц сканируется?",
    a: "Зависит от пресета и тарифа: Free — до 5 страниц, Pro — до 30, Business — до 50. Пресеты: быстро (1), стандарт (10), глубоко (30).",
  },
  {
    q: "Чем layer 1 отличается от LLM?",
    a: "Layer 1 — детерминированный парсер и правила: что найдено на сайте, с URL-доказательствами. LLM только дополняет формулировки и отраслевой контекст.",
  },
  {
    q: "Можно сразу сделать КП?",
    a: "Да. Из отчёта — переход в Proposal Generator с автозаполнением проблемы клиента на основе найденных рисков.",
  },
  {
    q: "Какие законы покрываются?",
    a: "152-ФЗ (ПД), 149-ФЗ (реклама/information), ЗоЗПП (оферта, реквизиты), маркировка рекламы — база рисков расширяется админом.",
  },
  {
    q: "Сколько времени занимает скан?",
    a: "От 30 секунд (1 страница) до нескольких минут (30 страниц). Прогресс чеклиста виден в реальном времени.",
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
        accent ? "bg-[#A32D2D] hover:bg-[#8a2626]" : "bg-[#0B0D0E] hover:bg-[#1a1d20]"
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
      <PrimaryButton href={LEGAL_SCAN_HREF} accent className="mt-6 shrink-0 sm:mt-0">
        {buttonLabel} <span aria-hidden>→</span>
      </PrimaryButton>
    </div>
  );
}

export function LegalScanLandingPage({ page }: Props) {
  return (
    <div className="tool-landing legal-scan-landing bg-white text-[#0B0D0E]">
      <header className="sticky top-0 z-50 border-b border-[#1A1D20] bg-[#0B0D0E] text-white">
        <div className="mx-auto flex h-16 max-w-[1200px] items-center justify-between gap-6 px-7">
          <LandingBrandLink labelClassName="text-[15px]" />
          <nav className="hidden items-center gap-7 text-sm text-[#C9CCD1] md:flex">
            <a href="#what" className="hover:text-white">Что проверяет</a>
            <a href="#steps" className="hover:text-white">Как работает</a>
            <a href="#checklist" className="hover:text-white">Чеклист</a>
            <a href="#document" className="hover:text-white">Структура</a>
            <a href="#faq" className="hover:text-white">FAQ</a>
            <Link
              href={LEGAL_SCAN_HREF}
              className="rounded-lg bg-[#A32D2D] px-3.5 py-2 font-medium text-white hover:bg-[#8a2626]"
            >
              Запустить скан →
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
              Legal-скан сайта{" "}
              <em className="font-mono not-italic" style={{ color: ACCENT }}>
                за 3 минуты
              </em>
            </h1>
            <p className="mt-6 max-w-[540px] text-[17px] leading-relaxed text-[#2A2D30]">
              Введите URL — получите отчёт по 152-ФЗ и смежным требованиям: риски, статьи закона,
              штрафы и рекомендации. С доказательствами по страницам.
            </p>
            <div className="mt-9 flex flex-wrap gap-3">
              <PrimaryButton href={LEGAL_SCAN_HREF} accent>
                Запустить скан <span aria-hidden>→</span>
              </PrimaryButton>
              <GhostButton href="#steps">Как это работает</GhostButton>
            </div>
            <div className="mt-7 flex flex-wrap gap-5 font-mono text-[12.5px] uppercase tracking-[0.08em] text-[#5A5D62]">
              <span><b className="font-medium text-[#0B0D0E]">12 проверок</b> · чеклист</span>
              <span><b className="font-medium text-[#0B0D0E]">Layer 1</b> · crawl + rules</span>
              <span><b className="font-medium text-[#0B0D0E]">→ КП</b> · proposal</span>
            </div>
          </div>

          <aside className="rounded-[20px] border border-[#E5E3DC] bg-white p-7 shadow-[0_30px_60px_-30px_rgba(11,13,14,0.18)]">
            <div className="mb-5 flex items-center justify-between border-b border-[#E5E3DC] pb-4 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">
              <span className="inline-flex items-center gap-1.5 font-medium text-[#0B0D0E]">
                <span className="grid h-5 w-5 place-items-center rounded-full bg-[#A32D2D] text-[11px] text-white">3</span>
                Отчёт
              </span>
              <span>Пример · scan</span>
            </div>
            <div className="grid grid-cols-2 gap-px overflow-hidden rounded-[14px] bg-[#E5E3DC]">
              {[
                { label: "Рисков", value: "7", big: true },
                { label: "Штрафы", value: FINE_VALUE, big: true },
                { label: "Высоких", value: "3", hint: "severity high" },
                { label: "Страниц", value: "10" },
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
            <div className="mt-[18px] flex items-center gap-2.5 rounded-[10px] bg-[#FCEBEB] px-3.5 py-3 text-[13.5px] text-[#A32D2D] before:h-[7px] before:w-[7px] before:shrink-0 before:rounded-full before:bg-[#A32D2D] before:content-['']">
              3 критичных риска — политика ПД, cookie, формы
            </div>
          </aside>
        </div>
      </section>

      <section className="bg-[#0B0D0E] px-7 py-[22px] text-white">
        <div className="mx-auto flex max-w-[1200px] flex-wrap items-center justify-between gap-7">
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">Модель проверки</span>
          <span className="font-mono text-[clamp(15px,1.6vw,19px)] tracking-[-0.01em]">
            <span className="text-[#F5A5A5]">Отчёт</span>
            <span className="px-1 text-[#7E8189]">=</span>
            <span className="text-[#F5A5A5]">Crawl</span>
            <span className="px-1 text-[#7E8189]">+</span>
            <span className="text-[#F5A5A5]">Checklist</span>
            <span className="px-1 text-[#7E8189]">→</span>
            <span className="text-[#F5A5A5]">Risks</span>
            <span className="px-1 text-[#7E8189]">+</span>
            <span className="text-[#F5A5A5]">Fines</span>
          </span>
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">Rules + LLM</span>
        </div>
      </section>

      <section id="what" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Что вы получите"
            title="Отчёт, который показывает клиенту цену бездействия."
            lead="Не общие советы — конкретные нарушения с URL-доказательствами, статьями закона и диапазоном штрафов."
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
              title="Проверьте сайт клиента сейчас"
              text="URL, отрасль и три клика — через минуту у вас отчёт с рисками и суммой штрафов."
              buttonLabel="Запустить скан"
            />
          </div>
        </div>
      </section>

      <section id="steps" className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Как работает"
            title="Три шага — потому что legal-риск без доказательств не продаётся."
            lead="Сначала параметры сайта, потом автоматический crawl и чеклист. На выходе — карточки рисков с штрафами и переход в КП."
          />
          <LegalScanLandingSteps />
          <div className="mt-10">
            <MidPageCta
              title="Готовы проверить URL?"
              text="Три шага — и у вас аргумент для продажи compliance-проекта клиенту."
              buttonLabel="Открыть сканер"
            />
          </div>
        </div>
      </section>

      <section id="checklist" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Чеклист"
            title="12 автоматических проверок — не ручной аудит."
            lead="Каждый пункт — правило с evidence: что найдено, на какой странице, какие данные извлечены."
          />
          <div className="grid gap-4 md:grid-cols-2">
            {checklistItems.map((item) => (
              <div key={item.name} className="grid grid-cols-[1fr_auto] items-center gap-4 rounded-[14px] border border-[#E5E3DC] bg-white px-6 py-5">
                <div>
                  <div className="mb-1.5 font-mono text-[11px] uppercase tracking-[0.08em] text-[#5A5D62]">{item.industry}</div>
                  <div className="mb-3 text-base font-medium">{item.name}</div>
                  <div className="flex flex-wrap gap-3.5 font-mono text-xs text-[#2A2D30]">
                    {item.stats.map((s) => (
                      <span key={s} className="inline-flex items-center gap-1.5 before:h-[5px] before:w-[5px] before:rounded-full before:bg-[#A32D2D] before:content-['']">{s}</span>
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
            lead="Агентство продаёт compliance. In-house — проверяет свой сайт до проверки регулятора."
          />
          <div className="grid overflow-hidden rounded-[20px] border border-[#E5E3DC] md:grid-cols-2">
            <div className="bg-white p-9">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">01 — Агентство</div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">Legal-аудит как вход в проект</h3>
              <p className="mt-3.5 text-[15px] text-[#2A2D30]">Отчёт со штрафами — аргумент для КП на доработку сайта. Один клик — Proposal с описанием проблем.</p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#2A2D30]">
                {["PDF для клиента", "Ссылки на страницы-нарушители", "База рисков под РФ"].map((item) => (
                  <li key={item} className="flex gap-2.5 before:text-[#5A5D62] before:content-['—']">{item}</li>
                ))}
              </ul>
            </div>
            <div className="bg-[#0B0D0E] p-9 text-white">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#9CA0A8]">02 — Бизнес</div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">Проверить свой сайт до претензии</h3>
              <p className="mt-3.5 text-[15px] text-[#C9CCD1]">Cookie, политика ПД, реквизиты — типовые причины штрафов. Чеклист покажет, что исправить первым.</p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#C9CCD1]">
                {["12 пунктов за один прогон", "Severity high/medium/low", "How to fix в каждой карточке"].map((item) => (
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
            title="Layer 1 + риски — прозрачная цепочка."
            lead="Сначала факты с сайта, потом сопоставление с базой legal_risks. LLM не выдумывает нарушения."
          />
          <div className="grid gap-10 lg:grid-cols-2">
            <div className="rounded-[20px] bg-[#0B0D0E] p-9 font-mono text-white">
              <h4 className="mb-5 font-sans text-[13px] font-medium uppercase tracking-[0.08em] text-[#9CA0A8]">Pipeline</h4>
              {[
                "Crawl + extract (HTML, forms, links)",
                "Findings → checklist (12 keys)",
                "Risk engine → legal_risks DB",
                "LLM enrich (industry note)",
                "Summary + fine totals",
                "PDF export + → Proposal",
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
            title="Что важно знать до скана."
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
              Один скан — и compliance становится коммерческим предложением.
            </h2>
            <p className="mt-4 max-w-[520px] text-[17px] text-[#C9CCD1]">
              URL, чеклист, риски со штрафами. Дальше — PDF клиенту или КП на исправление.
            </p>
            <div className="mt-7 flex flex-wrap gap-3">
              <Link href={LEGAL_SCAN_HREF} className="inline-flex items-center gap-2 rounded-[10px] bg-[#A32D2D] px-[22px] py-3.5 text-[15px] font-medium text-white hover:bg-[#8a2626]">
                Запустить скан <span aria-hidden>→</span>
              </Link>
              <Link href="/tools/proposal" className="inline-flex items-center gap-2 rounded-[10px] border border-[#2A2D30] px-[22px] py-3.5 text-[15px] font-medium text-white hover:border-[#4E5158]">
                Сразу КП
              </Link>
            </div>
          </div>
          <aside className="rounded-[20px] border border-[#1A1D20] bg-[#111417] p-7">
            <div className="mb-4 font-mono text-[11px] uppercase tracking-[0.1em] text-[#7E8189]">Связка инструментов</div>
            <div className="font-mono text-base leading-[1.9] text-white">
              <span className="text-[#F5A5A5]">Legal Scan</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#9FD89F]">КП</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#8FA1FF]">Калькулятор</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#E8C078]">Аудит</span>
            </div>
          </aside>
        </div>
      </section>

      <footer className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-12">
        <div className="mx-auto max-w-[1200px]">
          <div className="grid gap-9 md:grid-cols-[2fr_1fr_1fr_1fr]">
            <div>
              <LandingBrandLink />
              <p className="mt-3 max-w-[340px] text-sm text-[#5A5D62]">
                Legal-скан, ROI, стратегия и КП — полный цикл для digital-агентств.
              </p>
            </div>
            <div>
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">Инструменты</div>
              <ul className="space-y-2 text-sm text-[#2A2D30]">
                <li><Link href={LEGAL_SCAN_HREF} className="hover:text-[#0B0D0E]">Legal-скан</Link></li>
                <li><Link href="/tools/calculator" className="hover:text-[#0B0D0E]">Калькулятор ROI</Link></li>
                <li><Link href="/tools/proposal" className="hover:text-[#0B0D0E]">Генерация КП</Link></li>
                <li><Link href="/tools/audit" className="hover:text-[#0B0D0E]">Аудит автоматизации</Link></li>
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
            <span>Legal Scan v.1.0</span>
          </div>
        </div>
      </footer>
    </div>
  );
}
