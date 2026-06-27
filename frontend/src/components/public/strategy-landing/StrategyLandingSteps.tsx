type StepTextProps = {
  num: string;
  tag: string;
  title: string;
  desc: string;
  bullets: string[];
};

function StepText({ num, tag, title, desc, bullets }: StepTextProps) {
  return (
    <div className="step__text">
      <div className="step__num">
        <b>{num}</b> {tag}
      </div>
      <h3 className="step__title">{title}</h3>
      <p className="step__desc">{desc}</p>
      <ul className="step__list">
        {bullets.map((item) => (
          <li key={item}>{item}</li>
        ))}
      </ul>
    </div>
  );
}

function MockHead({ activeStep }: { activeStep: 1 | 2 | 3 }) {
  const steps = [
    { n: 1, label: "Контекст" },
    { n: 2, label: "Цели" },
    { n: 3, label: "Стратегия" },
  ] as const;

  return (
    <div className="mock__head">
      {steps.flatMap((step, index) => {
        const items = [];
        if (index > 0) {
          items.push(<span key={`sep-${step.n}`} className="mock__sep" aria-hidden />);
        }
        items.push(
          <span
            key={step.n}
            className={`mock__step${step.n === activeStep ? "" : " inactive"}`}
          >
            <span className="n">{step.n}</span> {step.label}
          </span>,
        );
        return items;
      })}
    </div>
  );
}

function Step1Mock() {
  return (
    <div className="mock">
      <MockHead activeStep={1} />
      <div className="chip-row">
        <span className="chip">Розница</span>
        <span className="chip active">B2B-услуги</span>
        <span className="chip">Производство</span>
        <span className="chip">IT</span>
        <span className="chip">Логистика</span>
      </div>
      <div className="mock-form">
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">Компания</div>
          <div className="mock-field__val">ООО «ТехПро»</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Размер</div>
          <div className="mock-field__val">51–200</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">AI-уровень</div>
          <div className="mock-field__val">Piloting</div>
        </div>
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">Описание бизнеса</div>
          <div className="mock-field__val mock-field__val--sm">
            B2B SaaS для автоматизации продаж, 120 сотрудников, региональный рынок
          </div>
        </div>
      </div>
    </div>
  );
}

function Step2Mock() {
  return (
    <div className="mock">
      <MockHead activeStep={2} />
      <div className="chip-row">
        <span className="chip active">Снизить расходы</span>
        <span className="chip">Ускорить процессы</span>
        <span className="chip">Рост продаж</span>
        <span className="chip">Аналитика</span>
      </div>
      <div className="mock-form">
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">Ключевые процессы</div>
          <div className="mock-field__val mock-field__val--sm">Продажи · Поддержка · Маркетинг</div>
        </div>
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">Боли</div>
          <div className="mock-field__val mock-field__val--sm">
            Ручная квалификация лидов, долгая подготовка КП, нет единой аналитики
          </div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Бюджет</div>
          <div className="mock-field__val">500K–5M</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Горизонт</div>
          <div className="mock-field__val">12 мес</div>
        </div>
      </div>
      <div className="mock-preview">
        <div className="mock-preview__row">
          <span>Связано с калькулятором</span>
          <b>2{"\u00a0"}процесса</b>
        </div>
        <div className="mock-preview__row">
          <span>NPV портфеля</span>
          <b>18.4{"\u00a0"}М{"\u00a0"}₽</b>
        </div>
      </div>
    </div>
  );
}

function Step3Mock() {
  return (
    <div className="mock">
      <MockHead activeStep={3} />
      <div className="mock-grid">
        <div className="mock-card">
          <b>#1 AI-квалификация лидов</b>
          <span>Impact: высокий · Effort: средний</span>
          <small>Q1 · CRM + LLM</small>
        </div>
        <div className="mock-card">
          <b>#2 Генерация КП</b>
          <span>Impact: высокий · Effort: низкий</span>
          <small>Q1 · шаблоны + ROI</small>
        </div>
        <div className="mock-card">
          <b>#3 Аналитика воронки</b>
          <span>Impact: средний · Effort: средний</span>
          <small>Q2 · BI + прогноз</small>
        </div>
        <div className="mock-card">
          <b>Roadmap · 4 квартала</b>
          <span>12 инициатив · 6 KPI</span>
          <small>PDF · Mermaid · 30 дней</small>
        </div>
      </div>
      <table className="mock-table">
        <thead>
          <tr>
            <th>KPI</th>
            <th>Цель</th>
            <th>Срок</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Время на КП</td>
            <td>−60%</td>
            <td>Q2</td>
          </tr>
          <tr>
            <td>Конверсия лидов</td>
            <td>+25%</td>
            <td>Q3</td>
          </tr>
          <tr>
            <td>Экономия ФОТ</td>
            <td>2.6{"\u00a0"}М{"\u00a0"}₽/год</td>
            <td>12 мес</td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}

export function StrategyLandingSteps() {
  return (
    <div className="steps">
      <article className="step">
        <StepText
          num="1"
          tag="Контекст"
          title={`Опишите компанию — отрасль, масштаб и текущий уровень AI`}
          desc={`Название, описание бизнеса, размер команды, выручка и позиция на рынке. Это фундамент стратегии: без контекста рекомендации будут абстрактными.`}
          bullets={[
            "8 отраслей и\u00a0типовых профилей зрелости",
            "Уровень AI: none → exploring → piloting → scaling",
            "IT-стек и\u00a0готовность команды к\u00a0изменениям",
          ]}
        />
        <div className="step__media">
          <Step1Mock />
        </div>
      </article>

      <article className="step step--reverse">
        <StepText
          num="2"
          tag="Цели"
          title="Цели, процессы и бюджет — что болит и к чему идёте"
          desc={`Выбираете главные цели, ключевые процессы и описываете боли. Можно привязать расчёты из калькулятора ROI — стратегия получит финансовое обоснование по каждому процессу.`}
          bullets={[
            "6 целей: от\u00a0снижения расходов до\u00a0роста продаж",
            "До\u00a010\u00a0процессов из\u00a0калькулятора или вручную",
            "Бюджет, горизонт и\u00a0режим отчёта: standard / consulting",
          ]}
        />
        <div className="step__media">
          <Step2Mock />
        </div>
      </article>

      <article className="step">
        <StepText
          num="3"
          tag="Стратегия"
          title={`Готовый документ — roadmap, KPI, риски и\u00a0план на\u00a030\u00a0дней`}
          desc={`На\u00a0выходе — структурированная стратегия: executive summary, топ-3 решения, roadmap по\u00a0кварталам, матрица приоритетов, бюджет и\u00a0метрики. Экспорт в\u00a0PDF для\u00a0совета или инвесторов.`}
          bullets={[
            "15+ разделов: от\u00a0диагностики до\u00a0data governance",
            "Mermaid-диаграммы архитектуры и\u00a0процессов",
            "Следующие 30\u00a0дней — конкретные шаги, не\u00a0абстракции",
          ]}
        />
        <div className="step__media">
          <Step3Mock />
        </div>
      </article>
    </div>
  );
}
