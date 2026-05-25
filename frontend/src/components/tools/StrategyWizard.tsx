"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchTools,
  getRun,
  isToolLimitError,
  listRuns,
  runStrategy,
  type RunListItem,
  type ToolListItem,
} from "@/lib/api";
import { isLimitReached } from "@/lib/tool-limits";
import {
  DEFAULT_STRATEGY_INPUT,
  MAX_CALCULATOR_LINKS,
  STRATEGY_GOAL_OPTIONS,
  STRATEGY_INDUSTRY_OPTIONS,
  STRATEGY_KEY_PROCESS_OPTIONS,
  STRATEGY_MARKET_OPTIONS,
  type StrategyInput,
  type StrategyOutput,
} from "@/lib/api-strategy";
import { ToolLimitBadge } from "@/components/dashboard/ToolLimitBadge";
import { ToolLimitExceededAlert } from "./ToolLimitExceededAlert";
import { StrategyStreamView } from "./StrategyStreamView";
import { StrategyResult } from "./StrategyResult";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

function StrategyWizardInner() {
  const t = useTranslations("strategy");
  const tLimits = useTranslations("toolLimits");
  const router = useRouter();
  const searchParams = useSearchParams();
  const runIdParam = searchParams.get("run");

  const [formStep, setFormStep] = useState<1 | 2>(1);
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [input, setInput] = useState<StrategyInput>(DEFAULT_STRATEGY_INPUT);
  const [runId, setRunId] = useState<string | null>(null);
  const [result, setResult] = useState<{ input: StrategyInput; output: StrategyOutput; runId: string } | null>(null);
  const [loading, setLoading] = useState(false);
  const [loadingRun, setLoadingRun] = useState(!!runIdParam);
  const [error, setError] = useState("");
  const [limitExceeded, setLimitExceeded] = useState(false);
  const [strategyTool, setStrategyTool] = useState<ToolListItem | null>(null);
  const [calcRuns, setCalcRuns] = useState<RunListItem[]>([]);

  function loadToolLimits() {
    return fetchTools()
      .then((data) => {
        const tool = data.tools.find((x) => x.slug === "strategy") ?? null;
        setStrategyTool(tool);
        if (tool) setLimitExceeded(isLimitReached(tool));
        return tool;
      })
      .catch(() => null);
  }

  useEffect(() => {
    void loadToolLimits();
    listRuns({ tool_slug: "calculator", limit: 50 })
      .then((data) => setCalcRuns(data.items.filter((r) => r.status === "done")))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!runIdParam) {
      setLoadingRun(false);
      return;
    }
    setLoadingRun(true);
    getRun(runIdParam)
      .then((run) => {
        if (run.status === "done" && run.input && run.output) {
          setResult({
            input: run.input as StrategyInput,
            output: run.output as StrategyOutput,
            runId: run.id,
          });
          setStep(3);
        } else if (run.status === "pending" || run.status === "processing") {
          setRunId(run.id);
          setStep(2);
        } else if (run.status === "error") {
          setError(run.error_msg || t("errors.loadFailed"));
        }
      })
      .catch(() => setError(t("errors.loadFailed")))
      .finally(() => setLoadingRun(false));
  }, [runIdParam, t]);

  function patch(partial: Partial<StrategyInput>) {
    setInput((prev) => ({ ...prev, ...partial }));
  }

  function toggleGoal(goal: string) {
    setInput((prev) => {
      const has = prev.main_goals.includes(goal);
      return {
        ...prev,
        main_goals: has ? prev.main_goals.filter((g) => g !== goal) : [...prev.main_goals, goal],
      };
    });
  }

  function toggleProcess(proc: string) {
    setInput((prev) => {
      const has = prev.key_processes.includes(proc);
      return {
        ...prev,
        key_processes: has ? prev.key_processes.filter((p) => p !== proc) : [...prev.key_processes, proc],
      };
    });
  }

  function toggleCalcRun(id: string) {
    setInput((prev) => {
      const ids = prev.calculator_run_ids ?? [];
      const has = ids.includes(id);
      if (has) {
        return { ...prev, calculator_run_ids: ids.filter((x) => x !== id) };
      }
      if (ids.length >= MAX_CALCULATOR_LINKS) return prev;
      return { ...prev, calculator_run_ids: [...ids, id] };
    });
  }

  const step1Valid =
    input.company_name.trim() &&
    input.business_description.trim() &&
    input.industry;

  const step2Valid =
    input.main_goals.length > 0 &&
    input.pain_points.trim();

  const strategyLimitReached = strategyTool ? isLimitReached(strategyTool) : limitExceeded;

  async function handleGenerate() {
    if (strategyLimitReached) {
      setLimitExceeded(true);
      return;
    }
    setError("");
    setLimitExceeded(false);
    setLoading(true);
    try {
      const payload: StrategyInput = {
        ...input,
        calculator_run_ids: input.calculator_run_ids?.length ? input.calculator_run_ids : undefined,
      };
      const data = await runStrategy(payload);
      setRunId(data.run_id);
      setStep(2);
      router.replace(`/tools/strategy?run=${data.run_id}`);
      void loadToolLimits();
    } catch (err) {
      if (isToolLimitError(err)) {
        setLimitExceeded(true);
        setError("");
      } else {
        setError(err instanceof Error ? err.message : t("errors.startFailed"));
      }
    } finally {
      setLoading(false);
    }
  }

  async function handleStreamComplete() {
    if (!runId) return;
    try {
      const run = await getRun(runId);
      if (run.status === "done" && run.input && run.output) {
        setResult({
          input: run.input as StrategyInput,
          output: run.output as StrategyOutput,
          runId: run.id,
        });
        setStep(3);
      } else if (run.status === "error") {
        setError(run.error_msg || t("errors.loadFailed"));
        setStep(1);
      }
    } catch {
      setError(t("errors.loadFailed"));
      setStep(1);
    }
  }

  function handleNew() {
    setInput(DEFAULT_STRATEGY_INPUT);
    setFormStep(1);
    setRunId(null);
    setResult(null);
    setStep(1);
    setError("");
    setLimitExceeded(false);
    void loadToolLimits();
    router.replace("/tools/strategy");
  }

  if (loadingRun) {
    return <p className="text-sm text-text2">{t("loadingRun")}</p>;
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-base font-medium">{t("title")}</h1>
          <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
        </div>
        {strategyTool && (
          <ToolLimitBadge
            tool={strategyTool}
            t={(key, values) => tLimits(key, values as Record<string, string | number> | undefined)}
          />
        )}
      </div>

      {(limitExceeded || strategyLimitReached) && step === 1 && (
        <ToolLimitExceededAlert toolName={t("title")} />
      )}

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}

      {step === 1 && (
        <div className="rounded-xl border border-border bg-bg p-6 space-y-6">
          <div className="flex gap-2">
            <FormStepBadge n={1} active={formStep === 1} done={formStep > 1} label={t("formSteps.company")} />
            <FormStepBadge n={2} active={formStep === 2} done={false} label={t("formSteps.goals")} />
          </div>

          {formStep === 1 && (
            <>
              <SectionLabel>{t("form.companySection")}</SectionLabel>
              <div className="grid gap-4 sm:grid-cols-2">
                <Field label={t("form.companyName")} className="sm:col-span-2">
                  <input className={fieldClass} value={input.company_name} onChange={(e) => patch({ company_name: e.target.value })} />
                </Field>
                <Field label={t("form.businessDescription")} className="sm:col-span-2">
                  <textarea
                    rows={3}
                    className={fieldClass}
                    placeholder={t("form.businessDescriptionPlaceholder")}
                    value={input.business_description}
                    onChange={(e) => patch({ business_description: e.target.value })}
                  />
                </Field>
                <Field label={t("form.industry")}>
                  <select className={fieldClass} value={input.industry} onChange={(e) => patch({ industry: e.target.value })}>
                    <option value="">{t("form.select")}</option>
                    {STRATEGY_INDUSTRY_OPTIONS.map((o) => (
                      <option key={o} value={o}>{t(`industries.${o}`)}</option>
                    ))}
                  </select>
                </Field>
                <Field label={t("form.companySize")}>
                  <select className={fieldClass} value={input.company_size} onChange={(e) => patch({ company_size: e.target.value as StrategyInput["company_size"] })}>
                    {(["1-10", "11-50", "51-200", "201-1000", "1000+"] as const).map((s) => (
                      <option key={s} value={s}>{s}</option>
                    ))}
                  </select>
                </Field>
                <Field label={t("form.revenue")}>
                  <select className={fieldClass} value={input.annual_revenue_range} onChange={(e) => patch({ annual_revenue_range: e.target.value })}>
                    <option value="до 10M">{t("revenue.up10m")}</option>
                    <option value="10-100M">{t("revenue.10to100m")}</option>
                    <option value="100M-1B">{t("revenue.100mto1b")}</option>
                    <option value="1B+">{t("revenue.1bplus")}</option>
                  </select>
                </Field>
                <Field label={t("form.marketPosition")}>
                  <select className={fieldClass} value={input.market_position} onChange={(e) => patch({ market_position: e.target.value })}>
                    {STRATEGY_MARKET_OPTIONS.map((m) => (
                      <option key={m} value={m}>{t(`market.${m}`)}</option>
                    ))}
                  </select>
                </Field>
                <Field label={t("form.aiLevel")}>
                  <select className={fieldClass} value={input.current_ai_level} onChange={(e) => patch({ current_ai_level: e.target.value as StrategyInput["current_ai_level"] })}>
                    {(["none", "exploring", "piloting", "scaling"] as const).map((l) => (
                      <option key={l} value={l}>{t(`aiLevels.${l}`)}</option>
                    ))}
                  </select>
                </Field>
              </div>
              <div className="flex justify-end pt-2">
                <button
                  type="button"
                  disabled={!step1Valid}
                  onClick={() => setFormStep(2)}
                  className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
                >
                  {t("form.next")}
                </button>
              </div>
            </>
          )}

          {formStep === 2 && (
            <>
              <SectionLabel>{t("form.goalsSection")}</SectionLabel>
              <div className="flex flex-wrap gap-2">
                {STRATEGY_GOAL_OPTIONS.map((goal) => (
                  <Chip key={goal} active={input.main_goals.includes(goal)} onClick={() => toggleGoal(goal)}>
                    {t(`goals.${goal}`)}
                  </Chip>
                ))}
              </div>

              <SectionLabel>{t("form.processesSection")}</SectionLabel>
              <div className="flex flex-wrap gap-2">
                {STRATEGY_KEY_PROCESS_OPTIONS.map((proc) => (
                  <Chip key={proc} active={input.key_processes.includes(proc)} onClick={() => toggleProcess(proc)}>
                    {t(`processes.${proc}`)}
                  </Chip>
                ))}
              </div>

              <Field label={t("form.painPoints")}>
                <textarea rows={4} className={fieldClass} value={input.pain_points} onChange={(e) => patch({ pain_points: e.target.value })} />
              </Field>

              <SectionLabel>{t("form.constraintsSection")}</SectionLabel>
              <div className="grid gap-4 sm:grid-cols-2">
                <Field label={t("form.dataMaturity")}>
                  <select className={fieldClass} value={input.data_maturity} onChange={(e) => patch({ data_maturity: e.target.value as StrategyInput["data_maturity"] })}>
                    {(["none", "basic", "team"] as const).map((d) => (
                      <option key={d} value={d}>{t(`dataMaturity.${d}`)}</option>
                    ))}
                  </select>
                </Field>
                <Field label={t("form.changeReadiness")}>
                  <select className={fieldClass} value={input.change_readiness} onChange={(e) => patch({ change_readiness: e.target.value as StrategyInput["change_readiness"] })}>
                    {(["low", "medium", "high"] as const).map((c) => (
                      <option key={c} value={c}>{t(`changeReadiness.${c}`)}</option>
                    ))}
                  </select>
                </Field>
                <Field label={t("form.budget")}>
                  <select className={fieldClass} value={input.budget_range} onChange={(e) => patch({ budget_range: e.target.value })}>
                    <option value="до 500K">{t("budget.up500k")}</option>
                    <option value="500K-5M">{t("budget.500kto5m")}</option>
                    <option value="5M+">{t("budget.5mplus")}</option>
                  </select>
                </Field>
                <Field label={t("form.timeline")}>
                  <select className={fieldClass} value={input.timeline} onChange={(e) => patch({ timeline: e.target.value as StrategyInput["timeline"] })}>
                    {(["3months", "6months", "1year", "2years"] as const).map((tl) => (
                      <option key={tl} value={tl}>{t(`timelines.${tl}`)}</option>
                    ))}
                  </select>
                </Field>
              </div>

              <Field label={t("form.existingTools")}>
                <textarea rows={3} className={fieldClass} value={input.existing_tools} onChange={(e) => patch({ existing_tools: e.target.value })} />
              </Field>

              <SectionLabel>{t("form.reportModeSection")}</SectionLabel>
              <div className="grid gap-3 sm:grid-cols-2">
                {(["consulting", "standard"] as const).map((mode) => (
                  <label
                    key={mode}
                    className={cn(
                      "cursor-pointer rounded-lg border p-4 transition-colors",
                      input.report_mode === mode ? "border-ai bg-ai-bg/40" : "border-border hover:border-border2"
                    )}
                  >
                    <input
                      type="radio"
                      name="report_mode"
                      className="sr-only"
                      checked={input.report_mode === mode}
                      onChange={() => patch({ report_mode: mode })}
                    />
                    <p className="text-sm font-medium text-text">{t(`reportMode.${mode}.title`)}</p>
                    <p className="mt-1 text-xs text-text2">{t(`reportMode.${mode}.desc`)}</p>
                  </label>
                ))}
              </div>

              {calcRuns.length > 0 && (
                <div>
                  <SectionLabel>{t("form.calculatorRuns")}</SectionLabel>
                  <p className="mb-3 text-xs text-text2">{t("form.calculatorRunsHint", { max: MAX_CALCULATOR_LINKS })}</p>
                  <div className="space-y-2 max-h-56 overflow-y-auto">
                    {calcRuns.map((r) => {
                      const selected = input.calculator_run_ids?.includes(r.id);
                      const atLimit = (input.calculator_run_ids?.length ?? 0) >= MAX_CALCULATOR_LINKS && !selected;
                      return (
                        <label
                          key={r.id}
                          className={cn(
                            "flex cursor-pointer items-start gap-3 rounded-lg border p-3 text-sm transition-colors",
                            selected ? "border-ai bg-ai-bg/40" : "border-border hover:border-border2",
                            atLimit && "cursor-not-allowed opacity-50"
                          )}
                        >
                          <input
                            type="checkbox"
                            className="mt-0.5"
                            checked={!!selected}
                            disabled={atLimit}
                            onChange={() => toggleCalcRun(r.id)}
                          />
                          <span className="flex-1">
                            <span className="font-medium text-text">{r.process_name || r.id.slice(0, 8)}</span>
                            <span className="mt-0.5 block text-xs text-text2">
                              {new Date(r.created_at).toLocaleDateString()}
                              {r.net_benefit_monthly != null && ` · ${Math.round(r.net_benefit_monthly).toLocaleString()} ₽/мес`}
                              {r.payback_months != null && ` · ${r.payback_months.toFixed(1)} мес`}
                            </span>
                          </span>
                        </label>
                      );
                    })}
                  </div>
                </div>
              )}

              <div className="flex justify-between pt-2">
                <button type="button" onClick={() => setFormStep(1)} className="rounded-lg border border-border2 px-4 py-2 text-sm text-text2">
                  {t("form.back")}
                </button>
                <button
                  type="button"
                  disabled={loading || !step2Valid || strategyLimitReached}
                  onClick={handleGenerate}
                  className="rounded-lg bg-ai px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
                >
                  {loading ? t("generating") : t("generate")}
                </button>
              </div>
              {strategyLimitReached && (
                <p className="text-xs text-text2 text-right">
                  {t("limits.noRunsLeft")}{" "}
                  <Link href="/billing" className="text-accent underline-offset-2 hover:underline">
                    {tLimits("upgradeCta")}
                  </Link>
                </p>
              )}
            </>
          )}
        </div>
      )}

      {step === 2 && runId && (
        <StrategyStreamView
          runId={runId}
          onComplete={handleStreamComplete}
          onError={(msg) => { setError(msg); setStep(1); }}
        />
      )}

      {step === 3 && result && (
        <div className="space-y-4">
          <button type="button" onClick={handleNew} className="text-sm text-accent hover:underline">
            {t("newStrategy")}
          </button>
          <StrategyResult input={result.input} output={result.output} runId={result.runId} />
        </div>
      )}
    </div>
  );
}

export function StrategyWizard() {
  return (
    <Suspense fallback={null}>
      <StrategyWizardInner />
    </Suspense>
  );
}

function FormStepBadge({ n, active, done, label }: { n: number; active: boolean; done: boolean; label: string }) {
  return (
    <div className={cn("flex items-center gap-2 rounded-lg px-3 py-1.5 text-xs", active ? "bg-bg2 font-medium text-text" : "text-text3")}>
      <span className={cn("flex h-5 w-5 items-center justify-center rounded-full text-[10px]", done ? "bg-green-100 text-green-800" : active ? "bg-ai text-white" : "bg-bg2")}>
        {done ? "✓" : n}
      </span>
      {label}
    </div>
  );
}

function Chip({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "rounded-lg border px-3 py-1.5 text-xs",
        active ? "border-ai bg-ai-bg font-medium text-ai" : "border-border2 text-text2 hover:border-border"
      )}
    >
      {children}
    </button>
  );
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <p className="border-b border-border pb-2 text-[10px] font-medium uppercase tracking-wider text-text3">
      {children}
    </p>
  );
}

function Field({ label, children, className }: { label: string; children: React.ReactNode; className?: string }) {
  return (
    <div className={className}>
      <label className="mb-1.5 block text-xs font-medium">{label}</label>
      {children}
    </div>
  );
}
