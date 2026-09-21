"use client";

import { Download, LoaderCircle, Save, Sparkles, X } from "lucide-react";
import { useEffect, useState } from "react";
import type { GeologicalJournalDocumentDetail, GeologicalJournalRow } from "@/lib/api-geological-journal";
import {
  exportGeologicalJournalCSV,
  exportGeologicalJournalJSON,
  exportGeologicalJournalXLSX,
  exportGeologicalJournalXML,
} from "@/lib/geological-journal-export";

const fields: Array<{ key: keyof GeologicalJournalRow; label: string; numeric?: boolean }> = [
  { key: "date", label: "Дата" },
  { key: "drilling_diameter_mm", label: "Диаметр, мм", numeric: true },
  { key: "depth_from_m", label: "Глубина от, м", numeric: true },
  { key: "depth_to_m", label: "Глубина до, м", numeric: true },
  { key: "drilling_run_m", label: "Проходка, м", numeric: true },
  { key: "core_recovery_m", label: "Выход керна, м", numeric: true },
  { key: "core_recovery_pct", label: "Выход керна, %", numeric: true },
  { key: "rock_description", label: "Описание породы" },
  { key: "sampling_interval", label: "Интервал опробования" },
  { key: "sample_number", label: "Номер пробы" },
  { key: "notes", label: "Примечания" },
];

function emptyRow(): GeologicalJournalRow {
  return {
    date: "",
    drilling_diameter_mm: null,
    depth_from_m: null,
    depth_to_m: null,
    drilling_run_m: null,
    core_recovery_m: null,
    core_recovery_pct: null,
    rock_description: "",
    sampling_interval: "",
    sample_number: "",
    notes: "",
    uncertainties: [],
  };
}

function displayValue(value: string | number | null | undefined) {
  return value ?? "";
}

type Props = {
  documentId: string;
  page: GeologicalJournalDocumentDetail["pages"][number] | null;
  imageUrl: string;
  busy?: boolean;
  onClose: () => void;
  onSave: (rows: GeologicalJournalRow[], ocrText: string) => Promise<void>;
  onReprocess: () => Promise<void>;
};

export function GeologicalJournalDocumentPageModal({
  documentId: _documentId,
  page,
  imageUrl,
  busy = false,
  onClose,
  onSave,
  onReprocess,
}: Props) {
  const [rows, setRows] = useState<GeologicalJournalRow[]>([]);
  const [ocrText, setOcrText] = useState("");
  const [saving, setSaving] = useState(false);
  const [processing, setProcessing] = useState(false);
  const [activeTab, setActiveTab] = useState<"table" | "ocr">("table");

  useEffect(() => {
    setRows(page?.table_result?.rows ?? []);
    setOcrText(page?.ocr_text ?? "");
    setActiveTab("table");
  }, [page]);

  if (!page) return null;

  function updateCell(index: number, key: keyof GeologicalJournalRow, value: string) {
    setRows((current) => current.map((row, rowIndex) => {
      if (rowIndex !== index) return row;
      const field = fields.find((item) => item.key === key);
      return { ...row, [key]: field?.numeric && value !== "" ? Number(value) : field?.numeric ? null : value };
    }));
  }

  async function save() {
    setSaving(true);
    try { await onSave(rows, ocrText); } finally { setSaving(false); }
  }

  async function reprocess() {
    if (!window.confirm("Повторное распознавание заменит текущий результат в редакторе. Продолжить?")) return;
    setProcessing(true);
    try { await onReprocess(); } finally { setProcessing(false); }
  }

  function exportRows(format: "csv" | "xlsx" | "json" | "xml") {
    const currentPage = page;
    if (rows.length === 0 || !currentPage) return;
    const filename = `geological-journal-page-${currentPage.page_number}.${format}`;
    if (format === "csv") exportGeologicalJournalCSV(rows, filename);
    if (format === "xlsx") exportGeologicalJournalXLSX(rows, filename);
    if (format === "json") exportGeologicalJournalJSON(rows, filename);
    if (format === "xml") exportGeologicalJournalXML(rows, filename);
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-3 sm:p-6" role="dialog" aria-modal="true" aria-label={`Редактор страницы ${page.page_number}`}>
      <div className="flex max-h-[94vh] w-full max-w-[1500px] flex-col overflow-hidden rounded-xl border border-border bg-bg shadow-2xl">
        <header className="flex flex-wrap items-center justify-between gap-3 border-b border-border px-4 py-3">
          <div>
            <h2 className="text-base font-semibold">Страница {page.page_number}</h2>
            <p className="mt-0.5 text-xs text-text3">{page.content_type || "Страница"} · поворот {page.orientation_degrees}° · уверенность {Math.round(page.orientation_confidence * 100)}%</p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <button type="button" onClick={() => void reprocess()} disabled={processing || saving || busy} className="inline-flex items-center gap-1.5 rounded-lg border border-ai/40 px-2.5 py-2 text-xs font-medium text-ai hover:bg-ai/10 disabled:opacity-40"><Sparkles size={14} />{processing ? "Распознавание…" : "Распознать заново"}</button>
            <button type="button" onClick={() => void save()} disabled={saving || processing || busy} className="inline-flex items-center gap-1.5 rounded-lg bg-text px-2.5 py-2 text-xs font-medium text-white disabled:opacity-40"><Save size={14} />{saving ? "Сохранение…" : "Сохранить"}</button>
            <button type="button" onClick={onClose} className="rounded p-2 text-text3 hover:bg-bg2" aria-label="Закрыть"><X size={18} /></button>
          </div>
        </header>
        <div className="grid min-h-0 flex-1 gap-4 overflow-auto p-4 lg:grid-cols-[minmax(280px,0.8fr)_minmax(0,1.4fr)]">
          <section className="min-h-[260px] rounded-lg border border-border2 bg-bg2 p-3">
            <h3 className="text-sm font-medium">Изображение страницы</h3>
            <div className="mt-3 max-h-[calc(94vh-170px)] overflow-auto rounded border border-border bg-white p-2"><img src={imageUrl} alt={`Страница ${page.page_number}`} className="h-auto w-full" /></div>
          </section>
          <section className="min-w-0 rounded-lg border border-border2">
            <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border px-3 py-2"><div className="flex gap-1"><button type="button" onClick={() => setActiveTab("table")} className={`rounded px-2.5 py-1.5 text-xs ${activeTab === "table" ? "bg-accent-bg font-medium text-accent" : "text-text3 hover:bg-bg2"}`}>Таблица ({rows.length})</button><button type="button" onClick={() => setActiveTab("ocr")} className={`rounded px-2.5 py-1.5 text-xs ${activeTab === "ocr" ? "bg-accent-bg font-medium text-accent" : "text-text3 hover:bg-bg2"}`}>Полный OCR</button></div><div className="flex flex-wrap gap-1"><button type="button" onClick={() => exportRows("xlsx")} disabled={!rows.length} className="inline-flex items-center gap-1 rounded border border-border2 px-2 py-1.5 text-[11px] disabled:opacity-40"><Download size={12} /> XLSX</button><button type="button" onClick={() => exportRows("csv")} disabled={!rows.length} className="inline-flex items-center gap-1 rounded border border-border2 px-2 py-1.5 text-[11px] disabled:opacity-40"><Download size={12} /> CSV</button><button type="button" onClick={() => exportRows("json")} disabled={!rows.length} className="inline-flex items-center gap-1 rounded border border-border2 px-2 py-1.5 text-[11px] disabled:opacity-40"><Download size={12} /> JSON</button><button type="button" onClick={() => exportRows("xml")} disabled={!rows.length} className="inline-flex items-center gap-1 rounded border border-border2 px-2 py-1.5 text-[11px] disabled:opacity-40"><Download size={12} /> XML</button></div></div>
            <div className="max-h-[calc(94vh-175px)] overflow-auto p-3">
              {activeTab === "ocr" ? <textarea value={ocrText} onChange={(event) => setOcrText(event.target.value)} className="min-h-[520px] w-full resize-y rounded border border-border2 bg-bg2 p-3 font-mono text-xs leading-5 outline-none focus:border-accent" placeholder="OCR-текст страницы" /> : <div className="min-w-[980px] overflow-auto rounded border border-border2"><table className="w-full text-left text-xs"><thead className="sticky top-0 bg-bg2"><tr><th className="px-2 py-2">№</th>{fields.map((field) => <th key={field.key} className="min-w-[130px] px-2 py-2">{field.label}</th>)}</tr></thead><tbody>{rows.map((row, index) => <tr key={index} className="border-t border-border"><td className="px-2 py-1.5 text-text3">{index + 1}</td>{fields.map((field) => <td key={field.key} className="px-1 py-1"><input type={field.numeric ? "number" : "text"} value={displayValue(row[field.key] as string | number | null)} onChange={(event) => updateCell(index, field.key, event.target.value)} className="w-full rounded border border-transparent bg-transparent px-1.5 py-1.5 outline-none focus:border-accent focus:bg-bg2" /></td>)}</tr>)}</tbody></table><button type="button" onClick={() => setRows((current) => [...current, emptyRow()])} className="m-2 rounded border border-border2 px-2.5 py-1.5 text-xs text-text2 hover:bg-bg2">Добавить строку</button>{rows.length === 0 && <p className="p-4 text-xs text-text3">Табличный результат ещё не создан. Нажмите «Распознать заново».</p>}</div>}
            </div>
          </section>
        </div>
        {processing && <div className="flex items-center gap-2 border-t border-border bg-bg2 px-4 py-2 text-xs text-text3"><LoaderCircle size={14} className="animate-spin" /> Страница отправлена на повторное распознавание…</div>}
      </div>
    </div>
  );
}
