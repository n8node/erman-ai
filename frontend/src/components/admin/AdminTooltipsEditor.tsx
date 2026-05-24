"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  bulkUpdateAdminTooltips,
  fetchAdminTooltips,
  type AdminTooltip,
} from "@/lib/api";

export function AdminTooltipsEditor() {
  const t = useTranslations("admin.tooltips");
  const [items, setItems] = useState<AdminTooltip[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    fetchAdminTooltips()
      .then((data) => setItems(data.items))
      .catch((err) =>
        setError(err instanceof Error ? err.message : t("loadFailed"))
      )
      .finally(() => setLoading(false));
  }, [t]);

  function updateItem(key: string, field: "text_ru" | "text_en", value: string) {
    setItems((prev) =>
      prev.map((item) => (item.key === key ? { ...item, [field]: value } : item))
    );
    setSuccess(false);
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess(false);
    try {
      const data = await bulkUpdateAdminTooltips(
        items.map(({ key, text_ru, text_en }) => ({ key, text_ru, text_en }))
      );
      setItems(data.items);
      setSuccess(true);
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
      <div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </div>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}
      {success && (
        <div className="rounded-md border border-green-200 bg-success-bg px-3 py-2 text-sm text-success">
          {t("saved")}
        </div>
      )}

      <div className="space-y-4">
        {items.map((item) => (
          <div
            key={item.key}
            className="rounded-xl border border-border bg-bg p-5 space-y-3"
          >
            <div>
              <p className="text-sm font-medium">{item.label}</p>
              <p className="text-[10px] font-mono text-text3">{item.key}</p>
            </div>
            <div className="grid gap-3 lg:grid-cols-2">
              <div>
                <label className="mb-1 block text-xs font-medium text-text2">
                  {t("textRu")}
                </label>
                <textarea
                  rows={3}
                  className="w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
                  value={item.text_ru}
                  onChange={(e) => updateItem(item.key, "text_ru", e.target.value)}
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-text2">
                  {t("textEn")}
                </label>
                <textarea
                  rows={3}
                  className="w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
                  value={item.text_en}
                  onChange={(e) => updateItem(item.key, "text_en", e.target.value)}
                />
              </div>
            </div>
          </div>
        ))}
      </div>

      <button
        type="button"
        onClick={handleSave}
        disabled={saving}
        className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-60"
      >
        {saving ? t("saving") : t("save")}
      </button>
    </div>
  );
}
