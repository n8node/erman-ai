"use client";

import { useTranslations } from "next-intl";
import type { StrategyInput, StrategyOutput } from "@/lib/api-strategy";
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

export function StrategyResult({ input, output, runId }: Props) {
  const t = useTranslations("strategy.result");

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-bg p-6">
        <p className="text-[10px] font-medium uppercase tracking-wider text-text3">{t("company")}</p>
        <h2 className="mt-1 text-base font-medium">{input.company_name}</h2>
        <p className="mt-1 text-sm text-text2">{input.industry}</p>
        <p className="mt-2 text-xs text-text3">Run: {runId.slice(0, 8)}…</p>
      </div>

      <Section title={t("executiveSummary")}>
        <p className="text-sm text-text2 whitespace-pre-wrap leading-relaxed">{output.executive_summary}</p>
      </Section>

      <Section title={t("currentSituation")}>
        <p className="text-sm text-text2 whitespace-pre-wrap leading-relaxed">{output.current_situation}</p>
      </Section>

      <Section title={t("solutions")}>
        <div className="space-y-4">
          {output.recommended_solutions?.map((s) => (
            <div key={s.priority} className="rounded-lg border border-border bg-bg2/50 p-4">
              <div className="flex items-start justify-between gap-2">
                <h3 className="text-sm font-medium text-text">{s.title}</h3>
                <span className="shrink-0 rounded bg-ai-bg px-2 py-0.5 text-[10px] font-medium text-ai">
                  #{s.priority}
                </span>
              </div>
              <p className="mt-2 text-sm text-text2">{s.description}</p>
              <p className="mt-2 text-xs text-text3">{s.rationale}</p>
            </div>
          ))}
        </div>
      </Section>

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

      <Section title={t("risks")}>
        <div className="space-y-2">
          {output.risks?.map((r, i) => (
            <div key={i} className={cn("rounded-lg border px-3 py-2 text-sm", severityClass(r.severity))}>
              <p className="font-medium">{r.risk}</p>
              <p className="mt-1 text-xs opacity-90">{r.mitigation}</p>
            </div>
          ))}
        </div>
      </Section>

      <Section title={t("metrics")}>
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead>
              <tr className="border-b border-border text-[10px] uppercase tracking-wider text-text3">
                <th className="pb-2 pr-4">{t("metricCol")}</th>
                <th className="pb-2 pr-4">{t("targetCol")}</th>
                <th className="pb-2">{t("timeCol")}</th>
              </tr>
            </thead>
            <tbody>
              {output.success_metrics?.map((m, i) => (
                <tr key={i} className="border-b border-border">
                  <td className="py-2 pr-4 text-text">{m.metric}</td>
                  <td className="py-2 pr-4 text-text2">{m.target}</td>
                  <td className="py-2 text-text2">{m.timeframe}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Section>

      <Section title={t("next30")}>
        <ol className="list-decimal space-y-2 pl-5 text-sm text-text2">
          {output.next_30_days?.map((step) => (
            <li key={step}>{step}</li>
          ))}
        </ol>
      </Section>
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
