"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { ChevronDown, Download, Lock } from "lucide-react";
import { exportStrategyPDF, getBillingPlan } from "@/lib/api";
import type { StrategyInput, StrategyOutput } from "@/lib/api-strategy";
import {
  StrategyMermaidDiagram,
  StrategyPriorityMatrix,
  StrategyROIBarChart,
} from "./StrategyVisuals";
import { cn } from "@/lib/utils";

type Props = {
  input: StrategyInput;
  output: StrategyOutput;
  runId: string;
};

function severityClass(severity: string) {
  if (severity === "high") return "bg-red-50 text-red-800 border-red-200";
  if (severity === "medium") return "bg-amber-50 text-amber-900 border-amber-200";
  return "bg-bg2 text-text2 border-border";
}

function Prose({ children }: { children: React.ReactNode }) {
  return <p className="text-sm text-text2 whitespace-pre-wrap leading-relaxed">{children}</p>;
}

export function StrategyResult({ input, output, runId }: Props) {
  const t = useTranslations("strategy.result");
  const [canPdf, setCanPdf] = useState(false);
  const [loadingPdf, setLoadingPdf] = useState(false);
  const [pdfError, setPdfError] = useState("");

  useEffect(() => {
    getBillingPlan()
      .then((plan) => setCanPdf(Boolean(plan?.features?.export_pdf)))
      .catch(() => undefined);
  }, []);

  async function handlePdf() {
    setPdfError("");
    setLoadingPdf(true);
    try {
      const blob = await exportStrategyPDF(runId);
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "strategy-report.pdf";
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      setPdfError(err instanceof Error ? err.message : t("pdfFailed"));
    } finally {
      setLoadingPdf(false);
    }
  }

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-bg p-6">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p className="text-[10px] font-medium uppercase tracking-wider text-text3">{t("company")}</p>
            <h2 className="mt-1 text-base font-medium">{input.company_name}</h2>
            <p className="mt-1 text-sm text-text2">{input.business_description}</p>
            {(input.calculator_contexts?.length ?? 0) > 0 && (
              <p className="mt-2 text-xs text-accent">
                {t("linkedCalculators", { count: input.calculator_contexts!.length })}
              </p>
            )}
          </div>
          <button
            type="button"
            onClick={canPdf ? handlePdf : undefined}
            disabled={!canPdf || loadingPdf}
            className={cn(
              "inline-flex items-center gap-2 rounded-lg border px-3 py-2 text-xs font-medium",
              canPdf
                ? "border-border2 text-text hover:bg-bg2"
                : "border-border text-text3 cursor-not-allowed"
            )}
          >
            {canPdf ? <Download className="h-3.5 w-3.5" /> : <Lock className="h-3.5 w-3.5" />}
            {loadingPdf ? t("pdfLoading") : canPdf ? t("exportPdf") : t("exportPdfLocked")}
          </button>
        </div>
        {pdfError && <p className="mt-2 text-xs text-red-700">{pdfError}</p>}
      </div>

      <Section title={t("executiveSummary")}>
        <Prose>{output.executive_summary}</Prose>
      </Section>

      {output.roi_summary && output.roi_summary.lines.length > 0 && (
        <Section title={t("roiSummary")}>
          <StrategyROIBarChart summary={output.roi_summary} />
          <DataTable
            headers={[t("processCol"), t("netBenefitCol"), t("paybackCol"), t("npvCol"), t("recCol")]}
            rows={output.roi_summary.lines.map((l) => [
              l.process_name,
              formatRub(l.net_benefit_monthly_rub),
              l.payback_months > 0 ? `${l.payback_months.toFixed(1)} ${t("months")}` : "—",
              formatRub(l.npv_rub),
              l.recommendation,
            ])}
          />
        </Section>
      )}

      <Section title={t("solutions")}>
        <div className="space-y-4">
          {output.recommended_solutions?.map((s) => (
            <div key={s.priority} className="rounded-lg border border-border bg-bg2/50 p-4">
              <div className="flex items-start justify-between gap-2">
                <h3 className="text-sm font-medium text-text">{s.title}</h3>
                <span className="shrink-0 rounded bg-ai-bg px-2 py-0.5 text-[10px] font-medium text-ai">#{s.priority}</span>
              </div>
              <p className="mt-2 text-sm text-text2">{s.description}</p>
              <p className="mt-2 text-xs text-text3">{s.rationale}</p>
            </div>
          ))}
        </div>
      </Section>

      {(output.priority_matrix?.length ?? 0) > 0 && (
        <Section title={t("priorityMatrix")}>
          <StrategyPriorityMatrix rows={output.priority_matrix!} />
        </Section>
      )}

      <Section title={t("roadmap")}>
        <div className="space-y-3">
          {output.roadmap?.map((phase, i) => (
            <div key={i} className="border-l-2 border-ai pl-4">
              <p className="text-sm font-medium text-text">{phase.phase}</p>
              <p className="text-xs text-text3">{phase.quarter}</p>
              <ul className="mt-2 list-disc space-y-1 pl-4 text-sm text-text2">
                {phase.initiatives?.map((item) => (
                  <li key={item}>{item}</li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </Section>

      {(output.diagrams?.length ?? 0) > 0 && (
        <Section title={t("diagrams")}>
          <div className="space-y-4">
            {output.diagrams!.map((d, i) => (
              <StrategyMermaidDiagram key={`${d.title}-${i}`} diagram={d} />
            ))}
          </div>
        </Section>
      )}

      <Section title={t("next30")}>
        <ol className="list-decimal space-y-2 pl-5 text-sm text-text2">
          {output.next_30_days?.map((step) => (
            <li key={step}>{step}</li>
          ))}
        </ol>
      </Section>

      <div className="rounded-xl border border-border bg-bg p-4">
        <p className="text-[10px] font-medium uppercase tracking-wider text-text3 mb-3">{t("fullReport")}</p>
        <div className="space-y-2">
          <Accordion title={t("currentSituation")}>
            <Prose>{output.current_situation}</Prose>
          </Accordion>
          {output.goals_and_rationale && (
            <Accordion title={t("goalsAndRationale")}>
              <Prose>{output.goals_and_rationale}</Prose>
            </Accordion>
          )}
          {(output.process_analysis?.length ?? 0) > 0 && (
            <Accordion title={t("processAnalysis")}>
              <div className="space-y-3">
                {output.process_analysis.map((p) => (
                  <div key={p.name} className="rounded-lg border border-border bg-bg2/30 p-3">
                    <p className="text-sm font-medium text-text">{p.name}</p>
                    <p className="mt-1 text-xs text-text2">{p.current_state}</p>
                    <p className="mt-2 text-xs text-text3">{p.ai_potential}</p>
                  </div>
                ))}
              </div>
            </Accordion>
          )}
          {(output.maturity_matrix?.length ?? 0) > 0 && (
            <Accordion title={t("maturityMatrix")}>
              <DataTable
                headers={[t("criterionCol"), t("currentCol"), t("targetLevelCol"), t("gapCol")]}
                rows={output.maturity_matrix!.map((r) => [r.criterion, r.current_level, r.target_level, r.gap])}
              />
            </Accordion>
          )}
          {output.data_and_infrastructure && (
            <Accordion title={t("dataAndInfrastructure")}>
              <Prose>{output.data_and_infrastructure}</Prose>
            </Accordion>
          )}
          {(output.ai_use_cases?.length ?? 0) > 0 && (
            <Accordion title={t("aiUseCases")}>
              <div className="space-y-2">
                {output.ai_use_cases.map((u) => (
                  <div key={u.priority} className="flex gap-2 text-sm">
                    <span className="shrink-0 font-mono text-xs text-text3">#{u.priority}</span>
                    <div>
                      <p className="font-medium text-text">{u.title}</p>
                      <p className="text-xs text-text2">{u.description}</p>
                    </div>
                  </div>
                ))}
              </div>
            </Accordion>
          )}
          {(output.implementation_plan?.length ?? 0) > 0 && (
            <Accordion title={t("implementationPlan")}>
              <div className="space-y-3">
                {output.implementation_plan.map((ph) => (
                  <div key={ph.title}>
                    <p className="text-sm font-medium text-text">{ph.title} · {ph.duration}</p>
                    <ul className="mt-1 list-disc pl-4 text-xs text-text2">
                      {ph.deliverables?.map((d) => (
                        <li key={d}>{d}</li>
                      ))}
                    </ul>
                  </div>
                ))}
              </div>
            </Accordion>
          )}
          {(output.tech_stack?.length ?? 0) > 0 && (
            <Accordion title={t("techStack")}>
              <DataTable
                headers={[t("layerCol"), t("toolCol"), t("roleCol"), t("statusCol")]}
                rows={output.tech_stack!.map((r) => [r.layer, r.tool, r.role, r.status])}
              />
            </Accordion>
          )}
          {(output.stakeholder_plan?.length ?? 0) > 0 && (
            <Accordion title={t("stakeholders")}>
              <DataTable
                headers={[t("roleCol"), t("respCol"), t("involvementCol")]}
                rows={output.stakeholder_plan!.map((r) => [r.role, r.responsibility, r.involvement])}
              />
            </Accordion>
          )}
          {output.team_and_training && (
            <Accordion title={t("teamAndTraining")}>
              <Prose>{output.team_and_training}</Prose>
            </Accordion>
          )}
          {output.architecture_overview && (
            <Accordion title={t("architectureOverview")}>
              <Prose>{output.architecture_overview}</Prose>
            </Accordion>
          )}
          {output.data_governance && (
            <Accordion title={t("dataGovernance")}>
              <Prose>{output.data_governance}</Prose>
            </Accordion>
          )}
          {output.ethics_and_compliance && (
            <Accordion title={t("ethicsAndCompliance")}>
              <Prose>{output.ethics_and_compliance}</Prose>
            </Accordion>
          )}
          {output.budget_overview?.summary && (
            <Accordion title={t("budgetOverview")}>
              <Prose>{output.budget_overview.summary}</Prose>
              {(output.budget_overview.lines?.length ?? 0) > 0 && (
                <DataTable
                  headers={[t("categoryCol"), t("amountCol"), t("notesCol")]}
                  rows={output.budget_overview.lines.map((l) => [l.category, l.amount_range, l.notes])}
                />
              )}
            </Accordion>
          )}
          {(output.budget_phases?.length ?? 0) > 0 && (
            <Accordion title={t("budgetPhases")}>
              <DataTable
                headers={[t("phaseCol"), "CAPEX", "OPEX/mo", t("cumulativeCol")]}
                rows={output.budget_phases!.map((p) => [p.phase, p.capex_rub, p.opex_monthly_rub, p.cumulative_rub])}
              />
            </Accordion>
          )}
          <Accordion title={t("metrics")}>
            <MetricsTable metrics={output.success_metrics} t={t} />
          </Accordion>
          {output.strategy_adjustment_plan && (
            <Accordion title={t("strategyAdjustment")}>
              <Prose>{output.strategy_adjustment_plan}</Prose>
            </Accordion>
          )}
          <Accordion title={t("risks")}>
            <div className="space-y-2">
              {output.risks?.map((r, i) => (
                <div key={i} className={cn("rounded-lg border px-3 py-2 text-sm", severityClass(r.severity))}>
                  <p className="font-medium">{r.risk}</p>
                  <p className="mt-1 text-xs opacity-90">{r.mitigation}</p>
                </div>
              ))}
            </div>
          </Accordion>
        </div>
      </div>
    </div>
  );
}

function formatRub(n: number) {
  return new Intl.NumberFormat("ru-RU", { maximumFractionDigits: 0 }).format(n) + " ₽";
}

function DataTable({ headers, rows }: { headers: string[]; rows: string[][] }) {
  return (
    <div className="overflow-x-auto">
      <table className="mt-2 w-full text-left text-xs">
        <thead>
          <tr className="border-b border-border text-[10px] uppercase tracking-wider text-text3">
            {headers.map((h) => (
              <th key={h} className="pb-2 pr-3 font-medium">{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} className="border-b border-border">
              {row.map((cell, j) => (
                <td key={j} className="py-2 pr-3 text-text2 align-top">{cell}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function MetricsTable({
  metrics,
  t,
}: {
  metrics: StrategyOutput["success_metrics"];
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[640px] text-left text-xs">
        <thead>
          <tr className="border-b border-border text-[10px] uppercase tracking-wider text-text3">
            <th className="pb-2 pr-4">{t("metricCol")}</th>
            <th className="pb-2 pr-4">{t("targetCol")}</th>
            <th className="pb-2">{t("timeCol")}</th>
          </tr>
        </thead>
        <tbody>
          {metrics?.map((m, i) => (
            <tr key={i} className="border-b border-border">
              <td className="py-2 pr-4 text-text">{m.metric}</td>
              <td className="py-2 pr-4 text-text2">{m.target}</td>
              <td className="py-2 text-text2">{m.timeframe}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="rounded-xl border border-border bg-bg p-6 space-y-3">
      <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">{title}</h2>
      {children}
    </section>
  );
}

function Accordion({ title, children }: { title: string; children: React.ReactNode }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="rounded-lg border border-border">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="flex w-full items-center justify-between px-4 py-3 text-left text-sm font-medium text-text hover:bg-bg2/50"
      >
        {title}
        <ChevronDown className={cn("h-4 w-4 text-text3 transition-transform", open && "rotate-180")} />
      </button>
      {open && <div className="border-t border-border px-4 py-3">{children}</div>}
    </div>
  );
}
