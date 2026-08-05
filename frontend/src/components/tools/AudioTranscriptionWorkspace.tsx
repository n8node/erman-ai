"use client";

import {
  AlertCircle,
  Check,
  Download,
  FileAudio,
  LoaderCircle,
  Trash2,
  Upload,
} from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { useEffect, useRef, useState } from "react";
import { intlLocale } from "@/i18n/intl-locale";
import {
  audioTranscriptionDownloadUrl,
  deleteAudioTranscriptionFile,
  getAudioTranscriptionFile,
  getAudioTranscriptionRun,
  listAudioTranscriptionFiles,
  uploadAudioTranscriptionFile,
  validateAudioTranscriptionFile,
  type AudioTranscriptionFile,
  type AudioTranscriptionOutput,
  type AudioTranscriptionRun,
} from "@/lib/api-audio-transcription";
import { ApiError } from "@/lib/api";
import { cn } from "@/lib/utils";

function parseTranscriptionOutput(raw: unknown): AudioTranscriptionOutput | null {
  if (!raw) return null;
  if (typeof raw === "string") {
    try {
      return parseTranscriptionOutput(JSON.parse(raw));
    } catch {
      return { text: raw, language: "", model: "", duration_sec: 0, chunk_count: 0 };
    }
  }
  if (typeof raw === "object" && raw !== null && "text" in raw) {
    const value = raw as AudioTranscriptionOutput;
    return {
      text: value.text ?? "",
      language: value.language ?? "",
      model: value.model ?? "",
      duration_sec: value.duration_sec ?? 0,
      chunk_count: value.chunk_count ?? 0,
      preview: value.preview,
    };
  }
  return null;
}

function formatBytes(bytes: number) {
  if (!bytes) return "—";
  if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function formatDuration(sec?: number | null) {
  if (!sec || sec <= 0) return "—";
  const m = Math.floor(sec / 60);
  const s = Math.round(sec % 60);
  return `${m}:${String(s).padStart(2, "0")}`;
}

export function AudioTranscriptionWorkspace() {
  const t = useTranslations("audioTranscription");
  const locale = useLocale();
  const [files, setFiles] = useState<AudioTranscriptionFile[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [output, setOutput] = useState<AudioTranscriptionOutput | null>(null);
  const [run, setRun] = useState<AudioTranscriptionRun | null>(null);
  const [loading, setLoading] = useState(true);
  const [uploadPercent, setUploadPercent] = useState<number | null>(null);
  const [dragActive, setDragActive] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [deleting, setDeleting] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    let cancelled = false;
    listAudioTranscriptionFiles()
      .then((data) => {
        if (!cancelled) setFiles(Array.isArray(data.items) ? data.items : []);
      })
      .catch((err) => {
        if (cancelled) return;
        if (err instanceof ApiError && err.status === 403) {
          setError(t("errors.accessDenied"));
          return;
        }
        setError(err instanceof Error ? err.message : t("loadFailed"));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- run once on mount
  }, []);

  useEffect(() => {
    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, []);

  async function reloadFiles() {
    const data = await listAudioTranscriptionFiles();
    setFiles(Array.isArray(data.items) ? data.items : []);
  }

  async function selectFile(id: string) {
    setSelectedId(id);
    setError("");
    setSuccess("");
    setOutput(null);
    setRun(null);
    try {
      const detail = await getAudioTranscriptionFile(id);
      const latest = detail.runs?.[0];
      if (latest?.status === "done") {
        const parsed = parseTranscriptionOutput(latest.output);
        if (parsed) {
          setOutput(parsed);
          setRun(latest as AudioTranscriptionRun);
        }
      } else if (latest && (latest.status === "pending" || latest.status === "processing")) {
        setRun(latest);
        startPolling(latest.id, id);
      } else if (latest?.status === "error") {
        setRun(latest);
        setError(latest.error_msg || t("transcriptionFailed"));
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t("loadFailed"));
    }
  }

  function startPolling(runId: string, fileId: string) {
    if (pollRef.current) clearInterval(pollRef.current);
    pollRef.current = setInterval(async () => {
      try {
        const current = await getAudioTranscriptionRun(runId);
        setRun(current);
        if (current.status === "done") {
          if (pollRef.current) clearInterval(pollRef.current);
          setOutput(parseTranscriptionOutput(current.output));
          setSuccess(t("transcriptionDone"));
          await reloadFiles();
          await selectFile(fileId);
        } else if (current.status === "error") {
          if (pollRef.current) clearInterval(pollRef.current);
          setError(current.error_msg || t("transcriptionFailed"));
          await reloadFiles();
        }
      } catch {
        // keep polling
      }
    }, 3000);
  }

  async function handleUpload(file: File) {
    setError("");
    setSuccess("");
    const validation = validateAudioTranscriptionFile(file);
    if (validation === "file_too_large") {
      setError(t("errors.fileTooLarge"));
      return;
    }
    if (validation === "invalid_type") {
      setError(t("errors.invalidType"));
      return;
    }
    setUploadPercent(0);
    const { promise } = uploadAudioTranscriptionFile(file, setUploadPercent);
    try {
      const result = await promise;
      setUploadPercent(null);
      await reloadFiles();
      setSelectedId(result.file.id);
      setSuccess(t("uploadStarted"));
      startPolling(result.run_id, result.file.id);
    } catch (err) {
      setUploadPercent(null);
      setError(err instanceof Error ? err.message : t("uploadFailed"));
    }
  }

  function onDrop(e: React.DragEvent) {
    e.preventDefault();
    setDragActive(false);
    const file = e.dataTransfer.files[0];
    if (file) void handleUpload(file);
  }

  async function handleDelete() {
    if (!selectedId || !confirm(t("deleteConfirm"))) return;
    setDeleting(true);
    setError("");
    try {
      await deleteAudioTranscriptionFile(selectedId);
      setSelectedId(null);
      setOutput(null);
      setRun(null);
      await reloadFiles();
      setSuccess(t("deleted"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("deleteFailed"));
    } finally {
      setDeleting(false);
    }
  }

  const selected = (files ?? []).find((f) => f.id === selectedId) || null;
  const isProcessing =
    run?.status === "pending" || run?.status === "processing";

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </div>

      {error && (
        <div className="flex items-start gap-2 rounded-lg border border-error/20 bg-[#fcebeb] px-3 py-2 text-sm text-[#a32d2d]">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}
      {success && (
        <div className="flex items-start gap-2 rounded-lg border border-success/20 bg-[#eaf3de] px-3 py-2 text-sm text-[#3b6d11]">
          <Check className="mt-0.5 h-4 w-4 shrink-0" />
          <span>{success}</span>
        </div>
      )}

      <div
        className={cn(
          "rounded-xl border border-dashed bg-bg p-8 text-center transition-colors",
          dragActive ? "border-accent bg-[#e6f1fb]" : "border-border2"
        )}
        onDragOver={(e) => {
          e.preventDefault();
          setDragActive(true);
        }}
        onDragLeave={() => setDragActive(false)}
        onDrop={onDrop}
      >
        <FileAudio className="mx-auto h-8 w-8 text-text3" />
        <p className="mt-3 text-sm font-medium text-text">{t("dropTitle")}</p>
        <p className="mt-1 text-xs text-text3">{t("dropHint")}</p>
        <button
          type="button"
          onClick={() => inputRef.current?.click()}
          className="mt-4 inline-flex items-center gap-2 rounded-lg bg-text px-4 py-2 text-sm text-white"
        >
          <Upload className="h-4 w-4" />
          {t("chooseFile")}
        </button>
        <input
          ref={inputRef}
          type="file"
          accept="audio/*,video/mp4,video/webm,.mp3,.wav,.ogg,.m4a,.webm"
          className="hidden"
          onChange={(e) => {
            const file = e.target.files?.[0];
            if (file) void handleUpload(file);
            e.target.value = "";
          }}
        />
        {uploadPercent !== null && (
          <div className="mx-auto mt-4 max-w-xs">
            <div className="h-2 overflow-hidden rounded-full bg-bg3">
              <div
                className="h-full bg-accent transition-all"
                style={{ width: `${uploadPercent}%` }}
              />
            </div>
            <p className="mt-2 text-xs text-text3">{t("uploading", { percent: uploadPercent })}</p>
          </div>
        )}
      </div>

      <div className="grid gap-6 lg:grid-cols-[280px_minmax(0,1fr)]">
        <div className="rounded-xl border border-border bg-bg">
          <div className="border-b border-border px-4 py-3 text-[10px] uppercase tracking-wider text-text3">
            {t("history")}
          </div>
          {loading ? (
            <p className="px-4 py-6 text-sm text-text2">{t("loading")}</p>
          ) : (files ?? []).length === 0 ? (
            <p className="px-4 py-6 text-sm text-text2">{t("empty")}</p>
          ) : (
            <ul className="max-h-[420px] overflow-y-auto">
              {(files ?? []).map((file) => (
                <li key={file.id}>
                  <button
                    type="button"
                    onClick={() => void selectFile(file.id)}
                    className={cn(
                      "w-full border-b border-border px-4 py-3 text-left last:border-0 hover:bg-bg2",
                      selectedId === file.id && "bg-bg2"
                    )}
                  >
                    <p className="truncate text-sm font-medium text-text">
                      {file.original_name}
                    </p>
                    <p className="mt-1 text-xs text-text3">
                      {formatBytes(file.size_bytes)} ·{" "}
                      {new Date(file.created_at).toLocaleString(intlLocale(locale))}
                    </p>
                    {file.latest_run_status && (
                      <p className="mt-1 text-xs text-text2">{file.latest_run_status}</p>
                    )}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="rounded-xl border border-border bg-bg p-5">
          {!selected ? (
            <p className="text-sm text-text2">{t("selectFile")}</p>
          ) : (
            <div className="space-y-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h2 className="text-sm font-medium text-text">{selected.original_name}</h2>
                  <p className="mt-1 text-xs text-text3">
                    {formatBytes(selected.size_bytes)} · {formatDuration(selected.duration_sec)}
                  </p>
                </div>
                <div className="flex gap-2">
                  {selected.has_transcript && (
                    <a
                      href={audioTranscriptionDownloadUrl(selected.id)}
                      className="inline-flex items-center gap-1 rounded-lg border border-border2 px-3 py-1.5 text-sm hover:bg-bg2"
                    >
                      <Download className="h-4 w-4" />
                      {t("downloadTxt")}
                    </a>
                  )}
                  <button
                    type="button"
                    disabled={deleting}
                    onClick={() => void handleDelete()}
                    className="inline-flex items-center gap-1 rounded-lg border border-border2 px-3 py-1.5 text-sm text-[#a32d2d] hover:bg-[#fcebeb] disabled:opacity-50"
                  >
                    <Trash2 className="h-4 w-4" />
                    {t("delete")}
                  </button>
                </div>
              </div>

              {isProcessing && (
                <div className="flex items-center gap-2 rounded-lg bg-bg2 px-3 py-2 text-sm text-text2">
                  <LoaderCircle className="h-4 w-4 animate-spin" />
                  {t("processing")}
                </div>
              )}

              {output && (
                <div className="space-y-2">
                  <div className="text-[10px] uppercase tracking-wider text-text3">
                    {t("result")}
                  </div>
                  <div className="max-h-[420px] overflow-y-auto rounded-lg border border-border bg-bg3 p-4 text-sm leading-relaxed text-text whitespace-pre-wrap">
                    {output.text}
                  </div>
                  <p className="text-xs text-text3">
                    {t("meta", {
                      chunks: output.chunk_count ?? 0,
                      model: output.model || "—",
                      language: output.language || "—",
                    })}
                  </p>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
