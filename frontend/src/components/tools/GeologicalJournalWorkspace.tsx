"use client";

import {
  AlertCircle,
  BookOpen,
  Check,
  ChevronRight,
  Clock3,
  FileImage,
  Images,
  LoaderCircle,
  PencilLine,
  Plus,
  RefreshCw,
  RotateCcw,
  Save,
  Trash2,
  Upload,
  X,
} from "lucide-react";
import Image from "next/image";
import { useLocale, useTranslations } from "next-intl";
import { useCallback, useEffect, useRef, useState } from "react";
import { intlLocale } from "@/i18n/intl-locale";
import { cn } from "@/lib/utils";
import {
  analyzeGeologicalJournalPage,
  currentPageResult,
  deleteGeologicalJournalPage,
  geologicalJournalExampleImageUrl,
  geologicalJournalPageImageUrl,
  getGeologicalJournalPage,
  getGeologicalJournalRun,
  listGeologicalJournalExamples,
  listGeologicalJournalPages,
  pickWorkspaceRun,
  runOutputRows,
  saveGeologicalJournalResult,
  uploadGeologicalJournalPage,
  validateGeologicalJournalImage,
  type GeologicalJournalExample,
  type GeologicalJournalPage,
  type GeologicalJournalPageDetail,
  type GeologicalJournalRow,
  type GeologicalJournalRun,
} from "@/lib/api-geological-journal";

type View = "upload" | "library" | "examples";

const emptyRow = (): GeologicalJournalRow => ({
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
});

const numericFields = new Set<keyof GeologicalJournalRow>([
  "drilling_diameter_mm",
  "depth_from_m",
  "depth_to_m",
  "drilling_run_m",
  "core_recovery_m",
  "core_recovery_pct",
]);

const columns: { key: keyof GeologicalJournalRow; width: string }[] = [
  { key: "date", width: "min-w-[124px]" },
  { key: "drilling_diameter_mm", width: "min-w-[132px]" },
  { key: "depth_from_m", width: "min-w-[110px]" },
  { key: "depth_to_m", width: "min-w-[110px]" },
  { key: "drilling_run_m", width: "min-w-[112px]" },
  { key: "core_recovery_m", width: "min-w-[116px]" },
  { key: "core_recovery_pct", width: "min-w-[116px]" },
  { key: "rock_description", width: "min-w-[260px]" },
  { key: "sampling_interval", width: "min-w-[160px]" },
  { key: "sample_number", width: "min-w-[132px]" },
  { key: "notes", width: "min-w-[220px]" },
];

function formatBytes(bytes: number) {
  if (!bytes) return "—";
  if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function GeologicalJournalWorkspace() {
  const t = useTranslations("geologicalJournal");
  const locale = useLocale();
  const [view, setView] = useState<View>("upload");
  const [pages, setPages] = useState<GeologicalJournalPage[]>([]);
  const [examples, setExamples] = useState<GeologicalJournalExample[]>([]);
  const [page, setPage] = useState<GeologicalJournalPageDetail | null>(null);
  const [rows, setRows] = useState<GeologicalJournalRow[]>([]);
  const [run, setRun] = useState<GeologicalJournalRun | null>(null);
  const [uploadPercent, setUploadPercent] = useState<number | null>(null);
  const [localPreview, setLocalPreview] = useState("");
  const [dragActive, setDragActive] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [previewExample, setPreviewExample] = useState<GeologicalJournalExample | null>(
    null
  );
  const abortUploadRef = useRef<(() => void) | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const date = useCallback(
    (value: string) =>
      new Date(value).toLocaleString(intlLocale(locale), {
        day: "2-digit",
        month: "short",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      }),
    [locale]
  );

  const loadCollections = useCallback(async () => {
    setLoading(true);
    try {
      const [pagesData, examplesData] = await Promise.all([
        listGeologicalJournalPages(),
        listGeologicalJournalExamples(),
      ]);
      setPages(pagesData.items ?? []);
      setExamples(examplesData.items ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.load"));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    void loadCollections();
    return () => abortUploadRef.current?.();
  }, [loadCollections]);

  useEffect(() => {
    if (!localPreview) return;
    return () => URL.revokeObjectURL(localPreview);
  }, [localPreview]);

  const openPage = useCallback(
    async (id: string) => {
      setLoading(true);
      setError("");
      setSuccess("");
      setLocalPreview("");
      setUploadPercent(null);
      try {
        const detail = await getGeologicalJournalPage(id);
        const result = currentPageResult(detail);
        const active = pickWorkspaceRun(detail.runs, result);
        setPage(detail);
        const activeRows = runOutputRows(active);
        setRows(activeRows.length > 0 ? activeRows : (result?.rows ?? []));
        setDirty(false);
        setRun(active);
        // Only surface a historical error when there is no saved table yet.
        // Library uses latest_result; Recognition must not disagree with it.
        if (active?.status === "error" && !(result?.rows?.length)) {
          setError(active.error_msg || t("errors.recognition"));
        }
        setView("upload");
      } catch (err) {
        setError(err instanceof Error ? err.message : t("errors.loadPage"));
      } finally {
        setLoading(false);
      }
    },
    [t]
  );

  const activeRunId = run?.id;
  const activeRunStatus = run?.status;
  const activePageId = page?.id;

  useEffect(() => {
    if (
      !activeRunId ||
      !activeRunStatus ||
      !["pending", "processing"].includes(activeRunStatus)
    ) {
      return;
    }
    let stopped = false;
    const poll = async () => {
      try {
        const next = await getGeologicalJournalRun(activeRunId);
        if (stopped) return;
        setRun(next);
        if (next.status === "done" && activePageId) {
          setError("");
          const detail = await getGeologicalJournalPage(activePageId);
          if (stopped) return;
          setPage(detail);
          setRows(
            runOutputRows(next).length > 0
              ? runOutputRows(next)
              : (currentPageResult(detail)?.rows ?? [])
          );
          setDirty(false);
          setPages((current) =>
            current.map((item) => (item.id === detail.id ? detail : item))
          );
        }
        if (next.status === "error") {
          setError(next.error_msg || t("errors.recognition"));
          // Keep the previously saved table visible (same source as Library).
          if (activePageId) {
            const detail = await getGeologicalJournalPage(activePageId);
            if (stopped) return;
            setPage(detail);
            const saved = currentPageResult(detail)?.rows ?? [];
            if (saved.length > 0) {
              setRows(saved);
              setDirty(false);
            }
            setPages((current) =>
              current.map((item) => (item.id === detail.id ? detail : item))
            );
          }
        }
      } catch {
        // A transient polling failure should not stop recognition tracking.
      }
    };
    void poll();
    const timer = window.setInterval(poll, 3000);
    return () => {
      stopped = true;
      window.clearInterval(timer);
    };
  }, [activePageId, activeRunId, activeRunStatus, t]);

  async function handleFile(file: File) {
    const validation = validateGeologicalJournalImage(file);
    if (validation) {
      setError(t(`errors.${validation}`));
      return;
    }
    setLocalPreview(URL.createObjectURL(file));
    setError("");
    setSuccess("");
    setUploadPercent(0);
    const request = uploadGeologicalJournalPage(file, setUploadPercent);
    abortUploadRef.current = request.abort;
    try {
      const response = await request.promise;
      setUploadPercent(100);
      const detail = await getGeologicalJournalPage(response.page.id);
      setPage(detail);
      setRows(currentPageResult(detail)?.rows ?? []);
      setRun({
        id: response.run_id,
        status: "pending",
        created_at: new Date().toISOString(),
      });
      setPages((current) => [response.page, ...current]);
      setDirty(false);
    } catch (err) {
      if (!(err instanceof DOMException && err.name === "AbortError")) {
        setError(err instanceof Error ? err.message : t("errors.upload"));
      }
      setUploadPercent(null);
    } finally {
      abortUploadRef.current = null;
    }
  }

  async function handleRetry() {
    if (!page) return;
    setError("");
    setSuccess("");
    setRows([]);
    try {
      const response = await analyzeGeologicalJournalPage(page.id);
      setRun({
        id: response.run_id,
        status: response.status || "pending",
        created_at: new Date().toISOString(),
      });
      const detail = await getGeologicalJournalPage(page.id);
      setPage(detail);
      setDirty(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.retry"));
    }
  }

  async function handleSave() {
    if (!page) return;
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      await saveGeologicalJournalResult(page.id, rows);
      const detail = await getGeologicalJournalPage(page.id);
      setPage(detail);
      setPages((current) =>
        current.map((item) => (item.id === detail.id ? detail : item))
      );
      setDirty(false);
      setSuccess(t("editor.saved"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.save"));
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(id: string) {
    if (!window.confirm(t("library.deleteConfirm"))) return;
    setDeleting(true);
    setError("");
    try {
      await deleteGeologicalJournalPage(id);
      setPages((current) => current.filter((item) => item.id !== id));
      if (page?.id === id) resetWorkspace();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.delete"));
    } finally {
      setDeleting(false);
    }
  }

  function resetWorkspace() {
    setPage(null);
    setRun(null);
    setRows([]);
    setLocalPreview("");
    setUploadPercent(null);
    setDirty(false);
    setSuccess("");
    setError("");
  }

  function updateRow(
    index: number,
    key: keyof GeologicalJournalRow,
    value: string
  ) {
    setRows((current) =>
      current.map((row, rowIndex) => {
        if (rowIndex !== index) return row;
        if (numericFields.has(key)) {
          return {
            ...row,
            [key]: value === "" ? null : Number(value),
          };
        }
        return { ...row, [key]: value };
      })
    );
    setDirty(true);
    setSuccess("");
  }

  const recognizing = Boolean(
    run && ["pending", "processing"].includes(run.status)
  );

  return (
    <div className="mx-auto max-w-[1500px] space-y-6">
      <header className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">{t("title")}</h1>
          <p className="mt-1 max-w-2xl text-sm text-text2">{t("subtitle")}</p>
        </div>
        {page && (
          <button
            type="button"
            onClick={resetWorkspace}
            className="inline-flex items-center justify-center gap-2 rounded-lg border border-border2 bg-bg px-3 py-2 text-sm text-text2 transition-colors duration-150 hover:bg-bg2 hover:text-text"
          >
            <Plus size={15} />
            {t("newPage")}
          </button>
        )}
      </header>

      <nav className="flex gap-1 overflow-x-auto border-b border-border">
        {(
          [
            ["upload", Upload],
            ["library", Images],
            ["examples", BookOpen],
          ] as const
        ).map(([id, Icon]) => (
          <button
            key={id}
            type="button"
            onClick={() => setView(id)}
            className={cn(
              "-mb-px inline-flex shrink-0 items-center gap-2 border-b-2 px-3 py-2.5 text-sm transition-colors duration-150",
              view === id
                ? "border-text font-medium text-text"
                : "border-transparent text-text2 hover:text-text"
            )}
          >
            <Icon size={15} />
            {t(`tabs.${id}`)}
            {id === "library" && pages.length > 0 && (
              <span className="rounded bg-bg2 px-1.5 py-0.5 text-[11px] text-text3">
                {pages.length}
              </span>
            )}
          </button>
        ))}
      </nav>

      {error && (
        <div
          role="alert"
          className="flex items-start gap-2 rounded-lg border border-red-200 bg-error-bg px-3 py-2.5 text-sm text-error"
        >
          <AlertCircle size={16} className="mt-0.5 shrink-0" />
          <span>{error}</span>
        </div>
      )}
      {success && (
        <div className="flex items-center gap-2 rounded-lg border border-green-200 bg-success-bg px-3 py-2.5 text-sm text-success">
          <Check size={16} />
          {success}
        </div>
      )}

      {view === "upload" && !page && (
        <UploadPanel
          active={dragActive}
          progress={uploadPercent}
          onActive={setDragActive}
          onFile={handleFile}
          onBrowse={() => fileInputRef.current?.click()}
          onCancel={() => abortUploadRef.current?.()}
          t={t}
        />
      )}

      {view === "upload" && page && (
        <div className="space-y-5">
          <RecognitionSteps
            uploadDone
            recognizing={recognizing}
            runStatus={run?.status}
            rowCount={rows.length}
            failed={run?.status === "error"}
            t={t}
          />

          {recognizing && (
            <ProcessingBanner run={run} t={t} />
          )}

          {run?.status === "done" && rows.length === 0 && !recognizing && (
            <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-warning/30 bg-warning-bg p-4">
              <div>
                <p className="text-sm font-medium text-warning">{t("status.emptyResult")}</p>
                <p className="mt-1 text-xs text-text2">{t("status.emptyResultHint")}</p>
              </div>
              <button
                type="button"
                onClick={() => void handleRetry()}
                className="inline-flex items-center gap-2 rounded-lg border border-border2 bg-bg px-3 py-2 text-sm font-medium text-text hover:bg-bg2"
              >
                <RotateCcw size={15} />
                {t("actions.retry")}
              </button>
            </div>
          )}

          {run?.status === "error" && (
            <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-red-200 bg-error-bg p-4">
              <div>
                <p className="text-sm font-medium text-error">{t("status.error")}</p>
                <p className="mt-1 text-xs text-text2">
                  {run.error_msg || t("errors.recognition")}
                </p>
                {rows.length > 0 && (
                  <p className="mt-1 text-xs text-text3">
                    {t("status.previousResultKept")}
                  </p>
                )}
              </div>
              <button
                type="button"
                onClick={() => void handleRetry()}
                className="inline-flex items-center gap-2 rounded-lg border border-red-300 bg-bg px-3 py-2 text-sm font-medium text-error hover:bg-red-50"
              >
                <RotateCcw size={15} />
                {t("actions.retry")}
              </button>
            </div>
          )}

          <div className="grid min-w-0 gap-5 xl:grid-cols-[minmax(320px,0.78fr)_minmax(0,1.45fr)]">
            <section className="min-w-0 rounded-xl border border-border bg-bg">
              <div className="flex items-center justify-between border-b border-border px-4 py-3">
                <div className="min-w-0">
                  <h2 className="truncate text-sm font-medium">{page.original_name}</h2>
                  <p className="mt-0.5 text-xs text-text3">
                    {formatBytes(page.size_bytes)}
                    {page.width > 0 && page.height > 0
                      ? ` · ${page.width} × ${page.height}`
                      : ""}
                  </p>
                </div>
                <FileImage size={17} className="shrink-0 text-text3" />
              </div>
              <div className="flex min-h-[420px] items-start justify-center overflow-auto bg-bg2 p-3 xl:max-h-[72vh]">
                {/* Authenticated API image; the browser sends the session cookie. */}
                <Image
                  src={localPreview || geologicalJournalPageImageUrl(page.id)}
                  alt={page.original_name}
                  width={page.width || 1600}
                  height={page.height || 1200}
                  unoptimized
                  className="h-auto max-w-full rounded border border-border bg-white object-contain"
                />
              </div>
            </section>

            <section className="min-w-0 rounded-xl border border-border bg-bg">
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border px-4 py-3">
                <div>
                  <h2 className="text-sm font-medium">{t("editor.title")}</h2>
                  <p className="mt-0.5 text-xs text-text3">
                    {t("editor.rows", { count: rows.length })}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => {
                      setRows((current) => [...current, emptyRow()]);
                      setDirty(true);
                    }}
                    disabled={recognizing}
                    className="inline-flex items-center gap-1.5 rounded-lg border border-border2 px-3 py-2 text-xs font-medium text-text2 hover:bg-bg2 disabled:opacity-50"
                  >
                    <Plus size={14} />
                    {t("editor.addRow")}
                  </button>
                  <button
                    type="button"
                    onClick={() => void handleSave()}
                    disabled={saving || recognizing || !dirty}
                    className="inline-flex items-center gap-1.5 rounded-lg bg-text px-3 py-2 text-xs font-medium text-white disabled:cursor-not-allowed disabled:opacity-40"
                  >
                    {saving ? (
                      <LoaderCircle size={14} className="animate-spin" />
                    ) : (
                      <Save size={14} />
                    )}
                    {saving ? t("editor.saving") : t("editor.save")}
                  </button>
                </div>
              </div>

              {rows.length === 0 ? (
                <div className="flex min-h-[300px] flex-col items-center justify-center px-6 text-center">
                  {recognizing ? (
                    <LoaderCircle size={24} className="animate-spin text-ai" />
                  ) : (
                    <PencilLine size={24} className="text-text3" />
                  )}
                  <p className="mt-3 text-sm font-medium">
                    {recognizing ? t("editor.waitingTitle") : t("editor.emptyTitle")}
                  </p>
                  <p className="mt-1 max-w-sm text-xs leading-5 text-text3">
                    {recognizing ? t("editor.waitingText") : t("editor.emptyText")}
                  </p>
                </div>
              ) : (
                <div className="max-h-[72vh] overflow-auto">
                  <table className="w-full border-collapse text-xs">
                    <thead className="sticky top-0 z-10 bg-bg2">
                      <tr className="border-b border-border text-left text-[10px] uppercase tracking-wide text-text3">
                        <th className="w-10 px-2 py-2.5 text-center font-medium">#</th>
                        {columns.map((column) => (
                          <th
                            key={column.key}
                            className={cn("px-2 py-2.5 font-medium", column.width)}
                          >
                            {t(`columns.${column.key}`)}
                          </th>
                        ))}
                        <th className="sticky right-0 w-10 bg-bg2 px-2 py-2.5" />
                      </tr>
                    </thead>
                    <tbody>
                      {rows.map((row, rowIndex) => (
                        <tr
                          key={rowIndex}
                          className={cn(
                            "border-b border-border align-top last:border-0",
                            row.uncertainties?.length > 0 && "bg-warning-bg/40"
                          )}
                        >
                          <td className="px-2 py-2 text-center text-text3">
                            <span className="inline-flex items-center gap-1">
                              {row.uncertainties?.length > 0 && (
                                <AlertCircle
                                  size={12}
                                  className="text-warning"
                                  aria-label={row.uncertainties.join(", ")}
                                />
                              )}
                              {rowIndex + 1}
                            </span>
                          </td>
                          {columns.map((column) => (
                            <td key={column.key} className="p-1.5">
                              {column.key === "rock_description" ||
                              column.key === "notes" ? (
                                <textarea
                                  rows={2}
                                  aria-label={t(`columns.${column.key}`)}
                                  value={String(row[column.key] ?? "")}
                                  onChange={(event) =>
                                    updateRow(rowIndex, column.key, event.target.value)
                                  }
                                  className="w-full resize-y rounded-md border border-transparent bg-transparent px-2 py-1.5 leading-5 outline-none hover:border-border2 focus:border-accent focus:bg-bg"
                                />
                              ) : (
                                <input
                                  type={numericFields.has(column.key) ? "number" : "text"}
                                  step="any"
                                  aria-label={t(`columns.${column.key}`)}
                                  value={String(row[column.key] ?? "")}
                                  onChange={(event) =>
                                    updateRow(rowIndex, column.key, event.target.value)
                                  }
                                  className="w-full rounded-md border border-transparent bg-transparent px-2 py-1.5 outline-none hover:border-border2 focus:border-accent focus:bg-bg"
                                />
                              )}
                            </td>
                          ))}
                          <td className="sticky right-0 bg-bg p-1.5">
                            <button
                              type="button"
                              aria-label={t("editor.deleteRow")}
                              onClick={() => {
                                setRows((current) =>
                                  current.filter((_, index) => index !== rowIndex)
                                );
                                setDirty(true);
                              }}
                              className="rounded p-1.5 text-text3 hover:bg-error-bg hover:text-error"
                            >
                              <Trash2 size={14} />
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </section>
          </div>

          <PageHistory
            runs={page.runs ?? []}
            versions={page.versions?.length ?? 0}
            date={date}
            onRetry={handleRetry}
            disabled={recognizing}
            t={t}
          />
        </div>
      )}

      {view === "library" && (
        <Library
          items={pages}
          loading={loading}
          deleting={deleting}
          date={date}
          onOpen={openPage}
          onDelete={handleDelete}
          t={t}
        />
      )}

      {view === "examples" && (
        <Examples
          items={examples}
          loading={loading}
          onPreview={setPreviewExample}
          t={t}
        />
      )}

      {previewExample && (
        <ExampleModal
          example={previewExample}
          onClose={() => setPreviewExample(null)}
          t={t}
        />
      )}

      <input
        ref={fileInputRef}
        type="file"
        accept="image/jpeg,image/png,image/webp,image/gif"
        className="hidden"
        onChange={(event) => {
          const file = event.target.files?.[0];
          if (file) void handleFile(file);
          event.target.value = "";
        }}
      />
    </div>
  );
}

function UploadPanel({
  active,
  progress,
  onActive,
  onFile,
  onBrowse,
  onCancel,
  t,
}: {
  active: boolean;
  progress: number | null;
  onActive: (active: boolean) => void;
  onFile: (file: File) => void;
  onBrowse: () => void;
  onCancel: () => void;
  t: ReturnType<typeof useTranslations>;
}) {
  const uploading = progress !== null && progress < 100;
  return (
    <section className="mx-auto max-w-3xl rounded-xl border border-border bg-bg p-4 sm:p-6">
      <div
        onDragEnter={(event) => {
          event.preventDefault();
          onActive(true);
        }}
        onDragOver={(event) => event.preventDefault()}
        onDragLeave={(event) => {
          event.preventDefault();
          if (event.currentTarget === event.target) onActive(false);
        }}
        onDrop={(event) => {
          event.preventDefault();
          onActive(false);
          const file = event.dataTransfer.files?.[0];
          if (file) onFile(file);
        }}
        className={cn(
          "flex min-h-[330px] flex-col items-center justify-center rounded-lg border border-dashed px-6 py-10 text-center transition-colors duration-150",
          active ? "border-accent bg-accent-bg" : "border-border2 bg-bg2/40"
        )}
      >
        <div className="flex h-12 w-12 items-center justify-center rounded-xl border border-border bg-bg">
          {uploading ? (
            <LoaderCircle size={22} className="animate-spin text-accent" />
          ) : (
            <Upload size={22} className="text-text2" />
          )}
        </div>
        <h2 className="mt-4 text-base font-medium">
          {uploading ? t("upload.uploading") : t("upload.title")}
        </h2>
        <p className="mt-2 max-w-md text-sm leading-6 text-text2">
          {uploading ? t("upload.keepOpen") : t("upload.description")}
        </p>
        {uploading ? (
          <div className="mt-6 w-full max-w-sm">
            <div className="mb-2 flex items-center justify-between text-xs">
              <span className="text-text2">{t("upload.progress")}</span>
              <span className="font-medium text-text">{progress}%</span>
            </div>
            <div
              className="h-1.5 overflow-hidden rounded-full bg-border"
              role="progressbar"
              aria-valuenow={progress}
              aria-valuemin={0}
              aria-valuemax={100}
            >
              <div
                className="h-full bg-accent transition-[width] duration-150"
                style={{ width: `${progress}%` }}
              />
            </div>
            <button
              type="button"
              onClick={onCancel}
              className="mt-4 text-xs text-text3 hover:text-error"
            >
              {t("upload.cancel")}
            </button>
          </div>
        ) : (
          <>
            <button
              type="button"
              onClick={onBrowse}
              className="mt-6 rounded-lg bg-text px-4 py-2.5 text-sm font-medium text-white"
            >
              {t("upload.choose")}
            </button>
            <p className="mt-4 text-xs text-text3">{t("upload.formats")}</p>
          </>
        )}
      </div>
      <div className="mt-4 grid gap-3 text-xs text-text2 sm:grid-cols-3">
        {(["quality", "privacy", "corrections"] as const).map((key) => (
          <div key={key} className="flex gap-2 rounded-lg bg-bg2 px-3 py-2.5">
            <Check size={14} className="mt-0.5 shrink-0 text-success" />
            {t(`upload.hints.${key}`)}
          </div>
        ))}
      </div>
    </section>
  );
}

function ProcessingBanner({
  run,
  t,
}: {
  run: GeologicalJournalRun | null;
  t: ReturnType<typeof useTranslations>;
}) {
  const [elapsed, setElapsed] = useState(0);
  useEffect(() => {
    if (!run?.created_at) return;
    const started = new Date(run.created_at).getTime();
    const tick = () => setElapsed(Math.max(0, Math.floor((Date.now() - started) / 1000)));
    tick();
    const timer = window.setInterval(tick, 1000);
    return () => window.clearInterval(timer);
  }, [run?.created_at, run?.id]);
  return (
    <div className="flex items-start gap-3 rounded-xl border border-[#d6d2f4] bg-ai-bg p-4">
      <LoaderCircle size={18} className="mt-0.5 shrink-0 animate-spin text-ai" />
      <div>
        <p className="text-sm font-medium text-ai">
          {t(`status.${run?.status === "pending" ? "pending" : "processing"}`)}
        </p>
        <p className="mt-1 text-xs text-text2">
          {t("status.polling")} {t("status.elapsed", { seconds: elapsed })}
        </p>
      </div>
    </div>
  );
}

function RecognitionSteps({
  uploadDone,
  recognizing,
  runStatus,
  rowCount,
  failed,
  t,
}: {
  uploadDone: boolean;
  recognizing: boolean;
  runStatus?: string;
  rowCount: number;
  failed: boolean;
  t: ReturnType<typeof useTranslations>;
}) {
  const emptyDone = runStatus === "done" && rowCount === 0;
  const states = [
    { done: uploadDone, active: false, failed: false },
    {
      done: (runStatus === "done" && rowCount > 0) || failed || emptyDone,
      active: recognizing,
      failed: failed || emptyDone,
    },
    {
      done: runStatus === "done" && rowCount > 0,
      active: false,
      failed: false,
    },
  ];
  return (
    <div className="overflow-x-auto rounded-xl border border-border bg-bg px-4 py-3">
      <ol className="flex min-w-[560px] items-center">
        {states.map((state, index) => (
          <li key={index} className="flex flex-1 items-center last:flex-none">
            <div className="flex items-center gap-2">
              <span
                className={cn(
                  "flex h-7 w-7 items-center justify-center rounded-full text-xs font-medium",
                  state.failed && "bg-error-bg text-error",
                  state.done && "bg-success-bg text-success",
                  state.active && !state.failed && "bg-ai-bg text-ai",
                  !state.done && !state.active && !state.failed && "bg-bg2 text-text3"
                )}
              >
                {state.failed ? (
                  <X size={14} />
                ) : state.done ? (
                  <Check size={14} />
                ) : state.active ? (
                  <LoaderCircle size={14} className="animate-spin" />
                ) : (
                  index + 1
                )}
              </span>
              <div>
                <p
                  className={cn(
                    "text-xs font-medium",
                    state.active || state.done ? "text-text" : "text-text3"
                  )}
                >
                  {t(`steps.${index + 1}.title`)}
                </p>
                <p className="text-[11px] text-text3">
                  {t(`steps.${index + 1}.description`)}
                </p>
              </div>
            </div>
            {index < states.length - 1 && (
              <ChevronRight size={15} className="mx-4 text-border2" />
            )}
          </li>
        ))}
      </ol>
    </div>
  );
}

function PageHistory({
  runs,
  versions,
  date,
  onRetry,
  disabled,
  t,
}: {
  runs: GeologicalJournalRun[];
  versions: number;
  date: (value: string) => string;
  onRetry: () => void;
  disabled: boolean;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <section className="rounded-xl border border-border bg-bg">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border px-4 py-3">
        <div>
          <h2 className="text-sm font-medium">{t("history.title")}</h2>
          <p className="mt-0.5 text-xs text-text3">
            {t("history.summary", { runs: runs.length, versions })}
          </p>
        </div>
        <button
          type="button"
          disabled={disabled}
          onClick={onRetry}
          className="inline-flex items-center gap-1.5 rounded-lg border border-border2 px-3 py-2 text-xs font-medium text-text2 hover:bg-bg2 disabled:opacity-50"
        >
          <RefreshCw size={14} />
          {t("actions.recognizeAgain")}
        </button>
      </div>
      {runs.length === 0 ? (
        <p className="px-4 py-5 text-sm text-text3">{t("history.empty")}</p>
      ) : (
        <div className="divide-y divide-border">
          {runs.map((item) => (
            <div
              key={item.id}
              className="flex flex-wrap items-center justify-between gap-2 px-4 py-3"
            >
              <div className="flex items-center gap-2">
                <Clock3 size={14} className="text-text3" />
                <span className="text-xs text-text2">{date(item.created_at)}</span>
                {item.model_used && (
                  <span className="hidden font-mono text-[11px] text-text3 sm:inline">
                    {item.model_used}
                  </span>
                )}
              </div>
              <StatusBadge status={item.status} t={t} />
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

function Library({
  items,
  loading,
  deleting,
  date,
  onOpen,
  onDelete,
  t,
}: {
  items: GeologicalJournalPage[];
  loading: boolean;
  deleting: boolean;
  date: (value: string) => string;
  onOpen: (id: string) => void;
  onDelete: (id: string) => void;
  t: ReturnType<typeof useTranslations>;
}) {
  if (loading) return <p className="text-sm text-text2">{t("loading")}</p>;
  if (items.length === 0) {
    return (
      <div className="rounded-xl border border-border bg-bg px-6 py-14 text-center">
        <Images size={28} className="mx-auto text-text3" />
        <h2 className="mt-3 text-sm font-medium">{t("library.emptyTitle")}</h2>
        <p className="mt-1 text-sm text-text2">{t("library.emptyText")}</p>
      </div>
    );
  }
  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      {items.map((item) => {
        const result = currentPageResult(item);
        return (
          <article
            key={item.id}
            className="group overflow-hidden rounded-xl border border-border bg-bg transition-colors duration-150 hover:border-border2"
          >
            <button
              type="button"
              onClick={() => void onOpen(item.id)}
              className="relative block aspect-[16/10] w-full overflow-hidden bg-bg2"
            >
              <Image
                src={geologicalJournalPageImageUrl(item.id)}
                alt={item.original_name}
                fill
                sizes="(max-width: 640px) 100vw, (max-width: 1280px) 50vw, 33vw"
                unoptimized
                className="h-full w-full object-cover transition-transform duration-200 group-hover:scale-[1.015]"
              />
            </button>
            <div className="p-4">
              <div className="flex items-start justify-between gap-3">
                <button
                  type="button"
                  onClick={() => void onOpen(item.id)}
                  className="min-w-0 text-left"
                >
                  <h2 className="truncate text-sm font-medium">{item.original_name}</h2>
                  <p className="mt-1 text-xs text-text3">{date(item.created_at)}</p>
                </button>
                <button
                  type="button"
                  disabled={deleting}
                  aria-label={t("library.delete")}
                  onClick={() => void onDelete(item.id)}
                  className="rounded p-1.5 text-text3 hover:bg-error-bg hover:text-error disabled:opacity-50"
                >
                  <Trash2 size={15} />
                </button>
              </div>
              <div className="mt-3 flex items-center justify-between border-t border-border pt-3 text-xs">
                <span className="text-text3">{formatBytes(item.size_bytes)}</span>
                <span
                  className={cn(
                    "rounded px-2 py-0.5",
                    result ? "bg-success-bg text-success" : "bg-bg2 text-text3"
                  )}
                >
                  {result
                    ? t("library.recognized", { count: result.rows?.length ?? 0 })
                    : t("library.notRecognized")}
                </span>
              </div>
            </div>
          </article>
        );
      })}
    </div>
  );
}

function Examples({
  items,
  loading,
  onPreview,
  t,
}: {
  items: GeologicalJournalExample[];
  loading: boolean;
  onPreview: (item: GeologicalJournalExample) => void;
  t: ReturnType<typeof useTranslations>;
}) {
  if (loading) return <p className="text-sm text-text2">{t("loading")}</p>;
  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-base font-medium">{t("examples.title")}</h2>
        <p className="mt-1 text-sm text-text2">{t("examples.subtitle")}</p>
      </div>
      {items.length === 0 ? (
        <div className="rounded-xl border border-border bg-bg px-6 py-12 text-center text-sm text-text2">
          {t("examples.empty")}
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {items.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => onPreview(item)}
              className="group overflow-hidden rounded-xl border border-border bg-bg text-left transition-colors duration-150 hover:border-border2"
            >
              <div className="relative aspect-[16/10] overflow-hidden bg-bg2">
                <Image
                  src={geologicalJournalExampleImageUrl(item.id)}
                  alt={item.title}
                  fill
                  sizes="(max-width: 640px) 100vw, (max-width: 1280px) 50vw, 33vw"
                  unoptimized
                  className="h-full w-full object-cover transition-transform duration-200 group-hover:scale-[1.015]"
                />
              </div>
              <div className="p-4">
                <h3 className="text-sm font-medium">{item.title}</h3>
                <p className="mt-1 line-clamp-2 text-xs leading-5 text-text2">
                  {item.description || t("examples.noDescription")}
                </p>
                <span className="mt-3 inline-flex items-center gap-1 text-xs font-medium text-accent">
                  {t("examples.preview")}
                  <ChevronRight size={13} />
                </span>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

function ExampleModal({
  example,
  onClose,
  t,
}: {
  example: GeologicalJournalExample;
  onClose: () => void;
  t: ReturnType<typeof useTranslations>;
}) {
  useEffect(() => {
    const close = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    window.addEventListener("keydown", close);
    return () => window.removeEventListener("keydown", close);
  }, [onClose]);
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/45 p-4"
      role="dialog"
      aria-modal="true"
      aria-label={example.title}
      onMouseDown={(event) => {
        if (event.currentTarget === event.target) onClose();
      }}
    >
      <div className="flex max-h-[92vh] w-full max-w-5xl flex-col overflow-hidden rounded-xl bg-bg shadow-xl">
        <div className="flex items-start justify-between gap-3 border-b border-border px-4 py-3">
          <div>
            <h2 className="text-sm font-medium">{example.title}</h2>
            {example.description && (
              <p className="mt-1 text-xs text-text2">{example.description}</p>
            )}
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label={t("examples.close")}
            className="rounded p-1.5 text-text3 hover:bg-bg2 hover:text-text"
          >
            <X size={18} />
          </button>
        </div>
        <div className="min-h-0 flex-1 overflow-auto bg-bg2 p-3">
          <Image
            src={geologicalJournalExampleImageUrl(example.id)}
            alt={example.title}
            width={1600}
            height={1200}
            unoptimized
            className="mx-auto h-auto max-w-full rounded border border-border bg-white"
          />
        </div>
        <div className="border-t border-border px-4 py-3">
          <p className="text-xs leading-5 text-text2">{t("examples.guidance")}</p>
        </div>
      </div>
    </div>
  );
}

function StatusBadge({
  status,
  t,
}: {
  status: string;
  t: ReturnType<typeof useTranslations>;
}) {
  const key = ["pending", "processing", "done", "error"].includes(status)
    ? status
    : "pending";
  return (
    <span
      className={cn(
        "rounded px-2 py-0.5 text-[11px] font-medium",
        key === "done" && "bg-success-bg text-success",
        key === "error" && "bg-error-bg text-error",
        key === "processing" && "bg-ai-bg text-ai",
        key === "pending" && "bg-bg2 text-text3"
      )}
    >
      {t(`status.${key}`)}
    </span>
  );
}
