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
    { n: 1, label: "Компания" },
    { n: 2, label: "Процессы" },
    { n: 3, label: "Аудит" },
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
        <span className="chip active">Снизить расходы</span>
        <span className="chip">Скорость</span>
        <span className="chip">Качество</span>
        <span className="chip">Quick wins</span>
      </div>
      <div className="mock-form">
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">Компания</div>
          <div className="mock-field__val">ООО «РитейлГруп»</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Размер</div>
          <div className="mock-field__val">51–200</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Бюджет</div>
          <div className="mock-field__val">500K–2M</div>
        </div>
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">IT-стек</div>
          <div className="mock-field__val mock-field__val--sm">1C · CRM · Excel · Helpdesk</div>
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
        <span className="chip active">Входящие лиды</span>
        <span className="chip">Обработка заказов</span>
        <span className="chip">Поддержка</span>
        <span className="chip">+ ещё</span>
      </div>
      <div className="mock-grid">
        <div className="mock-card">
          <b>Обработка заказов</b>
          <span>Операции · 2–3 FTE</span>
          <small>Score 8.7 · Quick win</small>
        </div>
        <div className="mock-card">
          <b>Входящие лиды</b>
          <span>Продажи · 1 FTE</span>
          <small>Score 7.9 · Quick win</small>
        </div>
        <div className="mock-card">
          <b>Клиентская поддержка</b>
          <span>Support · 4–10 FTE</span>
          <small>Score 6.2</small>
        </div>
        <div className="mock-card">
          <b>Выставление счетов</b>
          <span>Финансы · 2 FTE</span>
          <small>Score 5.8</small>
        </div>
      </div>
    </div>
  );
}

function Step3Mock() {
  return (
    <div className="mock">
      <MockHead activeStep={3} />
      <table className="mock-table">
        <thead>
          <tr>
            <th>#</th>
            <th>Процесс</th>
            <th>Score</th>
            <th>Экономия</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>1</td>
            <td>Обработка заказов</td>
            <td>8.7</td>
            <td><span className="delta-pos">420K{"\u00a0"}₽</span></td>
          </tr>
          <tr>
            <td>2</td>
            <td>Входящие лиды</td>
            <td>7.9</td>
            <td><span className="delta-pos">280K{"\u00a0"}₽</span></td>
          </tr>
          <tr>
            <td>3</td>
            <td>Поддержка</td>
            <td>6.2</td>
            <td><span className="delta-pos">190K{"\u00a0"}₽</span></td>
          </tr>
        </tbody>
      </table>
      <div className="mock-preview">
        <div className="mock-preview__row">
          <span>Quick wins</span>
          <b>3 процесса</b>
        </div>
        <div className="mock-preview__row">
          <span>Roadmap</span>
          <b>3 фазы · 6 мес</b>
        </div>
      </div>
    </div>
  );
}

export function AuditLandingSteps() {
  return (
    <div className="steps">
      <article className="step">
        <StepText
          num="1"
          tag="Компания"
          title="Цели, IT-ландшафт и критерии приоритизации"
          desc={`Название компании, главная цель автоматизации, размер команды и\u00a0бюджет. Критерии приоритета — экономия, скорость, quick wins — влияют на\u00a0ранжирование процессов.`}
          bullets={[
            "6 критериев приоритета — можно выбрать несколько",
            "IT-системы: CRM, ERP, 1C, таблицы, helpdesk",
            "Дублирование данных — фактор сложности интеграций",
          ]}
        />
        <div className="step__media">
          <Step1Mock />
        </div>
      </article>

      <article className="step step--reverse">
        <StepText
          num="2"
          tag="Процессы"
          title={`От 2 до 7 процессов — с\u00a0оценкой готовности к\u00a0автоматизации`}
          desc={`Для каждого процесса: частота, время, FTE, ставка, ошибки, системы и\u00a0узкие места. Есть готовые шаблоны — лиды, заказы, поддержка, HR, склад.`}
          bullets={[
            "8 шаблонов процессов + свой вариант",
            "Automation score — формула на\u00a0backend, не LLM",
            "Quick win — процессы с\u00a0высоким score и\u00a0низкой сложностью",
          ]}
        />
        <div className="step__media">
          <Step2Mock />
        </div>
      </article>

      <article className="step">
        <StepText
          num="3"
          tag="Аудит"
          title={`Приоритеты, roadmap и\u00a0quick wins — готовый отчёт`}
          desc={`На\u00a0выходе — ранжирование процессов по\u00a0automation score, оценка экономии, roadmap по\u00a0фазам, риски и\u00a0метрики для\u00a0отслеживания.`}
          bullets={[
            "Таблица приоритетов с\u00a0экономией и\u00a0окупаемостью",
            "Roadmap: фазы, процессы, deliverables",
            "Следующие шаги — конкретные действия после аудита",
          ]}
        />
        <div className="step__media">
          <Step3Mock />
        </div>
      </article>
    </div>
  );
}
