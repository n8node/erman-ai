"use client";

import { FileText, LoaderCircle, Play, Upload, Users, RefreshCw, Trash2, X, Eye, ZoomIn, ZoomOut } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  getGeologicalJournalDocument,
  chatGeologicalJournalDocument,
  deleteGeologicalJournalDocument,
  listGeologicalJournalDocuments,
  processGeologicalJournalDocumentLLM,
  setGeologicalJournalDocumentSharing,
  startGeologicalJournalDocumentAnalysis,
  uploadGeologicalJournalDocument,
  type GeologicalJournalDocument,
  type GeologicalJournalDocumentDetail,
} from "@/lib/api-geological-journal";

function formatBytes(value: number) {
  if (value < 1024 * 1024) return `${Math.round(value / 1024)} KB`;
  return `${(value / 1024 / 1024).toFixed(1)} MB`;
}

function statusLabel(status: string) {
  return ({ uploaded: "Загружен", queued: "В очереди", processing: "Обрабатывается", partially_done: "Частично готов", done: "Готово", error: "Ошибка" } as Record<string, string>)[status] ?? status;
}

export function GeologicalJournalDocumentsWorkspace() {
  const [documents, setDocuments] = useState<GeologicalJournalDocument[]>([]);
  const [active, setActive] = useState<GeologicalJournalDocumentDetail | null>(null);
  const [selectedPages, setSelectedPages] = useState<number[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [chatInput, setChatInput] = useState("");
  const [chatMessages, setChatMessages] = useState<Array<{ role: "user" | "assistant"; content: string; confidence?: string }>>([]);
  const [chatBusy, setChatBusy] = useState(false);
  const [previewPage, setPreviewPage] = useState<GeologicalJournalDocumentDetail["pages"][number] | null>(null);
  const [previewZoom, setPreviewZoom] = useState(1);
  const [uploadProgress, setUploadProgress] = useState<number | null>(null);
  const [uploadAbort, setUploadAbort] = useState<(() => void) | null>(null);
  const activePages = useMemo(() => active?.pages ?? [], [active?.pages]);

  const previewImageUrl = previewPage && active
    ? `/api/v1/tools/geological-journal/documents/${active.document.id}/pages/${previewPage.id}/assets/oriented`
    : "";

  const loadDocuments = useCallback(async () => {
    setLoading(true);
    try { setDocuments((await listGeologicalJournalDocuments()).items ?? []); }
    catch (err) { setError(err instanceof Error ? err.message : "Не удалось загрузить документы"); }
    finally { setLoading(false); }
  }, []);

  const loadActive = useCallback(async (id: string) => {
    try { setActive(await getGeologicalJournalDocument(id)); }
    catch (err) { setError(err instanceof Error ? err.message : "Не удалось загрузить документ"); }
  }, []);

  useEffect(() => { void loadDocuments(); }, [loadDocuments]);

  useEffect(() => {
    if (!active || !["queued", "processing", "partially_done"].includes(active.document.status)) return;
    const timer = window.setInterval(() => void loadActive(active.document.id), 3000);
    return () => window.clearInterval(timer);
  }, [active, loadActive]);

  const stats = useMemo(() => {
    const pages = activePages;
    return {
      done: pages.filter((page) => ["done", "needs_review"].includes(page.status)).length,
      review: pages.filter((page) => page.status === "needs_review").length,
      tables: pages.filter((page) => page.table_count > 0).length,
      text: pages.filter((page) => page.content_type === "free_text").length,
    };
  }, [activePages]);

  async function upload(file: File) {
    if (file.type !== "application/pdf") { setError("Выберите PDF-файл"); return; }
    setBusy(true); setError(""); setUploadProgress(0);
    const request = uploadGeologicalJournalDocument(file, setUploadProgress);
    setUploadAbort(() => request.abort);
    try { const document = await request.promise; setUploadProgress(100); await loadDocuments(); await loadActive(document.id); }
    catch (err) { setError(err instanceof Error ? err.message : "Не удалось загрузить PDF"); }
    finally { setBusy(false); setUploadAbort(null); window.setTimeout(() => setUploadProgress(null), 500); }
  }

  async function removeDocument(document: GeologicalJournalDocument) {
    if (!window.confirm(`Удалить документ «${document.original_name}» и всю историю его анализа?`)) return;
    setBusy(true); setError("");
    try { await deleteGeologicalJournalDocument(document.id); if (active?.document.id === document.id) setActive(null); await loadDocuments(); }
    catch (err) { setError(err instanceof Error ? err.message : "Не удалось удалить документ"); }
    finally { setBusy(false); }
  }

  async function startAnalysis() {
    if (!active) return;
    setBusy(true); setError("");
    try { await startGeologicalJournalDocumentAnalysis(active.document.id); await loadActive(active.document.id); await loadDocuments(); }
    catch (err) { setError(err instanceof Error ? err.message : "Не удалось запустить анализ"); }
    finally { setBusy(false); }
  }

  async function toggleSharing() {
    if (!active) return;
    setBusy(true);
    try { await setGeologicalJournalDocumentSharing(active.document.id, !active.document.is_shared); await loadActive(active.document.id); await loadDocuments(); }
    catch (err) { setError(err instanceof Error ? err.message : "Не удалось изменить доступ"); }
    finally { setBusy(false); }
  }

  async function processWithLLM() {
    if (!active || selectedPages.length === 0) return;
    setBusy(true); setError("");
    try { await processGeologicalJournalDocumentLLM(active.document.id, selectedPages); await loadActive(active.document.id); }
    catch (err) { setError(err instanceof Error ? err.message : "Не удалось обработать выбранные страницы"); }
    finally { setBusy(false); }
  }

  async function askChat() {
    if (!active || selectedPages.length === 0 || !chatInput.trim()) return;
    const question = chatInput.trim();
    setChatInput("");
    setChatMessages((current) => [...current, { role: "user", content: question }]);
    setChatBusy(true);
    try {
      const response = await chatGeologicalJournalDocument(active.document.id, selectedPages, question);
      setChatMessages((current) => [...current, { role: "assistant", content: response.content, confidence: response.confidence }]);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось получить ответ чата");
    } finally {
      setChatBusy(false);
    }
  }

  const progress = active?.document?.page_count ? Math.round((stats.done / active.document.page_count) * 100) : 0;

  return (
    <div className="mx-auto max-w-[1500px] space-y-6">
      <header>
        <h1 className="text-xl font-semibold tracking-tight">Геологические документы</h1>
        <p className="mt-1 text-sm text-text2">PDF анализируется постранично на сервере. LLM запускается только после выбора страниц.</p>
      </header>
      {error && <div role="alert" className="rounded-lg border border-red-200 bg-error-bg px-3 py-2 text-sm text-error">{error}</div>}
      <section className="rounded-xl border border-border bg-bg p-4">
        <label className="flex min-h-[130px] cursor-pointer flex-col items-center justify-center rounded-lg border border-dashed border-border2 bg-bg2/40 text-center hover:border-accent">
          <Upload size={22} className="text-text2" />
          <span className="mt-2 text-sm font-medium">Загрузить PDF-документ</span>
          <span className="mt-1 text-xs text-text3">До 250 МБ. После загрузки нажмите «Запустить анализ».</span>
          <input className="hidden" type="file" accept="application/pdf" disabled={busy} onChange={(event) => { const file = event.target.files?.[0]; if (file) void upload(file); event.target.value = ""; }} />
        </label>
        {uploadProgress !== null && <div className="mt-4 rounded-lg border border-border2 bg-bg2 p-3"><div className="flex items-center justify-between text-xs"><span>Загрузка PDF</span><span>{uploadProgress}%</span></div><div className="mt-2 h-2 overflow-hidden rounded-full bg-border"><div className="h-full bg-accent transition-[width]" style={{ width: `${uploadProgress}%` }} /></div><button type="button" onClick={() => uploadAbort?.()} className="mt-2 inline-flex items-center gap-1 text-xs text-text3 hover:text-error"><X size={13} /> Отменить загрузку</button></div>}
      </section>
      <div className="grid gap-5 lg:grid-cols-[300px_minmax(0,1fr)]">
        <section className="rounded-xl border border-border bg-bg p-3">
          <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-medium">История документов</h2><button type="button" onClick={() => void loadDocuments()} className="rounded p-1.5 text-text3 hover:bg-bg2"><RefreshCw size={15} /></button></div>
          {loading ? <LoaderCircle className="animate-spin text-text3" size={18} /> : documents.length === 0 ? <p className="text-xs text-text3">Документов пока нет.</p> : <div className="space-y-2">{documents.map((document) => <div key={document.id} className={`flex items-start gap-1 rounded-lg border p-2 ${active?.document.id === document.id ? "border-accent bg-accent-bg" : "border-border2"}`}><button type="button" onClick={() => { setSelectedPages([]); void loadActive(document.id); }} className="min-w-0 flex-1 p-1 text-left hover:bg-bg2"><div className="flex items-start gap-2"><FileText size={16} className="mt-0.5 shrink-0" /><span className="min-w-0 flex-1 truncate text-xs font-medium">{document.original_name}</span></div><div className="mt-1 text-[11px] text-text3">{statusLabel(document.status)} · {document.page_count || "?"} стр.</div></button><button type="button" aria-label="Удалить документ" onClick={() => void removeDocument(document)} disabled={busy} className="rounded p-1.5 text-text3 hover:bg-error-bg hover:text-error disabled:opacity-40"><Trash2 size={14} /></button></div>)}</div>}
        </section>
        <section className="min-w-0 rounded-xl border border-border bg-bg p-4">
          {!active ? <div className="flex min-h-[300px] items-center justify-center text-sm text-text3">Выберите документ из истории.</div> : <>
            <div className="flex flex-wrap items-start justify-between gap-3"><div><h2 className="text-base font-semibold">{active.document.original_name}</h2><p className="mt-1 text-xs text-text3">{formatBytes(active.document.size_bytes)} · {statusLabel(active.document.status)}</p></div><div className="flex flex-wrap gap-2"><button type="button" onClick={() => void startAnalysis()} disabled={busy || ["queued", "processing"].includes(active.document.status)} className="inline-flex items-center gap-1.5 rounded-lg bg-text px-3 py-2 text-xs font-medium text-white disabled:opacity-40"><Play size={14} /> Запустить анализ</button><button type="button" onClick={() => void toggleSharing()} disabled={busy} className={`inline-flex items-center gap-1.5 rounded-lg border px-3 py-2 text-xs font-medium ${active.document.is_shared ? "border-success bg-success-bg text-success" : "border-border2 text-text2"}`}><Users size={14} /> {active.document.is_shared ? "Доступно команде журнала" : "Сделать доступным команде"}</button></div></div>
            <div className="mt-5 rounded-lg bg-bg2 p-3"><div className="flex justify-between text-xs"><span>Прогресс страниц</span><span>{stats.done} / {active.document.page_count} · {progress}%</span></div><div className="mt-2 h-2 overflow-hidden rounded-full bg-border"><div className="h-full bg-accent transition-[width]" style={{ width: `${progress}%` }} /></div><div className="mt-2 grid grid-cols-3 gap-2 text-[11px] text-text3"><span>Таблицы: {stats.tables}</span><span>Текст: {stats.text}</span><span>Проверка: {stats.review}</span></div></div>
            <div className="mt-4 flex flex-wrap gap-2"><button type="button" onClick={() => setSelectedPages(activePages.map((page) => page.page_number))} className="rounded border border-border2 px-2 py-1 text-xs">Выбрать все страницы</button><button type="button" onClick={() => setSelectedPages(activePages.filter((page) => page.content_type === "table" || page.content_type === "mixed").map((page) => page.page_number))} className="rounded border border-border2 px-2 py-1 text-xs">Только таблицы</button><button type="button" onClick={() => void processWithLLM()} disabled={busy || selectedPages.length === 0} className="rounded bg-ai px-2 py-1 text-xs font-medium text-white disabled:opacity-40">Обработать выбранные через LLM</button><span className="self-center text-xs text-text3">Выбрано: {selectedPages.length}</span></div>
            <div className={`mt-4 grid gap-4 ${previewPage ? "lg:grid-cols-[minmax(0,1fr)_minmax(300px,0.72fr)]" : "grid-cols-1"}`}>
              <div className="max-h-[520px] overflow-auto rounded-lg border border-border2"><table className="w-full text-left text-xs"><thead className="sticky top-0 bg-bg2"><tr><th className="w-10 px-3 py-2">✓</th><th className="px-3 py-2">Страница</th><th className="px-3 py-2">Тип</th><th className="px-3 py-2">Поворот</th><th className="px-3 py-2">Статус</th><th className="px-3 py-2">Уверенность</th><th className="w-12 px-2 py-2" /></tr></thead><tbody>{activePages.map((page) => <tr key={page.id} className={`border-t border-border ${previewPage?.id === page.id ? "bg-accent-bg" : ""}`}><td className="px-3 py-2"><input type="checkbox" checked={selectedPages.includes(page.page_number)} onChange={(event) => setSelectedPages((current) => event.target.checked ? [...current, page.page_number].sort((a,b) => a-b) : current.filter((number) => number !== page.page_number))} /></td><td className="px-3 py-2 font-medium">{page.page_number}</td><td className="px-3 py-2">{page.content_type || "—"}</td><td className="px-3 py-2">{page.orientation_degrees}°</td><td className="px-3 py-2">{statusLabel(page.status)}</td><td className="px-3 py-2">{Math.round(page.orientation_confidence * 100)}%</td><td className="px-2 py-2"><button type="button" title="Показать страницу" onClick={() => { setPreviewPage(page); setPreviewZoom(1); }} className="rounded p-1.5 text-text3 hover:bg-accent-bg hover:text-accent"><Eye size={15} /></button></td></tr>)}</tbody></table></div>
              {previewPage && <aside className="min-w-0 rounded-lg border border-border2 bg-bg2 p-3"><div className="flex items-center justify-between gap-2"><div><h3 className="text-sm font-medium">Страница {previewPage.page_number}</h3><p className="mt-0.5 text-[11px] text-text3">{statusLabel(previewPage.status)} · поворот {previewPage.orientation_degrees}°</p></div><button type="button" onClick={() => setPreviewPage(null)} className="rounded p-1.5 text-text3 hover:bg-bg hover:text-text" aria-label="Закрыть превью"><X size={15} /></button></div><div className="mt-3 flex items-center justify-end gap-1"><button type="button" onClick={() => setPreviewZoom((value) => Math.max(0.5, value - 0.25))} className="rounded border border-border2 p-1.5 text-text3 hover:bg-bg"><ZoomOut size={14} /></button><span className="min-w-12 text-center text-[11px] text-text3">{Math.round(previewZoom * 100)}%</span><button type="button" onClick={() => setPreviewZoom((value) => Math.min(2.5, value + 0.25))} className="rounded border border-border2 p-1.5 text-text3 hover:bg-bg"><ZoomIn size={14} /></button></div><div className="mt-2 max-h-[430px] overflow-auto rounded border border-border bg-white p-2"><img src={previewImageUrl} alt={`Страница ${previewPage.page_number}`} className="h-auto max-w-none" style={{ width: `${previewZoom * 100}%` }} onError={(event) => { const image = event.currentTarget; if (!image.src.endsWith("/assets/preprocessed")) image.src = image.src.replace("/assets/oriented", "/assets/preprocessed"); else if (!image.src.endsWith("/assets/original")) image.src = image.src.replace("/assets/preprocessed", "/assets/original"); }} /></div></aside>}
            </div>
            <div className="mt-4 rounded-lg border border-border2 p-3"><div className="flex items-center justify-between"><h3 className="text-sm font-medium">Чат по выбранным страницам</h3><span className="text-xs text-text3">Источники: стр. {selectedPages.join(", ") || "—"}</span></div><div className="mt-3 max-h-56 space-y-2 overflow-auto">{chatMessages.length === 0 ? <p className="text-xs text-text3">Выберите страницы и задайте вопрос по распознанным данным.</p> : chatMessages.map((message, index) => <div key={`${message.role}-${index}`} className={`rounded-lg px-3 py-2 text-sm ${message.role === "user" ? "ml-8 bg-accent-bg" : "mr-8 bg-bg2"}`}><p className="whitespace-pre-wrap">{message.content}</p>{message.confidence && <p className="mt-1 text-[11px] text-text3">Уверенность: {message.confidence}</p>}</div>)}</div><div className="mt-3 flex gap-2"><input value={chatInput} onChange={(event) => setChatInput(event.target.value)} onKeyDown={(event) => { if (event.key === "Enter" && !event.shiftKey) { event.preventDefault(); void askChat(); } }} placeholder="Например: какие запасы меди указаны?" className="min-w-0 flex-1 rounded-lg border border-border2 bg-bg px-3 py-2 text-sm outline-none focus:border-accent" disabled={chatBusy || selectedPages.length === 0} /><button type="button" onClick={() => void askChat()} disabled={chatBusy || selectedPages.length === 0 || !chatInput.trim()} className="rounded-lg bg-text px-3 py-2 text-xs font-medium text-white disabled:opacity-40">{chatBusy ? "Ответ…" : "Спросить"}</button></div></div>
          </>}
        </section>
      </div>
    </div>
  );
}