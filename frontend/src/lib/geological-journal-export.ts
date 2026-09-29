import * as XLSX from "xlsx";
import type { GeologicalJournalRow } from "./api-geological-journal";

export type GeologicalJournalTableRow = Record<string, unknown>;

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

function cellValue(row: GeologicalJournalTableRow, key: GeologicalJournalExportColumn["key"]): string | number {
  const value = row[key];
  if (key === "uncertainties") return Array.isArray(value) ? value.join("; ") : "";
  return typeof value === "string" || typeof value === "number" ? value : "";
}

function rowsForLegacyExport(rows: GeologicalJournalRow[]) {
  return rows.map((row, index) => {
    const values: Record<string, string | number> = { "№": index + 1 };
    for (const column of EXPORT_COLUMNS) values[column.header] = cellValue(row, column.key);
    return values;
  });
}

function exportCellValue(value: unknown): string | number {
  if (Array.isArray(value)) return value.join("; ");
  if (typeof value === "string" || typeof value === "number") return value;
  if (typeof value === "boolean") return value ? "true" : "false";
  if (value && typeof value === "object") return JSON.stringify(value);
  return "";
}

function rowsForTableExport(rows: GeologicalJournalTableRow[], columns?: string[]) {
  const keys = columns?.length ? columns : Array.from(new Set(rows.flatMap((row) => Object.keys(row))));
  return {
    columns: keys,
    rows: rows.map((row, index) => {
      const values: Record<string, string | number> = { "№": index + 1 };
      for (const key of keys) values[key] = exportCellValue(row[key]);
      return values;
    }),
  };
}

function isDynamicTable(columns?: string[]): columns is string[] {
  return Boolean(columns?.length);
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

export function exportGeologicalJournalCSV(rows: GeologicalJournalRow[] | GeologicalJournalTableRow[], filename?: string, columns?: string[]) {
  const table = isDynamicTable(columns)
    ? rowsForTableExport(rows as GeologicalJournalTableRow[], columns)
    : { columns: EXPORT_COLUMNS.map((column) => column.header), rows: rowsForLegacyExport(rows as GeologicalJournalRow[]) };
  const headers = ["№", ...table.columns];
  const data = table.rows;
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

export function exportGeologicalJournalXLSX(rows: GeologicalJournalRow[] | GeologicalJournalTableRow[], filename?: string, columns?: string[]) {
  const table = isDynamicTable(columns)
    ? rowsForTableExport(rows as GeologicalJournalTableRow[], columns)
    : { columns: EXPORT_COLUMNS.map((column) => column.header), rows: rowsForLegacyExport(rows as GeologicalJournalRow[]) };
  const worksheet = XLSX.utils.json_to_sheet(table.rows, { header: ["№", ...table.columns], skipHeader: false });
  worksheet["!freeze"] = { xSplit: 0, ySplit: 1 };
  worksheet["!autofilter"] = { ref: worksheet["!ref"] ?? "A1:M1" };
  worksheet["!cols"] = [{ wch: 5 }, ...table.columns.map((column) => ({ wch: Math.min(Math.max(column.length + 2, 14), 38) }))];
  const workbook = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(workbook, worksheet, "Геологический журнал");
  XLSX.writeFile(workbook, filename ?? `geological-journal-${exportDate()}.xlsx`);
}

export function exportGeologicalJournalJSON(rows: GeologicalJournalRow[] | GeologicalJournalTableRow[], filename?: string, columns?: string[]) {
  const payload = JSON.stringify(isDynamicTable(columns) ? rowsForTableExport(rows as GeologicalJournalTableRow[], columns) : { rows }, null, 2);
  const blob = new Blob([payload], { type: "application/json;charset=utf-8" });
  downloadBlob(blob, filename ?? `geological-journal-${exportDate()}.json`);
}

export function exportGeologicalJournalXML(rows: GeologicalJournalRow[] | GeologicalJournalTableRow[], filename?: string, columns?: string[]) {
  const escapeXml = (value: unknown) =>
    String(value ?? "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&apos;");
  const tag = (name: string, value: unknown) =>
    `    <${name}>${escapeXml(value)}</${name}>`;
  const table = isDynamicTable(columns)
    ? rowsForTableExport(rows as GeologicalJournalTableRow[], columns)
    : rowsForTableExport(rows as GeologicalJournalTableRow[], EXPORT_COLUMNS.map((column) => column.key));
  const safeTagName = (name: string) => {
    const normalized = name.trim().replace(/[^A-Za-z0-9_.-]+/g, "_");
    return /^[A-Za-z_]/.test(normalized) ? normalized : `column_${normalized}`;
  };
  const xmlRows = table.rows
    .map((row, index) => {
      return [
        `  <row number="${index + 1}">`,
        ...table.columns.map((column) => tag(safeTagName(column), row[column])),
        "  </row>",
      ].join("\n");
    })
    .join("\n");
  const xml = `<?xml version="1.0" encoding="UTF-8"?>\n<geological_journal>\n${xmlRows}\n</geological_journal>\n`;
  const blob = new Blob([xml], { type: "application/xml;charset=utf-8" });
  downloadBlob(blob, filename ?? `geological-journal-${exportDate()}.xml`);
}
