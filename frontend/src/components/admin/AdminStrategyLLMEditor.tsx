"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { ModelPicker } from "@/components/admin/ModelPicker";
import { AdminLLMProxyFields } from "@/components/admin/AdminLLMProxyFields";
import {
  ModelPricingFields,
  modelsCountLabel,
  resolveModelPricing,
} from "@/components/admin/ModelPricingFields";
import {
  fetchAdminStrategyLLMSettings,
  testStrategyLLMConnection,
  updateAdminStrategyLLMSettings,
  DEFAULT_LLM_HTTP_PROXY,
  type LLMProvider,
  type LLMProviderPricing,
  type LLMProviderStatus,
  type StrategyLLMSettings,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const LLM_PROVIDERS: LLMProvider[] = ["yandex", "openrouter", "deepseek"];

const DEFAULT_SETTINGS: StrategyLLMSettings = {
  provider: "yandex",
  openrouter_model: "anthropic/claude-sonnet-4-5",
  deepseek_model: "deepseek-chat",
  yandex_model: "yandexgpt/latest",
  openrouter_proxy: DEFAULT_LLM_HTTP_PROXY,
  deepseek_proxy: DEFAULT_LLM_HTTP_PROXY,
  system_prompt: "",
  proposal_system_prompt: "",
  temperature: 0.7,
  max_tokens: 8192,
};

const DEFAULT_PRICING: Record<LLMProvider, LLMProviderPricing> = {
  yandex: { input_per_1k: 0.6, output_per_1k: 1.8, currency: "RUB" },
  openrouter: { input_per_1k: 0.003, output_per_1k: 0.015, currency: "USD" },
  deepseek: { input_per_1k: 0.014, output_per_1k: 0.028, currency: "USD" },
};

function activeModelLabel(settings: StrategyLLMSettings): string {
  switch (settings.provider) {
    case "deepseek":
      return settings.deepseek_model;
    case "yandex":
      return settings.yandex_model;
    default:
      return settings.openrouter_model;
  }
}

export function AdminStrategyLLMEditor() {
  const t = useTranslations("admin.strategyLlm");
  const [settings, setSettings] = useState<StrategyLLMSettings>(DEFAULT_SETTINGS);
  const [pricing, setPricing] = useState<Partial<Record<LLMProvider, LLMProviderPricing>>>(DEFAULT_PRICING);
  const [modelPricing, setModelPricing] = useState<Record<string, LLMProviderPricing>>({});
  const [defaultPrompt, setDefaultPrompt] = useState("");
  const [defaultProposalPrompt, setDefaultProposalPrompt] = useState("");
  const [providers, setProviders] = useState<LLMProviderStatus[]>([]);
  const [openrouterKeyInput, setOpenrouterKeyInput] = useState("");
  const [deepseekKeyInput, setDeepseekKeyInput] = useState("");
  const [yandexKeyInput, setYandexKeyInput] = useState("");
  const [yandexFolderInput, setYandexFolderInput] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState<LLMProvider | null>(null);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [testStatus, setTestStatus] = useState<Partial<Record<LLMProvider, string>>>({});

  function applyView(data: Awaited<ReturnType<typeof fetchAdminStrategyLLMSettings>>) {
    setSettings(data.settings);
    setDefaultPrompt(data.default_system_prompt);
    setDefaultProposalPrompt(data.default_proposal_system_prompt);
    setProviders(data.providers);
    setPricing({ ...DEFAULT_PRICING, ...data.pricing });
    setModelPricing(data.model_pricing ?? {});
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

  function patchModelPricing(modelId: string, provider: LLMProvider, partial: Partial<LLMProviderPricing>) {
    if (!modelId.trim()) return;
    setModelPricing((prev) => ({
      ...prev,
      [modelId]: {
        ...resolveModelPricing(modelId, provider, prev, pricing, DEFAULT_PRICING),
        ...partial,
      },
    }));
    setSuccess("");
  }

  function buildSavePayload() {
    return {
      settings,
      pricing,
      model_pricing: modelPricing,
      ...(openrouterKeyInput.trim() ? { openrouter_api_key: openrouterKeyInput.trim() } : {}),
      ...(deepseekKeyInput.trim() ? { deepseek_api_key: deepseekKeyInput.trim() } : {}),
      ...(yandexKeyInput.trim() ? { yandex_api_key: yandexKeyInput.trim() } : {}),
      ...(yandexFolderInput.trim() ? { yandex_folder_id: yandexFolderInput.trim() } : {}),
    };
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      const data = await updateAdminStrategyLLMSettings(buildSavePayload());
      applyView(data);
      setOpenrouterKeyInput("");
      setDeepseekKeyInput("");
      setYandexKeyInput("");
      setYandexFolderInput("");
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
      const saved = await updateAdminStrategyLLMSettings(buildSavePayload());
      applyView(saved);
      setOpenrouterKeyInput("");
      setDeepseekKeyInput("");
      setYandexKeyInput("");
      setYandexFolderInput("");

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
        if (provider === "yandex" && !result.models.includes(settings.yandex_model)) {
          patch({ yandex_model: result.models[0] });
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

  function handleResetProposalPrompt() {
    patch({ proposal_system_prompt: defaultProposalPrompt });
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

        <div className="rounded-lg border border-border bg-bg2/40 p-4 space-y-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <p className="text-sm font-medium">{t("providers.yandex")}</p>
            {yandexMeta?.configured ? (
              <span className="text-xs text-success">
                {t("keyConfigured")}: {yandexMeta.key_hint}
                {yandexMeta.folder_hint ? ` · ${t("folderConfigured")}: ${yandexMeta.folder_hint}` : ""}
              </span>
            ) : (
              <span className="text-xs text-warning">{t("yandexCredentialsMissing")}</span>
            )}
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("yandexApiKey")}</label>
            <input
              type="password"
              className={fieldClass}
              value={yandexKeyInput}
              placeholder={
                yandexMeta?.key_hint
                  ? t("keyPlaceholderExisting", { hint: yandexMeta.key_hint })
                  : t("keyPlaceholder")
              }
              onChange={(e) => {
                setYandexKeyInput(e.target.value);
                setSuccess("");
              }}
            />
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("yandexFolderId")}</label>
            <input
              type="text"
              className={fieldClass}
              value={yandexFolderInput}
              placeholder={
                yandexMeta?.folder_hint
                  ? t("folderPlaceholderExisting", { hint: yandexMeta.folder_hint })
                  : t("folderPlaceholder")
              }
              onChange={(e) => {
                setYandexFolderInput(e.target.value);
                setSuccess("");
              }}
            />
          </div>
          <button
            type="button"
            disabled={testing === "yandex"}
            onClick={() => handleTestConnection("yandex")}
            className="rounded-lg border border-border2 px-3 py-1.5 text-xs font-medium hover:bg-bg2 disabled:opacity-60"
          >
            {testing === "yandex" ? t("testing") : t("testConnection")}
          </button>
          {testStatus.yandex && (
            <p className={cn("text-xs", testStatus.yandex?.includes("connected") ? "text-success" : "text-text2")}>
              {testStatus.yandex}
            </p>
          )}
        </div>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("proxySection")}
        </h2>
        <p className="text-xs text-text3">{t("proxyScopeHint")}</p>
        <AdminLLMProxyFields
          providerLabel={t("providers.openrouter")}
          value={settings.openrouter_proxy ?? DEFAULT_LLM_HTTP_PROXY}
          onChange={(openrouter_proxy) => patch({ openrouter_proxy })}
        />
        <AdminLLMProxyFields
          providerLabel={t("providers.deepseek")}
          value={settings.deepseek_proxy ?? DEFAULT_LLM_HTTP_PROXY}
          onChange={(deepseek_proxy) => patch({ deepseek_proxy })}
        />
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("providerSection")}
        </h2>
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
        <div className="grid gap-4 sm:grid-cols-1 lg:grid-cols-2">
          <div>
            <ModelPicker
              label={t("yandexModel")}
              value={settings.yandex_model}
              models={yandexMeta?.models ?? []}
              onChange={(v) => patch({ yandex_model: v })}
              placeholder={t("modelsEmpty")}
            />
            <p className="mt-1 text-[11px] text-text3">
              {modelsCountLabel(yandexMeta?.models.length ?? 0, t("modelsAvailable"))}
            </p>
            <ModelPricingFields
              modelId={settings.yandex_model}
              provider="yandex"
              pricing={resolveModelPricing(settings.yandex_model, "yandex", modelPricing, pricing, DEFAULT_PRICING)}
              labels={{
                hint: t("modelPricingHint"),
                input: t("priceInputPer1k", { currency: "{currency}" }),
                output: t("priceOutputPer1k", { currency: "{currency}" }),
              }}
              onChange={(partial) => patchModelPricing(settings.yandex_model, "yandex", partial)}
            />
          </div>
          <div>
            <ModelPicker
              label={t("openrouterModel")}
              value={settings.openrouter_model}
              models={openrouterMeta?.models ?? []}
              onChange={(v) => patch({ openrouter_model: v })}
              placeholder={t("modelsEmpty")}
            />
            <p className="mt-1 text-[11px] text-text3">
              {modelsCountLabel(openrouterMeta?.models.length ?? 0, t("modelsAvailable"))}
            </p>
            <ModelPricingFields
              modelId={settings.openrouter_model}
              provider="openrouter"
              pricing={resolveModelPricing(settings.openrouter_model, "openrouter", modelPricing, pricing, DEFAULT_PRICING)}
              labels={{
                hint: t("modelPricingHint"),
                input: t("priceInputPer1k", { currency: "{currency}" }),
                output: t("priceOutputPer1k", { currency: "{currency}" }),
              }}
              onChange={(partial) => patchModelPricing(settings.openrouter_model, "openrouter", partial)}
            />
          </div>
          <div>
            <ModelPicker
              label={t("deepseekModel")}
              value={settings.deepseek_model}
              models={deepseekMeta?.models ?? []}
              onChange={(v) => patch({ deepseek_model: v })}
              placeholder={t("modelsEmpty")}
            />
            <p className="mt-1 text-[11px] text-text3">
              {modelsCountLabel(deepseekMeta?.models.length ?? 0, t("modelsAvailable"))}
            </p>
            <ModelPricingFields
              modelId={settings.deepseek_model}
              provider="deepseek"
              pricing={resolveModelPricing(settings.deepseek_model, "deepseek", modelPricing, pricing, DEFAULT_PRICING)}
              labels={{
                hint: t("modelPricingHint"),
                input: t("priceInputPer1k", { currency: "{currency}" }),
                output: t("priceOutputPer1k", { currency: "{currency}" }),
              }}
              onChange={(partial) => patchModelPricing(settings.deepseek_model, "deepseek", partial)}
            />
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
          <span className="font-medium text-text">{activeModelLabel(settings)}</span>
        </p>
        <p className="text-xs text-text3">{t("pricingHint")}</p>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-3">
        <div className="flex items-center justify-between gap-2">
          <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
            {t("strategyPromptSection")}
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

      <section className="rounded-xl border border-border bg-bg p-5 space-y-3">
        <div className="flex items-center justify-between gap-2">
          <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
            {t("proposalPromptSection")}
          </h2>
          <button
            type="button"
            onClick={handleResetProposalPrompt}
            className="text-xs text-accent underline hover:no-underline"
          >
            {t("resetProposalPrompt")}
          </button>
        </div>
        <p className="text-xs text-text3">{t("proposalPromptHint")}</p>
        <textarea
          rows={18}
          className={cn(fieldClass, "font-mono text-xs leading-relaxed")}
          value={settings.proposal_system_prompt}
          onChange={(e) => patch({ proposal_system_prompt: e.target.value })}
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
