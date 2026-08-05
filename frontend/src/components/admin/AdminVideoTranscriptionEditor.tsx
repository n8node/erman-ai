"use client";

import { AlertCircle, Check, LoaderCircle, Save, Settings2, Users } from "lucide-react";
import { useTranslations } from "next-intl";
import { useCallback, useEffect, useState } from "react";
import { cn } from "@/lib/utils";
import {
  fetchAdminVideoTranscriptionAccess,
  fetchAdminVideoTranscriptionSettings,
  updateAdminVideoTranscriptionAccess,
  updateAdminVideoTranscriptionSettings,
  type VideoTranscriptionAccessUser,
  type VideoTranscriptionSettings,
} from "@/lib/api-video-transcription";

type Tab = "settings" | "access";

const fieldClass =
  "w-full rounded-lg border border-border2 bg-bg px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const DEFAULT_SETTINGS: VideoTranscriptionSettings = {
  model: "general",
  language_code: "ru-RU",
  price_rub_per_minute: 0.16,
  text_normalization_enabled: true,
  literature_text: true,
  profanity_filter: false,
};

export function AdminVideoTranscriptionEditor() {
  const t = useTranslations("admin.videoTranscription");
  const [tab, setTab] = useState<Tab>("settings");
  const [settings, setSettings] = useState<VideoTranscriptionSettings>(DEFAULT_SETTINGS);
  const [users, setUsers] = useState<VideoTranscriptionAccessUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [busyId, setBusyId] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [settingsRes, accessRes] = await Promise.all([
        fetchAdminVideoTranscriptionSettings(),
        fetchAdminVideoTranscriptionAccess(),
      ]);
      setSettings({ ...DEFAULT_SETTINGS, ...settingsRes.settings });
      setUsers(Array.isArray(accessRes.items) ? accessRes.items : []);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    void load();
  }, [load]);

  async function saveSettings() {
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      const res = await updateAdminVideoTranscriptionSettings(settings);
      setSettings(res.settings);
      setSuccess(t("saved"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  async function toggleAccess(user: VideoTranscriptionAccessUser) {
    if (user.role === "superadmin") return;
    setBusyId(user.id);
    setError("");
    try {
      const enabled = !user.has_access;
      await updateAdminVideoTranscriptionAccess(user.id, enabled);
      setUsers((prev) =>
        prev.map((u) => (u.id === user.id ? { ...u, has_access: enabled } : u))
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : t("accessFailed"));
    } finally {
      setBusyId("");
    }
  }

  if (loading) {
    return <p className="text-sm text-text2">{t("loading")}</p>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </div>

      <div className="flex gap-1 border-b border-border">
        {(["settings", "access"] as Tab[]).map((key) => (
          <button
            key={key}
            type="button"
            onClick={() => setTab(key)}
            className={cn(
              "-mb-px border-b-2 px-3 py-2 text-sm",
              tab === key
                ? "border-text font-medium text-text"
                : "border-transparent text-text2 hover:text-text"
            )}
          >
            {key === "settings" ? t("tabSettings") : t("tabAccess")}
          </button>
        ))}
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

      {tab === "settings" && (
        <div className="max-w-xl space-y-4 rounded-xl border border-border bg-bg p-5">
          <div className="flex items-center gap-2 text-sm font-medium text-text">
            <Settings2 className="h-4 w-4" />
            {t("settingsTitle")}
          </div>
          <label className="block space-y-1">
            <span className="text-xs font-medium text-text2">{t("model")}</span>
            <select
              className={fieldClass}
              value={settings.model}
              onChange={(e) => setSettings((s) => ({ ...s, model: e.target.value }))}
            >
              <option value="general">general</option>
              <option value="general:rc">general:rc</option>
              <option value="deferred-general">deferred-general</option>
            </select>
          </label>
          <label className="block space-y-1">
            <span className="text-xs font-medium text-text2">{t("language")}</span>
            <select
              className={fieldClass}
              value={settings.language_code}
              onChange={(e) =>
                setSettings((s) => ({ ...s, language_code: e.target.value }))
              }
            >
              <option value="ru-RU">ru-RU</option>
              <option value="en-US">en-US</option>
              <option value="tr-TR">tr-TR</option>
            </select>
          </label>
          <div className="space-y-3 rounded-lg border border-border bg-bg2 p-4">
            <p className="text-xs font-medium uppercase tracking-wider text-text3">
              {t("formattingTitle")}
            </p>
            <label className="flex items-start gap-3 text-sm text-text">
              <input
                type="checkbox"
                className="mt-0.5"
                checked={settings.text_normalization_enabled}
                onChange={(e) =>
                  setSettings((s) => ({
                    ...s,
                    text_normalization_enabled: e.target.checked,
                    literature_text: e.target.checked ? s.literature_text : false,
                  }))
                }
              />
              <span>
                <span className="font-medium">{t("textNormalization")}</span>
                <span className="mt-1 block text-xs text-text2">{t("textNormalizationHint")}</span>
              </span>
            </label>
            <label className="flex items-start gap-3 text-sm text-text">
              <input
                type="checkbox"
                className="mt-0.5"
                checked={settings.literature_text}
                disabled={!settings.text_normalization_enabled}
                onChange={(e) =>
                  setSettings((s) => ({ ...s, literature_text: e.target.checked }))
                }
              />
              <span>
                <span className="font-medium">{t("literatureText")}</span>
                <span className="mt-1 block text-xs text-text2">{t("literatureTextHint")}</span>
              </span>
            </label>
            <label className="flex items-start gap-3 text-sm text-text">
              <input
                type="checkbox"
                className="mt-0.5"
                checked={settings.profanity_filter}
                onChange={(e) =>
                  setSettings((s) => ({ ...s, profanity_filter: e.target.checked }))
                }
              />
              <span>
                <span className="font-medium">{t("profanityFilter")}</span>
                <span className="mt-1 block text-xs text-text2">{t("profanityFilterHint")}</span>
              </span>
            </label>
          </div>
          <label className="block space-y-1">
            <span className="text-xs font-medium text-text2">{t("pricePerMinute")}</span>
            <input
              type="number"
              min={0}
              step={0.01}
              className={fieldClass}
              value={settings.price_rub_per_minute}
              onChange={(e) =>
                setSettings((s) => ({
                  ...s,
                  price_rub_per_minute: Number(e.target.value),
                }))
              }
            />
          </label>
          <button
            type="button"
            disabled={saving}
            onClick={() => void saveSettings()}
            className="inline-flex items-center gap-2 rounded-lg bg-text px-4 py-2 text-sm text-white disabled:opacity-50"
          >
            {saving ? (
              <LoaderCircle className="h-4 w-4 animate-spin" />
            ) : (
              <Save className="h-4 w-4" />
            )}
            {t("save")}
          </button>
        </div>
      )}

      {tab === "access" && (
        <div className="overflow-x-auto rounded-xl border border-border bg-bg">
          <div className="flex items-center gap-2 border-b border-border px-4 py-3 text-sm font-medium text-text">
            <Users className="h-4 w-4" />
            {t("accessTitle")}
          </div>
          <table className="w-full min-w-[520px] text-left text-sm">
            <thead>
              <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
                <th className="px-4 py-3 font-medium">{t("colEmail")}</th>
                <th className="px-4 py-3 font-medium">{t("colRole")}</th>
                <th className="px-4 py-3 font-medium">{t("colAccess")}</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-3">{user.email}</td>
                  <td className="px-4 py-3 text-text2">{user.role}</td>
                  <td className="px-4 py-3">
                    <button
                      type="button"
                      disabled={user.role === "superadmin" || busyId === user.id}
                      onClick={() => void toggleAccess(user)}
                      className={cn(
                        "rounded-full px-3 py-1 text-xs font-medium",
                        user.has_access
                          ? "bg-[#eaf3de] text-[#3b6d11]"
                          : "bg-bg2 text-text2",
                        user.role === "superadmin" && "opacity-60"
                      )}
                    >
                      {busyId === user.id ? (
                        <LoaderCircle className="inline h-3 w-3 animate-spin" />
                      ) : user.has_access ? (
                        t("accessOn")
                      ) : (
                        t("accessOff")
                      )}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
