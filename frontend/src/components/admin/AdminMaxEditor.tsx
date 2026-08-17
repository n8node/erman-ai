"use client";

import { Eye, EyeOff, Send } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchAdminMaxSettings,
  fetchAdminMaxStatus,
  sendAdminMaxTest,
  updateAdminMaxSettings,
  type MaxAdminView,
  type MaxBotStatus,
  type MaxRuntimeStatus,
  type MaxSettings,
} from "@/lib/api";
import { cn } from "@/lib/utils";
import { validateMaxBotToken } from "@/lib/bot-token-validation";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const DEFAULT_RUNTIME: MaxRuntimeStatus = {
  status: "misconfigured",
  message: "",
};

const DEFAULT_SETTINGS: MaxSettings = {
  enabled: false,
  bot_token: "",
  bot_username: "id504228241678_bot",
  notify_user_id: "",
  notify_chat_id: "",
  urgent_alerts_enabled: true,
};

function statusDotClass(status: MaxBotStatus): string {
  switch (status) {
    case "online":
      return "bg-[#3b6d11]";
    case "offline":
    case "misconfigured":
      return "bg-[#a32d2d]";
    default:
      return "bg-text3";
  }
}

export function AdminMaxEditor() {
  const t = useTranslations("admin.max");
  const [settings, setSettings] = useState<MaxSettings>(DEFAULT_SETTINGS);
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
  const [runtime, setRuntime] = useState<MaxRuntimeStatus>(DEFAULT_RUNTIME);

  function applyView(data: MaxAdminView) {
    setSettings(data.settings);
    setTokenSet(data.bot_token_set);
    setTokenHint(data.bot_token_hint || "");
    if (data.runtime) setRuntime(data.runtime);
  }

  const refreshStatus = useCallback(async () => {
    try {
      const st = await fetchAdminMaxStatus();
      setRuntime(st);
    } catch {
      /* keep last */
    }
  }, []);

  useEffect(() => {
    fetchAdminMaxSettings()
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

  function patch(partial: Partial<MaxSettings>) {
    setSettings((prev) => ({ ...prev, ...partial }));
    setSuccess("");
  }

  function assertMaxTokenInput(): boolean {
    const code = validateMaxBotToken(tokenInput);
    if (!code) return true;
    setError(t(code));
    return false;
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess("");
    if (!assertMaxTokenInput()) {
      setSaving(false);
      return;
    }
    try {
      const data = await updateAdminMaxSettings({
        settings,
        ...(tokenInput.trim() ? { bot_token: tokenInput.trim() } : {}),
      });
      applyView(data);
      setTokenInput("");
      setSuccess(t("saved"));
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
    if (!assertMaxTokenInput()) {
      setTesting(false);
      return;
    }
    try {
      if (tokenInput.trim()) {
        const saved = await updateAdminMaxSettings({
          settings,
          bot_token: tokenInput.trim(),
        });
        applyView(saved);
        setTokenInput("");
      }
      const result = await sendAdminMaxTest();
      setTestMessage(result.message);
      if (result.runtime) setRuntime(result.runtime);
      else await refreshStatus();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("testFailed"));
    } finally {
      setTesting(false);
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
          </div>
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
          {t("enable")}
        </label>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("botToken")}</label>
          <p className="mb-2 text-xs text-text3">{t("botTokenHint")}</p>
          <div className="relative">
            <input
              type={showToken ? "text" : "password"}
              autoComplete="new-password"
              name="erman-max-bot-token"
              data-1p-ignore
              data-lpignore="true"
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
            >
              {showToken ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          </div>
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("botUsername")}</label>
          <input
            type="text"
            value={settings.bot_username}
            onChange={(e) => patch({ bot_username: e.target.value })}
            className={fieldClass}
            placeholder="id504228241678_bot"
          />
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("notifyUserId")}</label>
          <p className="mb-2 text-xs text-text3">{t("notifyUserIdHint")}</p>
          <input
            type="text"
            value={settings.notify_user_id}
            onChange={(e) => patch({ notify_user_id: e.target.value })}
            className={fieldClass}
            placeholder="123456789"
          />
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("notifyChatId")}</label>
          <p className="mb-2 text-xs text-text3">{t("notifyChatIdHint")}</p>
          <input
            type="text"
            value={settings.notify_chat_id}
            onChange={(e) => patch({ notify_chat_id: e.target.value })}
            className={fieldClass}
            placeholder="-100…"
          />
        </div>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={settings.urgent_alerts_enabled}
            onChange={(e) => patch({ urgent_alerts_enabled: e.target.checked })}
            className="rounded border-border2"
          />
          {t("urgentAlerts")}
        </label>
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
