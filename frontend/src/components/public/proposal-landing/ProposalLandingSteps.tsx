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
    { n: 1, label: "Клиент" },
    { n: 2, label: "Решение" },
    { n: 3, label: "КП" },
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
        <span className="chip active">После контакта</span>
        <span className="chip">Cold outreach</span>
        <span className="chip">Проактивное</span>
      </div>
      <div className="mock-form">
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">Клиент</div>
          <div className="mock-field__val">ООО «ЛогистикПро»</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Контакт</div>
          <div className="mock-field__val">Иван Петров</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Отрасль</div>
          <div className="mock-field__val">Логистика</div>
        </div>
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">Задача клиента</div>
          <div className="mock-field__val mock-field__val--sm">
            Автоматизировать обработку заявок на отгрузку — 1500 ед/мес, много ручных ошибок
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
      <div className="mock-form">
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">Решение</div>
          <div className="mock-field__val">Автоматизация обработки заявок</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Срок</div>
          <div className="mock-field__val">8 нед</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Стоимость</div>
          <div className="mock-field__val">600{"\u00a0"}000{"\u00a0"}₽</div>
        </div>
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">Deliverables</div>
          <div className="mock-field__val mock-field__val--sm">
            CRM-интеграция · Workflow · Обучение · Документация
          </div>
        </div>
      </div>
      <div className="mock-preview">
        <div className="mock-preview__row">
          <span>Из калькулятора ROI</span>
          <b>Окупаемость 0.2{"\u00a0"}мес</b>
        </div>
        <div className="mock-preview__row">
          <span>Оплата</span>
          <b>50% / 50%</b>
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
          <b>Обращение</b>
          <span>Персонализированное intro</span>
          <small>§1 · greeting</small>
        </div>
        <div className="mock-card">
          <b>Scope of Work</b>
          <span>In / Out списки</span>
          <small>§4–5 · scope</small>
        </div>
        <div className="mock-card">
          <b>Этапы и сроки</b>
          <span>3 фазы · 8 нед</span>
          <small>§6 · timeline</small>
        </div>
        <div className="mock-card">
          <b>Стоимость + ROI</b>
          <span>600K · NPV из calc</span>
          <small>§7 · PDF/DOCX</small>
        </div>
      </div>
      <table className="mock-table">
        <thead>
          <tr>
            <th>Фаза</th>
            <th>Нед</th>
            <th>Результат</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Аудит и ТЗ</td>
            <td>2</td>
            <td><span className="delta-pos">ТЗ согласовано</span></td>
          </tr>
          <tr>
            <td>Разработка</td>
            <td>4</td>
            <td><span className="delta-pos">MVP в prod</span></td>
          </tr>
          <tr>
            <td>Запуск</td>
            <td>2</td>
            <td><span className="delta-pos">Обучение</span></td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}

export function ProposalLandingSteps() {
  return (
    <div className="steps">
      <article className="step">
        <StepText
          num="1"
          tag="Клиент"
          title="Компания, контакт и задача — плюс сценарий отправки"
          desc={`Три сценария: после первого контакта, cold outreach или проактивное предложение. Для cold — можно без цены. Контекст предыдущего разговора попадает в\u00a0обращение.`}
          bullets={[
            "После контакта — классическое КП с\u00a0ценой и\u00a0ROI",
            "Cold outreach — фокус на\u00a0проблеме, без прайса",
            "Проактивное — инициатива с\u00a0обоснованием ценности",
          ]}
        />
        <div className="step__media">
          <Step1Mock />
        </div>
      </article>

      <article className="step step--reverse">
        <StepText
          num="2"
          tag="Решение"
          title="Что предлагаете, сроки, стоимость — и\u00a0данные из калькулятора"
          desc={`Название решения, deliverables, бюджет и\u00a0график оплаты. Привязка расчёта ROI — КП автоматически получит окупаемость и\u00a0рекомендацию по\u00a0процессу.`}
          bullets={[
            "Deliverables — список того, что получит клиент",
            "4 схемы оплаты: 50/50, 100%, 30/70, milestone",
            "Данные отправителя — ваш бренд в\u00a0документе",
          ]}
        />
        <div className="step__media">
          <Step2Mock />
        </div>
      </article>

      <article className="step">
        <StepText
          num="3"
          tag="КП"
          title={`Готовое коммерческое предложение — PDF и\u00a0DOCX`}
          desc={`На\u00a0выходе — профессиональный документ: обращение, понимание задачи, scope, этапы, стоимость, почему мы и\u00a0следующий шаг. Экспорт в\u00a0PDF/DOCX на\u00a0Pro+.`}
          bullets={[
            "8 разделов — от\u00a0greeting до\u00a0next step",
            "Scope in/out — чёткие границы проекта",
            "ROI из калькулятора — в\u00a0блоке стоимости",
          ]}
        />
        <div className="step__media">
          <Step3Mock />
        </div>
      </article>
    </div>
  );
}
