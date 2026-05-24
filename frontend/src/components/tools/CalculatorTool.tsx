"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import type { CalculatorInput } from "@/lib/api";
import { runCalculator } from "@/lib/api";
import { CalculatorResult } from "./CalculatorResult";

const defaultInput: CalculatorInput = {
  process_name: "",
  hours_per_month: 40,
  hourly_rate_rub: 800,
  employees_count: 1,
  error_rate_pct: 5,
  error_cost_rub: 5000,
  automation_cost_rub: 300000,
  monthly_support_rub: 15000,
};

export function CalculatorTool() {
  const t = useTranslations("calculator");
  const [input, setInput] = useState<CalculatorInput>(defaultInput);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState<Awaited<
    ReturnType<typeof runCalculator>
  > | null>(null);

  function setField<K extends keyof CalculatorInput>(
    key: K,
    value: CalculatorInput[K]
  ) {
    setInput((prev) => ({ ...prev, [key]: value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const data = await runCalculator(input);
      setResult(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("calcFailed"));
    } finally {
      setLoading(false);
    }
  }

  const fieldClass =
    "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </div>

      <form
        onSubmit={handleSubmit}
        className="rounded-xl border border-border bg-bg p-6 space-y-6"
      >
        <section>
          <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
            {t("sections.process")}
          </h2>
          <div className="mt-4 grid gap-4 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <label className="mb-1.5 block text-xs font-medium">
                {t("fields.processName")}
              </label>
              <input
                required
                className={fieldClass}
                value={input.process_name}
                onChange={(e) => setField("process_name", e.target.value)}
              />
            </div>
            <div>
              <label className="mb-1.5 block text-xs font-medium">
                {t("fields.hoursPerMonth")}
              </label>
              <input
                type="number"
                min={0}
                step={0.1}
                required
                className={fieldClass}
                value={input.hours_per_month}
                onChange={(e) =>
                  setField("hours_per_month", Number(e.target.value))
                }
              />
            </div>
            <div>
              <label className="mb-1.5 block text-xs font-medium">
                {t("fields.employees")}
              </label>
              <input
                type="number"
                min={1}
                required
                className={fieldClass}
                value={input.employees_count}
                onChange={(e) =>
                  setField("employees_count", Number(e.target.value))
                }
              />
            </div>
            <div>
              <label className="mb-1.5 block text-xs font-medium">
                {t("fields.hourlyRate")}
              </label>
              <input
                type="number"
                min={0}
                required
                className={fieldClass}
                value={input.hourly_rate_rub}
                onChange={(e) =>
                  setField("hourly_rate_rub", Number(e.target.value))
                }
              />
            </div>
          </div>
        </section>

        <section className="border-t border-border pt-6">
          <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
            {t("sections.errors")}
          </h2>
          <div className="mt-4 grid gap-4 sm:grid-cols-2">
            <div>
              <label className="mb-1.5 block text-xs font-medium">
                {t("fields.errorRate")}
              </label>
              <input
                type="number"
                min={0}
                max={100}
                required
                className={fieldClass}
                value={input.error_rate_pct}
                onChange={(e) =>
                  setField("error_rate_pct", Number(e.target.value))
                }
              />
            </div>
            <div>
              <label className="mb-1.5 block text-xs font-medium">
                {t("fields.errorCost")}
              </label>
              <input
                type="number"
                min={0}
                required
                className={fieldClass}
                value={input.error_cost_rub}
                onChange={(e) =>
                  setField("error_cost_rub", Number(e.target.value))
                }
              />
            </div>
          </div>
        </section>

        <section className="border-t border-border pt-6">
          <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
            {t("sections.automation")}
          </h2>
          <div className="mt-4 grid gap-4 sm:grid-cols-2">
            <div>
              <label className="mb-1.5 block text-xs font-medium">
                {t("fields.automationCost")}
              </label>
              <input
                type="number"
                min={0}
                required
                className={fieldClass}
                value={input.automation_cost_rub}
                onChange={(e) =>
                  setField("automation_cost_rub", Number(e.target.value))
                }
              />
            </div>
            <div>
              <label className="mb-1.5 block text-xs font-medium">
                {t("fields.monthlySupport")}
              </label>
              <input
                type="number"
                min={0}
                required
                className={fieldClass}
                value={input.monthly_support_rub}
                onChange={(e) =>
                  setField("monthly_support_rub", Number(e.target.value))
                }
              />
            </div>
          </div>
        </section>

        {error && (
          <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
            {error}
          </div>
        )}

        <button
          type="submit"
          disabled={loading}
          className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-60"
        >
          {loading ? t("calculating") : t("calculate")}
        </button>
      </form>

      {result && (
        <CalculatorResult
          runId={result.run_id}
          input={result.input}
          output={result.output}
        />
      )}
    </div>
  );
}
