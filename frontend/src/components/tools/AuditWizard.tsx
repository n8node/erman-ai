"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchTools,
  getRun,
  isToolLimitError,
  runAudit,
  type ToolListItem,
} from "@/lib/api";
import {
  AUDIT_BOTTLENECK_OPTIONS,
  AUDIT_BUDGET_OPTIONS,
  AUDIT_COMPANY_SIZE_OPTIONS,
  AUDIT_DATA_DUPLICATION,
  AUDIT_DEPARTMENT_OPTIONS,
  AUDIT_GOAL_OPTIONS,
  AUDIT_HEADCOUNT_OPTIONS,
  AUDIT_IT_SYSTEMS,
  AUDIT_PRIORITY_CRITERIA,
  AUDIT_PROCESS_TEMPLATES,
  AUDIT_SYSTEMS,
  AUDIT_TIMELINE_OPTIONS,
  DEFAULT_AUDIT_INPUT,
  DEFAULT_AUDIT_PROCESS,
  ERROR_COST_RANGES,
  ERROR_RATE_RANGES,
  FTE_RANGES,
  FREQUENCY_RANGES,
  HOURLY_RANGES,
  MAX_AUDIT_PROCESSES,
  MIN_AUDIT_PROCESSES,
  MINUTES_RANGES,
  createProcessFromTemplate,
  isAuditInputValid,
  isAuditProcessValid,
  type AuditInput,
  type AuditOutput,
  type AuditProcessInput,
} from "@/lib/api-audit";
import { isLimitReached } from "@/lib/tool-limits";
import { ToolLimitBadge } from "@/components/dashboard/ToolLimitBadge";
import { ChoiceOrCustom, MultiChoiceWithOther, RangeSelect } from "@/components/ui/ChoiceFields";
import { ToolLimitExceededAlert } from "./ToolLimitExceededAlert";
import { AuditPollingView } from "./AuditPollingView";
import { AuditResult } from "./AuditResult";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

function AuditWizardInner() {
  const t = useTranslations("audit");
  const tLimits = useTranslations("toolLimits");
  const router = useRouter();
  const searchParams = useSearchParams();
  const runIdParam = searchParams.get("run");

  const [formStep, setFormStep] = useState<1 | 2 | 3>(1);
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [input, setInput] = useState<AuditInput>(DEFAULT_AUDIT_INPUT);
  const [runId, setRunId] = useState<string | null>(null);
  const [result, setResult] = useState<{ input: AuditInput; output: AuditOutput; runId: string } | null>(null);
  const [loading, setLoading] = useState(false);
  const [loadingRun, setLoadingRun] = useState(!!runIdParam);
  const [error, setError] = useState("");
  const [limitExceeded, setLimitExceeded] = useState(false);
  const [auditTool, setAuditTool] = useState<ToolListItem | null>(null);
  const [expandedProcess, setExpandedProcess] = useState(0);

  function loadToolLimits() {
    return fetchTools()
      .then((data) => {
        const tool = data.tools.find((x) => x.slug === "audit") ?? null;
        setAuditTool(tool);
        if (tool) setLimitExceeded(isLimitReached(tool));
        return tool;
      })
      .catch(() => null);
  }

  useEffect(() => {
    void loadToolLimits();
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
            input: run.input as unknown as AuditInput,
            output: run.output as unknown as AuditOutput,
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

  function patch(partial: Partial<AuditInput>) {
    setInput((prev) => ({ ...prev, ...partial }));
  }

  function patchProcess(index: number, partial: Partial<AuditProcessInput>) {
    setInput((prev) => {
      const processes = [...prev.processes];
      processes[index] = { ...processes[index], ...partial };
      return { ...prev, processes };
    });
  }

  function toggleCriteria(key: string) {
    setInput((prev) => {
      const has = prev.priority_criteria.includes(key);
      return {
        ...prev,
        priority_criteria: has ? prev.priority_criteria.filter((k) => k !== key) : [...prev.priority_criteria, key],
      };
    });
  }

  function addProcess(templateKey: string) {
    if (input.processes.length >= MAX_AUDIT_PROCESSES) return;
    const tpl = AUDIT_PROCESS_TEMPLATES.find((t) => t.key === templateKey);
    const proc = createProcessFromTemplate(
      templateKey,
      templateKey === "custom" ? "" : t(`templates.${templateKey}`),
      templateKey === "custom" ? "" : (tpl?.department ?? "")
    );
    setInput((prev) => ({ ...prev, processes: [...prev.processes, proc] }));
    setExpandedProcess(input.processes.length);
  }

  function removeProcess(index: number) {
    if (input.processes.length <= MIN_AUDIT_PROCESSES) return;
    setInput((prev) => ({
      ...prev,
      processes: prev.processes.filter((_, i) => i !== index),
    }));
    setExpandedProcess(Math.max(0, index - 1));
  }

  const step1Valid =
    input.company_name.trim() &&
    input.primary_goal.trim() &&
    input.priority_criteria.length > 0 &&
    (input.it_systems.length > 0 || (input.it_systems_other ?? "").trim());

  const step2Valid = input.processes.every(isAuditProcessValid);

  const auditLimitReached = auditTool ? isLimitReached(auditTool) : limitExceeded;

  async function handleGenerate() {
    if (!isAuditInputValid(input) || auditLimitReached) {
      if (auditLimitReached) setLimitExceeded(true);
      return;
    }
    setError("");
    setLimitExceeded(false);
    setLoading(true);
    try {
      const data = await runAudit(input);
      setRunId(data.run_id);
      setStep(2);
      router.replace(`/tools/audit?run=${data.run_id}`);
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
          input: run.input as unknown as AuditInput,
          output: run.output as unknown as AuditOutput,
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
    setInput(DEFAULT_AUDIT_INPUT);
    setFormStep(1);
    setRunId(null);
    setResult(null);
    setStep(1);
    setError("");
    setLimitExceeded(false);
    setExpandedProcess(0);
    void loadToolLimits();
    router.replace("/tools/audit");
  }

  const goalOptions = AUDIT_GOAL_OPTIONS.filter((g) => g !== "other").map((g) => ({
    value: g,
    label: t(`goals.${g}`),
  }));

  const deptOptions = AUDIT_DEPARTMENT_OPTIONS.filter((d) => d !== "other").map((d) => ({
    value: d,
    label: t(`departments.${d}`),
  }));

  const bottleneckOptions = AUDIT_BOTTLENECK_OPTIONS.filter((b) => b !== "other").map((b) => ({
    value: b,
    label: t(`bottlenecks.${b}`),
  }));

  const itSystemOptions = AUDIT_IT_SYSTEMS.map((s) => ({ value: s, label: t(`systems.${s}`) }));
  const processSystemOptions = AUDIT_SYSTEMS.filter((s) => s !== "other").map((s) => ({
    value: s,
    label: t(`systems.${s}`),
  }));

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
        {auditTool && (
          <ToolLimitBadge
            tool={auditTool}
            t={(key, values) => tLimits(key, values as Record<string, string | number> | undefined)}
          />
        )}
      </div>

      {(limitExceeded || auditLimitReached) && step === 1 && (
        <ToolLimitExceededAlert toolName={t("title")} />
      )}

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}

      {step === 1 && (
        <div className="rounded-xl border border-border bg-bg p-6 space-y-6">
          <div className="flex gap-2">
            <FormStepBadge n={1} active={formStep === 1} done={formStep > 1} label={t("formSteps.company")} />
            <FormStepBadge n={2} active={formStep === 2} done={formStep > 2} label={t("formSteps.processes")} />
            <FormStepBadge n={3} active={formStep === 3} done={false} label={t("formSteps.review")} />
          </div>

          {formStep === 1 && (
            <>
              <SectionLabel>{t("form.companySection")}</SectionLabel>
              <div className="grid gap-4 sm:grid-cols-2">
                <Field label={t("form.companyName")} className="sm:col-span-2">
                  <input className={fieldClass} value={input.company_name} onChange={(e) => patch({ company_name: e.target.value })} />
                </Field>
                <Field label={t("form.primaryGoal")}>
                  <ChoiceOrCustom
                    value={input.primary_goal}
                    onChange={(v) => patch({ primary_goal: v })}
                    options={goalOptions}
                    customLabel={t("form.select")}
                    placeholder={t("form.customGoal")}
                  />
                </Field>
                <Field label={t("form.companySize")}>
                  <select className={fieldClass} value={input.company_size} onChange={(e) => patch({ company_size: e.target.value })}>
                    {AUDIT_COMPANY_SIZE_OPTIONS.map((s) => (
                      <option key={s} value={s}>{s}</option>
                    ))}
                  </select>
                </Field>
                <Field label={t("form.headcount")}>
                  <select className={fieldClass} value={input.operational_headcount} onChange={(e) => patch({ operational_headcount: e.target.value })}>
                    {AUDIT_HEADCOUNT_OPTIONS.map((s) => (
                      <option key={s} value={s}>{t(`headcount.${s}`)}</option>
                    ))}
                  </select>
                </Field>
                <Field label={t("form.budget")}>
                  <select className={fieldClass} value={input.budget_range} onChange={(e) => patch({ budget_range: e.target.value })}>
                    {AUDIT_BUDGET_OPTIONS.map((b) => (
                      <option key={b} value={b}>{t(`budget.${b}`)}</option>
                    ))}
                  </select>
                </Field>
                <Field label={t("form.timeline")}>
                  <select className={fieldClass} value={input.decision_timeline} onChange={(e) => patch({ decision_timeline: e.target.value })}>
                    {AUDIT_TIMELINE_OPTIONS.map((tl) => (
                      <option key={tl} value={tl}>{t(`timeline.${tl}`)}</option>
                    ))}
                  </select>
                </Field>
                <Field label={t("form.dataDuplication")}>
                  <select className={fieldClass} value={input.data_duplication} onChange={(e) => patch({ data_duplication: e.target.value })}>
                    {AUDIT_DATA_DUPLICATION.map((d) => (
                      <option key={d} value={d}>{t(`duplication.${d}`)}</option>
                    ))}
                  </select>
                </Field>
              </div>

              <Field label={t("form.priorityCriteria")}>
                <div className="flex flex-wrap gap-2">
                  {AUDIT_PRIORITY_CRITERIA.map((c) => (
                    <Chip key={c} active={input.priority_criteria.includes(c)} onClick={() => toggleCriteria(c)}>
                      {t(`criteria.${c}`)}
                    </Chip>
                  ))}
                </div>
              </Field>

              <Field label={t("form.itSystems")}>
                <MultiChoiceWithOther
                  values={[...input.it_systems, ...(input.it_systems_other ? [input.it_systems_other] : [])]}
                  onChange={(vals) => {
                    const presetSet = new Set(AUDIT_IT_SYSTEMS as readonly string[]);
                    patch({
                      it_systems: vals.filter((v) => presetSet.has(v)),
                      it_systems_other: vals.find((v) => !presetSet.has(v)) ?? "",
                    });
                  }}
                  options={itSystemOptions}
                  otherLabel={t("form.otherSystem")}
                  otherPlaceholder={t("form.otherSystemPlaceholder")}
                />
              </Field>

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
              <SectionLabel>{t("form.processesSection")}</SectionLabel>
              <p className="text-xs text-text3">{t("form.processesHint", { min: MIN_AUDIT_PROCESSES, max: MAX_AUDIT_PROCESSES })}</p>

              <div className="flex flex-wrap gap-2">
                {AUDIT_PROCESS_TEMPLATES.map((tpl) => (
                  <button
                    key={tpl.key}
                    type="button"
                    disabled={input.processes.length >= MAX_AUDIT_PROCESSES}
                    onClick={() => addProcess(tpl.key)}
                    className="rounded-lg border border-border2 px-3 py-1.5 text-xs text-text2 hover:border-accent hover:text-accent disabled:opacity-50"
                  >
                    + {t(`templates.${tpl.key}`)}
                  </button>
                ))}
              </div>

              <div className="space-y-3">
                {input.processes.map((proc, idx) => (
                  <ProcessCard
                    key={idx}
                    index={idx}
                    proc={proc}
                    expanded={expandedProcess === idx}
                    onToggle={() => setExpandedProcess(expandedProcess === idx ? -1 : idx)}
                    onRemove={() => removeProcess(idx)}
                    canRemove={input.processes.length > MIN_AUDIT_PROCESSES}
                    onChange={(p) => patchProcess(idx, p)}
                    t={t}
                    deptOptions={deptOptions}
                    bottleneckOptions={bottleneckOptions}
                    processSystemOptions={processSystemOptions}
                    valid={isAuditProcessValid(proc)}
                  />
                ))}
              </div>

              <div className="flex justify-between pt-2">
                <button type="button" onClick={() => setFormStep(1)} className="rounded-lg border border-border2 px-4 py-2 text-sm text-text2">
                  {t("form.back")}
                </button>
                <button
                  type="button"
                  disabled={!step2Valid}
                  onClick={() => setFormStep(3)}
                  className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
                >
                  {t("form.next")}
                </button>
              </div>
            </>
          )}

          {formStep === 3 && (
            <>
              <SectionLabel>{t("form.reviewSection")}</SectionLabel>
              <div className="rounded-lg border border-border bg-bg2 p-4 text-sm space-y-2">
                <p><span className="text-text3">{t("form.companyName")}:</span> {input.company_name}</p>
                <p><span className="text-text3">{t("form.primaryGoal")}:</span> {input.primary_goal}</p>
                <p><span className="text-text3">{t("form.processesSection")}:</span> {input.processes.map((p) => p.name).join(", ")}</p>
              </div>
              <div className="flex justify-between pt-2">
                <button type="button" onClick={() => setFormStep(2)} className="rounded-lg border border-border2 px-4 py-2 text-sm text-text2">
                  {t("form.back")}
                </button>
                <button
                  type="button"
                  disabled={!isAuditInputValid(input) || loading || auditLimitReached}
                  onClick={() => void handleGenerate()}
                  className="rounded-lg bg-[#534ab7] px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
                >
                  {loading ? t("form.generating") : t("form.generate")}
                </button>
              </div>
            </>
          )}
        </div>
      )}

      {step === 2 && runId && (
        <AuditPollingView
          runId={runId}
          onComplete={() => void handlePollComplete()}
          onError={(msg) => {
            setError(msg);
            setStep(1);
          }}
        />
      )}

      {step === 3 && result && (
        <div className="space-y-4">
          <div className="flex justify-end">
            <button type="button" onClick={handleNew} className="rounded-lg border border-border2 px-4 py-2 text-sm text-text2">
              {t("form.newAudit")}
            </button>
          </div>
          <AuditResult input={result.input} output={result.output} runId={result.runId} />
        </div>
      )}
    </div>
  );
}

function ProcessCard({
  index,
  proc,
  expanded,
  onToggle,
  onRemove,
  canRemove,
  onChange,
  t,
  deptOptions,
  bottleneckOptions,
  processSystemOptions,
  valid,
}: {
  index: number;
  proc: AuditProcessInput;
  expanded: boolean;
  onToggle: () => void;
  onRemove: () => void;
  canRemove: boolean;
  onChange: (p: Partial<AuditProcessInput>) => void;
  t: ReturnType<typeof useTranslations>;
  deptOptions: { value: string; label: string }[];
  bottleneckOptions: { value: string; label: string }[];
  processSystemOptions: { value: string; label: string }[];
  valid: boolean;
}) {
  const rangeLabel = (prefix: string, key: string) => t(`${prefix}.${key}` as "ranges.frequency.1-10");

  return (
    <div className={cn("rounded-lg border", valid ? "border-border" : "border-warning/50")}>
      <button type="button" onClick={onToggle} className="flex w-full items-center justify-between px-4 py-3 text-left">
        <div>
          <span className="text-sm font-medium">{proc.name.trim() || t("form.processNumber", { n: index + 1 })}</span>
          {!valid && <span className="ml-2 text-xs text-warning">{t("form.incomplete")}</span>}
        </div>
        <span className="text-text3">{expanded ? "▲" : "▼"}</span>
      </button>
      {expanded && (
        <div className="border-t border-border px-4 py-4 space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <Field label={t("process.name")}>
              <input className={fieldClass} value={proc.name} onChange={(e) => onChange({ name: e.target.value })} />
            </Field>
            <Field label={t("process.department")}>
              <ChoiceOrCustom
                value={proc.department}
                onChange={(v) => onChange({ department: v })}
                options={deptOptions}
                customLabel={t("form.select")}
                placeholder={t("process.customDepartment")}
              />
            </Field>
            <Field label={t("process.frequency")}>
              <RangeSelect
                value={proc.frequency_range}
                onChange={(v) => onChange({ frequency_range: v })}
                options={FREQUENCY_RANGES.map((r) => ({ value: r, label: rangeLabel("ranges.frequency", r) }))}
              />
            </Field>
            <Field label={t("process.minutes")}>
              <RangeSelect
                value={proc.minutes_per_cycle_range}
                onChange={(v) => onChange({ minutes_per_cycle_range: v })}
                options={MINUTES_RANGES.map((r) => ({ value: r, label: rangeLabel("ranges.minutes", r) }))}
              />
            </Field>
            <Field label={t("process.fte")}>
              <RangeSelect
                value={proc.fte_range}
                onChange={(v) => onChange({ fte_range: v })}
                options={FTE_RANGES.map((r) => ({ value: r, label: rangeLabel("ranges.fte", r) }))}
              />
            </Field>
            <Field label={t("process.hourlyRate")}>
              <RangeSelect
                value={proc.hourly_rate_range}
                onChange={(v) => onChange({ hourly_rate_range: v })}
                options={HOURLY_RANGES.map((r) => ({ value: r, label: rangeLabel("ranges.hourly", r) }))}
              />
            </Field>
            <Field label={t("process.errorRate")}>
              <RangeSelect
                value={proc.error_rate_range}
                onChange={(v) => onChange({ error_rate_range: v })}
                options={ERROR_RATE_RANGES.map((r) => ({ value: r, label: rangeLabel("ranges.errorRate", r) }))}
              />
            </Field>
            <Field label={t("process.errorCost")}>
              <RangeSelect
                value={proc.error_cost_range}
                onChange={(v) => onChange({ error_cost_range: v })}
                options={ERROR_COST_RANGES.map((r) => ({ value: r, label: rangeLabel("ranges.errorCost", r) }))}
              />
            </Field>
          </div>

          <Field label={t("process.systems")}>
            <MultiChoiceWithOther
              values={[...proc.systems, ...(proc.systems_other ? [proc.systems_other] : [])]}
              onChange={(vals) => {
                const presetSet = new Set(AUDIT_SYSTEMS.filter((s) => s !== "other") as readonly string[]);
                onChange({
                  systems: vals.filter((v) => presetSet.has(v)),
                  systems_other: vals.find((v) => !presetSet.has(v)) ?? "",
                });
              }}
              options={processSystemOptions}
              otherLabel={t("form.otherSystem")}
              otherPlaceholder={t("form.otherSystemPlaceholder")}
            />
          </Field>

          <Field label={t("process.bottleneck")}>
            <ChoiceOrCustom
              value={proc.bottleneck}
              onChange={(v) => onChange({ bottleneck: v })}
              options={bottleneckOptions}
              customLabel={t("form.select")}
              placeholder={t("process.customBottleneck")}
            />
          </Field>

          <div className="grid gap-4 sm:grid-cols-3">
            <Field label={t("process.standardization")}>
              <select className={fieldClass} value={proc.standardization} onChange={(e) => onChange({ standardization: e.target.value as AuditProcessInput["standardization"] })}>
                {(["high", "medium", "low"] as const).map((v) => (
                  <option key={v} value={v}>{t(`levels.${v}`)}</option>
                ))}
              </select>
            </Field>
            <Field label={t("process.readiness")}>
              <select className={fieldClass} value={proc.automation_readiness} onChange={(e) => onChange({ automation_readiness: e.target.value as AuditProcessInput["automation_readiness"] })}>
                {(["high", "medium", "low"] as const).map((v) => (
                  <option key={v} value={v}>{t(`levels.${v}`)}</option>
                ))}
              </select>
            </Field>
            <Field label={t("process.impact")}>
              <select className={fieldClass} value={proc.expected_impact} onChange={(e) => onChange({ expected_impact: e.target.value as AuditProcessInput["expected_impact"] })}>
                {(["high", "medium", "low"] as const).map((v) => (
                  <option key={v} value={v}>{t(`levels.${v}`)}</option>
                ))}
              </select>
            </Field>
          </div>

          <Field label={t("process.integrationComplexity")}>
            <div className="flex gap-2">
              {([1, 2, 3, 4, 5] as const).map((n) => (
                <button
                  key={n}
                  type="button"
                  onClick={() => onChange({ integration_complexity: n })}
                  className={cn(
                    "h-9 w-9 rounded-lg border text-sm",
                    proc.integration_complexity === n ? "border-accent bg-accent/10 text-accent font-medium" : "border-border2 text-text2"
                  )}
                >
                  {n}
                </button>
              ))}
            </div>
          </Field>

          {canRemove && (
            <button type="button" onClick={onRemove} className="text-xs text-error hover:underline">
              {t("form.removeProcess")}
            </button>
          )}
        </div>
      )}
    </div>
  );
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return <p className="text-[10px] font-medium uppercase tracking-wider text-text3">{children}</p>;
}

function Field({ label, children, className }: { label: string; children: React.ReactNode; className?: string }) {
  return (
    <label className={cn("block space-y-1.5", className)}>
      <span className="text-xs font-medium text-text">{label}</span>
      {children}
    </label>
  );
}

function Chip({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "rounded-full border px-3 py-1 text-xs transition-colors",
        active ? "border-accent bg-accent/10 text-accent font-medium" : "border-border2 text-text2 hover:border-border"
      )}
    >
      {children}
    </button>
  );
}

function FormStepBadge({ n, active, done, label }: { n: number; active: boolean; done: boolean; label: string }) {
  return (
    <div className={cn("flex items-center gap-2 rounded-lg px-3 py-1.5 text-xs", active ? "bg-bg2 font-medium text-text" : "text-text3")}>
      <span className={cn("flex h-5 w-5 items-center justify-center rounded-full text-[10px]", done ? "bg-success text-white" : active ? "bg-text text-white" : "bg-bg2 text-text3")}>
        {done ? "✓" : n}
      </span>
      {label}
    </div>
  );
}

export function AuditWizard() {
  return (
    <Suspense fallback={<p className="text-sm text-text2">…</p>}>
      <AuditWizardInner />
    </Suspense>
  );
}
