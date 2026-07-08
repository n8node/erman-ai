"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

type DemoTab =
  | "timeline"
  | "fusion"
  | "tracks"
  | "taxonomy"
  | "qc"
  | "analytics"
  | "compare";

type CompareMode = "pronova" | "schlumberger" | "koa-rf" | "auto-ro";

const TABS: { id: DemoTab; label: string }[] = [
  { id: "timeline", label: "Таймлайн" },
  { id: "fusion", label: "Fusion" },
  { id: "tracks", label: "Треки" },
  { id: "taxonomy", label: "Таксономия" },
  { id: "qc", label: "QC" },
  { id: "analytics", label: "Аналитика" },
  { id: "compare", label: "Сравнение" },
];

const FIELD = "Северное-1";
const CLUSTER = "Куст 7.2";
const WELL = "СК-2847";

type OpGroup = "drill" | "ream" | "circ" | "trip" | "conn" | "other" | "bad";

type TimelineSeg = {
  from: number;
  to: number;
  code: string;
  label: string;
  group: OpGroup;
  confidence: number;
};

const TIMELINE: TimelineSeg[] = [
  { from: 1820, to: 1845, code: "1.1", label: "Бурение с вращением", group: "drill", confidence: 97.2 },
  { from: 1845, to: 1852, code: "3.1", label: "Статическая промывка", group: "circ", confidence: 94.8 },
  { from: 1852, to: 1868, code: "4.3", label: "Спуск", group: "trip", confidence: 96.1 },
  { from: 1868, to: 1874, code: "5.2", label: "Удержание в клиньях при СПО", group: "conn", confidence: 98.4 },
  { from: 1874, to: 1895, code: "1.2", label: "Направленное бурение", group: "drill", confidence: 95.7 },
  { from: 1895, to: 1900, code: "6.2", label: "Неудовл. качество данных", group: "bad", confidence: 0 },
  { from: 1900, to: 1918, code: "2.1", label: "Прямая проработка", group: "ream", confidence: 93.5 },
  { from: 1918, to: 1935, code: "4.6", label: "Подъём с промывкой", group: "trip", confidence: 97.8 },
];

const GROUP_COLORS: Record<OpGroup, string> = {
  drill: "#3B6D11",
  ream: "#0F6E56",
  circ: "#185FA5",
  trip: "#BA7517",
  conn: "#534AB7",
  other: "#5A5D62",
  bad: "#A32D2D",
};

const GROUP_BG: Record<OpGroup, string> = {
  drill: "#EAF3DE",
  ream: "#E1F5EE",
  circ: "#E6F1FB",
  trip: "#FAEEDA",
  conn: "#EEEDFE",
  other: "#F3F2EF",
  bad: "#FCEBEB",
};

const TAXONOMY = [
  {
    cat: "1. Бурение",
    pct: 62.4,
    items: [
      { code: "1.1", name: "Бурение с вращением", pct: 48.1 },
      { code: "1.2", name: "Направленное бурение", pct: 9.2 },
      { code: "1.3", name: "Направленное + осцилляция", pct: 5.1 },
    ],
  },
  {
    cat: "2. Проработка",
    pct: 6.8,
    items: [
      { code: "2.1", name: "Прямая", pct: 4.1 },
      { code: "2.2", name: "Обратная", pct: 2.7 },
    ],
  },
  {
    cat: "3. Промывка",
    pct: 8.3,
    items: [
      { code: "3.1", name: "Статическая", pct: 5.4 },
      { code: "3.2", name: "Статическая с вращением", pct: 2.9 },
    ],
  },
  {
    cat: "4. СПО",
    pct: 18.1,
    items: [
      { code: "4.1", name: "Подъём", pct: 4.2 },
      { code: "4.2", name: "Подъём с вращением", pct: 2.8 },
      { code: "4.3", name: "Спуск", pct: 5.1 },
      { code: "4.4", name: "Спуск с вращением", pct: 1.9 },
      { code: "4.5", name: "Спуск с промывкой", pct: 2.1 },
      { code: "4.6", name: "Подъём с промывкой", pct: 2.0 },
    ],
  },
  {
    cat: "5. Наращивание",
    pct: 3.2,
    items: [
      { code: "5.1", name: "Удержание при бурении", pct: 1.8 },
      { code: "5.2", name: "Удержание при СПО", pct: 1.4 },
    ],
  },
  {
    cat: "6. Другие",
    pct: 1.2,
    items: [
      { code: "6.1", name: "Данные отсутствуют", pct: 0.3 },
      { code: "6.2", name: "Неудовл. качество", pct: 0.4 },
      { code: "6.3", name: "Не определено", pct: 0.5 },
    ],
  },
];

const FUSION_CASES = [
  {
    algo: { code: "1.1", label: "Бурение с вращением", pct: 91 },
    ml: { code: "1.1", label: "Бурение с вращением", pct: 98 },
    fusion: { code: "1.1", label: "Бурение с вращением", pct: 97.5 },
    note: "Однозначный режим — слои согласны",
  },
  {
    algo: { code: "4.1", label: "Подъём", pct: 72 },
    ml: { code: "4.2", label: "Подъём с вращением", pct: 89 },
    fusion: { code: "4.2", label: "Подъём с вращением", pct: 86.4 },
    note: "Конфликт — Fusion учитывает RPM > порога",
  },
  {
    algo: { code: "6.2", label: "Неудовл. качество", pct: 100 },
    ml: { code: "6.2", label: "Неудовл. качество", pct: 100 },
    fusion: { code: "6.2", label: "Неудовл. качество", pct: 0 },
    note: "QC-блок — классификация приостановлена",
  },
];

const COMPARE_MODES: {
  id: CompareMode;
  label: string;
  ops: number;
  method: string;
  note: string;
}[] = [
  { id: "pronova", label: "ProNova", ops: 11, method: "Rules", note: "Много «не определено»" },
  {
    id: "schlumberger",
    label: "Schlumberger '06",
    ops: 13,
    method: "Rules",
    note: "Компактно, без QC",
  },
  { id: "koa-rf", label: "KOA-RF", ops: 33, method: "KOA + RF", note: "Мелкая мозаика" },
  { id: "auto-ro", label: "Авто РО", ops: 18, method: "Algo + ML", note: "Читаемо и детально" },
];

function fmt(n: number, digits = 1) {
  return n.toFixed(digits);
}

function depthMin() {
  return TIMELINE[0].from;
}

function depthMax() {
  return TIMELINE[TIMELINE.length - 1].to;
}

function segWidth(from: number, to: number) {
  const span = depthMax() - depthMin();
  return ((to - from) / span) * 100;
}

function segLeft(from: number) {
  const span = depthMax() - depthMin();
  return ((from - depthMin()) / span) * 100;
}

function TimelineView({ activeIdx }: { activeIdx: number }) {
  return (
    <div className="aro-timeline">
      <div className="aro-timeline__ruler">
        {TIMELINE.map((s, i) => (
          <span key={s.from} style={{ left: `${segLeft(s.from)}%` }}>
            {s.from}м
          </span>
        ))}
        <span style={{ right: 0 }}>{depthMax()}м</span>
      </div>
      <div className="aro-timeline__bar">
        {TIMELINE.map((s, i) => (
          <div
            key={`${s.from}-${s.code}`}
            className={`aro-timeline__seg${i === activeIdx ? " aro-timeline__seg--active" : ""}`}
            style={{
              left: `${segLeft(s.from)}%`,
              width: `${segWidth(s.from, s.to)}%`,
              backgroundColor: GROUP_BG[s.group],
              borderColor: GROUP_COLORS[s.group],
            }}
            title={`${s.code} ${s.label}`}
          >
            <span style={{ color: GROUP_COLORS[s.group] }}>{s.code}</span>
          </div>
        ))}
        <div
          className="aro-timeline__playhead"
          style={{ left: `${segLeft(TIMELINE[activeIdx].from) + segWidth(TIMELINE[activeIdx].from, TIMELINE[activeIdx].to) / 2}%` }}
        />
      </div>
      <div className="aro-timeline__detail">
        <div>
          <span className="aro-timeline__code" style={{ color: GROUP_COLORS[TIMELINE[activeIdx].group] }}>
            {TIMELINE[activeIdx].code}
          </span>
          <b>{TIMELINE[activeIdx].label}</b>
        </div>
        <div className="aro-timeline__meta">
          <span>
            {TIMELINE[activeIdx].from}–{TIMELINE[activeIdx].to} м
          </span>
          {TIMELINE[activeIdx].confidence > 0 && (
            <span className="aro-badge aro-badge--ok">{fmt(TIMELINE[activeIdx].confidence)}% уверенность</span>
          )}
          {TIMELINE[activeIdx].group === "bad" && (
            <span className="aro-badge aro-badge--bad">QC-блок</span>
          )}
        </div>
      </div>
    </div>
  );
}

function FusionView({ caseIdx, tick }: { caseIdx: number; tick: number }) {
  const c = FUSION_CASES[caseIdx];
  const wobble = Math.sin(tick * 0.3) * 0.5;

  return (
    <div className="aro-fusion">
      <div className="aro-fusion__channels">
        {[
          { name: "Вес на крюке", val: 142 + wobble, unit: "т" },
          { name: "RPM", val: caseIdx === 1 ? 42 + wobble : caseIdx === 2 ? 0 : 87 + wobble, unit: "об/мин" },
          { name: "Расход", val: 28 + wobble * 0.3, unit: "л/с" },
          { name: "WOB", val: caseIdx === 2 ? 0 : 10.2 + wobble * 0.2, unit: "т" },
          { name: "Момент", val: caseIdx === 2 ? 0 : 2.6 + wobble * 0.1, unit: "кН·м" },
        ].map((ch) => (
          <div key={ch.name} className="aro-fusion__ch">
            <span>{ch.name}</span>
            <b>
              {fmt(ch.val)}
              <small>{ch.unit}</small>
            </b>
          </div>
        ))}
      </div>
      <div className="aro-fusion__cols">
        {[
          { title: "Алгоритм", data: c.algo, color: "#185FA5" },
          { title: "ML-модель", data: c.ml, color: "#534AB7" },
          { title: "Fusion", data: c.fusion, color: "#3B6D11", highlight: true },
        ].map((col) => (
          <div
            key={col.title}
            className={`aro-fusion__col${col.highlight ? " aro-fusion__col--highlight" : ""}`}
          >
            <div className="aro-fusion__col-head" style={{ color: col.color }}>
              {col.title}
            </div>
            <div className="aro-fusion__col-code">{col.data.code}</div>
            <div className="aro-fusion__col-label">{col.data.label}</div>
            <div className="aro-fusion__col-pct">
              {col.data.pct > 0 ? `${fmt(col.data.pct)}%` : "—"}
            </div>
            <div className="aro-fusion__col-bar">
              <span style={{ width: `${col.data.pct}%`, backgroundColor: col.color }} />
            </div>
          </div>
        ))}
      </div>
      <p className="aro-fusion__note">{c.note}</p>
    </div>
  );
}

function TracksView({ activeIdx, tick }: { activeIdx: number; tick: number }) {
  const offset = tick * 2;
  const path = `M36,${70 - (offset % 15)} L100,${65 + (offset % 10)} L180,${55 + (offset % 12)} L260,${60 - (offset % 8)} L310,${50 + (offset % 6)}`;

  return (
    <div className="aro-tracks">
      {[
        { title: "Позиция тальблока", unit: "м", path },
        { title: "RPM", unit: "об/мин", path: path.replace(/70/g, "80") },
        { title: "Расход", unit: "л/с", path: path.replace(/65/g, "75") },
        { title: "WOB", unit: "т", path: path.replace(/55/g, "68") },
      ].map((tr) => (
        <div key={tr.title} className="aro-track-card">
          <div className="aro-track-card__head">
            <span>{tr.title}</span>
            <span>{tr.unit}</span>
          </div>
          <div className="aro-track-card__body">
            <div className="aro-track-card__bands">
              {TIMELINE.map((s) => (
                <div
                  key={`${tr.title}-${s.from}`}
                  style={{
                    left: `${segLeft(s.from)}%`,
                    width: `${segWidth(s.from, s.to)}%`,
                    backgroundColor: GROUP_BG[s.group],
                    opacity: s.from === TIMELINE[activeIdx].from ? 1 : 0.55,
                  }}
                />
              ))}
            </div>
            <svg viewBox="0 0 320 80" className="aro-track-card__svg" aria-hidden>
              <path d={tr.path} fill="none" stroke="#0B0D0E" strokeWidth="2" strokeLinecap="round" />
            </svg>
          </div>
        </div>
      ))}
    </div>
  );
}

function TaxonomyView() {
  const [open, setOpen] = useState<string | null>("1. Бурение");

  return (
    <div className="aro-taxonomy">
      {TAXONOMY.map((cat) => (
        <div key={cat.cat} className="aro-taxonomy__cat">
          <button
            type="button"
            className="aro-taxonomy__cat-head"
            onClick={() => setOpen(open === cat.cat ? null : cat.cat)}
          >
            <span>{open === cat.cat ? "▼" : "▶"} {cat.cat}</span>
            <span className="aro-taxonomy__cat-pct">{fmt(cat.pct)}%</span>
          </button>
          {open === cat.cat && (
            <div className="aro-taxonomy__items">
              {cat.items.map((item) => (
                <div key={item.code} className="aro-taxonomy__item">
                  <span className="aro-taxonomy__item-code">{item.code}</span>
                  <span>{item.name}</span>
                  <span className="aro-taxonomy__item-pct">{fmt(item.pct)}%</span>
                </div>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}

function QcView() {
  return (
    <div className="aro-qc">
      <div className="aro-qc__map">
        {TIMELINE.map((s) => (
          <div
            key={`qc-${s.from}`}
            className={`aro-qc__zone aro-qc__zone--${s.group}`}
            style={{
              left: `${segLeft(s.from)}%`,
              width: `${segWidth(s.from, s.to)}%`,
            }}
            title={s.label}
          />
        ))}
      </div>
      <div className="aro-qc__legend">
        <span className="aro-qc__legend-item aro-qc__legend-item--ok">Норма</span>
        <span className="aro-qc__legend-item aro-qc__legend-item--bad">6.2 Плохие данные</span>
        <span className="aro-qc__legend-item aro-qc__legend-item--miss">6.1 Нет данных</span>
      </div>
      <table className="aro-qc__table">
        <thead>
          <tr>
            <th>Интервал</th>
            <th>Статус</th>
            <th>Причина</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>1820–1895 м</td>
            <td>
              <span className="aro-badge aro-badge--ok">OK</span>
            </td>
            <td>Все 10 каналов в норме</td>
          </tr>
          <tr>
            <td>1895–1900 м</td>
            <td>
              <span className="aro-badge aro-badge--bad">6.2</span>
            </td>
            <td>Обрыв WOB, пропуск &gt;30 сек</td>
          </tr>
          <tr>
            <td>1900–1935 м</td>
            <td>
              <span className="aro-badge aro-badge--ok">OK</span>
            </td>
            <td>Восстановление каналов</td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}

function AnalyticsView() {
  return (
    <div className="aro-analytics">
      <div className="aro-analytics__grid">
        {[
          { label: "Бурение", pct: 62.4, time: "11ч 20м", color: GROUP_COLORS.drill },
          { label: "СПО", pct: 18.1, time: "3ч 17м", color: GROUP_COLORS.trip },
          { label: "Промывка", pct: 8.3, time: "1ч 31м", color: GROUP_COLORS.circ },
          { label: "Проработка", pct: 6.8, time: "1ч 14м", color: GROUP_COLORS.ream },
          { label: "Наращивание", pct: 3.2, time: "35м", color: GROUP_COLORS.conn },
          { label: "Другие", pct: 1.2, time: "13м", color: GROUP_COLORS.other },
        ].map((row) => (
          <div key={row.label} className="aro-analytics__card">
            <div className="aro-analytics__card-head">
              <span style={{ color: row.color }}>{row.label}</span>
              <b>{fmt(row.pct)}%</b>
            </div>
            <div className="aro-analytics__bar">
              <span style={{ width: `${row.pct}%`, backgroundColor: row.color }} />
            </div>
            <small>{row.time}</small>
          </div>
        ))}
      </div>
      <div className="aro-analytics__kpi">
        <div>
          <span>Точность классификации</span>
          <b>97.5%</b>
        </div>
        <div>
          <span>Доля «не определено»</span>
          <b>0.5%</b>
        </div>
        <div>
          <span>Дискретизация</span>
          <b>10 сек</b>
        </div>
      </div>
    </div>
  );
}

function CompareView({ mode, setMode }: { mode: CompareMode; setMode: (m: CompareMode) => void }) {
  const blocks = useMemo(() => {
    const count =
      mode === "pronova" ? 11 : mode === "schlumberger" ? 13 : mode === "koa-rf" ? 33 : 18;
    return Array.from({ length: Math.min(count, 24) }, (_, i) => ({
      w: 3 + (i % 5) * 2 + (mode === "koa-rf" ? 1 : 4),
      g: (["drill", "circ", "trip", "conn", "ream", "other"] as OpGroup[])[i % 6],
    }));
  }, [mode]);

  const selected = COMPARE_MODES.find((m) => m.id === mode)!;

  return (
    <div className="aro-compare">
      <div className="aro-compare__modes">
        {COMPARE_MODES.map((m) => (
          <button
            key={m.id}
            type="button"
            className={`aro-compare__mode${mode === m.id ? " aro-compare__mode--active" : ""}`}
            onClick={() => setMode(m.id)}
          >
            <b>{m.label}</b>
            <span>
              {m.ops} ops · {m.method}
            </span>
          </button>
        ))}
      </div>
      <div className="aro-compare__preview">
        <div className="aro-compare__preview-head">
          <span>
            {selected.label} — {selected.ops} операций
          </span>
          <span>{selected.note}</span>
        </div>
        <div className={`aro-compare__mosaic aro-compare__mosaic--${mode}`}>
          {blocks.map((b, i) => (
            <div
              key={i}
              style={{
                flex: `${b.w} 0 0`,
                backgroundColor: GROUP_BG[b.g],
                borderColor: GROUP_COLORS[b.g],
              }}
            />
          ))}
        </div>
        {mode === "auto-ro" && (
          <p className="aro-compare__accuracy">
            Точность: <b>97.5%</b> · 6 категорий · 18 операций · QC-слой
          </p>
        )}
      </div>
    </div>
  );
}

export function AutoRoDemoShowcase() {
  const [tab, setTab] = useState<DemoTab>("timeline");
  const [tick, setTick] = useState(0);
  const [autoRotate, setAutoRotate] = useState(true);
  const [compareMode, setCompareMode] = useState<CompareMode>("auto-ro");

  const activeIdx = tick % TIMELINE.length;
  const fusionIdx = tick % FUSION_CASES.length;

  useEffect(() => {
    const id = window.setInterval(() => setTick((t) => t + 1), 2200);
    return () => window.clearInterval(id);
  }, []);

  useEffect(() => {
    if (!autoRotate) return;
    const order: DemoTab[] = ["timeline", "fusion", "tracks", "taxonomy", "qc", "analytics", "compare"];
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
    <div className="aro-demo" id="demo">
      <div className="aro-demo__chrome">
        <div className="aro-demo__sidebar">
          <div className="aro-demo__sidebar-head">Скважины</div>
          <div className="aro-demo__search">Поиск…</div>
          <div className="aro-demo__well aro-demo__well--active">
            <b>{WELL}</b>
            <span>
              {FIELD} · {CLUSTER}
            </span>
          </div>
          <div className="aro-demo__well">
            <b>СК-2811</b>
            <span>Северное-1 · Куст 4.1</span>
          </div>
          <div className="aro-demo__well">
            <b>СК-2790</b>
            <span>Восточное-2 · Куст 3.0</span>
          </div>
        </div>

        <div className="aro-demo__main">
          <div className="aro-demo__header">
            <div>
              <div className="aro-demo__crumb">Главная / Авто РО</div>
              <h3 className="aro-demo__title">Авто РО</h3>
              <div className="aro-demo__meta">
                <span>{FIELD}</span>
                <span>{CLUSTER}</span>
                <span>{WELL}</span>
                <span className="aro-demo__updated">Обновлено: сейчас</span>
              </div>
            </div>
            <div className="aro-demo__header-actions">
              <span className="aro-badge aro-badge--ok">97.5% точность</span>
              <span className="aro-badge aro-badge--module">Модуль АБУ</span>
            </div>
          </div>

          <div className="aro-demo__filters">
            <span className="aro-filter">Интервал: 1820–1935 м</span>
            <span className="aro-filter">Дискретизация: 10 сек ▾</span>
            <span className="aro-filter">Качество: OK ▾</span>
          </div>

          <div className="aro-demo__tabs" role="tablist">
            {TABS.map((t) => (
              <button
                key={t.id}
                type="button"
                role="tab"
                aria-selected={tab === t.id}
                className={`aro-demo__tab${tab === t.id ? " aro-demo__tab--active" : ""}`}
                onClick={() => onTab(t.id)}
              >
                {t.label}
              </button>
            ))}
            <button
              type="button"
              className={`aro-demo__autoplay${autoRotate ? " aro-demo__autoplay--on" : ""}`}
              onClick={() => setAutoRotate((v) => !v)}
              title="Автопереключение вкладок"
            >
              {autoRotate ? "● LIVE" : "○ PAUSE"}
            </button>
          </div>

          <div className="aro-demo__body" key={tab}>
            {tab === "timeline" && <TimelineView activeIdx={activeIdx} />}
            {tab === "fusion" && <FusionView caseIdx={fusionIdx} tick={tick} />}
            {tab === "tracks" && <TracksView activeIdx={activeIdx} tick={tick} />}
            {tab === "taxonomy" && <TaxonomyView />}
            {tab === "qc" && <QcView />}
            {tab === "analytics" && <AnalyticsView />}
            {tab === "compare" && <CompareView mode={compareMode} setMode={setCompareMode} />}
          </div>
        </div>
      </div>
    </div>
  );
}
