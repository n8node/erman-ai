"use client";

import { Eye, EyeOff, RefreshCw, Send } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchAdminTelegramSettings,
  fetchAdminTelegramStatus,
  restartAdminTelegramBot,
  sendAdminTelegramTest,
  updateAdminTelegramSettings,
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

  function applyView(data: TelegramAdminView) {
    setSettings(data.settings);
    setTokenSet(data.bot_token_set);
    setTokenHint(data.bot_token_hint || "");
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
    const id = setInterval(refreshStatus, 5000);
    return () => clearInterval(id);
  }, [loading, refreshStatus]);

  function patch(partial: Partial<TelegramSettings>) {
    setSettings((prev) => ({ ...prev, ...partial }));
    setSuccess("");
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
