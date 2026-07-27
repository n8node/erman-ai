"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { ModelPicker } from "@/components/admin/ModelPicker";
import {
  fetchAdminLegalScanLLMSettings,
  updateAdminLegalScanLLMSettings,
  type LLMProvider,
  type LegalScanLLMSettings,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const LLM_PROVIDERS: LLMProvider[] = ["yandex", "openrouter", "deepseek"];

const DEFAULT_SETTINGS: LegalScanLLMSettings = {
  provider: "yandex",
  openrouter_model: "google/gemini-flash-1.5-8b",
  deepseek_model: "deepseek-chat",
  yandex_model: "yandexgpt-lite/latest",
  system_prompt: "",
  temperature: 0.3,
  max_tokens: 4096,
};

function activeModelLabel(settings: LegalScanLLMSettings): string {
  switch (settings.provider) {
    case "deepseek":
      return settings.deepseek_model;
    case "yandex":
      return settings.yandex_model;
    default:
      return settings.openrouter_model;
  }
}

export function AdminLegalScanLLMEditor() {
  const t = useTranslations("admin.legalScanLlm");
  const [settings, setSettings] = useState<LegalScanLLMSettings>(DEFAULT_SETTINGS);
  const [defaultPrompt, setDefaultPrompt] = useState("");
  const [providers, setProviders] = useState<
    Awaited<ReturnType<typeof fetchAdminLegalScanLLMSettings>>["providers"]
  >([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    fetchAdminLegalScanLLMSettings()
      .then((data) => {
        setSettings(data.settings);
        setDefaultPrompt(data.default_system_prompt);
        setProviders(data.providers);
      })
      .catch((err) => setError(err instanceof Error ? err.message : t("loadFailed")))
      .finally(() => setLoading(false));
  }, [t]);

  function patch(partial: Partial<LegalScanLLMSettings>) {
    setSettings((prev) => ({ ...prev, ...partial }));
    setSuccess("");
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      const data = await updateAdminLegalScanLLMSettings(settings);
      setSettings(data.settings);
      setProviders(data.providers);
      setSuccess(t("saved"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  const openrouterMeta = providers.find((p) => p.id === "openrouter");
  const deepseekMeta = providers.find((p) => p.id === "deepseek");
  const yandexMeta = providers.find((p) => p.id === "yandex");

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
          {success}
        </div>
      )}

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("providerSection")}
        </h2>
        <p className="text-xs text-text3">{t("keysHint")}</p>
        <div className="flex flex-wrap gap-2">
          {LLM_PROVIDERS.map((id) => {
            const meta = providers.find((p) => p.id === id);
            return (
              <button
                key={id}
                type="button"
                onClick={() => patch({ provider: id })}
                className={cn(
                  "rounded-lg border px-4 py-2 text-sm",
                  settings.provider === id
                    ? "border-ai bg-ai-bg font-medium text-ai"
                    : "border-border2 text-text2 hover:border-border"
                )}
              >
                {t(`providers.${id}`)}
                {meta && !meta.configured && (
                  <span className="ml-2 text-xs text-warning">({t("keyMissing")})</span>
                )}
              </button>
            );
          })}
        </div>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("modelsSection")}
        </h2>
        <div className="grid gap-4 sm:grid-cols-2">
          <ModelPicker
            label={t("yandexModel")}
            value={settings.yandex_model}
            models={yandexMeta?.models ?? []}
            onChange={(v) => patch({ yandex_model: v })}
            placeholder={t("modelsEmpty")}
          />
          <ModelPicker
            label={t("openrouterModel")}
            value={settings.openrouter_model}
            models={openrouterMeta?.models ?? []}
            onChange={(v) => patch({ openrouter_model: v })}
            placeholder={t("modelsEmpty")}
          />
          <ModelPicker
            label={t("deepseekModel")}
            value={settings.deepseek_model}
            models={deepseekMeta?.models ?? []}
            onChange={(v) => patch({ deepseek_model: v })}
            placeholder={t("modelsEmpty")}
          />
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("temperature")}</label>
            <input
              type="number"
              step={0.1}
              min={0}
              max={2}
              className={fieldClass}
              value={settings.temperature}
              onChange={(e) => patch({ temperature: Number(e.target.value) })}
            />
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("maxTokens")}</label>
            <input
              type="number"
              step={256}
              min={256}
              max={8192}
              className={fieldClass}
              value={settings.max_tokens}
              onChange={(e) => patch({ max_tokens: Number(e.target.value) })}
            />
          </div>
        </div>
        <p className="text-xs text-text3">
          {t("activeModel")}:{" "}
          <span className="font-medium text-text">{activeModelLabel(settings)}</span>
        </p>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-3">
        <div className="flex items-center justify-between gap-2">
          <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
            {t("promptSection")}
          </h2>
          <button
            type="button"
            onClick={() => patch({ system_prompt: defaultPrompt })}
            className="text-xs text-accent underline hover:no-underline"
          >
            {t("resetPrompt")}
          </button>
        </div>
        <textarea
          rows={14}
          className={cn(fieldClass, "font-mono text-xs leading-relaxed")}
          value={settings.system_prompt}
          onChange={(e) => patch({ system_prompt: e.target.value })}
        />
      </section>

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
