"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchAdminCalculatorBudgetConfig,
  updateAdminCalculatorBudgetConfig,
} from "@/lib/api";
import {
  BUDGET_CONFIG_SECTIONS,
  DEFAULT_CALCULATOR_BUDGET_CONFIG,
  isIntegerBudgetField,
  type BudgetConfigFieldKey,
  type CalculatorBudgetConfig,
} from "@/lib/calculator-budget-config";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

export function AdminCalculatorBudgetEditor() {
  const t = useTranslations("admin.calculatorBudget");
  const [config, setConfig] = useState<CalculatorBudgetConfig>(DEFAULT_CALCULATOR_BUDGET_CONFIG);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    fetchAdminCalculatorBudgetConfig()
      .then((data) => setConfig(data.config))
      .catch((err) =>
        setError(err instanceof Error ? err.message : t("loadFailed"))
      )
      .finally(() => setLoading(false));
  }, [t]);

  function updateField(key: BudgetConfigFieldKey, raw: string) {
    const value = isIntegerBudgetField(key) ? parseInt(raw, 10) : parseFloat(raw);
    if (Number.isNaN(value)) return;
    setConfig((prev) => ({ ...prev, [key]: value }));
    setSuccess(false);
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess(false);
    try {
      const data = await updateAdminCalculatorBudgetConfig(config);
      setConfig(data.config);
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  function handleReset() {
    setConfig(DEFAULT_CALCULATOR_BUDGET_CONFIG);
    setSuccess(false);
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

      <div className="space-y-6">
        {BUDGET_CONFIG_SECTIONS.map(({ sectionKey, fields }) => (
          <section
            key={sectionKey}
            className="rounded-xl border border-border bg-bg p-5 space-y-4"
          >
            <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
              {t(`sections.${sectionKey}`)}
            </h2>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {fields.map((key) => (
                <div key={key}>
                  <label className="mb-1.5 block text-xs font-medium text-text">
                    {t(`fields.${key}`)}
                  </label>
                  <input
                    type="number"
                    step={isIntegerBudgetField(key) ? 1 : key.includes("mult") ? 0.01 : 1}
                    className={fieldClass}
                    value={config[key]}
                    onChange={(e) => updateField(key, e.target.value)}
                  />
                </div>
              ))}
            </div>
          </section>
        ))}
      </div>

      <div className="flex flex-wrap gap-2">
        <button
          type="button"
          onClick={handleSave}
          disabled={saving}
          className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-60"
        >
          {saving ? t("saving") : t("save")}
        </button>
        <button
          type="button"
          onClick={handleReset}
          disabled={saving}
          className="rounded-lg border border-border2 px-4 py-2 text-sm text-text2 hover:text-text"
        >
          {t("resetDefaults")}
        </button>
      </div>
    </div>
  );
}
