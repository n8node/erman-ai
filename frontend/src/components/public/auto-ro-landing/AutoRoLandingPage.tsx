import Link from "next/link";
import type { PublicPage } from "@/lib/api-public-pages";
import { LandingBrandLink } from "@/components/public/LandingBrandLink";
import { AutoRoLandingSteps } from "@/components/public/auto-ro-landing/AutoRoLandingSteps";
import { AutoRoDemoShowcase } from "@/components/public/auto-ro-landing/AutoRoDemoShowcase";

const DEMO_HREF = "#demo";
const DISCUSS_HREF = "/discuss";
const RTK_HREF = "/auto-rtk";
const ACCENT = "#185FA5";
const ACCENT_SOFT = "#E6F1FB";
const ML_ACCENT = "#534AB7";

type Props = {
  page: PublicPage;
};

const outcomes = [
  {
    icon: "01 — OPS",
    name: "6 категорий / 18 операций",
    desc: "Полная таксономия от бурения до СПО и наращивания — без раздувания до 33 типов и без упрощения до 11.",
    value: "18",
  },
  {
    icon: "02 — ML",
    name: "Точность 97.5%",
    desc: "Гибрид алгоритмов и собственной ML-модели. Fusion-слой на уровне лучших ensemble-моделей отрасли.",
    value: "97.5%",
  },
  {
    icon: "03 — CH",
    name: "10 каналов",
    desc: "Стандартная наземная телеметрия: тальблок, WOB, расход, давление, RPM, момент, вес на крюке и др.",
    value: "10",
  },
  {
    icon: "04 — QC",
    name: "QC + честный fallback",
    desc: "6.1 данные отсутствуют, 6.2 плохое качество, 6.3 не определено — не один catch-all «другая операция».",
    value: "3",
  },
];

const moduleItems = [
  { tag: "Ingest", name: "Приём телеметрии", stats: ["WITSML", "OPC", "10 кан."] },
  { tag: "QC", name: "Контроль качества", stats: ["6.1", "6.2", "Gate"] },
  { tag: "Rules", name: "Rule Engine", stats: ["Пороги", "Schlumberger", "Быстро"] },
  { tag: "ML", name: "ML Classifier", stats: ["Ensemble", "10 сек", "97.5%"] },
  { tag: "Fusion", name: "Fusion Layer", stats: ["Algo+ML", "Confidence", "QC"] },
  { tag: "Timeline", name: "Таймлайн операций", stats: ["18 ops", "Цвета", "Drill-down"] },
  { tag: "KPI", name: "KPI-слой", stats: ["Микро-KPI", "НПВ", "СПО"] },
  { tag: "Export", name: "Экспорт и API", stats: ["Excel", "API", "РТК"] },
];

const compareItems = [
  {
    name: "ProNova",
    ops: "11 ops",
    method: "Только правила",
    pro: "Компактная таксономия, объяснимость",
    con: "Много «не определено», KPI отдельно от операций",
    us: "18 операций + KPI в едином дереве",
  },
  {
    name: "Schlumberger '06",
    ops: "13 ops",
    method: "Минимальные датчики",
    pro: "4–5 каналов, отраслевой стандарт",
    con: "Нет ML, нет QC-слоя",
    us: "Те же датчики + ML + QC 6.1–6.3",
  },
  {
    name: "KOA-RF",
    ops: "33 ops · 95.65%",
    method: "KOA + Random Forest",
    pro: "Высокая точность, гибрид",
    con: "Раздутые названия, трудно читать таймлайн",
    us: "97.5% при 18 операциях — читаемо",
  },
  {
    name: "Galina V. (бенчмарк)",
    ops: "10 ops",
    method: "RF / AdaBoost / RUSBoost",
    pro: "Доказанная точность ML-класса",
    con: "Только ML, без rule-слоя",
    us: "ML + Rules: скорость и точность",
  },
];

const pipelineSections = [
  { key: "01", title: "Телеметрия", text: "10 каналов, дискретизация 1–10 сек, WITSML/OPC." },
  { key: "02", title: "QC-гейт", text: "Валидация каналов → 6.1 / 6.2 до классификации." },
  { key: "03", title: "Rule Engine", text: "Пороговые правила по Hook, RPM, Flow, WOB." },
  { key: "04", title: "ML Classifier", text: "Собственная модель, ensemble-класс точности." },
  { key: "05", title: "Fusion", text: "Слияние Algo + ML, confidence scoring." },
  { key: "06", title: "Отчёт", text: "Таймлайн, аналитика, API, интеграция с Авто РТК." },
];

const faqItems = [
  {
    q: "Почему 18 операций, а не 11 как у ProNova?",
    a: "Больше детализации без потери читаемости. Наращивание и варианты СПО — в основном дереве, не только в KPI.",
  },
  {
    q: "Почему не 33 как у KOA-RF?",
    a: "Из отраслевого анализа: раздутые названия не добавляют аналитической ценности. KPI-слой закрывает микро-метрики.",
  },
  {
    q: "Какая реальная точность?",
    a: "97.5% на реальных данных буровой установки. Fusion-слой объединяет алгоритмический и ML-результат.",
  },
  {
    q: "Какие датчики нужны минимум?",
    a: "Как в патенте Schlumberger: вес на крюке, положение тальблока, давление, момент/обороты. Остальные каналы повышают точность.",
  },
  {
    q: "Это отдельный продукт или часть системы?",
    a: "Отдельный модуль единой платформы автоматизированной буровой. Работает автономно и интегрируется с модулем Авто РТК.",
  },
  {
    q: "Что с «не определено»?",
    a: "Три уровня: 6.1 нет данных, 6.2 плохое качество, 6.3 не определено — не один catch-all как у конкурентов.",
  },
];

function PrimaryButton({
  href,
  children,
  className = "",
  accent = false,
  ml = false,
}: {
  href: string;
  children: React.ReactNode;
  className?: string;
  accent?: boolean;
  ml?: boolean;
}) {
  const bg = ml ? ML_ACCENT : accent ? ACCENT : "#0B0D0E";
  const hover = ml ? "hover:bg-[#433da0]" : accent ? "hover:bg-[#134d88]" : "hover:bg-[#1a1d20]";

  return (
    <Link
      href={href}
      className={`inline-flex items-center gap-2 rounded-[10px] px-[22px] py-3.5 text-[15px] font-medium text-white transition hover:-translate-y-px ${hover} ${className}`}
      style={accent || ml ? { backgroundColor: bg } : undefined}
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

export function AutoRoLandingPage({ page }: Props) {
  return (
    <div className="tool-landing auto-ro-landing bg-white text-[#0B0D0E]">
      <header className="sticky top-0 z-50 border-b border-[#1A1D20] bg-[#0B0D0E] text-white">
        <div className="mx-auto flex h-16 max-w-[1200px] items-center justify-between gap-6 px-7">
          <LandingBrandLink labelClassName="text-[15px]" />
          <nav className="hidden items-center gap-7 text-sm text-[#C9CCD1] md:flex">
            <a href="#demo" className="hover:text-white">
              Демо
            </a>
            <a href="#what" className="hover:text-white">
              Возможности
            </a>
            <a href="#steps" className="hover:text-white">
              Сценарии
            </a>
            <a href="#compare" className="hover:text-white">
              Сравнение
            </a>
            <a href="#faq" className="hover:text-white">
              FAQ
            </a>
            <Link
              href={DEMO_HREF}
              className="rounded-lg px-3.5 py-2 font-medium text-white hover:opacity-90"
              style={{ backgroundColor: ACCENT }}
            >
              Смотреть демо →
            </Link>
          </nav>
        </div>
      </header>

      <section className="border-b border-[#E5E3DC] bg-white pb-16 pt-[88px]">
        <div className="mx-auto grid max-w-[1200px] gap-16 px-7 lg:grid-cols-[1.05fr_1fr] lg:items-center">
          <div>
            <p className="mb-7 flex flex-wrap items-center gap-2 font-mono text-xs uppercase tracking-[0.1em] text-[#5A5D62]">
              <span>Кейс</span>
              <span className="text-[#8A8E94]">/</span>
              <span>Нефтедобыча</span>
              <span className="text-[#8A8E94]">/</span>
              <span className="font-medium text-[#0B0D0E]">{page.title}</span>
            </p>
            <h1 className="text-[clamp(36px,5.5vw,72px)] font-semibold leading-[0.98] tracking-[-0.035em]">
              Авто РО —{" "}
              <em className="font-mono not-italic" style={{ color: ML_ACCENT }}>
                18 операций
              </em>{" "}
              из телеметрии
            </h1>
            <p className="mt-6 max-w-[540px] text-[17px] leading-relaxed text-[#2A2D30]">
              Модуль единой системы автоматизированной буровой: гибрид алгоритмов и ML-модели
              классифицирует операции с точностью 97.5%. 6 категорий, QC-слой, таймлайн и аналитика
              рейса.
            </p>
            <div className="mt-9 flex flex-wrap gap-3">
              <PrimaryButton href={DEMO_HREF} ml>
                Смотреть демо <span aria-hidden>→</span>
              </PrimaryButton>
              <GhostButton href={DISCUSS_HREF}>Обсудить проект</GhostButton>
            </div>
            <div className="mt-7 flex flex-wrap gap-5 font-mono text-[12.5px] uppercase tracking-[0.08em] text-[#5A5D62]">
              <span>
                <b className="font-medium text-[#0B0D0E]">97.5%</b> · точность
              </span>
              <span>
                <b className="font-medium text-[#0B0D0E]">18 ops</b> · 6 категорий
              </span>
              <span>
                <b className="font-medium text-[#0B0D0E]">Algo+ML</b> · fusion
              </span>
            </div>
          </div>

          <aside className="rounded-[20px] border border-[#E5E3DC] bg-white p-7 shadow-[0_30px_60px_-30px_rgba(11,13,14,0.18)]">
            <div className="mb-5 flex items-center justify-between border-b border-[#E5E3DC] pb-4 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">
              <span className="inline-flex items-center gap-1.5 font-medium text-[#0B0D0E]">
                <span
                  className="grid h-5 w-5 place-items-center rounded-full text-[11px] text-white"
                  style={{ backgroundColor: ML_ACCENT }}
                >
                  ●
                </span>
                Текущая операция
              </span>
              <span className="aro-badge aro-badge--ok">LIVE</span>
            </div>
            <div className="mb-4 font-mono text-[11px] text-[#5A5D62]">
              Северное-1 · Куст 7.2 · СК-2847 · 1847 м
            </div>
            <div
              className="mb-4 rounded-[12px] px-4 py-3"
              style={{ backgroundColor: "#EAF3DE", color: "#3B6D11" }}
            >
              <div className="font-mono text-[11px] uppercase tracking-[0.08em] opacity-80">
                1.1 · Бурение
              </div>
              <div className="text-[22px] font-semibold tracking-[-0.02em]">Бурение с вращением</div>
            </div>
            <div className="grid grid-cols-2 gap-px overflow-hidden rounded-[14px] bg-[#E5E3DC]">
              {[
                { label: "Fusion", value: "97.5%", big: true },
                { label: "Алгоритм", value: "91%", big: true },
                { label: "ML-модель", value: "98%" },
                { label: "Каналов", value: "10/10" },
              ].map((m) => (
                <div key={m.label} className="flex flex-col gap-2 bg-white p-[18px]">
                  <div className="font-mono text-[11px] uppercase tracking-[0.1em] text-[#5A5D62]">
                    {m.label}
                  </div>
                  <div
                    className={`whitespace-nowrap font-mono font-medium tracking-[-0.02em] text-[#0B0D0E] ${m.big ? "text-[26px]" : "text-[22px]"}`}
                  >
                    {m.value}
                  </div>
                </div>
              ))}
            </div>
            <div
              className="mt-[18px] flex items-center gap-2.5 rounded-[10px] px-3.5 py-3 text-[13.5px]"
              style={{ backgroundColor: ACCENT_SOFT, color: ACCENT }}
            >
              Модуль АБУ · интеграция с{" "}
              <Link href={RTK_HREF} className="font-medium underline hover:opacity-80">
                Авто РТК
              </Link>
            </div>
          </aside>
        </div>
      </section>

      <section className="bg-[#0B0D0E] px-7 py-[22px] text-white">
        <div className="mx-auto flex max-w-[1200px] flex-wrap items-center justify-between gap-7">
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">
            Пайплайн
          </span>
          <span className="font-mono text-[clamp(13px,1.4vw,17px)] tracking-[-0.01em]">
            <span className="text-[#7EC8F2]">10 каналов</span>
            <span className="px-1 text-[#7E8189]">→</span>
            <span className="text-[#7EC8F2]">Rules</span>
            <span className="px-1 text-[#7E8189]">+</span>
            <span className="text-[#C4B5FD]">ML</span>
            <span className="px-1 text-[#7E8189]">→</span>
            <span className="text-[#9FD89F]">18 ops · 97.5%</span>
          </span>
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">
            Модуль единой АБУ
          </span>
        </div>
      </section>

      <section id="demo" className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Интерактивное демо"
            title="Классификация операций — как на буровой."
            lead="Переключайте вкладки: таймлайн, fusion algo+ML, треки, таксономия, QC и сравнение с рынком. Данные обезличены."
          />
          <AutoRoDemoShowcase />
        </div>
      </section>

      <section id="what" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Что реализовано"
            title="6 категорий, 18 операций, 97.5% точности."
            lead="Модуль встроен в контур автоматизированной буровой. Инженер видит текущую операцию и полный таймлайн рейса."
          />
          <div className="grid gap-px overflow-hidden rounded-[20px] border border-[#E5E3DC] bg-[#E5E3DC] sm:grid-cols-2 xl:grid-cols-4">
            {outcomes.map((item) => (
              <div key={item.icon} className="flex min-h-[220px] flex-col gap-3.5 bg-white p-7">
                <div className="font-mono text-xs tracking-[0.08em]" style={{ color: ACCENT }}>
                  {item.icon}
                </div>
                <h3 className="text-[15px] font-medium">{item.name}</h3>
                <p className="text-[13.5px] leading-relaxed text-[#5A5D62]">{item.desc}</p>
                <div className="mt-auto whitespace-nowrap font-mono text-[38px] font-medium leading-none tracking-[-0.03em]">
                  {item.value}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section id="steps" className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Типичные сценарии"
            title="От телеметрии до отчёта по рейсу."
            lead="Подключение каналов, классификация в реальном времени и экспорт аналитики."
          />
          <AutoRoLandingSteps />
        </div>
      </section>

      <section id="modules" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Модули системы"
            title="Восемь функциональных блоков."
            lead="От приёма данных до экспорта и интеграции с Авто РТК."
          />
          <div className="grid gap-4 md:grid-cols-2">
            {moduleItems.map((item) => (
              <div
                key={item.name}
                className="grid grid-cols-[1fr_auto] items-center gap-4 rounded-[14px] border border-[#E5E3DC] bg-white px-6 py-5"
              >
                <div>
                  <div className="mb-1.5 font-mono text-[11px] uppercase tracking-[0.08em] text-[#5A5D62]">
                    {item.tag}
                  </div>
                  <div className="mb-3 text-base font-medium">{item.name}</div>
                  <div className="flex flex-wrap gap-3.5 font-mono text-xs text-[#2A2D30]">
                    {item.stats.map((s) => (
                      <span key={s} className="inline-flex items-center gap-1.5">
                        <span
                          className="inline-block h-[5px] w-[5px] rounded-full"
                          style={{ backgroundColor: ACCENT }}
                        />
                        {s}
                      </span>
                    ))}
                  </div>
                </div>
                <div className="grid h-[38px] w-[38px] place-items-center rounded-full border border-[#E5E3DC] font-mono text-[#5A5D62]">
                  →
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Для кого"
            title="Два ключевых пользователя на буровой."
            lead="Инженер видит операцию в реальном времени. Аналитик — распределение времени по рейсу."
          />
          <div className="grid overflow-hidden rounded-[20px] border border-[#E5E3DC] md:grid-cols-2">
            <div className="bg-white p-9">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">
                01 — Инженер
              </div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">
                Операция на экране — без ручного рапорта
              </h3>
              <p className="mt-3.5 text-[15px] text-[#2A2D30]">
                Текущая операция, fusion algo+ML, таймлайн и QC-статусы — для оперативных решений
                на буровой.
              </p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#2A2D30]">
                {["LIVE классификация", "18 типов операций", "Алерт при 6.2"].map((item) => (
                  <li key={item} className="flex gap-2.5 before:text-[#5A5D62] before:content-['—']">
                    {item}
                  </li>
                ))}
              </ul>
            </div>
            <div className="bg-[#0B0D0E] p-9 text-white">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#9CA0A8]">
                02 — Аналитик
              </div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">
                Распределение времени и скрытый НПВ
              </h3>
              <p className="mt-3.5 text-[15px] text-[#C9CCD1]">
                Сводка по 6 категориям, KPI-слой, сравнение скважин. Экспорт в Excel и API для BI.
              </p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#C9CCD1]">
                {["97.5% точность", "Доля «не определено»", "Интеграция с Авто РТК"].map((item) => (
                  <li key={item} className="flex gap-2.5 before:text-[#7E8189] before:content-['—']">
                    {item}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </section>

      <section id="compare" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Сравнение с рынком"
            title="ProNova · Schlumberger · KOA-RF · Авто РО."
            lead="Отраслевой анализ: почему 18 операций и гибрид algo+ML — оптимальный баланс точности и читаемости."
          />
          <div className="grid gap-4 md:grid-cols-2">
            {compareItems.map((item) => (
              <div
                key={item.name}
                className="rounded-[14px] border border-[#E5E3DC] bg-white p-6"
              >
                <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
                  <h3 className="text-base font-medium">{item.name}</h3>
                  <span className="font-mono text-[11px] text-[#5A5D62]">
                    {item.ops} · {item.method}
                  </span>
                </div>
                <p className="text-[13px] text-[#3B6D11]">
                  <b>+</b> {item.pro}
                </p>
                <p className="mt-1 text-[13px] text-[#A32D2D]">
                  <b>−</b> {item.con}
                </p>
                <p
                  className="mt-3 rounded-[8px] px-3 py-2 text-[13px]"
                  style={{ backgroundColor: ACCENT_SOFT, color: ACCENT }}
                >
                  <b>Авто РО:</b> {item.us}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section id="pipeline" className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Архитектура"
            title="От телеметрии до интеграции с Авто РТК."
            lead="Модуль единой платформы АБУ — автономный, но связанный с контуром план/факт."
          />
          <div className="grid gap-10 lg:grid-cols-2">
            <div className="rounded-[20px] bg-[#0B0D0E] p-9 font-mono text-white">
              <h4 className="mb-5 font-sans text-[13px] font-medium uppercase tracking-[0.08em] text-[#9CA0A8]">
                Pipeline
              </h4>
              {[
                "10 telemetry channels (1–10 sec)",
                "QC gate → 6.1 missing / 6.2 bad quality",
                "Rule Engine (threshold classifier)",
                "ML model inference (97.5% accuracy)",
                "Fusion + confidence scoring",
                "18 ops in 6 categories → timeline",
                "Export + API + Auto RTK integration",
              ].map((eq) => (
                <div
                  key={eq}
                  className="border-t border-[#1A1D20] py-4 text-[15px] leading-[2.2] text-[#C9CCD1] first:border-0 first:pt-0"
                >
                  {eq}
                </div>
              ))}
            </div>
            <div className="divide-y divide-[#E5E3DC] border-t border-[#E5E3DC]">
              {pipelineSections.map((row) => (
                <div key={row.key} className="grid grid-cols-[64px_1fr] gap-4 py-[18px]">
                  <div className="font-mono text-lg" style={{ color: ACCENT }}>
                    {row.key}
                  </div>
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
            title="Что важно знать о модуле."
            lead="Точность, таксономия, датчики и место в единой системе АБУ."
          />
          <div className="divide-y divide-[#E5E3DC] border-b border-[#E5E3DC]">
            {faqItems.map((item) => (
              <details key={item.q} className="group py-6">
                <summary className="flex cursor-pointer list-none items-center justify-between text-[17px] font-medium marker:content-none">
                  {item.q}
                  <span className="font-mono text-xl text-[#5A5D62] transition group-open:rotate-45">
                    +
                  </span>
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
              18 операций · 97.5% · модуль единой АБУ.
            </h2>
            <p className="mt-4 max-w-[520px] text-[17px] text-[#C9CCD1]">
              Авто РО — автоклассификация буровых операций. Работает автономно и в связке с{" "}
              <Link href={RTK_HREF} className="text-[#C4B5FD] underline hover:opacity-80">
                Авто РТК
              </Link>
              .
            </p>
            <div className="mt-7 flex flex-wrap gap-3">
              <Link
                href={DEMO_HREF}
                className="inline-flex items-center gap-2 rounded-[10px] px-[22px] py-3.5 text-[15px] font-medium text-white hover:opacity-90"
                style={{ backgroundColor: ML_ACCENT }}
              >
                Смотреть демо <span aria-hidden>→</span>
              </Link>
              <Link
                href={DISCUSS_HREF}
                className="inline-flex items-center gap-2 rounded-[10px] border border-[#2A2D30] px-[22px] py-3.5 text-[15px] font-medium text-white hover:border-[#4E5158]"
              >
                Обсудить проект
              </Link>
            </div>
          </div>
          <aside className="rounded-[20px] border border-[#1A1D20] bg-[#111417] p-7">
            <div className="mb-4 font-mono text-[11px] uppercase tracking-[0.1em] text-[#7E8189]">
              Экосистема АБУ
            </div>
            <div className="font-mono text-base leading-[1.9] text-white">
              <span className="text-[#7EC8F2]">Телеметрия</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#C4B5FD]">Авто РО</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#9FD89F]">Авто РТК</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#E8C078]">Отчёт</span>
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
                Авто РО — модуль автоклассификации операций в единой системе автоматизированной
                буровой.
              </p>
            </div>
            <div>
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">
                Кейс
              </div>
              <ul className="space-y-2 text-sm text-[#2A2D30]">
                <li>
                  <a href="#demo" className="hover:text-[#0B0D0E]">
                    Демо интерфейса
                  </a>
                </li>
                <li>
                  <a href="#compare" className="hover:text-[#0B0D0E]">
                    Сравнение
                  </a>
                </li>
                <li>
                  <Link href={RTK_HREF} className="hover:text-[#0B0D0E]">
                    Модуль Авто РТК
                  </Link>
                </li>
              </ul>
            </div>
            <div>
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">
                Компания
              </div>
              <ul className="space-y-2 text-sm text-[#2A2D30]">
                <li>
                  <Link href="/about" className="hover:text-[#0B0D0E]">
                    О нас
                  </Link>
                </li>
                <li>
                  <Link href={DISCUSS_HREF} className="hover:text-[#0B0D0E]">
                    Обсудить проект
                  </Link>
                </li>
              </ul>
            </div>
            <div>
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">
                Связь
              </div>
              <ul className="space-y-2 text-sm text-[#2A2D30]">
                <li>
                  <a href="https://t.me/ermanai" className="hover:text-[#0B0D0E]">
                    @ermanai
                  </a>
                </li>
                <li>
                  <a href="https://t.me/ermanai_bot" className="hover:text-[#0B0D0E]">
                    @ermanai_bot
                  </a>
                </li>
              </ul>
            </div>
          </div>
          <div className="mt-8 flex flex-wrap justify-between gap-3 border-t border-[#E5E3DC] pt-8 font-mono text-xs tracking-[0.06em] text-[#5A5D62]">
            <span>1982 — 2026 · ERMAN AI</span>
            <span>Авто РО · Case Study</span>
          </div>
        </div>
      </footer>
    </div>
  );
}
