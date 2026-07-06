import Link from "next/link";
import type { PublicPage } from "@/lib/api-public-pages";
import { LandingBrandLink } from "@/components/public/LandingBrandLink";
import { AutoRtkLandingSteps } from "@/components/public/auto-rtk-landing/AutoRtkLandingSteps";
import { AutoRtkDemoShowcase } from "@/components/public/auto-rtk-landing/AutoRtkDemoShowcase";

const DEMO_HREF = "#demo";
const DISCUSS_HREF = "/discuss";
const ACCENT = "#185FA5";
const ACCENT_SOFT = "#E6F1FB";

type Props = {
  page: PublicPage;
};

const outcomes = [
  {
    icon: "01 — TREE",
    name: "Иерархия активов",
    desc: "Месторождение → куст → скважина. Единая модель данных для всех объектов бурения на площадке.",
    value: "3 ур.",
  },
  {
    icon: "02 — PLAN",
    name: "План vs факт",
    desc: "Сравнение по нагрузке, перепаду, моменту и ROP с цветовой индикацией отклонений от уставок.",
    value: "4+",
  },
  {
    icon: "03 — DEPTH",
    name: "Интервалы 50 м",
    desc: "Детализация по глубине: режим ротор/слайд, проходка, лимиты и фактические значения в таблице.",
    value: "50 м",
  },
  {
    icon: "04 — FLOW",
    name: "Согласование РТК",
    desc: "Редактирование плана, авто-сохранение, локальное согласование и рассылка Excel на email.",
    value: "3 стат.",
  },
];

const moduleItems = [
  { tag: "Панель", name: "Панель управления", stats: ["Сводка", "Последние", "Добавить"] },
  { tag: "Структура", name: "Иерархия и рассылка", stats: ["Кусты", "Email РТК", "Статусы"] },
  { tag: "Карточка", name: "Карточка скважины", stats: ["Общее", "Графики", "Таблица"] },
  { tag: "SyncDrill", name: "Графики параметров", stats: ["ROP", "WOB", "dP", "Torque"] },
  { tag: "Таблица", name: "Интервальный отчёт", stats: ["50 м", "↑↓", "Excel"] },
  { tag: "План", name: "Редактирование РТК", stats: ["Уставки", "Лимиты", "Auto-save"] },
  { tag: "Фильтры", name: "Секция и интервал", stats: ["Геология", "Метры", "Сброс"] },
  { tag: "Workflow", name: "Согласование", stats: ["Approve", "Email", "Excel"] },
];

const scenarios = [
  {
    key: "01",
    title: "Создание новой скважины",
    text: "Добавьте скважину через форму на панели управления. Импортируйте или сгенерируйте план РТК. Проверьте данные на вкладках «Общее», «Графики» и «Таблица».",
  },
  {
    key: "02",
    title: "Контроль бурения",
    text: "Откройте карточку, выберите секцию и интервал глубины. Следите за KPI и отклонениями на вкладке «Общее», детализируйте на графиках или в таблице.",
  },
  {
    key: "03",
    title: "Обновление плана",
    text: "Включите режим «Редактировать план», внесите изменения в таблицу с авто-сохранением. Согласуйте локально или отправьте Excel на согласование.",
  },
];

const pipelineSections = [
  { key: "01", title: "Иерархия", text: "Месторождение, куст, скважина — единая структура и статусы планов." },
  { key: "02", title: "Импорт плана", text: "Загрузка РТК из PDF, таблиц или ручной ввод уставок и ограничений." },
  { key: "03", title: "Телеметрия", text: "Приём фактических параметров бурения и сопоставление с планом." },
  { key: "04", title: "Аналитика", text: "KPI, графики SyncDrill, таблица по интервалам 50 м." },
  { key: "05", title: "Фильтрация", text: "Секция, интервал по метрам или датам, экспорт в Excel." },
  { key: "06", title: "Согласование", text: "Workflow: черновик → на согласовании → согласован + email." },
];

const faqItems = [
  {
    q: "Откуда берётся план РТК?",
    a: "План импортируется из PDF, таблицы Excel или вводится вручную в режиме редактирования. Уставки и ограничения задаются по интервалам глубины.",
  },
  {
    q: "Как читать отклонения на вкладке «Общее»?",
    a: "Зелёные значения — в пределах нормы относительно уставки. Оранжевые и красные — превышение или недобор. Процент соответствия показывает, насколько факт близок к целевому коридору.",
  },
  {
    q: "Что означают статусы плана?",
    a: "«Черновик» — план в работе. «На согласовании» — отправлен по email. «Согласован» — утверждён и доступен для контроля бурения.",
  },
  {
    q: "Как работает авто-сохранение?",
    a: "При редактировании таблицы плана значение сохраняется автоматически, когда поле теряет фокус. Не нужно нажимать «Сохранить» после каждой ячейки.",
  },
  {
    q: "Можно экспортировать данные?",
    a: "Да. Табличные данные — в Excel. Графики SyncDrill — в PDF или PNG. При согласовании Excel-файл уходит получателям по email.",
  },
  {
    q: "Это часть какой системы?",
    a: "Авто РТК — модуль автоматизированной буровой установки. Веб-сервис для инженеров и технологов, интегрированный в контур управления бурением.",
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
        accent ? "hover:bg-[#134d88]" : "bg-[#0B0D0E] hover:bg-[#1a1d20]"
      } ${className}`}
      style={accent ? { backgroundColor: ACCENT } : undefined}
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

export function AutoRtkLandingPage({ page }: Props) {
  return (
    <div className="tool-landing auto-rtk-landing bg-white text-[#0B0D0E]">
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
            <a href="#modules" className="hover:text-white">
              Модули
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
              Авто РТК —{" "}
              <em className="font-mono not-italic" style={{ color: ACCENT }}>
                цифровой контур
              </em>{" "}
              бурения
            </h1>
            <p className="mt-6 max-w-[540px] text-[17px] leading-relaxed text-[#2A2D30]">
              Веб-сервис для работы с режимно-технологическими картами: иерархия скважин, контроль
              план/факт по ключевым параметрам, графики, таблицы и согласование планов с рассылкой.
            </p>
            <div className="mt-9 flex flex-wrap gap-3">
              <PrimaryButton href={DEMO_HREF} accent>
                Смотреть демо <span aria-hidden>→</span>
              </PrimaryButton>
              <GhostButton href={DISCUSS_HREF}>Обсудить проект</GhostButton>
            </div>
            <div className="mt-7 flex flex-wrap gap-5 font-mono text-[12.5px] uppercase tracking-[0.08em] text-[#5A5D62]">
              <span>
                <b className="font-medium text-[#0B0D0E]">3 уровня</b> · иерархия
              </span>
              <span>
                <b className="font-medium text-[#0B0D0E]">4 вкладки</b> · карточка
              </span>
              <span>
                <b className="font-medium text-[#0B0D0E]">Excel</b> · согласование
              </span>
            </div>
          </div>

          <aside className="rounded-[20px] border border-[#E5E3DC] bg-white p-7 shadow-[0_30px_60px_-30px_rgba(11,13,14,0.18)]">
            <div className="mb-5 flex items-center justify-between border-b border-[#E5E3DC] pb-4 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">
              <span className="inline-flex items-center gap-1.5 font-medium text-[#0B0D0E]">
                <span
                  className="grid h-5 w-5 place-items-center rounded-full text-[11px] text-white"
                  style={{ backgroundColor: ACCENT }}
                >
                  ✓
                </span>
                Карточка скважины
              </span>
              <span className="rtk-badge rtk-badge--ok">Согласован</span>
            </div>
            <div className="mb-4 font-mono text-[11px] text-[#5A5D62]">
              Северное-1 · Куст 7.2 · СК-2847
            </div>
            <div className="grid grid-cols-2 gap-px overflow-hidden rounded-[14px] bg-[#E5E3DC]">
              {[
                { label: "Проходка", value: "1920 м", big: true },
                { label: "ROP факт", value: "74 м/ч", big: true },
                { label: "Нагрузка", value: "10.03 т", hint: "уст. 13.2" },
                { label: "Соответствие", value: "72.7%" },
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
                  {m.hint && <div className="text-xs text-[#5A5D62]">{m.hint}</div>}
                </div>
              ))}
            </div>
            <div
              className="mt-[18px] flex items-center gap-2.5 rounded-[10px] px-3.5 py-3 text-[13.5px] before:h-[7px] before:w-[7px] before:shrink-0 before:rounded-full before:content-['']"
              style={{ backgroundColor: ACCENT_SOFT, color: ACCENT }}
            >
              <span className="before:inline-block before:h-[7px] before:w-[7px] before:rounded-full before:bg-[#185FA5]" />
              План vs факт — 4 параметра в пределах коридора
            </div>
          </aside>
        </div>
      </section>

      <section className="bg-[#0B0D0E] px-7 py-[22px] text-white">
        <div className="mx-auto flex max-w-[1200px] flex-wrap items-center justify-between gap-7">
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">
            Модель данных
          </span>
          <span className="font-mono text-[clamp(14px,1.5vw,18px)] tracking-[-0.01em]">
            <span className="text-[#7EC8F2]">РТК</span>
            <span className="px-1 text-[#7E8189]">=</span>
            <span className="text-[#7EC8F2]">План</span>
            <span className="px-1 text-[#7E8189]">+</span>
            <span className="text-[#7EC8F2]">Факт</span>
            <span className="px-1 text-[#7E8189]">→</span>
            <span className="text-[#9FD89F]">Согласование</span>
          </span>
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">
            Месторождение → Скважина
          </span>
        </div>
      </section>

      <section id="demo" className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Интерактивное демо"
            title="Интерфейс, который работает как на буровой."
            lead="Переключайте вкладки, смотрите обновление параметров в реальном времени. Все данные обезличены — это демонстрация возможностей системы."
          />
          <AutoRtkDemoShowcase />
        </div>
      </section>

      <section id="what" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Что реализовано"
            title="Полный цикл работы с РТК — от структуры до согласования."
            lead="Модуль встроен в контур автоматизированной буровой. Инженер видит план, факт и отклонения в одном окне."
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
            title="Три рабочих процесса — как в документации."
            lead="Создание скважины, контроль бурения по план/факт и обновление плана с согласованием."
          />
          <AutoRtkLandingSteps />
        </div>
      </section>

      <section id="modules" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Модули системы"
            title="Восемь функциональных блоков — по разделам документации."
            lead="Каждый модуль закрывает отдельную задачу инженера или технолога на буровой."
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
                      <span
                        key={s}
                        className="inline-flex items-center gap-1.5 before:h-[5px] before:w-[5px] before:rounded-full before:content-['']"
                        style={{ ["--tw-before-bg" as string]: ACCENT }}
                      >
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
            lead="Инженер-технолог контролирует параметры. Руководитель проекта согласует планы."
          />
          <div className="grid overflow-hidden rounded-[20px] border border-[#E5E3DC] md:grid-cols-2">
            <div className="bg-white p-9">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">
                01 — Инженер
              </div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">
                Контроль бурения в реальном времени
              </h3>
              <p className="mt-3.5 text-[15px] text-[#2A2D30]">
                Карточка скважины, фильтры по секции, графики SyncDrill и таблица по интервалам —
                всё для оперативных решений на буровой.
              </p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#2A2D30]">
                {["KPI план/факт", "4 параметра SyncDrill", "Экспорт Excel"].map((item) => (
                  <li key={item} className="flex gap-2.5 before:text-[#5A5D62] before:content-['—']">
                    {item}
                  </li>
                ))}
              </ul>
            </div>
            <div className="bg-[#0B0D0E] p-9 text-white">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#9CA0A8]">
                02 — Руководитель
              </div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">
                Согласование планов РТК
              </h3>
              <p className="mt-3.5 text-[15px] text-[#C9CCD1]">
                Редактирование уставок, локальное согласование или рассылка Excel на email с
                комментарием. Статус виден по каждой скважине.
              </p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#C9CCD1]">
                {["Workflow статусов", "Email-рассылка", "Иерархия объектов"].map((item) => (
                  <li key={item} className="flex gap-2.5 before:text-[#7E8189] before:content-['—']">
                    {item}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </section>

      <section id="pipeline" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Архитектура"
            title="От структуры данных до согласования."
            lead="Цепочка модулей, которую мы реализовали для клиента в нефтедобыче."
          />
          <div className="grid gap-10 lg:grid-cols-2">
            <div className="rounded-[20px] bg-[#0B0D0E] p-9 font-mono text-white">
              <h4 className="mb-5 font-sans text-[13px] font-medium uppercase tracking-[0.08em] text-[#9CA0A8]">
                Pipeline
              </h4>
              {[
                "Field → Cluster → Well hierarchy",
                "RTK plan import (PDF / Excel / manual)",
                "Telemetry → fact parameters",
                "Plan vs fact analytics + charts",
                "Filters by section & depth interval",
                "Approval workflow + email + Excel",
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
          <div className="mt-12 grid gap-4 md:grid-cols-3">
            {scenarios.map((s) => (
              <div key={s.key} className="rounded-[14px] border border-[#E5E3DC] bg-[#F5F4EF] p-6">
                <div className="mb-2 font-mono text-sm" style={{ color: ACCENT }}>
                  {s.key}
                </div>
                <h4 className="text-base font-medium">{s.title}</h4>
                <p className="mt-2 text-[14px] leading-relaxed text-[#5A5D62]">{s.text}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section id="faq" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Частые вопросы"
            title="Что важно знать о модуле."
            lead="Ответы по импорту планов, статусам, экспорту и интеграции в буровой контур."
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
              Цифровой контур бурения — от плана до согласования.
            </h2>
            <p className="mt-4 max-w-[520px] text-[17px] text-[#C9CCD1]">
              Кейс для нефтедобычи: иерархия скважин, план/факт, графики и workflow согласования РТК.
            </p>
            <div className="mt-7 flex flex-wrap gap-3">
              <Link
                href={DEMO_HREF}
                className="inline-flex items-center gap-2 rounded-[10px] px-[22px] py-3.5 text-[15px] font-medium text-white hover:opacity-90"
                style={{ backgroundColor: ACCENT }}
              >
                Смотреть демо <span aria-hidden>→</span>
              </Link>
              <Link
                href="/discuss"
                className="inline-flex items-center gap-2 rounded-[10px] border border-[#2A2D30] px-[22px] py-3.5 text-[15px] font-medium text-white hover:border-[#4E5158]"
              >
                Обсудить проект
              </Link>
            </div>
          </div>
          <aside className="rounded-[20px] border border-[#1A1D20] bg-[#111417] p-7">
            <div className="mb-4 font-mono text-[11px] uppercase tracking-[0.1em] text-[#7E8189]">
              Стек решения
            </div>
            <div className="font-mono text-base leading-[1.9] text-white">
              <span className="text-[#7EC8F2]">Web UI</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#9FD89F]">API</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#E8C078]">Telemetry</span>
              <span className="text-[#7E8189]"> → </span>
              <span className="text-[#F5A5A5]">Excel / Email</span>
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
                Кейс Авто РТК — модуль автоматизированной буровой для нефтедобычи.
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
                  <a href="#modules" className="hover:text-[#0B0D0E]">
                    Модули
                  </a>
                </li>
                <li>
                  <a href="#faq" className="hover:text-[#0B0D0E]">
                    FAQ
                  </a>
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
                  <Link href="/discuss" className="hover:text-[#0B0D0E]">
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
            <span>Авто РТК · Case Study</span>
          </div>
        </div>
      </footer>
    </div>
  );
}
