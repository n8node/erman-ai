import * as XLSX from "xlsx";
import type { GeologicalJournalRow } from "./api-geological-journal";

export type GeologicalJournalExportColumn = {
  key: keyof Pick<
    GeologicalJournalRow,
    | "date"
    | "drilling_diameter_mm"
    | "depth_from_m"
    | "depth_to_m"
    | "drilling_run_m"
    | "core_recovery_m"
    | "core_recovery_pct"
    | "rock_description"
    | "sampling_interval"
    | "sample_number"
    | "notes"
    | "uncertainties"
  >;
  header: string;
};

const EXPORT_COLUMNS: GeologicalJournalExportColumn[] = [
  { key: "date", header: "Дата" },
  { key: "drilling_diameter_mm", header: "Диаметр, мм" },
  { key: "depth_from_m", header: "Глубина от, м" },
  { key: "depth_to_m", header: "Глубина до, м" },
  { key: "drilling_run_m", header: "Проходка, м" },
  { key: "core_recovery_m", header: "Выход керна, м" },
  { key: "core_recovery_pct", header: "Выход керна, %" },
  { key: "rock_description", header: "Описание породы" },
  { key: "sampling_interval", header: "Интервал опробования" },
  { key: "sample_number", header: "Номер пробы" },
  { key: "notes", header: "Примечания" },
  { key: "uncertainties", header: "Неопределённости" },
];

function cellValue(row: GeologicalJournalRow, key: GeologicalJournalExportColumn["key"]): string | number {
  const value = row[key];
  if (key === "uncertainties") return Array.isArray(value) ? value.join("; ") : "";
  return typeof value === "string" || typeof value === "number" ? value : "";
}

function rowsForExport(rows: GeologicalJournalRow[]) {
  return rows.map((row, index) => {
    const values: Record<string, string | number> = { "№": index + 1 };
    for (const column of EXPORT_COLUMNS) values[column.header] = cellValue(row, column.key);
    return values;
  });
}

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  window.setTimeout(() => URL.revokeObjectURL(url), 0);
}

function exportDate() {
  return new Date().toISOString().slice(0, 10);
}

export function exportGeologicalJournalCSV(rows: GeologicalJournalRow[], filename?: string) {
  const headers = ["№", ...EXPORT_COLUMNS.map((column) => column.header)];
  const data = rowsForExport(rows);
  const escape = (value: string | number) => {
    const text = String(value);
    return /[";\r\n]/.test(text) ? `"${text.replaceAll('"', '""')}"` : text;
  };
  const csv = [headers, ...data.map((row) => headers.map((header) => escape(row[header])))]
    .map((line) => line.join(";"))
    .join("\r\n");
  const blob = new Blob(["\uFEFF", csv], { type: "text/csv;charset=utf-8" });
  downloadBlob(blob, filename ?? `geological-journal-${exportDate()}.csv`);
}

export function exportGeologicalJournalXLSX(rows: GeologicalJournalRow[], filename?: string) {
  const worksheet = XLSX.utils.json_to_sheet(rowsForExport(rows), { skipHeader: false });
  worksheet["!freeze"] = { xSplit: 0, ySplit: 1 };
  worksheet["!autofilter"] = { ref: worksheet["!ref"] ?? "A1:M1" };
  worksheet["!cols"] = [
    { wch: 5 },
    { wch: 14 },
    { wch: 14 },
    { wch: 15 },
    { wch: 15 },
    { wch: 13 },
    { wch: 15 },
    { wch: 15 },
    { wch: 38 },
    { wch: 22 },
    { wch: 16 },
    { wch: 28 },
    { wch: 32 },
  ];
  const workbook = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(workbook, worksheet, "Геологический журнал");
  XLSX.writeFile(workbook, filename ?? `geological-journal-${exportDate()}.xlsx`);
}

export function exportGeologicalJournalJSON(rows: GeologicalJournalRow[], filename?: string) {
  const payload = JSON.stringify({ rows }, null, 2);
  const blob = new Blob([payload], { type: "application/json;charset=utf-8" });
  downloadBlob(blob, filename ?? `geological-journal-${exportDate()}.json`);
}

export function exportGeologicalJournalXML(rows: GeologicalJournalRow[], filename?: string) {
  const escapeXml = (value: string | number | null | undefined) =>
    String(value ?? "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&apos;");
  const tag = (name: string, value: string | number | null | undefined) =>
    `    <${name}>${escapeXml(value)}</${name}>`;
  const xmlRows = rows
    .map((row, index) => {
      const uncertainties = row.uncertainties ?? [];
      return [
        `  <row number="${index + 1}">`,
        tag("date", row.date),
        tag("drilling_diameter_mm", row.drilling_diameter_mm),
        tag("depth_from_m", row.depth_from_m),
        tag("depth_to_m", row.depth_to_m),
        tag("drilling_run_m", row.drilling_run_m),
        tag("core_recovery_m", row.core_recovery_m),
        tag("core_recovery_pct", row.core_recovery_pct),
        tag("rock_description", row.rock_description),
        tag("sampling_interval", row.sampling_interval),
        tag("sample_number", row.sample_number),
        tag("notes", row.notes),
        "    <uncertainties>",
        ...uncertainties.map((item) => tag("item", item)),
        "    </uncertainties>",
        "  </row>",
      ].join("\n");
    })
    .join("\n");
  const xml = `<?xml version="1.0" encoding="UTF-8"?>\n<geological_journal>\n${xmlRows}\n</geological_journal>\n`;
  const blob = new Blob([xml], { type: "application/xml;charset=utf-8" });
  downloadBlob(blob, filename ?? `geological-journal-${exportDate()}.xml`);
}
