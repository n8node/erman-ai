"use client";

import { Eye, EyeOff, RefreshCw, Send } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  adminTelegramStartImageUrl,
  fetchAdminTelegramQueue,
  fetchAdminTelegramSettings,
  fetchAdminTelegramStatus,
  restartAdminTelegramBot,
  retryAdminTelegramQueueItem,
  sendAdminTelegramTest,
  type TelegramNotificationRecord,
  type TelegramNotificationStatus,
  updateAdminTelegramSettings,
  uploadAdminTelegramStartImage,
  type TelegramAdminView,
  type TelegramBotStatus,
  type TelegramRuntimeStatus,
  type TelegramSettings,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const templateClass =
  "w-full min-h-[100px] rounded-lg border border-border2 bg-[#e6f1fb]/60 px-3 py-2 text-sm font-mono outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const DEFAULT_RUNTIME: TelegramRuntimeStatus = {
  status: "starting",
  message: "",
  supervisor_running: false,
};
const QUEUE_PAGE_SIZE = 10;

function statusDotClass(status: TelegramBotStatus): string {
  switch (status) {
    case "online":
      return "bg-[#3b6d11]";
    case "offline":
    case "misconfigured":
      return "bg-[#a32d2d]";
    case "starting":
      return "bg-[#ba7517] animate-pulse";
    default:
      return "bg-text3";
  }
}

const DEFAULT_SETTINGS: TelegramSettings = {
  enabled: false,
  chat_id: "",
  start_enabled: false,
  start_text: "",
  support_enabled: false,
  support_forum_chat_id: "",
  dashboard_url: "https://erman.ai/dashboard/",
  urgent_enabled: true,
  urgent_email: "erman.ai@yandex.ru",
  urgent_instruction: "",
  notify_registration: true,
  registration_template: "",
  notify_email_verified: true,
  email_verified_template: "",
  notify_payment: true,
  payment_template: "",
};

export function AdminTelegramEditor() {
  const t = useTranslations("admin.telegram");
  const [settings, setSettings] = useState<TelegramSettings>(DEFAULT_SETTINGS);
  const [tokenInput, setTokenInput] = useState("");
  const [tokenHint, setTokenHint] = useState("");
  const [tokenSet, setTokenSet] = useState(false);
  const [showToken, setShowToken] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [testMessage, setTestMessage] = useState("");
  const [runtime, setRuntime] = useState<TelegramRuntimeStatus>(DEFAULT_RUNTIME);
  const [restarting, setRestarting] = useState(false);
  const [startImageConfigured, setStartImageConfigured] = useState(false);
  const [startTextRunes, setStartTextRunes] = useState(0);
  const [startTextLimit, setStartTextLimit] = useState(4096);
  const [startCaptionLimit, setStartCaptionLimit] = useState(1024);
  const [imageCacheBust, setImageCacheBust] = useState(0);
  const [uploadingImage, setUploadingImage] = useState(false);
  const [queueItems, setQueueItems] = useState<TelegramNotificationRecord[]>([]);
  const [queueTotal, setQueueTotal] = useState(0);
  const [queueOffset, setQueueOffset] = useState(0);
  const [queueStatus, setQueueStatus] = useState<"" | TelegramNotificationStatus>("");
  const [queueLoading, setQueueLoading] = useState(false);
  const [queueError, setQueueError] = useState("");
  const [retryingQueueID, setRetryingQueueID] = useState("");

  function applyView(data: TelegramAdminView) {
    setSettings(data.settings);
    setTokenSet(data.bot_token_set);
    setTokenHint(data.bot_token_hint || "");
    setStartImageConfigured(data.start_image_configured);
    setStartTextRunes(data.start_text_runes);
    setStartTextLimit(data.start_text_limit || 4096);
    setStartCaptionLimit(data.start_caption_limit || 1024);
    if (data.runtime) setRuntime(data.runtime);
  }

  const refreshStatus = useCallback(async () => {
    try {
      const st = await fetchAdminTelegramStatus();
      setRuntime(st);
    } catch {
      /* keep last known status */
    }
  }, []);

  const refreshQueue = useCallback(
    async (next?: { offset?: number; status?: "" | TelegramNotificationStatus }) => {
      const offset = next?.offset ?? queueOffset;
      const status = next?.status ?? queueStatus;
      setQueueLoading(true);
      if (!next) setQueueError("");
      try {
        const data = await fetchAdminTelegramQueue({
          limit: QUEUE_PAGE_SIZE,
          offset,
          ...(status ? { status } : {}),
        });
        setQueueItems(data.items);
        setQueueTotal(data.total);
        setQueueOffset(offset);
        setQueueStatus(status);
      } catch (err) {
        setQueueError(err instanceof Error ? err.message : t("queueLoadFailed"));
      } finally {
        setQueueLoading(false);
      }
    },
    [queueOffset, queueStatus, t]
  );

  useEffect(() => {
    fetchAdminTelegramSettings()
      .then(applyView)
      .catch((err) =>
        setError(err instanceof Error ? err.message : t("loadFailed"))
      )
      .finally(() => setLoading(false));
  }, [t]);

  useEffect(() => {
    if (loading) return;
    void refreshQueue({ offset: 0 });
  }, [loading, refreshQueue]);

  useEffect(() => {
    if (loading) return;
    const id = setInterval(refreshStatus, 5000);
    return () => clearInterval(id);
  }, [loading, refreshStatus]);

  useEffect(() => {
    if (loading) return;
    const id = setInterval(() => {
      void refreshQueue();
    }, 10000);
    return () => clearInterval(id);
  }, [loading, refreshQueue]);

  function patch(partial: Partial<TelegramSettings>) {
    setSettings((prev) => ({ ...prev, ...partial }));
    setSuccess("");
  }

  async function handleImageUpload(file: File) {
    setUploadingImage(true);
    setError("");
    try {
      const data = await uploadAdminTelegramStartImage(file);
      applyView(data);
      setImageCacheBust(Date.now());
      setSuccess(t("imageUploaded"));
      if (data.runtime) setRuntime(data.runtime);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("imageUploadFailed"));
    } finally {
      setUploadingImage(false);
    }
  }

  async function handleRemoveImage() {
    setSaving(true);
    setError("");
    try {
      const data = await updateAdminTelegramSettings({
        settings,
        clear_start_image: true,
        ...(tokenInput.trim() ? { bot_token: tokenInput.trim() } : {}),
      });
      applyView(data);
      setImageCacheBust(0);
      setSuccess(t("imageRemoved"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      const data = await updateAdminTelegramSettings({
        settings,
        ...(tokenInput.trim() ? { bot_token: tokenInput.trim() } : {}),
      });
      applyView(data);
      setTokenInput("");
      setSuccess(t("saved"));
      if (data.runtime) setRuntime(data.runtime);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  async function handleTestSend() {
    setTesting(true);
    setTestMessage("");
    setError("");
    try {
      if (tokenInput.trim()) {
        const saved = await updateAdminTelegramSettings({
          settings,
          bot_token: tokenInput.trim(),
        });
        applyView(saved);
        setTokenInput("");
      }
      const result = await sendAdminTelegramTest();
      setTestMessage(result.message);
      if (result.runtime) setRuntime(result.runtime);
      else await refreshStatus();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("testFailed"));
    } finally {
      setTesting(false);
    }
  }

  async function handleRestart() {
    setRestarting(true);
    setError("");
    try {
      const st = await restartAdminTelegramBot();
      setRuntime(st);
      setSuccess(t("restarted"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("restartFailed"));
    } finally {
      setRestarting(false);
    }
  }

  async function handleQueueRetry(id: string) {
    setRetryingQueueID(id);
    setQueueError("");
    try {
      await retryAdminTelegramQueueItem(id);
      setSuccess(t("queueRetryQueued"));
      await refreshQueue();
    } catch (err) {
      setQueueError(err instanceof Error ? err.message : t("queueRetryFailed"));
    } finally {
      setRetryingQueueID("");
    }
  }

  function queueStatusLabel(status: TelegramNotificationStatus) {
    switch (status) {
      case "pending":
        return t("queueStatusPending");
      case "processing":
        return t("queueStatusProcessing");
      case "sent":
        return t("queueStatusSent");
      case "failed":
        return t("queueStatusFailed");
      default:
        return status;
    }
  }

  function queueStatusBadge(status: TelegramNotificationStatus) {
    switch (status) {
      case "sent":
        return "bg-[#eaf3de] text-[#3b6d11]";
      case "failed":
        return "bg-[#fcebeb] text-[#a32d2d]";
      case "processing":
        return "bg-[#e6f1fb] text-[#185fa5]";
      default:
        return "bg-[#faeeda] text-[#ba7517]";
    }
  }

  const queuePage = Math.floor(queueOffset / QUEUE_PAGE_SIZE) + 1;
  const queuePages = Math.max(1, Math.ceil(queueTotal / QUEUE_PAGE_SIZE));

  const lastCheckLabel =
    runtime.last_check_at &&
    new Date(runtime.last_check_at).toLocaleString(undefined, {
      dateStyle: "short",
      timeStyle: "short",
    });

  if (loading) {
    return <p className="text-sm text-text2">{t("loading")}</p>;
  }

  return (
    <div className="space-y-6">
      <section className="rounded-xl border border-border bg-bg p-5">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="flex items-start gap-3">
            <span
              className={cn(
                "mt-1.5 h-2.5 w-2.5 shrink-0 rounded-full",
                statusDotClass(runtime.status)
              )}
              aria-hidden
            />
            <div>
              <p className="text-sm font-medium">{t("runtimeTitle")}</p>
              <p className="mt-1 text-sm text-text2">{runtime.message}</p>
              {runtime.bot_username && (
                <p className="mt-1 text-xs text-text3">@{runtime.bot_username}</p>
              )}
              {runtime.last_error &&
                (runtime.status === "offline" ||
                  runtime.status === "misconfigured") && (
                  <p className="mt-2 text-xs text-red-800">{runtime.last_error}</p>
                )}
              {lastCheckLabel && (
                <p className="mt-2 text-xs text-text3">
                  {t("lastCheck", { time: lastCheckLabel })}
                </p>
              )}
              {!runtime.supervisor_running && (
                <p className="mt-1 text-xs text-amber-800">{t("supervisorStopped")}</p>
              )}
            </div>
          </div>
          <button
            type="button"
            disabled={restarting}
            onClick={handleRestart}
            className="inline-flex items-center gap-2 rounded-lg border border-border2 px-3 py-2 text-sm hover:bg-bg2 disabled:opacity-60"
          >
            <RefreshCw className={cn("h-4 w-4", restarting && "animate-spin")} />
            {restarting ? t("restarting") : t("restart")}
          </button>
        </div>
        <p className="mt-3 text-xs text-text3">{t("runtimeHint")}</p>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-sm font-medium">{t("queueTitle")}</h2>
            <p className="text-xs text-text3">{t("queueHint")}</p>
          </div>
          <div className="flex items-center gap-2">
            <select
              value={queueStatus}
              onChange={(e) =>
                void refreshQueue({
                  status: e.target.value as "" | TelegramNotificationStatus,
                  offset: 0,
                })
              }
              className="rounded-lg border border-border2 px-2 py-1.5 text-xs"
            >
              <option value="">{t("queueStatusAll")}</option>
              <option value="pending">{t("queueStatusPending")}</option>
              <option value="processing">{t("queueStatusProcessing")}</option>
              <option value="sent">{t("queueStatusSent")}</option>
              <option value="failed">{t("queueStatusFailed")}</option>
            </select>
            <button
              type="button"
              onClick={() => void refreshQueue()}
              className="rounded-lg border border-border2 px-3 py-1.5 text-xs hover:bg-bg2"
            >
              {t("queueRefresh")}
            </button>
          </div>
        </div>

        {queueError && (
          <p className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-800">
            {queueError}
          </p>
        )}

        <div className="overflow-x-auto rounded-lg border border-border">
          <table className="min-w-full text-sm">
            <thead className="bg-bg2 text-xs uppercase tracking-wide text-text3">
              <tr>
                <th className="px-3 py-2 text-left">{t("queueColCreated")}</th>
                <th className="px-3 py-2 text-left">{t("queueColKind")}</th>
                <th className="px-3 py-2 text-left">{t("queueColStatus")}</th>
                <th className="px-3 py-2 text-left">{t("queueColAttempts")}</th>
                <th className="px-3 py-2 text-left">{t("queueColError")}</th>
                <th className="px-3 py-2 text-right">{t("queueColActions")}</th>
              </tr>
            </thead>
            <tbody>
              {queueItems.map((item) => (
                <tr key={item.id} className="border-t border-border">
                  <td className="px-3 py-2 text-xs text-text2">
                    {new Date(item.created_at).toLocaleString()}
                  </td>
                  <td className="px-3 py-2 font-mono text-xs">{item.kind}</td>
                  <td className="px-3 py-2">
                    <span
                      className={cn(
                        "rounded-full px-2 py-1 text-xs font-medium",
                        queueStatusBadge(item.status)
                      )}
                    >
                      {queueStatusLabel(item.status)}
                    </span>
                  </td>
                  <td className="px-3 py-2 text-xs text-text2">{item.attempt_count}</td>
                  <td className="max-w-[320px] px-3 py-2 text-xs text-red-800">
                    {item.last_error || "—"}
                  </td>
                  <td className="px-3 py-2 text-right">
                    {item.status === "failed" && (
                      <button
                        type="button"
                        disabled={retryingQueueID === item.id}
                        onClick={() => void handleQueueRetry(item.id)}
                        className="rounded-md border border-border2 px-2 py-1 text-xs hover:bg-bg2 disabled:opacity-60"
                      >
                        {retryingQueueID === item.id ? t("queueRetrying") : t("queueRetry")}
                      </button>
                    )}
                  </td>
                </tr>
              ))}
              {!queueLoading && queueItems.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-3 py-6 text-center text-sm text-text3">
                    {t("queueEmpty")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <div className="flex items-center justify-between text-xs text-text3">
          <span>{t("queuePagination", { page: queuePage, total: queuePages, count: queueTotal })}</span>
          <div className="flex gap-2">
            <button
              type="button"
              disabled={queueOffset <= 0 || queueLoading}
              onClick={() => void refreshQueue({ offset: Math.max(0, queueOffset - QUEUE_PAGE_SIZE) })}
              className="rounded border border-border2 px-2 py-1 disabled:opacity-50"
            >
              {t("queuePrev")}
            </button>
            <button
              type="button"
              disabled={queueOffset + QUEUE_PAGE_SIZE >= queueTotal || queueLoading}
              onClick={() => void refreshQueue({ offset: queueOffset + QUEUE_PAGE_SIZE })}
              className="rounded border border-border2 px-2 py-1 disabled:opacity-50"
            >
              {t("queueNext")}
            </button>
          </div>
        </div>
      </section>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}
      {success && (
        <div className="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-900">
          {success}
        </div>
      )}

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={settings.enabled}
            onChange={(e) => patch({ enabled: e.target.checked })}
            className="rounded border-border2"
          />
          {t("enableNotifications")}
        </label>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("botToken")}</label>
          <p className="mb-2 text-xs text-text3">{t("botTokenHint")}</p>
          <div className="relative">
            <input
              type={showToken ? "text" : "password"}
              autoComplete="off"
              value={tokenInput}
              onChange={(e) => setTokenInput(e.target.value)}
              placeholder={
                tokenSet
                  ? t("tokenPlaceholderExisting", { hint: tokenHint || "••••" })
                  : t("tokenPlaceholder")
              }
              className={cn(fieldClass, "pr-10")}
            />
            <button
              type="button"
              onClick={() => setShowToken((v) => !v)}
              className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-text3 hover:text-text"
              aria-label={showToken ? t("hideToken") : t("showToken")}
            >
              {showToken ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          </div>
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("chatId")}</label>
          <p className="mb-2 text-xs text-text3">{t("chatIdHint")}</p>
          <input
            type="text"
            value={settings.chat_id}
            onChange={(e) => patch({ chat_id: e.target.value })}
            className={fieldClass}
            placeholder="639160984"
          />
        </div>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-xs font-medium uppercase tracking-wide text-text3">
          {t("startSection")}
        </h2>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={settings.start_enabled}
            onChange={(e) => patch({ start_enabled: e.target.checked })}
            className="rounded border-border2"
          />
          {t("startEnabled")}
        </label>
        <p className="text-xs text-text3">{t("startHint")}</p>
        <div>
          <div className="mb-1.5 flex items-center justify-between gap-2">
            <label className="text-xs font-medium">{t("startText")}</label>
            <span
              className={cn(
                "text-xs",
                startTextRunes >
                  (startImageConfigured ? startCaptionLimit : startTextLimit)
                  ? "text-red-800"
                  : "text-text3"
              )}
            >
              {startTextRunes} /{" "}
              {startImageConfigured ? startCaptionLimit : startTextLimit}
              {startImageConfigured && startTextRunes > startCaptionLimit
                ? ` (+ ${t("startOverflowHint")})`
                : ""}
            </span>
          </div>
          <textarea
            value={settings.start_text}
            onChange={(e) => {
              patch({ start_text: e.target.value });
              setStartTextRunes([...e.target.value].length);
            }}
            className={templateClass}
            rows={6}
          />
        </div>
        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("startImage")}</label>
          <p className="mb-2 text-xs text-text3">{t("startImageHint")}</p>
          {startImageConfigured && (
            <div className="mb-3 overflow-hidden rounded-lg border border-border">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={adminTelegramStartImageUrl(imageCacheBust)}
                alt=""
                className="max-h-48 w-full object-contain bg-bg2"
              />
            </div>
          )}
          <div className="flex flex-wrap gap-2">
            <label className="inline-flex cursor-pointer items-center rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2">
              <input
                type="file"
                accept="image/jpeg,image/png,image/webp"
                className="sr-only"
                disabled={uploadingImage}
                onChange={(e) => {
                  const f = e.target.files?.[0];
                  if (f) void handleImageUpload(f);
                  e.target.value = "";
                }}
              />
              {uploadingImage ? t("imageUploading") : t("startImageUpload")}
            </label>
            {startImageConfigured && (
              <button
                type="button"
                disabled={saving}
                onClick={handleRemoveImage}
                className="rounded-lg border border-border2 px-4 py-2 text-sm text-text2 hover:bg-bg2 disabled:opacity-60"
              >
                {t("startImageRemove")}
              </button>
            )}
          </div>
        </div>
        {runtime.polling_running && (
          <p className="text-xs text-green-800">{t("startPollingActive")}</p>
        )}
        <p className="text-xs text-text3">{t("startButtonsHint")}</p>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-xs font-medium uppercase tracking-wide text-text3">
          {t("urgentSection")}
        </h2>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={settings.urgent_enabled}
            onChange={(e) => patch({ urgent_enabled: e.target.checked })}
            className="rounded border-border2"
          />
          {t("urgentEnabled")}
        </label>
        <p className="text-xs text-text3">{t("urgentHint")}</p>
        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("urgentEmail")}</label>
          <input
            type="email"
            value={settings.urgent_email}
            onChange={(e) => patch({ urgent_email: e.target.value })}
            className={fieldClass}
            placeholder="erman.ai@yandex.ru"
          />
        </div>
        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("urgentInstruction")}</label>
          <textarea
            value={settings.urgent_instruction}
            onChange={(e) => patch({ urgent_instruction: e.target.value })}
            className={templateClass}
            rows={8}
          />
        </div>
        <p className="text-xs text-text3">{t("urgentLimitHint")}</p>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-xs font-medium uppercase tracking-wide text-text3">
          {t("supportSection")}
        </h2>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={settings.support_enabled}
            onChange={(e) => patch({ support_enabled: e.target.checked })}
            className="rounded border-border2"
          />
          {t("supportEnabled")}
        </label>
        <p className="text-xs text-text3">{t("supportHint")}</p>
        <div>
          <label className="mb-1.5 block text-xs font-medium">
            {t("supportForumChatId")}
          </label>
          <p className="mb-2 text-xs text-text3">{t("supportForumChatIdHint")}</p>
          <input
            type="text"
            value={settings.support_forum_chat_id}
            onChange={(e) => patch({ support_forum_chat_id: e.target.value })}
            className={fieldClass}
            placeholder="-1001234567890"
          />
        </div>
        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("dashboardUrl")}</label>
          <p className="mb-2 text-xs text-text3">{t("dashboardUrlHint")}</p>
          <input
            type="url"
            value={settings.dashboard_url}
            onChange={(e) => patch({ dashboard_url: e.target.value })}
            className={fieldClass}
            placeholder="https://erman.ai/dashboard/"
          />
        </div>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-3">
        <label className="flex items-center gap-2 text-sm font-medium">
          <input
            type="checkbox"
            checked={settings.notify_registration}
            onChange={(e) => patch({ notify_registration: e.target.checked })}
            className="rounded border-border2"
          />
          {t("notifyRegistration")}
        </label>
        <p className="text-xs text-text3">{t("registrationVars")}</p>
        <textarea
          value={settings.registration_template}
          onChange={(e) => patch({ registration_template: e.target.value })}
          className={templateClass}
          rows={5}
        />
        <p className="text-xs text-text3">{t("emailVerifiedVars")}</p>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={settings.notify_email_verified}
            onChange={(e) => patch({ notify_email_verified: e.target.checked })}
            className="rounded border-border2"
          />
          {t("notifyEmailVerified")}
        </label>
        <textarea
          value={settings.email_verified_template}
          onChange={(e) => patch({ email_verified_template: e.target.value })}
          className={templateClass}
          rows={4}
        />
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-3">
        <label className="flex items-center gap-2 text-sm font-medium">
          <input
            type="checkbox"
            checked={settings.notify_payment}
            onChange={(e) => patch({ notify_payment: e.target.checked })}
            className="rounded border-border2"
          />
          {t("notifyPayment")}
        </label>
        <p className="text-xs text-text3">{t("paymentVars")}</p>
        <textarea
          value={settings.payment_template}
          onChange={(e) => patch({ payment_template: e.target.value })}
          className={templateClass}
          rows={5}
        />
      </section>

      <section className="rounded-xl border border-accent/30 bg-[#e6f1fb]/40 p-5 space-y-3">
        <div className="flex items-center gap-2">
          <Send className="h-4 w-4 text-text" />
          <h2 className="text-sm font-medium">{t("verifySection")}</h2>
        </div>
        <p className="text-sm text-accent">{t("verifyHint")}</p>
        <button
          type="button"
          disabled={testing}
          onClick={handleTestSend}
          className="inline-flex items-center gap-2 rounded-lg border border-accent px-4 py-2 text-sm text-accent hover:bg-[#e6f1fb] disabled:opacity-60"
        >
          <Send className="h-4 w-4" />
          {testing ? t("testing") : t("testSend")}
        </button>
        {testMessage && (
          <p
            className={cn(
              "text-sm",
              testMessage.includes("отправлен") ? "text-green-800" : "text-text2"
            )}
          >
            {testMessage}
          </p>
        )}
      </section>

      <button
        type="button"
        disabled={saving}
        onClick={handleSave}
        className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-60"
      >
        {saving ? t("saving") : t("save")}
      </button>
    </div>
  );
}
