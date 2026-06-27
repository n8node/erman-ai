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
    { n: 1, label: "Сайт" },
    { n: 2, label: "Скан" },
    { n: 3, label: "Отчёт" },
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
      <div className="mock-form">
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">URL сайта</div>
          <div className="mock-field__val mock-field__val--sm">https://example-shop.ru</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Отрасль</div>
          <div className="mock-field__val">E-commerce</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Размер</div>
          <div className="mock-field__val">Малый бизнес</div>
        </div>
      </div>
      <div className="chip-row">
        <span className="chip active">Формы</span>
        <span className="chip">Реклама</span>
        <span className="chip active">Онлайн-оплата</span>
        <span className="chip">Иностр. сервисы</span>
      </div>
      <div className="mock-preview">
        <div className="mock-preview__row">
          <span>Глубина скана</span>
          <b>Стандарт · 10{"\u00a0"}стр.</b>
        </div>
      </div>
    </div>
  );
}

function Step2Mock() {
  return (
    <div className="mock">
      <MockHead activeStep={2} />
      <table className="mock-table">
        <thead>
          <tr>
            <th>Проверка</th>
            <th>Статус</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>SSL / HTTPS</td>
            <td><span className="mock-status mock-status--ok">OK</span></td>
          </tr>
          <tr>
            <td>Политика ПД</td>
            <td><span className="mock-status mock-status--risk">Риск</span></td>
          </tr>
          <tr>
            <td>Cookie-баннер</td>
            <td><span className="mock-status mock-status--risk">Риск</span></td>
          </tr>
          <tr>
            <td>Реквизиты ИНН</td>
            <td><span className="mock-status mock-status--ok">OK</span></td>
          </tr>
          <tr>
            <td>Согласие у форм</td>
            <td><span className="mock-status mock-status--risk">Риск</span></td>
          </tr>
        </tbody>
      </table>
      <div className="mock-preview">
        <div className="mock-preview__row">
          <span>Страниц проверено</span>
          <b>10 / 10</b>
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
          <b>Нет политики ПД</b>
          <span>152-ФЗ · высокий</span>
          <small>до 6{"\u00a0"}000{"\u00a0"}000{"\u00a0"}₽</small>
        </div>
        <div className="mock-card">
          <b>Cookie без согласия</b>
          <span>149-ФЗ · средний</span>
          <small>до 300{"\u00a0"}000{"\u00a0"}₽</small>
        </div>
        <div className="mock-card">
          <b>Формы без чекбокса</b>
          <span>152-ФЗ · высокий</span>
          <small>Как исправить →</small>
        </div>
        <div className="mock-card">
          <b>Итого штрафов</b>
          <span>7 рисков · 3 высоких</span>
          <small>PDF · → КП</small>
        </div>
      </div>
    </div>
  );
}

export function LegalScanLandingSteps() {
  return (
    <div className="steps">
      <article className="step">
        <StepText
          num="1"
          tag="Сайт"
          title="URL, отрасль и признаки сайта — что проверять"
          desc={`Вводите адрес сайта клиента, отрасль и\u00a0размер бизнеса. Отметьте формы, рекламу, оплату и\u00a0иностранные сервисы — сканер сузит проверку.`}
          bullets={[
            "6 отраслей: e-commerce, услуги, медицина, финансы…",
            "3 пресета глубины: 1, 10 или 30\u00a0страниц",
            "Лимит страниц зависит от\u00a0тарифа (Free — 5)",
          ]}
        />
        <div className="step__media">
          <Step1Mock />
        </div>
      </article>

      <article className="step step--reverse">
        <StepText
          num="2"
          tag="Скан"
          title={`Обход сайта и\u00a012 автоматических проверок`}
          desc={`Сканер обходит страницы, ищет документы, формы, трекеры и\u00a0реквизиты. Layer\u00a01 — детерминированный: без LLM, с\u00a0доказательствами и\u00a0ссылками на\u00a0страницы.`}
          bullets={[
            "SSL, политика ПД, cookie, оферта, реквизиты",
            "Трекеры, erid, шифрование форм",
            "Прогресс в\u00a0реальном времени — чеклист по пунктам",
          ]}
        />
        <div className="step__media">
          <Step2Mock />
        </div>
      </article>

      <article className="step">
        <StepText
          num="3"
          tag="Отчёт"
          title={`Риски, статьи закона, штрафы и\u00a0как исправить`}
          desc={`На\u00a0выходе — карточки рисков с\u00a0severity, ссылкой на\u00a0статью, диапазоном штрафа и\u00a0рекомендацией. Суммарная оценка штрафов и\u00a0экспорт PDF.`}
          bullets={[
            "База рисков — управляется суперадмином",
            "LLM дополняет контекст по\u00a0отрасли",
            "Переход в\u00a0Proposal Generator — КП по\u00a0найденным рискам",
          ]}
        />
        <div className="step__media">
          <Step3Mock />
        </div>
      </article>
    </div>
  );
}
