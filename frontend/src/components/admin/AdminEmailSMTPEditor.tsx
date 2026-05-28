"use client";

import { Send } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchAdminSMTPSettings,
  sendAdminSMTPTest,
  updateAdminSMTPSettings,
  type SMTPAdminView,
  type SMTPEncryption,
  type SMTPSettings,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const DEFAULT_SETTINGS: SMTPSettings = {
  enabled: false,
  from_email: "",
  from_name: "Erman AI",
  force_from_email: true,
  force_from_name: true,
  reply_to_from_email: false,
  host: "",
  port: 465,
  encryption: "ssl",
  auto_tls: true,
  auth: true,
  username: "",
};

export function AdminEmailSMTPEditor() {
  const t = useTranslations("admin.emailSmtp");
  const [settings, setSettings] = useState<SMTPSettings>(DEFAULT_SETTINGS);
  const [passwordInput, setPasswordInput] = useState("");
  const [passwordHint, setPasswordHint] = useState("");
  const [passwordSet, setPasswordSet] = useState(false);
  const [yandexHost, setYandexHost] = useState("smtp.yandex.ru");
  const [yandexPort, setYandexPort] = useState(465);
  const [testTo, setTestTo] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [testMessage, setTestMessage] = useState("");

  function applyView(data: SMTPAdminView) {
    setSettings(data.settings);
    setPasswordSet(data.password_set);
    setPasswordHint(data.password_hint || "");
    setYandexHost(data.yandex_preset_host);
    setYandexPort(data.yandex_preset_port);
  }

  useEffect(() => {
    fetchAdminSMTPSettings()
      .then(applyView)
      .catch((err) =>
        setError(err instanceof Error ? err.message : t("loadFailed"))
      )
      .finally(() => setLoading(false));
  }, [t]);

  function patch(partial: Partial<SMTPSettings>) {
    setSettings((prev) => ({ ...prev, ...partial }));
    setSuccess("");
  }

  function applyYandexPreset() {
    patch({
      host: yandexHost,
      port: yandexPort,
      encryption: "ssl",
      auto_tls: true,
      auth: true,
    });
    setSuccess("");
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      const data = await updateAdminSMTPSettings({
        settings,
        ...(passwordInput.trim() ? { password: passwordInput.trim() } : {}),
      });
      applyView(data);
      setPasswordInput("");
      setSuccess(t("saved"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  async function handleTestSend() {
    if (!testTo.trim()) return;
    setTesting(true);
    setTestMessage("");
    setError("");
    try {
      if (passwordInput.trim()) {
        const saved = await updateAdminSMTPSettings({
          settings,
          password: passwordInput.trim(),
        });
        applyView(saved);
        setPasswordInput("");
      }
      const result = await sendAdminSMTPTest(testTo.trim());
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

      <section className="rounded-xl border border-border bg-bg p-5">
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={settings.enabled}
            onChange={(e) => patch({ enabled: e.target.checked })}
            className="rounded border-border2"
          />
          {t("enableSending")}
        </label>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-3">
        <h2 className="text-xs font-medium uppercase tracking-wide text-text3">
          {t("presetSection")}
        </h2>
        <button
          type="button"
          onClick={applyYandexPreset}
          className="rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2"
        >
          {t("yandexPreset")}
        </button>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-xs font-medium uppercase tracking-wide text-text3">
          {t("senderSection")}
        </h2>
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("fromEmail")}</label>
            <input
              type="email"
              value={settings.from_email}
              onChange={(e) => patch({ from_email: e.target.value })}
              className={fieldClass}
            />
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("fromName")}</label>
            <input
              type="text"
              value={settings.from_name}
              onChange={(e) => patch({ from_name: e.target.value })}
              className={fieldClass}
            />
          </div>
        </div>
        <div className="flex flex-col gap-2 text-sm">
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={settings.force_from_email}
              onChange={(e) => patch({ force_from_email: e.target.checked })}
            />
            {t("forceFromEmail")}
          </label>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={settings.force_from_name}
              onChange={(e) => patch({ force_from_name: e.target.checked })}
            />
            {t("forceFromName")}
          </label>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={settings.reply_to_from_email}
              onChange={(e) => patch({ reply_to_from_email: e.target.checked })}
            />
            {t("replyToFromEmail")}
          </label>
        </div>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-xs font-medium uppercase tracking-wide text-text3">
          {t("connectionSection")}
        </h2>
        <div className="grid gap-4 sm:grid-cols-[1fr_120px]">
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("host")}</label>
            <input
              type="text"
              value={settings.host}
              onChange={(e) => patch({ host: e.target.value })}
              className={fieldClass}
            />
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("port")}</label>
            <input
              type="number"
              min={1}
              max={65535}
              value={settings.port}
              onChange={(e) => patch({ port: Number(e.target.value) || 465 })}
              className={fieldClass}
            />
          </div>
        </div>
        <div>
          <span className="mb-2 block text-xs font-medium">{t("encryption")}</span>
          <div className="flex gap-4 text-sm">
            {(["none", "ssl", "tls"] as SMTPEncryption[]).map((enc) => (
              <label key={enc} className="flex items-center gap-1.5 uppercase">
                <input
                  type="radio"
                  name="encryption"
                  checked={settings.encryption === enc}
                  onChange={() => patch({ encryption: enc })}
                />
                {enc}
              </label>
            ))}
          </div>
        </div>
        <div className="flex flex-col gap-2 text-sm">
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={settings.auto_tls}
              onChange={(e) => patch({ auto_tls: e.target.checked })}
            />
            {t("autoTls")}
          </label>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={settings.auth}
              onChange={(e) => patch({ auth: e.target.checked })}
            />
            {t("auth")}
          </label>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("username")}</label>
            <input
              type="text"
              autoComplete="off"
              value={settings.username}
              onChange={(e) => patch({ username: e.target.value })}
              className={fieldClass}
            />
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("password")}</label>
            <input
              type="password"
              autoComplete="new-password"
              value={passwordInput}
              onChange={(e) => setPasswordInput(e.target.value)}
              placeholder={
                passwordSet
                  ? t("passwordPlaceholderExisting", { hint: passwordHint || "••••" })
                  : t("passwordPlaceholder")
              }
              className={fieldClass}
            />
          </div>
        </div>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-3">
        <h2 className="text-xs font-medium uppercase tracking-wide text-text3">
          {t("testSection")}
        </h2>
        <div className="flex flex-wrap gap-2">
          <input
            type="email"
            value={testTo}
            onChange={(e) => setTestTo(e.target.value)}
            placeholder={t("testEmailPlaceholder")}
            className={cn(fieldClass, "min-w-[200px] flex-1")}
          />
          <button
            type="button"
            disabled={testing || !testTo.trim()}
            onClick={handleTestSend}
            className="inline-flex items-center gap-2 rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2 disabled:opacity-60"
          >
            <Send className="h-4 w-4" />
            {testing ? t("testing") : t("testSend")}
          </button>
        </div>
        {testMessage && (
          <p className={cn("text-sm", testMessage.includes("отправлено") || testMessage.toLowerCase().includes("sent") ? "text-green-800" : "text-text2")}>
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
