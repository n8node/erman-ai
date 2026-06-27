import Link from "next/link";
import type { PublicPage } from "@/lib/api-public-pages";
import { CalculatorLandingSteps } from "@/components/public/calculator-landing/CalculatorLandingSteps";

const CALCULATOR_HREF = "/tools/calculator";

type Props = {
  page: PublicPage;
};

const outcomes = [
  {
    icon: "01 — ₽/МЕС",
    name: "Чистая выгода в месяц",
    desc: "Сколько денег процесс начинает приносить после внедрения, после вычета поддержки и амортизации внедрения.",
    value: "2 663 667 ₽",
  },
  {
    icon: "02 — МЕС",
    name: "Срок окупаемости",
    desc: "Через сколько месяцев бюджет внедрения вернётся накопленной выгодой. Меньше 12 — зелёный сигнал.",
    value: "0.2 мес",
  },
  {
    icon: "03 — FTE",
    name: "Эквивалент полных ставок",
    desc: "Сколько людей высвобождает автоматизация. Не ради сокращений — ради перевода рук на работу, которая растит выручку.",
    value: "7.94",
  },
  {
    icon: "04 — %",
    name: "ROI за горизонт",
    desc: "Возврат инвестиций за выбранный период (по умолчанию 12 мес). Здесь становится видно, что считать «успехом».",
    value: "5227.3%",
  },
];

const templates = [
  { industry: "Digital-маркетинг", name: "Обработка лидов с рекламы", stats: ["4 шага", "1200 ед/мес", "75% авто"] },
  { industry: "E-commerce и retail", name: "Обработка возвратов", stats: ["5 шагов", "400 ед/мес", "70% авто"] },
  { industry: "B2B-продажи", name: "Квалификация входящих", stats: ["4 шага", "300 ед/мес", "65% авто"] },
  { industry: "HR и рекрутинг", name: "Скрининг резюме", stats: ["3 шага", "800 ед/мес", "80% авто"] },
  { industry: "Финансы и бухгалтерия", name: "Сверка первичных документов", stats: ["4 шага", "2500 ед/мес", "85% авто"] },
  { industry: "Клиентская поддержка", name: "Первая линия по типовым", stats: ["3 шага", "5000 ед/мес", "78% авто"] },
  { industry: "Юридический и документооборот", name: "Подготовка типовых договоров", stats: ["5 шагов", "120 ед/мес", "60% авто"] },
  { industry: "Недвижимость", name: "Обработка заявок с площадок", stats: ["4 шага", "900 ед/мес", "72% авто"] },
  { industry: "IT и digital-агентства", name: "Подготовка КП и ТЗ", stats: ["4 шага", "80 ед/мес", "55% авто"] },
  { industry: "Производство и логистика", name: "Обработка заявок на отгрузку", stats: ["5 шагов", "1500 ед/мес", "70% авто"] },
];

const faqItems = [
  {
    q: "Это бесплатно?",
    a: "Да. Калькулятор открыт для всех — без регистрации и лимитов на расчёты. Платное — только дальнейшие шаги: КП, аудит, внедрение.",
  },
  {
    q: "Откуда берутся цифры в шаблонах?",
    a: "Из 100 000+ реальных автоматизированных процессов, которые прошли через наши инструменты и команды. Это медианы, не маркетинг.",
  },
  {
    q: "Если у нас нестандартный процесс?",
    a: "Опишите его вручную — можно добавить любое число шагов со своими ставками и временем. Шаблоны нужны для скорости, но не обязательны.",
  },
  {
    q: "Что включает «бюджет внедрения»?",
    a: "Разовые затраты: аудит процесса, разработка, интеграции с вашими системами, обучение команды. Помесячные расходы идут в Om (поддержка).",
  },
  {
    q: "Можно поделиться расчётом с коллегами?",
    a: "Да. На шаге «Результат» есть кнопка «Публичная ссылка» — открывается копия расчёта без вашего входа.",
  },
  {
    q: "Что делать после расчёта?",
    a: "Если ROI устраивает — нажмите «Создать КП», и команда Erman AI пришлёт коммерческое предложение по этому расчёту.",
  },
];

function PrimaryButton({
  href,
  children,
  className = "",
}: {
  href: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <Link
      href={href}
      className={`inline-flex items-center gap-2 rounded-[10px] bg-[#0B0D0E] px-[22px] py-3.5 text-[15px] font-medium text-white transition hover:-translate-y-px hover:bg-[#1a1d20] ${className}`}
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
      <span className="text-[#2D3FE5]">●</span> {children}
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
    <div className="rounded-[20px] border border-[#E5E3DC] bg-[#EBEDFE] px-8 py-10 sm:flex sm:items-center sm:justify-between sm:gap-8">
      <div className="max-w-xl">
        <h3 className="text-xl font-semibold tracking-[-0.02em] text-[#0B0D0E]">{title}</h3>
        <p className="mt-2 text-[15px] leading-relaxed text-[#2A2D30]">{text}</p>
      </div>
      <PrimaryButton href={CALCULATOR_HREF} className="mt-6 shrink-0 sm:mt-0">
        {buttonLabel} <span aria-hidden>→</span>
      </PrimaryButton>
    </div>
  );
}

export function CalculatorLandingPage({ page }: Props) {
  return (
    <div className="calculator-landing bg-white text-[#0B0D0E]">
      <header className="sticky top-0 z-50 border-b border-[#1A1D20] bg-[#0B0D0E] text-white">
        <div className="mx-auto flex h-16 max-w-[1200px] items-center justify-between gap-6 px-7">
          <Link href="/" className="flex items-center gap-2.5 font-semibold">
            <span className="grid h-[30px] w-[30px] place-items-center rounded-full bg-white text-sm font-bold text-[#0B0D0E]">
              E
            </span>
            <span className="text-[15px]">Erman AI</span>
          </Link>
          <nav className="hidden items-center gap-7 text-sm text-[#C9CCD1] md:flex">
            <a href="#what" className="hover:text-white">Что считает</a>
            <a href="#steps" className="hover:text-white">Как работает</a>
            <a href="#templates" className="hover:text-white">Шаблоны</a>
            <a href="#model" className="hover:text-white">Модель</a>
            <a href="#faq" className="hover:text-white">FAQ</a>
            <Link
              href={CALCULATOR_HREF}
              className="rounded-lg bg-white px-3.5 py-2 font-medium text-[#0B0D0E] hover:bg-[#F3F1EA]"
            >
              Открыть калькулятор →
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
              Посчитайте выгоду от автоматизации{" "}
              <em className="font-mono not-italic text-[#2D3FE5]">за 3 минуты</em>
            </h1>
            <p className="mt-6 max-w-[540px] text-[17px] leading-relaxed text-[#2A2D30]">
              Финансовая модель Erman AI: введите процесс — получите чистую выгоду в месяц, срок
              окупаемости, эквивалент FTE и ROI на горизонте. Без таблиц и консультантов. Без
              регистрации.
            </p>
            <div className="mt-9 flex flex-wrap gap-3">
              <PrimaryButton href={CALCULATOR_HREF}>
                Начать расчёт <span aria-hidden>→</span>
              </PrimaryButton>
              <GhostButton href="#steps">Как это работает</GhostButton>
            </div>
            <div className="mt-7 flex flex-wrap gap-5 font-mono text-[12.5px] uppercase tracking-[0.08em] text-[#5A5D62]">
              <span><b className="font-medium text-[#0B0D0E]">3 шага</b> · процесс → финансы → результат</span>
              <span><b className="font-medium text-[#0B0D0E]">10 отраслей</b> · готовые шаблоны</span>
              <span><b className="font-medium text-[#0B0D0E]">Бесплатно</b> · без регистрации</span>
            </div>
          </div>

          <aside className="rounded-[20px] border border-[#E5E3DC] bg-white p-7 shadow-[0_30px_60px_-30px_rgba(11,13,14,0.18)]">
            <div className="mb-5 flex items-center justify-between border-b border-[#E5E3DC] pb-4 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">
              <span className="inline-flex items-center gap-1.5 font-medium text-[#0B0D0E]">
                <span className="grid h-5 w-5 place-items-center rounded-full bg-[#0B0D0E] text-[11px] text-white">3</span>
                Результат
              </span>
              <span>ROI · пример</span>
            </div>
            <div className="grid grid-cols-2 gap-px overflow-hidden rounded-[14px] bg-[#E5E3DC]">
              {[
                { label: "Чистая выгода / мес", value: "2 663 667 ₽", big: true },
                { label: "Окупаемость", value: "0.2 мес", big: true },
                { label: "FTE", value: "7.94", hint: "эквивалент полных ставок" },
                { label: "ROI за горизонт", value: "5227.3%" },
              ].map((m) => (
                <div key={m.label} className="flex flex-col gap-2 bg-white p-[18px]">
                  <div className="font-mono text-[11px] uppercase tracking-[0.1em] text-[#5A5D62]">{m.label}</div>
                  <div className={`font-mono font-medium tracking-[-0.02em] text-[#0B0D0E] ${m.big ? "text-[30px]" : "text-[26px]"}`}>
                    {m.value}
                  </div>
                  {m.hint && <div className="text-xs text-[#5A5D62]">{m.hint}</div>}
                </div>
              ))}
            </div>
            <div className="mt-[18px] flex items-center gap-2.5 rounded-[10px] bg-[#E6F1E8] px-3.5 py-3 text-[13.5px] text-[#156F3F] before:h-[7px] before:w-[7px] before:shrink-0 before:rounded-full before:bg-[#156F3F] before:content-['']">
              Рекомендуется автоматизировать — окупаемость &lt; 12 мес, положительный ROI
            </div>
          </aside>
        </div>
      </section>

      <section className="bg-[#0B0D0E] px-7 py-[22px] text-white">
        <div className="mx-auto flex max-w-[1200px] flex-wrap items-center justify-between gap-7">
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">Финансовая модель</span>
          <span className="font-mono text-[clamp(15px,1.6vw,19px)] tracking-[-0.01em]">
            <span className="text-[#8FA1FF]">Чистая выгода</span>
            <span className="px-1 text-[#7E8189]">=</span>
            <span className="text-[#8FA1FF]">Hm</span>
            <span className="text-[#7E8189]">·</span>
            <span className="text-[#8FA1FF]">Ch</span>
            <span className="px-1 text-[#7E8189]">−</span>
            <span className="text-[#8FA1FF]">Om</span>
            <span className="px-1 text-[#7E8189]">−</span>
            <span className="text-[#8FA1FF]">I₀</span>
            <span className="text-[#7E8189]">/</span>
            <span className="text-[#8FA1FF]">t</span>
          </span>
          <span className="font-mono text-[11.5px] uppercase tracking-[0.14em] text-[#9CA0A8]">Прозрачно. Без чёрных ящиков.</span>
        </div>
      </section>

      <section id="what" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Что вы получите"
            title="Четыре цифры, которые отвечают на вопрос «стоит ли»."
            lead="Каждая метрика рассчитана по явной формуле. Можно открыть детализацию, изменить любую переменную и увидеть, как смещается результат — в реальном времени."
          />
          <div className="grid gap-px overflow-hidden rounded-[20px] border border-[#E5E3DC] bg-[#E5E3DC] sm:grid-cols-2 xl:grid-cols-4">
            {outcomes.map((item) => (
              <div key={item.icon} className="flex min-h-[220px] flex-col gap-3.5 bg-white p-7">
                <div className="font-mono text-xs tracking-[0.08em] text-[#2D3FE5]">{item.icon}</div>
                <h3 className="text-[15px] font-medium">{item.name}</h3>
                <p className="text-[13.5px] leading-relaxed text-[#5A5D62]">{item.desc}</p>
                <div className="mt-auto font-mono text-[38px] font-medium leading-none tracking-[-0.03em]">{item.value}</div>
              </div>
            ))}
          </div>
          <div className="mt-10">
            <MidPageCta
              title="Узнайте свои цифры за 3 минуты"
              text="Откройте калькулятор, выберите шаблон отрасли или опишите свой процесс — результат появится сразу, без регистрации."
              buttonLabel="Рассчитать ROI"
            />
          </div>
        </div>
      </section>

      <section id="steps" className="border-t border-[#E5E3DC] bg-[#F5F4EF] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Как работает"
            title="Три шага — потому что считать ROI без процесса бессмысленно."
            lead="Сначала калькулятор разбирает процесс по операциям и считает экономию часов. Только потом — деньги. И только потом — выводы. Каждый шаг можно сохранить как черновик."
          />
          <CalculatorLandingSteps />
          <div className="mt-10">
            <MidPageCta
              title="Готовы посчитать свой процесс?"
              text="Три шага в калькуляторе — и у вас будут цифры для разговора с CFO, советом или клиентом."
              buttonLabel="Открыть калькулятор"
            />
          </div>
        </div>
      </section>

      <section id="templates" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Шаблоны"
            title="10 отраслей — не нужно описывать процесс с нуля."
            lead="Каждый шаблон собран на основе реальных проектов: количество шагов, средний объём, типовая доля автоматизации."
          />
          <div className="grid gap-4 md:grid-cols-2">
            {templates.map((tpl) => (
              <div key={tpl.name} className="grid grid-cols-[1fr_auto] items-center gap-4 rounded-[14px] border border-[#E5E3DC] bg-white px-6 py-5">
                <div>
                  <div className="mb-1.5 font-mono text-[11px] uppercase tracking-[0.08em] text-[#5A5D62]">{tpl.industry}</div>
                  <div className="mb-3 text-base font-medium">{tpl.name}</div>
                  <div className="flex flex-wrap gap-3.5 font-mono text-xs text-[#2A2D30]">
                    {tpl.stats.map((s) => (
                      <span key={s} className="inline-flex items-center gap-1.5 before:h-[5px] before:w-[5px] before:rounded-full before:bg-[#2D3FE5] before:content-['']">{s}</span>
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
            title="Один калькулятор — две роли."
            lead="Бизнес считает, стоит ли вкладываться. Интегратор — обосновывает цену и границы проекта."
          />
          <div className="grid overflow-hidden rounded-[20px] border border-[#E5E3DC] md:grid-cols-2">
            <div className="bg-white p-9">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">01 — Бизнес</div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">Прежде чем платить за внедрение — посчитайте, окупится ли</h3>
              <p className="mt-3.5 text-[15px] text-[#2A2D30]">Калькулятор отвечает на вопрос CFO: «через сколько месяцев это вернёт деньги и что получим за год».</p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#2A2D30]">
                {["Защитить бюджет внедрения цифрами", "Сравнить несколько процессов", "Запросить КП из расчёта"].map((item) => (
                  <li key={item} className="flex gap-2.5 before:text-[#5A5D62] before:content-['—']">{item}</li>
                ))}
              </ul>
            </div>
            <div className="bg-[#0B0D0E] p-9 text-white">
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#9CA0A8]">02 — Интегратор</div>
              <h3 className="text-[26px] font-semibold leading-[1.1] tracking-[-0.02em]">Превратить разговор о цене в разговор об ROI</h3>
              <p className="mt-3.5 text-[15px] text-[#C9CCD1]">Когда клиент видит окупаемость 0.2 месяца, торг про скидку 10% теряет смысл.</p>
              <ul className="mt-5 space-y-2.5 text-[14.5px] text-[#C9CCD1]">
                {["Готовые шаблоны под 10 отраслей", "Публичная ссылка с расчётом", "Экспорт PDF под бренд-бук"].map((item) => (
                  <li key={item} className="flex gap-2.5 before:text-[#7E8189] before:content-['—']">{item}</li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </section>

      <section id="model" className="border-t border-[#E5E3DC] px-7 py-24">
        <div className="mx-auto max-w-[1200px]">
          <SectionHead
            eyebrow="Финансовая модель"
            title="Никаких чёрных ящиков. Только формулы."
            lead="Калькулятор — это не ML-магия. Это явная финансовая модель Erman AI, основанная на опыте более чем 100 000 автоматизированных процессов."
          />
          <div className="grid gap-10 lg:grid-cols-2">
            <div className="rounded-[20px] bg-[#0B0D0E] p-9 font-mono text-white">
              <h4 className="mb-5 font-sans text-[13px] font-medium uppercase tracking-[0.08em] text-[#9CA0A8]">Уравнения модели</h4>
              {[
                "Чистая выгода/мес = Hm·Ch − Om − I₀/t",
                "Окупаемость = I₀ / (Hm·Ch − Om)",
                "FTE = Hm / (22·8)",
                "ROI = ((Hm·Ch − Om)·t − I₀) / I₀ · 100%",
              ].map((eq) => (
                <div key={eq} className="border-t border-[#1A1D20] py-4 first:border-0 first:pt-0 text-[15px] leading-[2.2] text-[#C9CCD1]">{eq}</div>
              ))}
            </div>
            <div className="divide-y divide-[#E5E3DC] border-t border-[#E5E3DC]">
              {[
                { key: "Hm", title: "Экономия часов в месяц", text: "Считается из объёма процесса × минут на единицу × % автоматизации." },
                { key: "Ch", title: "Полная стоимость часа", text: "ЗП + налоги + бонусы + накладные + рабочие часы." },
                { key: "I₀", title: "Бюджет внедрения, разово", text: "Аудит, разработка, обучение, интеграции." },
                { key: "Om", title: "Поддержка, ₽/мес", text: "VPS, лицензии, домен, тех. поддержка." },
                { key: "t", title: "Горизонт оценки, мес", text: "6 — для осторожных, 12 — стандарт, 24 — для стратегических проектов." },
              ].map((row) => (
                <div key={row.key} className="grid grid-cols-[64px_1fr] gap-4 py-[18px]">
                  <div className="font-mono text-lg text-[#2D3FE5]">{row.key}</div>
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
            title="Что важно знать до расчёта."
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
              Один расчёт — и разговор про автоматизацию становится разговором про деньги.
            </h2>
            <p className="mt-4 max-w-[520px] text-[17px] text-[#C9CCD1]">
              Три минуты, четыре цифры, ноль регистраций. Если результат вам подходит — нажмёте «Создать КП», и мы продолжим.
            </p>
            <div className="mt-7 flex flex-wrap gap-3">
              <Link href={CALCULATOR_HREF} className="inline-flex items-center gap-2 rounded-[10px] bg-white px-[22px] py-3.5 text-[15px] font-medium text-[#0B0D0E] hover:bg-[#F3F1EA]">
                Открыть калькулятор <span aria-hidden>→</span>
              </Link>
              <Link href="/discuss" className="inline-flex items-center gap-2 rounded-[10px] border border-[#2A2D30] px-[22px] py-3.5 text-[15px] font-medium text-white hover:border-[#4E5158]">
                Обсудить проект
              </Link>
            </div>
          </div>
          <aside className="rounded-[20px] border border-[#1A1D20] bg-[#111417] p-7">
            <div className="mb-4 font-mono text-[11px] uppercase tracking-[0.1em] text-[#7E8189]">Финансовая модель</div>
            <div className="font-mono text-base leading-[1.9] text-white">
              <span className="text-[#8FA1FF]">ROI</span>
              <span className="text-[#7E8189]"> = (</span>
              <span className="text-[#8FA1FF]">Hm</span>
              <span className="text-[#7E8189]">·</span>
              <span className="text-[#8FA1FF]">Ch</span>
              <span className="text-[#7E8189]"> − </span>
              <span className="text-[#8FA1FF]">Om</span>
              <span className="text-[#7E8189]">)·</span>
              <span className="text-[#8FA1FF]">t</span>
              <span className="text-[#7E8189]"> − I₀ / I₀ · 100%</span>
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
                Внедрение AI и автоматизация бизнес-процессов. 100 000+ работающих процессов на наших инструментах.
              </p>
            </div>
            <div>
              <div className="mb-3.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-[#5A5D62]">Инструменты</div>
              <ul className="space-y-2 text-sm text-[#2A2D30]">
                <li><Link href={CALCULATOR_HREF} className="hover:text-[#0B0D0E]">Калькулятор ROI</Link></li>
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
            <span>Финансовая модель v.1.0</span>
          </div>
        </div>
      </footer>
    </div>
  );
}
