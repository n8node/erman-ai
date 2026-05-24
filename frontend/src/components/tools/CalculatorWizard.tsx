"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import type { CalculatorInput, ProcessStep } from "@/lib/api";
import { getRun, runCalculator } from "@/lib/api";
import {
  clearCalculatorDraft,
  loadCalculatorDraft,
  saveCalculatorDraft,
} from "@/lib/calculator-draft";
import {
  defaultCalculatorInput,
  previewCalculator,
  resolveHoursSaved,
  totalMinutesPerUnit,
} from "@/lib/calculator";
import { CalculatorResult } from "./CalculatorResult";
import { HelpTooltip, LabelWithHelp } from "@/components/ui/HelpTooltip";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

export function CalculatorWizard() {
  const t = useTranslations("calculator");
  const router = useRouter();
  const searchParams = useSearchParams();
  const runIdParam = searchParams.get("run");

  const [step, setStep] = useState(1);
  const [input, setInput] = useState<CalculatorInput>(defaultCalculatorInput);
  const [showExpert, setShowExpert] = useState(false);
  const [loading, setLoading] = useState(false);
  const [loadingRun, setLoadingRun] = useState(!!runIdParam);
  const [error, setError] = useState("");
  const [draftBanner, setDraftBanner] = useState(false);
  const [initialized, setInitialized] = useState(false);
  const [result, setResult] = useState<Awaited<
    ReturnType<typeof runCalculator>
  > | null>(null);
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (runIdParam) {
      setLoadingRun(true);
      getRun(runIdParam)
        .then((run) => {
          if (run.input && run.output) {
            setResult({
              run_id: run.id,
              input: run.input,
              output: run.output,
            });
            setStep(3);
          } else {
            setError(t("history.legacyRun"));
          }
        })
        .catch(() => setError(t("history.runNotFound")))
        .finally(() => {
          setLoadingRun(false);
          setInitialized(true);
        });
      return;
    }

    const draft = loadCalculatorDraft();
    if (draft) {
      setInput(draft.input);
      setStep(draft.step);
      setShowExpert(draft.showExpert);
      setDraftBanner(true);
    }
    setInitialized(true);
  }, [runIdParam, t]);

  useEffect(() => {
    if (!initialized || runIdParam || step === 3) return;
    if (saveTimer.current) clearTimeout(saveTimer.current);
    saveTimer.current = setTimeout(() => {
      saveCalculatorDraft({ input, step: step as 1 | 2, showExpert });
    }, 400);
    return () => {
      if (saveTimer.current) clearTimeout(saveTimer.current);
    };
  }, [input, step, showExpert, initialized, runIdParam]);

  function patch(partial: Partial<CalculatorInput>) {
    setInput((prev) => ({ ...prev, ...partial }));
  }

  function updateStep(idx: number, partial: Partial<ProcessStep>) {
    setInput((prev) => {
      const steps = [...(prev.process_steps || [])];
      steps[idx] = { ...steps[idx], ...partial };
      return { ...prev, process_steps: steps };
    });
  }

  function addStep() {
    setInput((prev) => ({
      ...prev,
      process_steps: [
        ...(prev.process_steps || []),
        { operation: "", executor: "", minutes_per_unit: 0, rate_rub_per_min: 0 },
      ],
    }));
  }

  function removeStep(idx: number) {
    setInput((prev) => ({
      ...prev,
      process_steps: (prev.process_steps || []).filter((_, i) => i !== idx),
    }));
  }

  const hmPreview = useMemo(() => {
    if (input.hm_mode === "direct") return input.hours_saved_month;
    return resolveHoursSaved(input);
  }, [input]);

  const financePreview = useMemo(
    () => previewCalculator({ ...input, hours_saved_month: hmPreview }),
    [input, hmPreview]
  );

  async function handleCalculate() {
    setError("");
    setLoading(true);
    const payload = {
      ...input,
      hours_saved_month: hmPreview,
    };
    try {
      const data = await runCalculator(payload);
      clearCalculatorDraft();
      setDraftBanner(false);
      setResult(data);
      setStep(3);
      router.replace(`/tools/calculator?run=${data.run_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("calcFailed"));
    } finally {
      setLoading(false);
    }
  }

  function handleNewCalculation() {
    clearCalculatorDraft();
    setDraftBanner(false);
    setResult(null);
    setInput(defaultCalculatorInput);
    setStep(1);
    setShowExpert(false);
    setError("");
    router.replace("/tools/calculator");
  }

  function dismissDraft() {
    clearCalculatorDraft();
    setDraftBanner(false);
    setInput(defaultCalculatorInput);
    setStep(1);
    setShowExpert(false);
  }

  const totalMin = totalMinutesPerUnit(input.process_steps || []);
  const costPerUnit = (input.process_steps || []).reduce(
    (s, r) => s + r.minutes_per_unit * r.rate_rub_per_min,
    0
  );

  if (loadingRun) {
    return (
      <div className="mx-auto max-w-5xl">
        <p className="text-sm text-text2">{t("history.loadingRun")}</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </div>

      {draftBanner && step !== 3 && (
        <div className="flex items-center justify-between rounded-lg border border-accent bg-accent-bg px-4 py-3 text-sm text-accent">
          <span>{t("draft.restored")}</span>
          <button type="button" onClick={dismissDraft} className="text-xs underline hover:no-underline">
            {t("draft.discard")}
          </button>
        </div>
      )}

      {error && step !== 3 && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}

      <WizardStepper step={step} labels={[t("wizard.step1"), t("wizard.step2"), t("wizard.step3")]} />

      {step === 1 && (
        <div className="rounded-xl border border-border bg-bg p-6 space-y-6">
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("fields.processName")}</label>
            <input
              required
              className={fieldClass}
              value={input.process_name}
              onChange={(e) => patch({ process_name: e.target.value })}
            />
          </div>

          <div className="flex items-center gap-2">
            <ModeButton
              active={input.hm_mode === "process"}
              onClick={() => patch({ hm_mode: "process" })}
              label={t("wizard.modeProcess")}
            />
            <ModeButton
              active={input.hm_mode === "direct"}
              onClick={() => patch({ hm_mode: "direct" })}
              label={t("wizard.modeDirect")}
            />
            <HelpTooltip tooltipKey="calculator.wizard.hm_mode_direct" />
          </div>

          {input.hm_mode === "process" ? (
            <>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead>
                    <tr className="border-b border-border">
                      <th className="pb-2 pr-2 normal-case">
                        <span className="inline-flex items-center gap-1 whitespace-nowrap text-[10px] font-medium uppercase tracking-wider text-text3">
                          {t("wizard.colOperation")}
                          <HelpTooltip tooltipKey="calculator.wizard.col_operation" />
                        </span>
                      </th>
                      <th className="pb-2 pr-2 normal-case">
                        <span className="inline-flex items-center gap-1 whitespace-nowrap text-[10px] font-medium uppercase tracking-wider text-text3">
                          {t("wizard.colExecutor")}
                          <HelpTooltip tooltipKey="calculator.wizard.col_executor" />
                        </span>
                      </th>
                      <th className="pb-2 pr-2 normal-case">
                        <span className="inline-flex items-center gap-1 whitespace-nowrap text-[10px] font-medium uppercase tracking-wider text-text3">
                          {t("wizard.colMinutes")}
                          <HelpTooltip tooltipKey="calculator.wizard.col_minutes" />
                        </span>
                      </th>
                      <th className="pb-2 pr-2 normal-case">
                        <span className="inline-flex items-center gap-1 whitespace-nowrap text-[10px] font-medium uppercase tracking-wider text-text3">
                          {t("wizard.colRate")}
                          <HelpTooltip tooltipKey="calculator.wizard.col_rate" />
                        </span>
                      </th>
                      <th className="pb-2 normal-case">
                        <span className="inline-flex items-center gap-1 whitespace-nowrap text-[10px] font-medium uppercase tracking-wider text-text3">
                          {t("wizard.colCost")}
                          <HelpTooltip tooltipKey="calculator.wizard.col_cost" />
                        </span>
                      </th>
                      <th className="pb-2 w-8" />
                    </tr>
                  </thead>
                  <tbody>
                    {(input.process_steps || []).map((row, idx) => (
                      <tr key={idx} className="border-b border-border">
                        <td className="py-2 pr-2">
                          <input className={fieldClass} value={row.operation} onChange={(e) => updateStep(idx, { operation: e.target.value })} />
                        </td>
                        <td className="py-2 pr-2">
                          <input className={fieldClass} value={row.executor || ""} onChange={(e) => updateStep(idx, { executor: e.target.value })} />
                        </td>
                        <td className="py-2 pr-2">
                          <input type="number" min={0} className={fieldClass} value={row.minutes_per_unit} onChange={(e) => updateStep(idx, { minutes_per_unit: Number(e.target.value) })} />
                        </td>
                        <td className="py-2 pr-2">
                          <input type="number" min={0} className={fieldClass} value={row.rate_rub_per_min} onChange={(e) => updateStep(idx, { rate_rub_per_min: Number(e.target.value) })} />
                        </td>
                        <td className="py-2 text-text2">
                          {Math.round(row.minutes_per_unit * row.rate_rub_per_min)} ₽
                        </td>
                        <td className="py-2">
                          {(input.process_steps?.length || 0) > 1 && (
                            <button type="button" className="text-text3 hover:text-red-600" onClick={() => removeStep(idx)}>×</button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <button type="button" onClick={addStep} className="text-sm text-accent hover:underline">{t("wizard.addStep")}</button>
              <div className="rounded-lg bg-bg2 px-4 py-3 text-sm text-text2">
                {t("wizard.totalPerUnit")}: {totalMin} {t("wizard.minShort")} · {costPerUnit} ₽
              </div>
              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <label className="mb-1.5 block text-xs font-medium">
                    <LabelWithHelp
                      label={t("wizard.unitsPerMonth")}
                      tooltipKey="calculator.wizard.units_per_month"
                    />
                  </label>
                  <input type="number" min={0} className={fieldClass} value={input.units_per_month} onChange={(e) => patch({ units_per_month: Number(e.target.value) })} />
                </div>
                <div>
                  <label className="mb-1.5 block text-xs font-medium">
                    <LabelWithHelp
                      label={t("wizard.automationPct")}
                      tooltipKey="calculator.wizard.automation_pct"
                    />
                  </label>
                  <input type="number" min={0} max={100} className={fieldClass} value={input.automation_pct} onChange={(e) => patch({ automation_pct: Number(e.target.value) })} />
                </div>
              </div>
            </>
          ) : (
            <div>
              <label className="mb-1.5 block text-xs font-medium">{t("wizard.hmDirect")}</label>
              <input type="number" min={0} className={fieldClass} value={input.hours_saved_month} onChange={(e) => patch({ hours_saved_month: Number(e.target.value) })} />
              <p className="mt-1 text-xs text-text3">{t("wizard.hmHint")}</p>
            </div>
          )}

          <div className="grid gap-4 sm:grid-cols-2 border-t border-border pt-4">
            <div>
              <label className="mb-1.5 block text-xs font-medium">
                <LabelWithHelp
                  label={t("wizard.errorRateBefore")}
                  tooltipKey="calculator.wizard.error_rate_before"
                />
              </label>
              <input type="number" min={0} max={100} className={fieldClass} value={input.error_rate_before_pct} onChange={(e) => patch({ error_rate_before_pct: Number(e.target.value) })} />
            </div>
            <div>
              <label className="mb-1.5 block text-xs font-medium">
                <LabelWithHelp
                  label={t("wizard.throughputBefore")}
                  tooltipKey="calculator.wizard.throughput_before"
                />
              </label>
              <input type="number" min={0} className={fieldClass} value={input.throughput_before_per_day} onChange={(e) => patch({ throughput_before_per_day: Number(e.target.value) })} />
            </div>
          </div>

          <div className="rounded-lg border border-accent bg-accent-bg px-4 py-3 text-sm">
            <span className="inline-flex items-center gap-1.5 font-medium text-accent">
              Hm:
              <HelpTooltip tooltipKey="calculator.wizard.hm_preview" />
            </span>{" "}
            {Math.round(hmPreview)} {t("wizard.hoursMonth")}
          </div>

          <div className="flex justify-end">
            <button type="button" onClick={() => setStep(2)} disabled={!input.process_name || hmPreview <= 0} className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-50">
              {t("wizard.next")}
            </button>
          </div>
        </div>
      )}

      {step === 2 && (
        <div className="grid gap-6 lg:grid-cols-5">
          <div className="lg:col-span-3 rounded-xl border border-border bg-bg p-6 space-y-4">
            <div>
              <label className="mb-1.5 block text-xs font-medium">
                <LabelWithHelp
                  label={t("wizard.hmLabel")}
                  tooltipKey="calculator.wizard.hm_label"
                />
              </label>
              <input type="number" min={0} className={fieldClass} value={Math.round(hmPreview * 100) / 100} onChange={(e) => patch({ hours_saved_month: Number(e.target.value), hm_mode: "direct" })} />
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field label={t("fields.ch")} hint={t("hints.ch")} tooltipKey="calculator.wizard.ch" type="number" value={input.hourly_cost_loaded} onChange={(v) => patch({ hourly_cost_loaded: v })} />
              <Field label={t("fields.om")} hint={t("hints.om")} tooltipKey="calculator.wizard.om" type="number" value={input.monthly_solution_cost} onChange={(v) => patch({ monthly_solution_cost: v })} />
              <Field label={t("fields.capex")} hint={t("hints.capex")} tooltipKey="calculator.wizard.capex" type="number" value={input.capex} onChange={(v) => patch({ capex: v })} />
              <div>
                <label className="mb-1.5 block text-xs font-medium">
                  <LabelWithHelp
                    label={t("fields.horizon")}
                    tooltipKey="calculator.wizard.horizon"
                  />
                </label>
                <select className={fieldClass} value={input.horizon_months} onChange={(e) => patch({ horizon_months: Number(e.target.value) })}>
                  <option value={12}>12</option>
                  <option value={24}>24</option>
                  <option value={36}>36</option>
                </select>
              </div>
            </div>

            <button type="button" onClick={() => setShowExpert(!showExpert)} className="text-xs text-accent hover:underline">
              {showExpert ? t("wizard.hideExpert") : t("wizard.showExpert")}
            </button>

            {showExpert && (
              <div className="grid gap-4 sm:grid-cols-2 border-t border-border pt-4">
                <Field label={t("fields.sm")} tooltipKey="calculator.wizard.sm" type="number" value={input.other_benefit_monthly} onChange={(v) => patch({ other_benefit_monthly: v })} />
                <Field label={t("fields.utilization")} tooltipKey="calculator.wizard.utilization" type="number" value={input.utilization} onChange={(v) => patch({ utilization: v })} step={0.01} />
                <Field label={t("fields.discount")} tooltipKey="calculator.wizard.discount" type="number" value={input.discount_rate_annual} onChange={(v) => patch({ discount_rate_annual: v })} step={0.01} />
              </div>
            )}

            {error && <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>}

            <div className="flex gap-2 pt-2">
              <button type="button" onClick={() => setStep(1)} className="rounded-lg border border-border2 px-4 py-2 text-sm">{t("wizard.back")}</button>
              <button type="button" onClick={handleCalculate} disabled={loading} className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-60">
                {loading ? t("calculating") : t("calculate")}
              </button>
            </div>
          </div>

          <div className="lg:col-span-2 rounded-xl border border-border bg-bg p-5 space-y-3 h-fit">
            <p className="text-[10px] font-medium uppercase tracking-wider text-text3">{t("wizard.preview")}</p>
            <PreviewRow label={t("metrics.netBenefit")} value={formatRub(financePreview.net_benefit_monthly)} />
            <PreviewRow label={t("metrics.payback")} value={formatPayback(financePreview.payback_months, t)} />
            <PreviewRow label={t("metrics.fte")} value={financePreview.fte.toFixed(2)} tooltipKey="calculator.wizard.fte" />
            <PreviewRow label={t("metrics.roi")} value={`${financePreview.roi_horizon_pct.toFixed(1)}%`} />
          </div>
        </div>
      )}

      {step === 3 && result && (
        <div className="space-y-4">
          <button type="button" onClick={handleNewCalculation} className="text-sm text-accent hover:underline">{t("wizard.recalculate")}</button>
          <CalculatorResult runId={result.run_id} input={result.input} output={result.output} />
        </div>
      )}
    </div>
  );
}

function WizardStepper({ step, labels }: { step: number; labels: string[] }) {
  return (
    <div className="flex justify-center">
      <div className="flex items-center">
        {labels.map((label, i) => {
          const n = i + 1;
          return (
            <div key={label} className="flex items-center">
              <div className="flex items-center gap-2">
                <div
                  className={cn(
                    "flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-medium",
                    n <= step ? "bg-text text-white" : "bg-bg2 text-text3"
                  )}
                >
                  {n}
                </div>
                <span
                  className={cn(
                    "hidden text-xs sm:inline whitespace-nowrap",
                    n === step ? "font-medium text-text" : "text-text3"
                  )}
                >
                  {label}
                </span>
              </div>
              {i < labels.length - 1 && (
                <div className="mx-4 h-px w-12 bg-border sm:mx-6 sm:w-20" />
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

function ModeButton({ active, onClick, label }: { active: boolean; onClick: () => void; label: string }) {
  return (
    <button type="button" onClick={onClick} className={cn("rounded-lg border px-3 py-2 text-xs", active ? "border-text bg-bg2 font-medium" : "border-border text-text2")}>{label}</button>
  );
}

function Field({ label, hint, tooltipKey, value, onChange, type = "text", step }: { label: string; hint?: string; tooltipKey?: string; value: number; onChange: (v: number) => void; type?: string; step?: number }) {
  return (
    <div>
      <label className="mb-1.5 block text-xs font-medium">
        {tooltipKey ? (
          <LabelWithHelp label={label} tooltipKey={tooltipKey} />
        ) : (
          label
        )}
      </label>
      <input type={type} min={0} step={step} className={fieldClass} value={value} onChange={(e) => onChange(Number(e.target.value))} />
      {hint && <p className="mt-1 text-[10px] text-text3">{hint}</p>}
    </div>
  );
}

function PreviewRow({ label, value, tooltipKey }: { label: string; value: string; tooltipKey?: string }) {
  return (
    <div>
      <p className="text-[10px] font-medium uppercase tracking-wider text-text3">
        {tooltipKey ? (
          <span className="inline-flex items-center gap-1 normal-case">
            <span className="uppercase tracking-wider">{label}</span>
            <HelpTooltip tooltipKey={tooltipKey} />
          </span>
        ) : (
          label
        )}
      </p>
      <p className="text-lg font-medium">{value}</p>
    </div>
  );
}

function formatRub(n: number) {
  if (!isFinite(n)) return "—";
  return new Intl.NumberFormat("ru-RU").format(Math.round(n)) + " ₽";
}

function formatPayback(n: number, t: ReturnType<typeof useTranslations>) {
  if (!isFinite(n) || n > 1e6) return "—";
  return `${n.toFixed(1)} ${t("metrics.months")}`;
}
