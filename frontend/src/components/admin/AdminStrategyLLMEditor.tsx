"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { ModelPicker } from "@/components/admin/ModelPicker";
import {
  fetchAdminStrategyLLMSettings,
  testStrategyLLMConnection,
  updateAdminStrategyLLMSettings,
  type LLMProvider,
  type LLMProviderStatus,
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
  const [providers, setProviders] = useState<LLMProviderStatus[]>([]);
  const [openrouterKeyInput, setOpenrouterKeyInput] = useState("");
  const [deepseekKeyInput, setDeepseekKeyInput] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState<LLMProvider | null>(null);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [testStatus, setTestStatus] = useState<Partial<Record<LLMProvider, string>>>({});

  function applyView(data: Awaited<ReturnType<typeof fetchAdminStrategyLLMSettings>>) {
    setSettings(data.settings);
    setDefaultPrompt(data.default_system_prompt);
    setProviders(data.providers);
  }

  useEffect(() => {
    fetchAdminStrategyLLMSettings()
      .then(applyView)
      .catch((err) =>
        setError(err instanceof Error ? err.message : t("loadFailed"))
      )
      .finally(() => setLoading(false));
  }, [t]);

  function patch(partial: Partial<StrategyLLMSettings>) {
    setSettings((prev) => ({ ...prev, ...partial }));
    setSuccess("");
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      const data = await updateAdminStrategyLLMSettings({
        settings,
        ...(openrouterKeyInput.trim() ? { openrouter_api_key: openrouterKeyInput.trim() } : {}),
        ...(deepseekKeyInput.trim() ? { deepseek_api_key: deepseekKeyInput.trim() } : {}),
      });
      applyView(data);
      setOpenrouterKeyInput("");
      setDeepseekKeyInput("");
      setSuccess(t("saved"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  async function handleTestConnection(provider: LLMProvider) {
    setTesting(provider);
    setError("");
    setTestStatus((prev) => ({ ...prev, [provider]: "" }));
    try {
      if (
        (provider === "openrouter" && openrouterKeyInput.trim()) ||
        (provider === "deepseek" && deepseekKeyInput.trim())
      ) {
        const saved = await updateAdminStrategyLLMSettings({
          settings,
          ...(openrouterKeyInput.trim() ? { openrouter_api_key: openrouterKeyInput.trim() } : {}),
          ...(deepseekKeyInput.trim() ? { deepseek_api_key: deepseekKeyInput.trim() } : {}),
        });
        applyView(saved);
        setOpenrouterKeyInput("");
        setDeepseekKeyInput("");
      }

      const result = await testStrategyLLMConnection(provider);
      if (!result.ok) {
        setTestStatus((prev) => ({ ...prev, [provider]: result.message }));
        return;
      }

      setTestStatus((prev) => ({ ...prev, [provider]: result.message }));
      const refreshed = await fetchAdminStrategyLLMSettings();
      applyView(refreshed);
      if (result.models.length > 0) {
        if (provider === "openrouter" && !result.models.includes(settings.openrouter_model)) {
          patch({ openrouter_model: result.models[0] });
        }
        if (provider === "deepseek" && !result.models.includes(settings.deepseek_model)) {
          patch({ deepseek_model: result.models[0] });
        }
      }
    } catch (err) {
      setTestStatus((prev) => ({
        ...prev,
        [provider]: err instanceof Error ? err.message : t("connectionFailed"),
      }));
    } finally {
      setTesting(null);
    }
  }

  function handleResetPrompt() {
    patch({ system_prompt: defaultPrompt });
  }

  const openrouterMeta = providers.find((p) => p.id === "openrouter");
  const deepseekMeta = providers.find((p) => p.id === "deepseek");

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
          {t("keysSection")}
        </h2>
        {(["openrouter", "deepseek"] as LLMProvider[]).map((id) => {
          const meta = providers.find((p) => p.id === id);
          const keyValue = id === "openrouter" ? openrouterKeyInput : deepseekKeyInput;
          const setKey = id === "openrouter" ? setOpenrouterKeyInput : setDeepseekKeyInput;
          return (
            <div key={id} className="rounded-lg border border-border bg-bg2/40 p-4 space-y-2">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="text-sm font-medium">{t(`providers.${id}`)}</p>
                {meta?.configured ? (
                  <span className="text-xs text-success">{t("keyConfigured")}: {meta.key_hint}</span>
                ) : (
                  <span className="text-xs text-warning">{t("keyMissing")}</span>
                )}
              </div>
              <input
                type="password"
                className={fieldClass}
                value={keyValue}
                placeholder={meta?.key_hint ? t("keyPlaceholderExisting", { hint: meta.key_hint }) : t("keyPlaceholder")}
                onChange={(e) => {
                  setKey(e.target.value);
                  setSuccess("");
                }}
              />
              <button
                type="button"
                disabled={testing === id}
                onClick={() => handleTestConnection(id)}
                className="rounded-lg border border-border2 px-3 py-1.5 text-xs font-medium hover:bg-bg2 disabled:opacity-60"
              >
                {testing === id ? t("testing") : t("testConnection")}
              </button>
              {testStatus[id] && (
                <p className={cn("text-xs", testStatus[id]?.includes("connected") ? "text-success" : "text-text2")}>
                  {testStatus[id]}
                </p>
              )}
            </div>
          );
        })}
      </section>

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
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("modelsSection")}
        </h2>
        <div className="grid gap-4 sm:grid-cols-2">
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
