"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchTools,
  fetchMe,
  getRun,
  isToolLimitError,
  listRuns,
  runProposal,
  type RunListItem,
  type ToolListItem,
} from "@/lib/api";
import {
  DEFAULT_PROPOSAL_INPUT,
  PAYMENT_SCHEDULE_OPTIONS,
  PROPOSAL_INDUSTRY_OPTIONS,
  PROPOSAL_SCENARIOS,
  type ProposalInput,
  type ProposalOutput,
  type ProposalScenario,
} from "@/lib/api-proposal";
import { isLimitReached } from "@/lib/tool-limits";
import { ToolLimitBadge } from "@/components/dashboard/ToolLimitBadge";
import { HelpTooltip, LabelWithHelp } from "@/components/ui/HelpTooltip";
import { ToolLimitExceededAlert } from "./ToolLimitExceededAlert";
import { ProposalPollingView } from "./ProposalPollingView";
import { ProposalResult } from "./ProposalResult";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

function scenarioDefaults(scenario: ProposalScenario): Partial<ProposalInput> {
  if (scenario === "cold_outreach") {
    return { include_pricing: false };
  }
  return { include_pricing: true };
}

function ProposalWizardInner() {
  const t = useTranslations("proposal");
  const tLimits = useTranslations("toolLimits");
  const router = useRouter();
  const searchParams = useSearchParams();
  const runIdParam = searchParams.get("run");
  const calcRunParam = searchParams.get("calculator_run_id");

  const [input, setInput] = useState<ProposalInput>(DEFAULT_PROPOSAL_INPUT);
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [runId, setRunId] = useState<string | null>(null);
  const [result, setResult] = useState<{ input: ProposalInput; output: ProposalOutput; runId: string } | null>(null);
  const [loading, setLoading] = useState(false);
  const [loadingRun, setLoadingRun] = useState(!!runIdParam);
  const [error, setError] = useState("");
  const [limitExceeded, setLimitExceeded] = useState(false);
  const [proposalTool, setProposalTool] = useState<ToolListItem | null>(null);
  const [calcRuns, setCalcRuns] = useState<RunListItem[]>([]);

  function loadToolLimits() {
    return fetchTools()
      .then((data) => {
        const tool = data.tools.find((x) => x.slug === "proposal") ?? null;
        setProposalTool(tool);
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
    fetchMe()
      .then((user) => {
        if (user?.email) {
          setInput((prev) => ({ ...prev, sender_email: prev.sender_email || user.email }));
        }
      })
      .catch(() => undefined);
  }, []);

  useEffect(() => {
    if (!calcRunParam) return;
    setInput((prev) => ({
      ...prev,
      calculator_run_id: calcRunParam,
      proposal_scenario: "proactive_offer",
      include_pricing: true,
    }));
    getRun(calcRunParam)
      .then((run) => {
        if (run.status !== "done" || !run.input || !run.output) return;
        const calcIn = run.input as { process_name?: string; capex?: number };
        const calcOut = run.output as { net_benefit_monthly?: number };
        setInput((prev) => ({
          ...prev,
          calculator_run_id: calcRunParam,
          proposal_scenario: "proactive_offer",
          include_pricing: true,
          solution_name: prev.solution_name || `Автоматизация: ${calcIn.process_name || "процесс"}`,
          project_cost_rub: prev.project_cost_rub || calcIn.capex || 0,
          client_problem:
            prev.client_problem ||
            `Процесс «${calcIn.process_name || ""}»: потенциальная экономия ${Math.round(calcOut.net_benefit_monthly || 0).toLocaleString("ru-RU")} ₽/мес после автоматизации.`,
        }));
      })
      .catch(() => undefined);
  }, [calcRunParam]);

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
            input: run.input as unknown as ProposalInput,
            output: run.output as unknown as ProposalOutput,
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

  function patch(partial: Partial<ProposalInput>) {
    setInput((prev) => ({ ...prev, ...partial }));
  }

  function selectScenario(scenario: ProposalScenario) {
    setInput((prev) => ({
      ...prev,
      proposal_scenario: scenario,
      ...scenarioDefaults(scenario),
    }));
  }

  function setDeliverable(index: number, value: string) {
    setInput((prev) => {
      const next = [...prev.deliverables];
      next[index] = value;
      return { ...prev, deliverables: next };
    });
  }

  function addDeliverable() {
    setInput((prev) => ({ ...prev, deliverables: [...prev.deliverables, ""] }));
  }

  function removeDeliverable(index: number) {
    setInput((prev) => ({
      ...prev,
      deliverables: prev.deliverables.filter((_, i) => i !== index),
    }));
  }

  const scenario = input.proposal_scenario;
  const pricingRequired = input.include_pricing;

  const formValid =
    input.client_company.trim() &&
    input.client_problem.trim() &&
    input.solution_name.trim() &&
    input.solution_description.trim() &&
    input.deliverables.some((d) => d.trim()) &&
    (!pricingRequired || input.project_cost_rub > 0) &&
    input.timeline_weeks > 0 &&
    (!pricingRequired || input.payment_schedule.trim()) &&
    input.sender_company.trim() &&
    input.sender_contact.trim() &&
    input.sender_email.trim() &&
    (scenario !== "after_contact" || input.prior_contact_summary?.trim()) &&
    (scenario !== "cold_outreach" || input.problem_source?.trim());

  const proposalLimitReached = proposalTool ? isLimitReached(proposalTool) : limitExceeded;

  async function handleGenerate() {
    if (proposalLimitReached) {
      setLimitExceeded(true);
      return;
    }
    setError("");
    setLimitExceeded(false);
    setLoading(true);
    try {
      const payload: ProposalInput = {
        ...input,
        deliverables: input.deliverables.map((d) => d.trim()).filter(Boolean),
        calculator_run_id: input.calculator_run_id || undefined,
        prior_contact_summary: input.prior_contact_summary?.trim() || undefined,
        problem_source: input.problem_source?.trim() || undefined,
      };
      const data = await runProposal(payload);
      setRunId(data.run_id);
      setStep(2);
      router.replace(`/tools/proposal?run=${data.run_id}`);
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

  async function handlePollComplete() {
    if (!runId) return;
    try {
      const run = await getRun(runId);
      if (run.status === "done" && run.input && run.output) {
        setResult({
          input: run.input as unknown as ProposalInput,
          output: run.output as unknown as ProposalOutput,
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
    setInput({ ...DEFAULT_PROPOSAL_INPUT });
    setRunId(null);
    setResult(null);
    setStep(1);
    setError("");
    setLimitExceeded(false);
    void loadToolLimits();
    fetchMe()
      .then((user) => {
        if (user?.email) setInput((prev) => ({ ...prev, sender_email: user.email }));
      })
      .catch(() => undefined);
    router.replace("/tools/proposal");
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
        {proposalTool && (
          <ToolLimitBadge
            tool={proposalTool}
            t={(key, values) => tLimits(key, values as Record<string, string | number> | undefined)}
          />
        )}
      </div>

      {(limitExceeded || proposalLimitReached) && (
        <ToolLimitExceededAlert toolName={t("title")} />
      )}

      {step === 1 && (
        <div className="space-y-6">
          <FormSection
            label={
              <LabelWithHelp label={t("form.scenarioSection")} tooltipKey="proposal.wizard.scenario" />
            }
          >
            <div className="grid gap-3 sm:grid-cols-3">
              {PROPOSAL_SCENARIOS.map((key) => (
                <button
                  key={key}
                  type="button"
                  onClick={() => selectScenario(key)}
                  className={cn(
                    "rounded-xl border p-4 text-left transition-colors",
                    scenario === key
                      ? "border-ai bg-ai-bg"
                      : "border-border hover:border-border2"
                  )}
                >
                  <div className="flex items-start justify-between gap-2">
                    <p className="text-sm font-medium text-text">{t(`scenarios.${key}.title`)}</p>
                    <HelpTooltip tooltipKey={`proposal.wizard.scenario_${key}`} />
                  </div>
                  <p className="mt-1.5 text-xs text-text2 leading-relaxed">{t(`scenarios.${key}.description`)}</p>
                </button>
              ))}
            </div>
          </FormSection>

          <FormSection label={t("form.clientSection")}>
            <Field label={t("form.clientCompany")} required tooltipKey="proposal.wizard.client_company">
              <input className={fieldClass} value={input.client_company} onChange={(e) => patch({ client_company: e.target.value })} />
            </Field>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field
                label={t("form.clientContact")}
                required={scenario === "after_contact"}
                tooltipKey="proposal.wizard.client_contact"
              >
                <input className={fieldClass} value={input.client_contact} onChange={(e) => patch({ client_contact: e.target.value })} />
              </Field>
              <Field label={t("form.clientIndustry")} tooltipKey="proposal.wizard.client_industry">
                <select className={fieldClass} value={input.client_industry} onChange={(e) => patch({ client_industry: e.target.value })}>
                  <option value="">{t("form.select")}</option>
                  {PROPOSAL_INDUSTRY_OPTIONS.map((key) => (
                    <option key={key} value={key}>{t(`industries.${key}`)}</option>
                  ))}
                </select>
              </Field>
            </div>

            {scenario === "after_contact" && (
              <Field label={t("form.priorContactSummary")} required tooltipKey="proposal.wizard.prior_contact_summary">
                <textarea
                  rows={3}
                  className={fieldClass}
                  value={input.prior_contact_summary || ""}
                  onChange={(e) => patch({ prior_contact_summary: e.target.value })}
                  placeholder={t("form.priorContactPlaceholder")}
                />
              </Field>
            )}

            {scenario === "cold_outreach" && (
              <Field label={t("form.problemSource")} required tooltipKey="proposal.wizard.problem_source">
                <textarea
                  rows={3}
                  className={fieldClass}
                  value={input.problem_source || ""}
                  onChange={(e) => patch({ problem_source: e.target.value })}
                  placeholder={t("form.problemSourcePlaceholder")}
                />
              </Field>
            )}

            <Field label={t("form.clientProblem")} required tooltipKey="proposal.wizard.client_problem">
              <textarea rows={4} className={fieldClass} value={input.client_problem} onChange={(e) => patch({ client_problem: e.target.value })} />
            </Field>
          </FormSection>

          <FormSection label={t("form.solutionSection")}>
            {calcRuns.length > 0 && (
              <Field label={t("form.calculatorRun")} tooltipKey="proposal.wizard.calculator_run">
                <select
                  className={fieldClass}
                  value={input.calculator_run_id || ""}
                  onChange={(e) => {
                    const id = e.target.value || undefined;
                    patch({
                      calculator_run_id: id,
                      ...(id ? { proposal_scenario: "proactive_offer" as const, include_pricing: true } : {}),
                    });
                  }}
                >
                  <option value="">{t("form.noCalculator")}</option>
                  {calcRuns.map((r) => (
                    <option key={r.id} value={r.id}>
                      {r.process_name || r.id.slice(0, 8)} — {new Date(r.created_at).toLocaleDateString()}
                    </option>
                  ))}
                </select>
              </Field>
            )}
            <Field label={t("form.solutionName")} required tooltipKey="proposal.wizard.solution_name">
              <input className={fieldClass} value={input.solution_name} onChange={(e) => patch({ solution_name: e.target.value })} />
            </Field>
            <Field label={t("form.solutionDescription")} required tooltipKey="proposal.wizard.solution_description">
              <textarea rows={3} className={fieldClass} value={input.solution_description} onChange={(e) => patch({ solution_description: e.target.value })} />
            </Field>
            <Field label={t("form.deliverables")} required tooltipKey="proposal.wizard.deliverables">
              <div className="space-y-2">
                {input.deliverables.map((d, i) => (
                  <div key={i} className="flex gap-2">
                    <input className={fieldClass} value={d} onChange={(e) => setDeliverable(i, e.target.value)} placeholder={t("form.deliverablePlaceholder")} />
                    {input.deliverables.length > 1 && (
                      <button type="button" onClick={() => removeDeliverable(i)} className="shrink-0 rounded-lg border border-border2 px-3 text-text3 hover:bg-bg2">×</button>
                    )}
                  </div>
                ))}
                <button type="button" onClick={addDeliverable} className="text-xs text-accent hover:underline">{t("form.addDeliverable")}</button>
              </div>
            </Field>

            <Field label={t("form.includePricing")} tooltipKey="proposal.wizard.include_pricing">
              <label className="flex cursor-pointer items-center gap-2 text-sm text-text">
                <input
                  type="checkbox"
                  checked={input.include_pricing}
                  onChange={(e) => patch({ include_pricing: e.target.checked })}
                  className="rounded border-border2"
                />
                {t("form.includePricingLabel")}
              </label>
            </Field>

            <div className="grid gap-4 sm:grid-cols-3">
              {pricingRequired && (
                <Field label={t("form.projectCost")} required tooltipKey="proposal.wizard.project_cost">
                  <input type="number" min={1} className={fieldClass} value={input.project_cost_rub || ""} onChange={(e) => patch({ project_cost_rub: Number(e.target.value) })} />
                </Field>
              )}
              <Field label={t("form.timelineWeeks")} required tooltipKey="proposal.wizard.timeline_weeks">
                <input type="number" min={1} max={104} className={fieldClass} value={input.timeline_weeks} onChange={(e) => patch({ timeline_weeks: Number(e.target.value) })} />
              </Field>
              {pricingRequired && (
                <Field label={t("form.paymentSchedule")} required tooltipKey="proposal.wizard.payment_schedule">
                  <select className={fieldClass} value={input.payment_schedule} onChange={(e) => patch({ payment_schedule: e.target.value })}>
                    {PAYMENT_SCHEDULE_OPTIONS.map((opt) => (
                      <option key={opt} value={opt}>{opt}</option>
                    ))}
                  </select>
                </Field>
              )}
            </div>
          </FormSection>

          <FormSection label={t("form.senderSection")}>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field label={t("form.senderCompany")} required tooltipKey="proposal.wizard.sender_company">
                <input className={fieldClass} value={input.sender_company} onChange={(e) => patch({ sender_company: e.target.value })} />
              </Field>
              <Field label={t("form.senderContact")} required tooltipKey="proposal.wizard.sender_contact">
                <input className={fieldClass} value={input.sender_contact} onChange={(e) => patch({ sender_contact: e.target.value })} />
              </Field>
              <Field label={t("form.senderPhone")} tooltipKey="proposal.wizard.sender_phone">
                <input className={fieldClass} value={input.sender_phone} onChange={(e) => patch({ sender_phone: e.target.value })} />
              </Field>
              <Field label={t("form.senderEmail")} required tooltipKey="proposal.wizard.sender_email">
                <input type="email" className={fieldClass} value={input.sender_email} onChange={(e) => patch({ sender_email: e.target.value })} />
              </Field>
            </div>
          </FormSection>

          {error && (
            <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
          )}

          <div className="flex justify-end">
            <button
              type="button"
              onClick={handleGenerate}
              disabled={!formValid || loading || proposalLimitReached}
              className={cn(
                "rounded-lg px-5 py-2.5 text-sm font-medium text-white disabled:opacity-50",
                "bg-[#534ab7] hover:opacity-90"
              )}
            >
              {loading ? t("generating") : t("generate")}
            </button>
          </div>
        </div>
      )}

      {step === 2 && runId && (
        <ProposalPollingView
          runId={runId}
          onComplete={handlePollComplete}
          onError={(msg) => {
            setError(msg);
            setStep(1);
          }}
        />
      )}

      {step === 3 && result && (
        <div className="space-y-4">
          <button type="button" onClick={handleNew} className="text-sm text-accent hover:underline">
            {t("newProposal")}
          </button>
          <ProposalResult input={result.input} output={result.output} runId={result.runId} />
        </div>
      )}
    </div>
  );
}

function FormSection({ label, children }: { label: React.ReactNode; children: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-border bg-bg p-5 space-y-4">
      <div className="text-[10px] font-medium uppercase tracking-wider text-text3 border-b border-border pb-3">{label}</div>
      {children}
    </div>
  );
}

function Field({
  label,
  required,
  tooltipKey,
  children,
}: {
  label: string;
  required?: boolean;
  tooltipKey?: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label className="mb-1.5 block text-xs font-medium text-text2">
        {tooltipKey ? (
          <LabelWithHelp label={`${label}${required ? " *" : ""}`} tooltipKey={tooltipKey} />
        ) : (
          <>
            {label}
            {required ? " *" : ""}
          </>
        )}
      </label>
      {children}
    </div>
  );
}

export function ProposalWizard() {
  return (
    <Suspense fallback={<p className="text-sm text-text2">…</p>}>
      <ProposalWizardInner />
    </Suspense>
  );
}
