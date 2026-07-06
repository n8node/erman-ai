"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

type DemoTab = "dashboard" | "card" | "charts" | "table" | "approval";

const TABS: { id: DemoTab; label: string }[] = [
  { id: "dashboard", label: "Панель" },
  { id: "card", label: "Карточка" },
  { id: "charts", label: "Графики" },
  { id: "table", label: "Таблица" },
  { id: "approval", label: "Согласование" },
];

const FIELD = "Северное-1";
const CLUSTER = "Куст 7.2";
const WELL = "СК-2847";

function fmt(n: number, digits = 2) {
  return n.toFixed(digits);
}

function Delta({ value, unit = "" }: { value: number; unit?: string }) {
  const pos = value >= 0;
  return (
    <span className={`rtk-delta ${pos ? "rtk-delta--pos" : "rtk-delta--neg"}`}>
      {pos ? "+" : ""}
      {fmt(value)}
      {unit} {pos ? "↑" : "↓"}
    </span>
  );
}

function MiniSparkline({ seed }: { seed: number }) {
  const points = useMemo(() => {
    const pts: number[] = [];
    let v = 40 + (seed % 20);
    for (let i = 0; i < 12; i++) {
      v += Math.sin(i * 0.8 + seed) * 8 + Math.cos(i * 1.1 + seed * 0.7) * 2;
      pts.push(Math.max(8, Math.min(92, v)));
    }
    return pts;
  }, [seed]);

  const d = points
    .map((y, i) => `${i === 0 ? "M" : "L"}${(i / 11) * 100},${100 - y}`)
    .join(" ");

  return (
    <svg viewBox="0 0 100 100" className="rtk-spark" preserveAspectRatio="none" aria-hidden>
      <path d={d} fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
    </svg>
  );
}

function DepthChart({
  title,
  unit,
  planY,
  factPath,
  limitY,
  tick,
}: {
  title: string;
  unit: string;
  planY: number;
  factPath: string;
  limitY: number;
  tick: number;
}) {
  return (
    <div className="rtk-chart-card">
      <div className="rtk-chart-card__head">
        <span>{title}</span>
        <span className="rtk-chart-card__unit">{unit}</span>
      </div>
      <svg viewBox="0 0 320 140" className="rtk-chart-svg" aria-hidden>
        <defs>
          <linearGradient id={`plan-${title}`} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#185FA5" stopOpacity="0.25" />
            <stop offset="100%" stopColor="#185FA5" stopOpacity="0.04" />
          </linearGradient>
        </defs>
        {[0, 1, 2, 3, 4].map((i) => (
          <line
            key={i}
            x1="36"
            y1={24 + i * 24}
            x2="310"
            y2={24 + i * 24}
            stroke="#E8E7E3"
            strokeWidth="1"
          />
        ))}
        <text x="4" y="28" className="rtk-chart-axis">
          {2400 + tick * 50}
        </text>
        <text x="4" y="76" className="rtk-chart-axis">
          {2200 + tick * 50}
        </text>
        <text x="4" y="124" className="rtk-chart-axis">
          {2000 + tick * 50}
        </text>
        <rect
          x="36"
          y={planY - 18}
          width="274"
          height="36"
          fill={`url(#plan-${title})`}
          rx="2"
        />
        <line
          x1="36"
          y1={limitY}
          x2="310"
          y2={limitY}
          stroke="#A32D2D"
          strokeWidth="1.5"
          strokeDasharray="5 4"
          opacity="0.7"
        />
        <path
          d={factPath}
          fill="none"
          stroke="#3B6D11"
          strokeWidth="2.5"
          strokeLinecap="round"
          className="rtk-chart-line"
        />
        <path
          d="M36,70 L310,70"
          fill="none"
          stroke="#185FA5"
          strokeWidth="1.5"
          opacity="0.5"
        />
      </svg>
    </div>
  );
}

function DashboardView() {
  return (
    <div className="rtk-demo-dashboard">
      <div className="rtk-stat-grid">
        {[
          { label: "Месторождения", value: "3", sub: "Активные площадки" },
          { label: "Кусты", value: "12", sub: "Узлы группировки" },
          { label: "Скважины", value: "47", sub: "В системе" },
          { label: "Добавить", value: "+", sub: "Новая цепочка", action: true },
        ].map((s) => (
          <div key={s.label} className={`rtk-stat-card${s.action ? " rtk-stat-card--action" : ""}`}>
            <div className="rtk-stat-card__label">{s.label}</div>
            <div className="rtk-stat-card__value">{s.value}</div>
            <div className="rtk-stat-card__sub">{s.sub}</div>
          </div>
        ))}
      </div>
      <div className="rtk-recent">
        <div className="rtk-recent__head">Последние скважины</div>
        {[
          { well: WELL, cluster: CLUSTER, status: "Согласован" },
          { well: "СК-2811", cluster: "Куст 4.1", status: "На согласовании" },
          { well: "СК-2790", cluster: "Куст 7.2", status: "Черновик" },
        ].map((row) => (
          <div key={row.well} className="rtk-recent__row">
            <div>
              <b>{row.well}</b>
              <span>{row.cluster}</span>
            </div>
            <span
              className={`rtk-badge rtk-badge--${
                row.status === "Согласован"
                  ? "ok"
                  : row.status === "На согласовании"
                    ? "pending"
                    : "draft"
              }`}
            >
              {row.status}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

function CardView({ live }: { live: LiveMetrics }) {
  return (
    <div className="rtk-card-view">
      <div className="rtk-card-kpi">
        <div className="rtk-card-kpi__main">
          <span>Проходка, м</span>
          <div>
            <small>план 1850</small>
            <b>{Math.round(live.penetration)}</b>
          </div>
        </div>
        <div className="rtk-card-kpi__mini">
          <span>Загрузка системы</span>
          <MiniSparkline seed={live.tick} />
          <b>{fmt(live.utilization, 1)}%</b>
        </div>
        <div className="rtk-card-kpi__mini">
          <span>Колебания</span>
          <MiniSparkline seed={live.tick + 7} />
          <b>{fmt(live.oscillation, 2)}</b>
        </div>
      </div>
      <div className="rtk-param-grid">
        {[
          { name: "Нагрузка, т", fact: live.wob, set: 13.2, limit: 18.5, pct: 72.7 },
          { name: "Перепад, атм", fact: live.dp, set: 22.1, limit: 28.0, pct: 89.5 },
          { name: "Момент, кН·м", fact: live.torque, set: 3.1, limit: 4.8, pct: 84.2 },
          { name: "Скорость, м/ч", fact: live.rop, set: 68.0, limit: 95.0, pct: live.ropPct },
        ].map((p) => (
          <div key={p.name} className="rtk-param-card">
            <div className="rtk-param-card__title">{p.name}</div>
            <div className="rtk-param-card__fact">
              {fmt(p.fact)}
              <Delta value={p.fact - p.set} />
            </div>
            <div className="rtk-param-card__row">
              <span>Уставка</span>
              <b>{fmt(p.set)}</b>
            </div>
            <div className="rtk-param-card__row">
              <span>Ограничение</span>
              <b>{fmt(p.limit)}</b>
            </div>
            <div className="rtk-param-card__pct">
              <div className="rtk-param-card__bar">
                <span style={{ width: `${Math.min(100, p.pct)}%` }} />
              </div>
              <small>{fmt(p.pct, 1)}% / 100%</small>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

type LiveMetrics = {
  tick: number;
  penetration: number;
  utilization: number;
  oscillation: number;
  wob: number;
  dp: number;
  torque: number;
  rop: number;
  ropPct: number;
};

function ChartsView({ tick }: { tick: number }) {
  const offset = tick * 3;
  const paths = [
    `M36,${90 - offset % 20} Q120,${60 + (offset % 15)} 200,${75 - (offset % 10)} T310,${55 + (offset % 12)}`,
    `M36,70 L80,${68 + (offset % 8)} L140,${72 - (offset % 6)} L200,${65 + (offset % 10)} L260,${70 - (offset % 5)} L310,${68 + (offset % 7)}`,
    `M36,${85 - (offset % 12)} L100,${85 - (offset % 12)} L160,${55 + (offset % 8)} L220,${55 + (offset % 8)} L310,${40 + (offset % 6)}`,
    `M36,${75 + Math.sin(offset * 0.2) * 8} L90,${70 + Math.sin(offset * 0.3) * 10} L150,${85} L210,${60 + Math.sin(offset * 0.25) * 12} L310,${72}`,
  ];

  return (
    <div className="rtk-charts-grid">
      <DepthChart title="ROP" unit="м/ч" planY={70} factPath={paths[0]} limitY={42} tick={tick} />
      <DepthChart title="Момент" unit="кН·м" planY={68} factPath={paths[1]} limitY={38} tick={tick + 1} />
      <DepthChart title="Нагрузка" unit="т" planY={72} factPath={paths[2]} limitY={44} tick={tick + 2} />
      <DepthChart title="Перепад" unit="атм" planY={66} factPath={paths[3]} limitY={40} tick={tick + 3} />
    </div>
  );
}

const TABLE_ROWS = [
  { from: 0, to: 50, mode: "Ротор", prog: 50, wob: [12, 10.2, 10.8], dp: [18, 17.1, 16.4], torque: [2.4, 2.1, 2.3], rop: [62, 58, 71] },
  { from: 50, to: 100, mode: "Ротор", prog: 50, wob: [12, 11.5, 12.1], dp: [18, 19.2, 18.8], torque: [2.4, 2.8, 2.6], rop: [62, 64, 78] },
  { from: 100, to: 150, mode: "Слайд", prog: 48, wob: [11, 9.8, 9.2], dp: [16, 15.1, 14.6], torque: [2.2, 1.9, 2.0], rop: [45, 42, 52] },
  { from: 150, to: 200, mode: "Ротор", prog: 50, wob: [13, 12.4, 13.1], dp: [20, 21.3, 20.7], torque: [2.6, 2.9, 3.1], rop: [68, 70, 82] },
  { from: 200, to: 250, mode: "Ротор", prog: 50, wob: [13, 13.8, 14.2], dp: [20, 22.5, 23.1], torque: [2.6, 3.2, 3.4], rop: [68, 72, 69] },
];

function TableView({ jitter }: { jitter: number }) {
  return (
    <div className="rtk-table-wrap">
      <table className="rtk-data-table">
        <thead>
          <tr>
            <th colSpan={2}>Интервал, м</th>
            <th>Режим</th>
            <th>Проходка</th>
            <th colSpan={3}>Нагрузка, т</th>
            <th colSpan={3}>Перепад, атм</th>
            <th colSpan={3}>Момент, кН·м</th>
            <th colSpan={2}>Скорость, м/ч</th>
          </tr>
          <tr className="rtk-data-table__sub">
            <th>от</th>
            <th>до</th>
            <th />
            <th />
            <th>огр.</th>
            <th>уст.</th>
            <th>факт</th>
            <th>огр.</th>
            <th>уст.</th>
            <th>факт</th>
            <th>огр.</th>
            <th>уст.</th>
            <th>факт</th>
            <th>уст.</th>
            <th>факт</th>
          </tr>
        </thead>
        <tbody>
          {TABLE_ROWS.map((row) => {
            const j = (jitter % 5) * 0.04;
            const wobFact = row.wob[2] + j;
            const dpFact = row.dp[2] + j * 0.5;
            const torqueFact = row.torque[2] + j * 0.3;
            const ropFact = row.rop[2] + j * 2;
            return (
              <tr key={row.from}>
                <td>{row.from}</td>
                <td>{row.to}</td>
                <td>{row.mode}</td>
                <td>{row.prog}</td>
                <td>{row.wob[0]}</td>
                <td>{row.wob[1]}</td>
                <td>
                  {fmt(wobFact, 1)}
                  <Delta value={wobFact - row.wob[1]} />
                </td>
                <td>{row.dp[0]}</td>
                <td>{row.dp[1]}</td>
                <td>
                  {fmt(dpFact, 1)}
                  <Delta value={dpFact - row.dp[1]} />
                </td>
                <td>{row.torque[0]}</td>
                <td>{row.torque[1]}</td>
                <td>
                  {fmt(torqueFact, 1)}
                  <Delta value={torqueFact - row.torque[1]} />
                </td>
                <td>{row.rop[1]}</td>
                <td>
                  {fmt(ropFact, 1)}
                  <Delta value={ropFact - row.rop[1]} />
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function ApprovalView() {
  return (
    <div className="rtk-approval">
      <div className="rtk-approval__toolbar">
        <span className="rtk-approval__title">Редактирование плана РТК</span>
        <div className="rtk-approval__actions">
          <button type="button" className="rtk-btn rtk-btn--ghost">
            К просмотру
          </button>
          <button type="button" className="rtk-btn rtk-btn--ghost">
            Согласовать
          </button>
          <button type="button" className="rtk-btn rtk-btn--primary">
            Отправить на согласование
          </button>
        </div>
      </div>
      <div className="rtk-approval__form">
        <label>
          <span>Email получателей</span>
          <div className="rtk-approval__input">tech@example.com; supervisor@example.com</div>
        </label>
        <label>
          <span>Комментарий</span>
          <div className="rtk-approval__textarea">Обновлён план на интервал 200–350 м</div>
        </label>
      </div>
      <div className="rtk-table-wrap rtk-table-wrap--compact">
        <table className="rtk-data-table rtk-data-table--plan">
          <thead>
            <tr>
              <th colSpan={2}>Интервал</th>
              <th colSpan={2}>Нагрузка</th>
              <th colSpan={2}>Перепад</th>
              <th colSpan={2}>Момент</th>
              <th colSpan={2}>Скорость</th>
            </tr>
            <tr className="rtk-data-table__sub">
              <th>от</th>
              <th>до</th>
              <th>уст.</th>
              <th>огр.</th>
              <th>уст.</th>
              <th>огр.</th>
              <th>уст.</th>
              <th>огр.</th>
              <th>уст.</th>
              <th>огр.</th>
            </tr>
          </thead>
          <tbody>
            {[
              [0, 50, 12, 18, 18, 24, 2.4, 3.6, 62, 88],
              [50, 100, 12, 18, 18, 24, 2.4, 3.6, 62, 88],
              [100, 150, 11, 16, 16, 22, 2.2, 3.2, 45, 72],
            ].map(([from, to, ...vals]) => (
              <tr key={from}>
                <td>{from}</td>
                <td>{to}</td>
                {vals.map((v, i) => (
                  <td key={i} className={i % 2 === 0 ? "rtk-cell-editable" : ""}>
                    {v}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export function AutoRtkDemoShowcase() {
  const [tab, setTab] = useState<DemoTab>("card");
  const [tick, setTick] = useState(0);
  const [autoRotate, setAutoRotate] = useState(true);

  const live: LiveMetrics = useMemo(() => {
    const wobble = Math.sin(tick * 0.35) * 0.4;
    return {
      tick,
      penetration: 1920 + tick * 2.3 + wobble * 12,
      utilization: 78.4 + Math.sin(tick * 0.2) * 3,
      oscillation: 0.42 + Math.cos(tick * 0.15) * 0.08,
      wob: 10.03 + wobble,
      dp: 19.78 + wobble * 0.6,
      torque: 2.61 + wobble * 0.3,
      rop: 74.04 + wobble * 4,
      ropPct: 108 + Math.sin(tick * 0.25) * 6,
    };
  }, [tick]);

  useEffect(() => {
    const id = window.setInterval(() => setTick((t) => t + 1), 2200);
    return () => window.clearInterval(id);
  }, []);

  useEffect(() => {
    if (!autoRotate) return;
    const order: DemoTab[] = ["dashboard", "card", "charts", "table", "approval"];
    const id = window.setInterval(() => {
      setTab((current) => {
        const idx = order.indexOf(current);
        return order[(idx + 1) % order.length];
      });
    }, 8000);
    return () => window.clearInterval(id);
  }, [autoRotate]);

  const onTab = useCallback((id: DemoTab) => {
    setAutoRotate(false);
    setTab(id);
  }, []);

  return (
    <div className="rtk-demo" id="demo">
      <div className="rtk-demo__chrome">
        <div className="rtk-demo__sidebar">
          <div className="rtk-demo__sidebar-head">Скважины</div>
          <div className="rtk-demo__search">Поиск…</div>
          <div className="rtk-demo__well rtk-demo__well--active">
            <b>{WELL}</b>
            <span>
              {FIELD} · {CLUSTER}
            </span>
          </div>
          <div className="rtk-demo__well">
            <b>СК-2811</b>
            <span>Северное-1 · Куст 4.1</span>
          </div>
          <div className="rtk-demo__well">
            <b>СК-2790</b>
            <span>Восточное-2 · Куст 3.0</span>
          </div>
        </div>

        <div className="rtk-demo__main">
          <div className="rtk-demo__header">
            <div>
              <div className="rtk-demo__crumb">Главная / Авто РТК</div>
              <h3 className="rtk-demo__title">Авто РТК</h3>
              <div className="rtk-demo__meta">
                <span>{FIELD}</span>
                <span>{CLUSTER}</span>
                <span>{WELL}</span>
                <span className="rtk-demo__updated">Обновлено: сейчас</span>
              </div>
            </div>
            <div className="rtk-demo__header-actions">
              <button type="button" className="rtk-btn rtk-btn--ghost rtk-btn--sm">
                Редактировать
              </button>
              <span className="rtk-badge rtk-badge--ok">Согласован</span>
            </div>
          </div>

          <div className="rtk-demo__filters">
            <span className="rtk-filter">Все секции ▾</span>
            <span className="rtk-filter">Интервал: метры ▾</span>
            <span className="rtk-filter">0 — 2660</span>
            <button type="button" className="rtk-btn rtk-btn--ghost rtk-btn--xs">
              Сбросить
            </button>
          </div>

          <div className="rtk-demo__tabs" role="tablist">
            {TABS.map((t) => (
              <button
                key={t.id}
                type="button"
                role="tab"
                aria-selected={tab === t.id}
                className={`rtk-demo__tab${tab === t.id ? " rtk-demo__tab--active" : ""}`}
                onClick={() => onTab(t.id)}
              >
                {t.label}
              </button>
            ))}
            <button
              type="button"
              className={`rtk-demo__autoplay${autoRotate ? " rtk-demo__autoplay--on" : ""}`}
              onClick={() => setAutoRotate((v) => !v)}
              title="Автопереключение вкладок"
            >
              {autoRotate ? "● LIVE" : "○ PAUSE"}
            </button>
          </div>

          <div className="rtk-demo__body" key={tab}>
            {tab === "dashboard" && <DashboardView />}
            {tab === "card" && <CardView live={live} />}
            {tab === "charts" && <ChartsView tick={tick} />}
            {tab === "table" && <TableView jitter={tick} />}
            {tab === "approval" && <ApprovalView />}
          </div>
        </div>
      </div>
    </div>
  );
}
