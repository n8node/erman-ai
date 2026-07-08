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

function Step1Mock() {
  return (
    <div className="mock aro-step-mock">
      <div className="aro-step-mock__head">Источник телеметрии</div>
      <p className="aro-step-mock__hint">10 каналов · дискретизация 10 сек</p>
      <div className="mock-form">
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">Протокол</div>
          <div className="mock-field__val">WITSML / OPC ▾</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Скважина</div>
          <div className="mock-field__val">СК-2847</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Каналов</div>
          <div className="mock-field__val">10 / 10</div>
        </div>
      </div>
      <div className="aro-step-mock__channels">
        {["Тальблок", "WOB", "Расход", "Давление", "RPM"].map((ch) => (
          <span key={ch} className="aro-step-mock__chip">
            {ch} ✓
          </span>
        ))}
      </div>
    </div>
  );
}

function Step2Mock() {
  return (
    <div className="mock aro-step-mock">
      <div className="aro-step-mock__live">
        <span className="aro-step-mock__live-dot" />
        LIVE · 1.1 Бурение с вращением
      </div>
      <div className="aro-step-mock__fusion-row">
        <div>
          <small>Алгоритм</small>
          <b>91%</b>
        </div>
        <div>
          <small>ML</small>
          <b>98%</b>
        </div>
        <div className="aro-step-mock__fusion-main">
          <small>Fusion</small>
          <b>97.5%</b>
        </div>
      </div>
      <div className="aro-step-mock__timeline-mini">
        {[
          { code: "1.1", w: 40, color: "#3B6D11" },
          { code: "3.1", w: 12, color: "#185FA5" },
          { code: "4.3", w: 20, color: "#BA7517" },
          { code: "5.2", w: 10, color: "#534AB7" },
          { code: "1.2", w: 18, color: "#3B6D11" },
        ].map((s) => (
          <span key={s.code} style={{ flex: s.w, backgroundColor: s.color }} title={s.code} />
        ))}
      </div>
    </div>
  );
}

function Step3Mock() {
  return (
    <div className="mock aro-step-mock">
      <div className="aro-step-mock__head">Отчёт по рейсу</div>
      <table className="mock-table aro-step-table">
        <thead>
          <tr>
            <th>Категория</th>
            <th>Доля</th>
            <th>Время</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>1. Бурение</td>
            <td>62.4%</td>
            <td>11ч 20м</td>
          </tr>
          <tr>
            <td>4. СПО</td>
            <td>18.1%</td>
            <td>3ч 17м</td>
          </tr>
          <tr>
            <td>6. Другие</td>
            <td>1.2%</td>
            <td>13м</td>
          </tr>
        </tbody>
      </table>
      <div className="aro-step-mock__actions">
        <span className="aro-btn aro-btn--ghost aro-btn--xs">Excel</span>
        <span className="aro-btn aro-btn--primary aro-btn--xs">API</span>
      </div>
    </div>
  );
}

export function AutoRoLandingSteps() {
  return (
    <div className="steps">
      <article className="step">
        <StepText
          num="1"
          tag="Данные"
          title="Подключение 10 каналов телеметрии"
          desc="Приём потока с буровой: позиция тальблока, WOB, расход, давление, RPM, момент, вес на крюке и другие. Валидация каналов и маппинг на внутреннюю схему модуля."
          bullets={[
            "WITSML, OPC или файловый импорт",
            "QC-гейт: 6.1 отсутствие / 6.2 плохое качество",
            "Дискретизация 1–10 сек",
          ]}
        />
        <div className="step__media">
          <Step1Mock />
        </div>
      </article>

      <article className="step step--reverse">
        <StepText
          num="2"
          tag="Классификация"
          title="Fusion: алгоритмы + ML в реальном времени"
          desc="Rule Engine закрывает однозначные режимы. ML-модель уточняет пограничные случаи. Fusion-слой выдаёт итоговую операцию из 18 типов с уверенностью до 97.5%."
          bullets={[
            "18 операций в 6 категориях",
            "Прозрачный разбор: Algo · ML · Fusion",
            "Обновление таймлайна каждые 10 сек",
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
          title="Аналитика рейса и экспорт"
          desc="Распределение времени по категориям, KPI-слой, доля «не определено». Экспорт в Excel и API для интеграции с модулем Авто РТК и внешними BI-системами."
          bullets={[
            "Сводка: бурение / СПО / промывка / наращивание",
            "Целевой KPI: «не определено» < 3%",
            "Экспорт: Excel · API · суточный рапорт",
          ]}
        />
        <div className="step__media">
          <Step3Mock />
        </div>
      </article>
    </div>
  );
}
