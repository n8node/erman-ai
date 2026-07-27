"use client";

import type { LLMProvider, LLMProviderPricing } from "@/lib/api";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

type Props = {
  modelId: string;
  provider: LLMProvider;
  pricing: LLMProviderPricing;
  labels: {
    input: string;
    output: string;
    hint: string;
  };
  onChange: (partial: Partial<LLMProviderPricing>) => void;
};

export function ModelPricingFields({ modelId, provider, pricing, labels, onChange }: Props) {
  if (!modelId.trim()) {
    return null;
  }

  const currency = pricing.currency ?? (provider === "yandex" ? "RUB" : "USD");

  return (
    <div className="mt-2 rounded-lg border border-dashed border-border bg-bg2/30 p-3 space-y-2">
      <p className="text-[10px] uppercase tracking-wider text-text3">{labels.hint}</p>
      <p className="truncate font-mono text-[11px] text-text2" title={modelId}>
        {modelId}
      </p>
      <div className="grid gap-2 sm:grid-cols-2">
        <div>
          <label className="mb-1 block text-xs font-medium">{labels.input.replace("{currency}", currency)}</label>
          <input
            type="number"
            step={0.0001}
            min={0}
            className={fieldClass}
            value={pricing.input_per_1k}
            onChange={(e) => onChange({ input_per_1k: Number(e.target.value), currency })}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium">{labels.output.replace("{currency}", currency)}</label>
          <input
            type="number"
            step={0.0001}
            min={0}
            className={fieldClass}
            value={pricing.output_per_1k}
            onChange={(e) => onChange({ output_per_1k: Number(e.target.value), currency })}
          />
        </div>
      </div>
    </div>
  );
}

export function resolveModelPricing(
  modelId: string,
  provider: LLMProvider,
  modelPricing: Record<string, LLMProviderPricing>,
  providerDefaults: Partial<Record<LLMProvider, LLMProviderPricing>>,
  fallbacks: Record<LLMProvider, LLMProviderPricing>
): LLMProviderPricing {
  const stored = modelPricing[modelId];
  if (stored && (stored.input_per_1k > 0 || stored.output_per_1k > 0)) {
    return {
      currency: stored.currency ?? fallbacks[provider].currency,
      input_per_1k: stored.input_per_1k,
      output_per_1k: stored.output_per_1k,
    };
  }
  const providerStored = providerDefaults[provider];
  if (providerStored && (providerStored.input_per_1k > 0 || providerStored.output_per_1k > 0)) {
    return providerStored;
  }
  return fallbacks[provider];
}

export function modelsCountLabel(count: number, template: string): string {
  return template.replace("{count}", String(count));
}
