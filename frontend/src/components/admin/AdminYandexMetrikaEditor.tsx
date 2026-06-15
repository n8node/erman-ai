"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchAdminYandexMetrikaSettings,
  updateAdminYandexMetrikaSettings,
  type YandexMetrikaAdminView,
  type YandexMetrikaSettings,
} from "@/lib/api";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm font-mono outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const DEFAULT_SETTINGS: YandexMetrikaSettings = {
  enabled: false,
  counter_code: "",
};

export function AdminYandexMetrikaEditor() {
  const t = useTranslations("admin.yandexMetrika");
  const [settings, setSettings] = useState<YandexMetrikaSettings>(DEFAULT_SETTINGS);
  const [updatedAt, setUpdatedAt] = useState<string | undefined>();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  function applyView(data: YandexMetrikaAdminView) {
    setSettings(data.settings);
    setUpdatedAt(data.updated_at);
  }

  useEffect(() => {
    fetchAdminYandexMetrikaSettings()
      .then(applyView)
      .catch((err) =>
        setError(err instanceof Error ? err.message : t("loadFailed"))
      )
      .finally(() => setLoading(false));
  }, [t]);

  function patch(partial: Partial<YandexMetrikaSettings>) {
    setSettings((prev) => ({ ...prev, ...partial }));
    setSuccess("");
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      const data = await updateAdminYandexMetrikaSettings({ settings });
      applyView(data);
      setSuccess(t("saved"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return <p className="text-sm text-text2">{t("loading")}</p>;
  }

  return (
    <div className="space-y-6">
      {error && (
        <div className="rounded-lg border border-error/20 bg-error/10 px-4 py-3 text-sm text-error">
          {error}
        </div>
      )}
      {success && (
        <div className="rounded-lg border border-success/20 bg-success/10 px-4 py-3 text-sm text-success">
          {success}
        </div>
      )}

      <section className="rounded-xl border border-border bg-bg p-6 space-y-4">
        <div>
          <h2 className="text-sm font-medium">{t("sectionTitle")}</h2>
          <p className="mt-1 text-sm text-text2">{t("sectionHint")}</p>
        </div>

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
          <label className="mb-1.5 block text-xs font-medium text-text2">
            {t("counterCode")}
          </label>
          <textarea
            value={settings.counter_code}
            onChange={(e) => patch({ counter_code: e.target.value })}
            rows={16}
            spellCheck={false}
            placeholder={t("counterCodePlaceholder")}
            className={fieldClass}
          />
          <p className="mt-1.5 text-xs text-text3">{t("counterCodeHint")}</p>
        </div>

        {updatedAt && (
          <p className="text-xs text-text3">
            {t("updatedAt", { time: new Date(updatedAt).toLocaleString() })}
          </p>
        )}

        <button
          type="button"
          onClick={handleSave}
          disabled={saving}
          className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
        >
          {saving ? t("saving") : t("save")}
        </button>
      </section>
    </div>
  );
}
