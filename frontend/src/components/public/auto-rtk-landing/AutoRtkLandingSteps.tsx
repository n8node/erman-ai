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
    <div className="mock rtk-step-mock">
      <div className="rtk-step-mock__head">Добавление скважины</div>
      <p className="rtk-step-mock__hint">Месторождение → куст → номер скважины</p>
      <div className="mock-form">
        <div className="mock-field mock-field--wide">
          <div className="mock-field__label">Месторождение</div>
          <div className="mock-field__val">Северное-1 ▾</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Куст</div>
          <div className="mock-field__val">7.2</div>
        </div>
        <div className="mock-field">
          <div className="mock-field__label">Скважина</div>
          <div className="mock-field__val">СК-2847</div>
        </div>
      </div>
      <div className="rtk-step-mock__actions">
        <span className="rtk-btn rtk-btn--ghost rtk-btn--xs">Отмена</span>
        <span className="rtk-btn rtk-btn--primary rtk-btn--xs">Создать</span>
      </div>
    </div>
  );
}

function Step2Mock() {
  return (
    <div className="mock rtk-step-mock">
      <div className="rtk-step-mock__tabs">
        <span className="active">Общее</span>
        <span>Графики</span>
        <span>Таблица</span>
      </div>
      <div className="rtk-step-mock__kpi-row">
        <div>
          <small>Проходка</small>
          <b>1920 м</b>
          <span className="rtk-delta rtk-delta--pos">+70 ↑</span>
        </div>
        <div>
          <small>Нагрузка</small>
          <b>10.03 т</b>
          <span className="rtk-delta rtk-delta--neg">-3.17 ↓</span>
        </div>
      </div>
      <div className="mock-grid">
        <div className="mock-card">
          <b>Факт</b>
          <span>ROP 74 м/ч</span>
          <small>Момент 2.61</small>
        </div>
        <div className="mock-card">
          <b>Уставки</b>
          <span>ROP 68 м/ч</span>
          <small>Момент 3.10</small>
        </div>
        <div className="mock-card">
          <b>Ограничения</b>
          <span>ROP 95 м/ч</span>
          <small>Момент 4.80</small>
        </div>
        <div className="mock-card">
          <b>Соответствие</b>
          <span>72.7% / 100%</span>
          <small>4 параметра</small>
        </div>
      </div>
    </div>
  );
}

function Step3Mock() {
  return (
    <div className="mock rtk-step-mock">
      <div className="rtk-step-mock__head">Согласование плана РТК</div>
      <table className="mock-table rtk-step-table">
        <thead>
          <tr>
            <th>Интервал</th>
            <th>Нагрузка</th>
            <th>ROP</th>
            <th>Статус</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>0–50 м</td>
            <td>12 / 18</td>
            <td>62 / 88</td>
            <td><span className="mock-status mock-status--ok">OK</span></td>
          </tr>
          <tr>
            <td>50–100 м</td>
            <td>12 / 18</td>
            <td>62 / 88</td>
            <td><span className="mock-status mock-status--ok">OK</span></td>
          </tr>
          <tr>
            <td>100–150 м</td>
            <td>11 / 16</td>
            <td>45 / 72</td>
            <td><span className="mock-status mock-status--risk">Изм.</span></td>
          </tr>
        </tbody>
      </table>
      <div className="rtk-step-mock__actions">
        <span className="rtk-btn rtk-btn--ghost rtk-btn--xs">К просмотру</span>
        <span className="rtk-btn rtk-btn--primary rtk-btn--xs">Отправить на согласование</span>
      </div>
    </div>
  );
}

export function AutoRtkLandingSteps() {
  return (
    <div className="steps">
      <article className="step">
        <StepText
          num="1"
          tag="Структура"
          title="Месторождение → куст → скважина за три поля"
          desc="Создайте иерархию активов: месторождение, куст и номер скважины. Редактируйте структуру, настраивайте рассылку для согласования РТК и отслеживайте статус плана по каждой скважине."
          bullets={[
            "Панель управления со сводкой по объектам",
            "Модальное окно добавления скважины",
            "Управление кустами и email-рассылкой",
          ]}
        />
        <div className="step__media">
          <Step1Mock />
        </div>
      </article>

      <article className="step step--reverse">
        <StepText
          num="2"
          tag="Контроль"
          title="Карточка скважины: план и факт в реальном времени"
          desc="Вкладки «Общее», «Графики» и «Таблица» показывают проходку, нагрузку, перепад, момент и скорость. Фильтры по геологической секции и интервалу глубины сужают анализ до нужного участка."
          bullets={[
            "KPI-карточки: факт, уставки, ограничения, % соответствия",
            "Цветовая индикация отклонений от плана",
            "Фильтры по секции и интервалу (метры / даты)",
          ]}
        />
        <div className="step__media">
          <Step2Mock />
        </div>
      </article>

      <article className="step">
        <StepText
          num="3"
          tag="Согласование"
          title="Редактирование плана и workflow согласования"
          desc="Режим редактирования РТК с авто-сохранением по интервалам 50 м. Согласуйте локально или отправьте Excel на email с комментарием — статус меняется на «На согласовании» или «Согласован»."
          bullets={[
            "Таблица уставок и ограничений по глубине",
            "Кнопки: К просмотру · Согласовать · Отправить",
            "Экспорт в Excel и рассылка получателям",
          ]}
        />
        <div className="step__media">
          <Step3Mock />
        </div>
      </article>
    </div>
  );
}
