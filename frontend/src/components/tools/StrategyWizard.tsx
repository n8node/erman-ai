"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchTools,
  getRun,
  listRuns,
  runStrategy,
  type RunDetail,
  type RunListItem,
} from "@/lib/api";
import {
  DEFAULT_STRATEGY_INPUT,
  STRATEGY_GOAL_OPTIONS,
  STRATEGY_INDUSTRY_OPTIONS,
  type StrategyInput,
  type StrategyOutput,
} from "@/lib/api-strategy";
import { RunStatusPoller } from "./RunStatusPoller";
import { StrategyResult } from "./StrategyResult";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

function StrategyWizardInner() {
  const t = useTranslations("strategy");
  const router = useRouter();
  const searchParams = useSearchParams();
  const runIdParam = searchParams.get("run");

  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [input, setInput] = useState<StrategyInput>(DEFAULT_STRATEGY_INPUT);
  const [runId, setRunId] = useState<string | null>(null);
  const [result, setResult] = useState<{ input: StrategyInput; output: StrategyOutput; runId: string } | null>(null);
  const [loading, setLoading] = useState(false);
  const [loadingRun, setLoadingRun] = useState(!!runIdParam);
  const [error, setError] = useState("");
  const [runsLeft, setRunsLeft] = useState<string>("");
  const [calcRuns, setCalcRuns] = useState<RunListItem[]>([]);

  useEffect(() => {
    fetchTools()
      .then((data) => {
        const tool = data.tools.find((x) => x.slug === "strategy");
        if (tool) {
          setRunsLeft(
            tool.runs_limit === -1
              ? t("limits.unlimited")
              : t("limits.remaining", {
                  left: Math.max(0, tool.runs_limit - tool.runs_used),
                  total: tool.runs_limit,
                })
          );
        }
      })
      .catch(() => {});
    listRuns({ tool_slug: "calculator", limit: 20 })
      .then((data) => setCalcRuns(data.items))
      .catch(() => {});
  }, [t]);

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
        } else {
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
        main_goals: has
          ? prev.main_goals.filter((g) => g !== goal)
          : [...prev.main_goals, goal],
      };
    });
  }

  async function handleGenerate() {
    setError("");
    setLoading(true);
    try {
      const payload = {
        ...input,
        calculator_run_id: input.calculator_run_id || undefined,
      };
      const data = await runStrategy(payload);
      setRunId(data.run_id);
      setStep(2);
      router.replace(`/tools/strategy?run=${data.run_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.startFailed"));
    } finally {
      setLoading(false);
    }
  }

  function handlePollComplete(run: RunDetail & { output?: StrategyOutput }) {
    if (!run.input || !run.output) {
      setError(t("errors.loadFailed"));
      return;
    }
    setResult({
      input: run.input as StrategyInput,
      output: run.output as StrategyOutput,
      runId: run.id,
    });
    setStep(3);
  }

  function handleNew() {
    setInput(DEFAULT_STRATEGY_INPUT);
    setRunId(null);
    setResult(null);
    setStep(1);
    setError("");
    router.replace("/tools/strategy");
  }

  if (loadingRun) {
    return <p className="text-sm text-text2">{t("loadingRun")}</p>;
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-base font-medium">{t("title")}</h1>
          <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
        </div>
        {runsLeft && (
          <span className="rounded-lg bg-ai-bg px-3 py-1 text-xs font-medium text-ai">{runsLeft}</span>
        )}
      </div>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}

      {step === 1 && (
        <div className="rounded-xl border border-border bg-bg p-6 space-y-6">
          <SectionLabel>{t("form.companySection")}</SectionLabel>
          <div className="grid gap-4 sm:grid-cols-2">
            <Field label={t("form.companyName")}>
              <input className={fieldClass} value={input.company_name} onChange={(e) => patch({ company_name: e.target.value })} />
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
            <Field label={t("form.aiLevel")}>
              <select className={fieldClass} value={input.current_ai_level} onChange={(e) => patch({ current_ai_level: e.target.value as StrategyInput["current_ai_level"] })}>
                {(["none", "exploring", "piloting", "scaling"] as const).map((l) => (
                  <option key={l} value={l}>{t(`aiLevels.${l}`)}</option>
                ))}
              </select>
            </Field>
          </div>

          <SectionLabel>{t("form.goalsSection")}</SectionLabel>
          <div className="flex flex-wrap gap-2">
            {STRATEGY_GOAL_OPTIONS.map((goal) => (
              <button
                key={goal}
                type="button"
                onClick={() => toggleGoal(goal)}
                className={cn(
                  "rounded-lg border px-3 py-1.5 text-xs",
                  input.main_goals.includes(goal)
                    ? "border-ai bg-ai-bg font-medium text-ai"
                    : "border-border2 text-text2 hover:border-border"
                )}
              >
                {t(`goals.${goal}`)}
              </button>
            ))}
          </div>

          <Field label={t("form.painPoints")}>
            <textarea rows={4} className={fieldClass} value={input.pain_points} onChange={(e) => patch({ pain_points: e.target.value })} />
          </Field>

          <SectionLabel>{t("form.constraintsSection")}</SectionLabel>
          <div className="grid gap-4 sm:grid-cols-2">
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

          {calcRuns.length > 0 && (
            <Field label={t("form.calculatorRun")}>
              <select
                className={fieldClass}
                value={input.calculator_run_id || ""}
                onChange={(e) => patch({ calculator_run_id: e.target.value || null })}
              >
                <option value="">{t("form.noCalculatorRun")}</option>
                {calcRuns.map((r) => (
                  <option key={r.id} value={r.id}>
                    {r.process_name || r.id.slice(0, 8)} — {new Date(r.created_at).toLocaleDateString()}
                  </option>
                ))}
              </select>
            </Field>
          )}

          <div className="flex justify-end pt-2">
            <button
              type="button"
              disabled={loading || !input.company_name || !input.industry || input.main_goals.length === 0 || !input.pain_points}
              onClick={handleGenerate}
              className="rounded-lg bg-ai px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
            >
              {loading ? t("generating") : t("generate")}
            </button>
          </div>
        </div>
      )}

      {step === 2 && runId && (
        <RunStatusPoller runId={runId} onComplete={handlePollComplete} onError={(msg) => { setError(msg); setStep(1); }} />
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

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <p className="border-b border-border pb-2 text-[10px] font-medium uppercase tracking-wider text-text3">
      {children}
    </p>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="mb-1.5 block text-xs font-medium">{label}</label>
      {children}
    </div>
  );
}
