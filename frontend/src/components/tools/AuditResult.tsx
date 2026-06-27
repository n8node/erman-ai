"use client";

import { useTranslations } from "next-intl";
import type { AuditInput, AuditOutput } from "@/lib/api-audit";
import { cn } from "@/lib/utils";

type Props = {
  input: AuditInput;
  output: AuditOutput;
  runId: string;
};

function formatRub(n: number) {
  return new Intl.NumberFormat("ru-RU").format(Math.round(n)) + " ₽";
}

function Prose({ children }: { children: React.ReactNode }) {
  return <p className="text-sm text-text2 whitespace-pre-wrap leading-relaxed">{children}</p>;
}

export function AuditResult({ input, output }: Props) {
  const t = useTranslations("audit.result");

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-bg p-6">
        <p className="text-[10px] font-medium uppercase tracking-wider text-text3">{t("document")}</p>
        <h2 className="mt-1 text-base font-medium">{input.company_name}</h2>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
        <div className="mt-4 grid gap-4 sm:grid-cols-3">
          <MetricCard label={t("metrics.totalCost")} value={formatRub(output.total_monthly_cost_rub)} />
          <MetricCard label={t("metrics.totalSavings")} value={formatRub(output.total_monthly_savings_est_rub)} />
          <MetricCard label={t("metrics.processCount")} value={String(output.priority_ranking.length)} />
        </div>
      </div>

      <Section title={t("sections.summary")}>
        <Prose>{output.executive_summary}</Prose>
      </Section>

      <Section title={t("sections.context")}>
        <Prose>{output.company_context}</Prose>
      </Section>

      {output.quick_wins.length > 0 && (
        <Section title={t("sections.quickWins")}>
          <ul className="list-disc space-y-1 pl-5 text-sm text-text2">
            {output.quick_wins.map((w) => (
              <li key={w}>{w}</li>
            ))}
          </ul>
        </Section>
      )}

      <Section title={t("sections.priority")}>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[640px] text-left text-sm">
            <thead>
              <tr className="border-b border-border text-[10px] uppercase tracking-wider text-text3">
                <th className="py-2 pr-3">{t("table.rank")}</th>
                <th className="py-2 pr-3">{t("table.process")}</th>
                <th className="py-2 pr-3">{t("table.score")}</th>
                <th className="py-2 pr-3">{t("table.savings")}</th>
                <th className="py-2">{t("table.quickWin")}</th>
              </tr>
            </thead>
            <tbody>
              {output.priority_ranking.map((row) => (
                <tr key={row.rank} className="border-b border-border last:border-0">
                  <td className="py-3 pr-3 font-medium">{row.rank}</td>
                  <td className="py-3 pr-3">
                    <div>{row.process_name}</div>
                    <div className="mt-0.5 text-xs text-text3">{row.rationale}</div>
                  </td>
                  <td className="py-3 pr-3">{row.automation_score.toFixed(1)}</td>
                  <td className="py-3 pr-3">{formatRub(row.monthly_savings_est_rub)}</td>
                  <td className="py-3">{row.quick_win ? t("table.yes") : "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Section>

      {output.roadmap.length > 0 && (
        <Section title={t("sections.roadmap")}>
          <div className="space-y-4">
            {output.roadmap.map((phase) => (
              <div key={phase.phase} className="rounded-lg border border-border p-4">
                <div className="flex flex-wrap items-baseline gap-2">
                  <h3 className="text-sm font-medium">{phase.phase}</h3>
                  <span className="text-xs text-text3">{phase.period}</span>
                </div>
                {phase.processes.length > 0 && (
                  <p className="mt-2 text-xs text-text2">
                    {t("roadmap.processes")}: {phase.processes.join(", ")}
                  </p>
                )}
                <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-text2">
                  {phase.deliverables.map((d) => (
                    <li key={d}>{d}</li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </Section>
      )}

      {output.risks.length > 0 && (
        <Section title={t("sections.risks")}>
          <div className="space-y-3">
            {output.risks.map((r) => (
              <div key={r.risk} className="rounded-lg border border-border p-4">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">{r.risk}</span>
                  <SeverityBadge severity={r.severity} t={t} />
                </div>
                <p className="mt-1 text-sm text-text2">{r.mitigation}</p>
              </div>
            ))}
          </div>
        </Section>
      )}

      {output.next_steps.length > 0 && (
        <Section title={t("sections.nextSteps")}>
          <ol className="list-decimal space-y-1 pl-5 text-sm text-text2">
            {output.next_steps.map((s) => (
              <li key={s}>{s}</li>
            ))}
          </ol>
        </Section>
      )}

      {output.metrics_to_track.length > 0 && (
        <Section title={t("sections.metrics")}>
          <ul className="list-disc space-y-1 pl-5 text-sm text-text2">
            {output.metrics_to_track.map((m) => (
              <li key={m}>{m}</li>
            ))}
          </ul>
        </Section>
      )}
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-border bg-bg p-6">
      <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">{title}</h2>
      <div className="mt-3">{children}</div>
    </div>
  );
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border p-4">
      <p className="text-[10px] uppercase tracking-wider text-text3">{label}</p>
      <p className="mt-1 text-lg font-medium">{value}</p>
    </div>
  );
}

function SeverityBadge({ severity, t }: { severity: string; t: ReturnType<typeof useTranslations> }) {
  const cls =
    severity === "high"
      ? "bg-error/10 text-error"
      : severity === "medium"
        ? "bg-warning/10 text-warning"
        : "bg-bg2 text-text3";
  return (
    <span className={cn("rounded px-1.5 py-0.5 text-[10px] uppercase", cls)}>
      {t(`severity.${severity}` as "severity.high")}
    </span>
  );
}
