"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchAdminStrategyLLMSettings,
  updateAdminStrategyLLMSettings,
  type LLMProvider,
  type StrategyLLMSettings,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const DEFAULT_SETTINGS: StrategyLLMSettings = {
  provider: "openrouter",
  openrouter_model: "anthropic/claude-sonnet-4-5",
  deepseek_model: "deepseek-chat",
  system_prompt: "",
  temperature: 0.7,
  max_tokens: 8192,
};

export function AdminStrategyLLMEditor() {
  const t = useTranslations("admin.strategyLlm");
  const [settings, setSettings] = useState<StrategyLLMSettings>(DEFAULT_SETTINGS);
  const [defaultPrompt, setDefaultPrompt] = useState("");
  const [providers, setProviders] = useState<
    { id: LLMProvider; configured: boolean; suggested_models: string[] }[]
  >([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    fetchAdminStrategyLLMSettings()
      .then((data) => {
        setSettings(data.settings);
        setDefaultPrompt(data.default_system_prompt);
        setProviders(data.providers);
      })
      .catch((err) =>
        setError(err instanceof Error ? err.message : t("loadFailed"))
      )
      .finally(() => setLoading(false));
  }, [t]);

  function patch(partial: Partial<StrategyLLMSettings>) {
    setSettings((prev) => ({ ...prev, ...partial }));
    setSuccess(false);
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess(false);
    try {
      const data = await updateAdminStrategyLLMSettings(settings);
      setSettings(data.settings);
      setDefaultPrompt(data.default_system_prompt);
      setProviders(data.providers);
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  function handleResetPrompt() {
    patch({ system_prompt: defaultPrompt });
  }

  const activeProvider = providers.find((p) => p.id === settings.provider);

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

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("providerSection")}
        </h2>
        <div className="flex flex-wrap gap-2">
          {(["openrouter", "deepseek"] as LLMProvider[]).map((id) => {
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
        {activeProvider && !activeProvider.configured && (
          <p className="text-xs text-warning">{t("keyMissingHint")}</p>
        )}
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("modelsSection")}
        </h2>
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("openrouterModel")}</label>
            <input
              className={fieldClass}
              value={settings.openrouter_model}
              onChange={(e) => patch({ openrouter_model: e.target.value })}
              list="openrouter-models"
            />
            <datalist id="openrouter-models">
              {providers
                .find((p) => p.id === "openrouter")
                ?.suggested_models.map((m) => (
                  <option key={m} value={m} />
                ))}
            </datalist>
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("deepseekModel")}</label>
            <input
              className={fieldClass}
              value={settings.deepseek_model}
              onChange={(e) => patch({ deepseek_model: e.target.value })}
              list="deepseek-models"
            />
            <datalist id="deepseek-models">
              {providers
                .find((p) => p.id === "deepseek")
                ?.suggested_models.map((m) => (
                  <option key={m} value={m} />
                ))}
            </datalist>
          </div>
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
              max={32000}
              className={fieldClass}
              value={settings.max_tokens}
              onChange={(e) => patch({ max_tokens: Number(e.target.value) })}
            />
          </div>
        </div>
        <p className="text-xs text-text3">
          {t("activeModel")}:{" "}
          <span className="font-medium text-text">
            {settings.provider === "deepseek"
              ? settings.deepseek_model
              : settings.openrouter_model}
          </span>
        </p>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-3">
        <div className="flex items-center justify-between gap-2">
          <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
            {t("promptSection")}
          </h2>
          <button
            type="button"
            onClick={handleResetPrompt}
            className="text-xs text-accent underline hover:no-underline"
          >
            {t("resetPrompt")}
          </button>
        </div>
        <textarea
          rows={18}
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
