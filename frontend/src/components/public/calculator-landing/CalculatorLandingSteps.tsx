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
    { n: 1, label: "Процесс" },
    { n: 2, label: "Финансы" },
    { n: 3, label: "Результат" },
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
        <span className="chip active">Digital-маркетинг</span>
        <span className="chip">E-commerce</span>
        <span className="chip">B2B-продажи</span>
        <span className="chip">HR</span>
        <span className="chip">Финансы</span>
        <span className="chip">Поддержка</span>
      </div>
      <div className="mock-grid">
        <div className="mock-card">
          <b>Обработка лидов с{"\u00a0"}рекламы</b>
          <span>Квалификация и{"\u00a0"}передача в{"\u00a0"}CRM</span>
          <small>4 шага · 1200 ед/мес · 75%</small>
        </div>
        <div className="mock-card">
          <b>Еженедельный отчёт</b>
          <span>Сбор данных и{"\u00a0"}подготовка отчёта</span>
          <small>4 шага · 20 ед/мес · 70%</small>
        </div>
        <div className="mock-card">
          <b>Бриф на{"\u00a0"}кампанию</b>
          <span>Сбор ТЗ и{"\u00a0"}постановка задач</span>
          <small>4 шага · 12 ед/мес · 55%</small>
        </div>
        <div className="mock-card">
          <b>Контент-план</b>
          <span>Идеи, тексты, публикация</span>
          <small>4 шага · 40 ед/мес · 60%</small>
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
        <span className="chip">Простой</span>
        <span className="chip active">Стандарт</span>
        <span className="chip">Сложный</span>
      </div>
      <div className="mock-form">
        <div className="mock-field">
          <div className="mock-field__label">Hm — экономия часов/мес</div>
          <div className="mock-field__val">1066,67</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Ch — стоимость часа, ₽</div>
          <div className="mock-field__val">2500</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">I₀ — бюджет внедрения, ₽</div>
          <div className="mock-field__val">600 000</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Om — поддержка, ₽/мес</div>
          <div className="mock-field__val">18 000</div>
        </div>
      </div>
      <div className="mock-preview">
        <div className="mock-preview__row">
          <span>Предпросмотр · выгода/мес</span>
          <b>2 663 667{"\u00a0"}₽</b>
        </div>
        <div className="mock-preview__row">
          <span>Окупаемость</span>
          <b>0.2{"\u00a0"}мес</b>
        </div>
        <div className="mock-preview__row">
          <span>ROI за{"\u00a0"}12{"\u00a0"}мес</span>
          <b>5227.3%</b>
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
            <th>KPI</th>
            <th>До</th>
            <th>После</th>
            <th>Δ</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Время / ед.</td>
            <td>10 мин</td>
            <td>2 мин</td>
            <td>
              <span className="delta-pos">−80%</span>
            </td>
          </tr>
          <tr>
            <td>Объём / сотр.</td>
            <td>20/день</td>
            <td>100/день</td>
            <td>
              <span className="delta-pos">5.0×</span>
            </td>
          </tr>
          <tr>
            <td>% ошибок</td>
            <td>~5%</td>
            <td>0%</td>
            <td>
              <span className="delta-pos">устранены</span>
            </td>
          </tr>
          <tr>
            <td>Часы / мес</td>
            <td>1333 ч</td>
            <td>267 ч</td>
            <td>
              <span className="delta-pos">−1067 ч</span>
            </td>
          </tr>
          <tr>
            <td>Экономия{"\u00a0"}ФОТ</td>
            <td>—</td>
            <td>2{"\u00a0"}666{"\u00a0"}667{"\u00a0"}₽</td>
            <td>
              <span className="delta-pos">+2.6М</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}

export function CalculatorLandingSteps() {
  return (
    <div className="steps">
      <article className="step">
        <StepText
          num="1"
          tag="Процесс"
          title={`Разберите процесс на\u00a0операции — или возьмите готовый шаблон`}
          desc={`Выбираете отрасль из\u00a010\u00a0готовых пакетов, либо описываете свой процесс вручную: операция, исполнитель, минуты на\u00a0единицу, ставка ₽/мин. Калькулятор сам считает Hm — экономию часов в\u00a0месяц.`}
          bullets={[
            "10 отраслей с\u00a0готовыми шаблонами — от\u00a0Digital-маркетинга до\u00a0логистики",
            "Любое число шагов внутри процесса, со\u00a0своими ставками",
            "Параметры объёма: единиц/мес, % автоматизации, % ошибок",
          ]}
        />
        <div className="step__media">
          <Step1Mock />
        </div>
      </article>

      <article className="step step--reverse">
        <StepText
          num="2"
          tag="Финансы"
          title="Стоимость часа, бюджет внедрения, горизонт — модель сама подставит оценку"
          desc={`Выбираете уровень внедрения — Простой, Стандарт или Сложный, — и\u00a0калькулятор подставляет рекомендуемый бюджет и\u00a0стоимость поддержки. Можно править вручную. Превью результата обновляется на\u00a0лету.`}
          bullets={[
            "Ch — полная стоимость часа (ЗП\u00a0+ налоги\u00a0+ бонусы\u00a0+ накладные)",
            "I₀ — бюджет внедрения: аудит, разработка, обучение\u00a0— разово",
            "Om — поддержка решения: VPS, домен, лицензии — помесячно",
            "Горизонт оценки — 6, 12 или 24\u00a0месяца",
          ]}
        />
        <div className="step__media">
          <Step2Mock />
        </div>
      </article>

      <article className="step">
        <StepText
          num="3"
          tag="Результат"
          title={`Готовый финансовый отчёт — с\u00a0KPI «до/после» и\u00a0возможностью отправить КП`}
          desc={`На\u00a0выходе — 4 ключевых цифры, таблица KPI «до/после», TCO и\u00a0NPV по\u00a0запросу. Можно сгенерировать КП, опубликовать ссылку для\u00a0коллег или экспортировать\u00a0PDF.`}
          bullets={[
            "Сравнение по\u00a06\u00a0KPI: время на\u00a0единицу, объём, ошибки, ФОТ",
            "Публичная ссылка — поделиться расчётом с\u00a0CFO или командой",
            "Экспорт PDF и\u00a0генерация КП в\u00a0один клик",
          ]}
        />
        <div className="step__media">
          <Step3Mock />
        </div>
      </article>
    </div>
  );
}
