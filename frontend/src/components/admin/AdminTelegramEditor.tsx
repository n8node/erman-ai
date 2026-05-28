"use client";

import { Eye, EyeOff, Send } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchAdminTelegramSettings,
  sendAdminTelegramTest,
  updateAdminTelegramSettings,
  type TelegramAdminView,
  type TelegramSettings,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const templateClass =
  "w-full min-h-[100px] rounded-lg border border-border2 bg-[#e6f1fb]/60 px-3 py-2 text-sm font-mono outline-none focus:border-accent focus:ring-1 focus:ring-accent";

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

  function applyView(data: TelegramAdminView) {
    setSettings(data.settings);
    setTokenSet(data.bot_token_set);
    setTokenHint(data.bot_token_hint || "");
  }

  useEffect(() => {
    fetchAdminTelegramSettings()
      .then(applyView)
      .catch((err) =>
        setError(err instanceof Error ? err.message : t("loadFailed"))
      )
      .finally(() => setLoading(false));
  }, [t]);

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
    } catch (err) {
      setError(err instanceof Error ? err.message : t("testFailed"));
    } finally {
      setTesting(false);
    }
  }

  if (loading) {
    return <p className="text-sm text-text2">{t("loading")}</p>;
  }

  return (
    <div className="space-y-6">
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
