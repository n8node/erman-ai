"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Download, Lock } from "lucide-react";
import { exportProposalPDF, getBillingPlan } from "@/lib/api";
import type { ProposalInput, ProposalOutput } from "@/lib/api-proposal";
import { cn } from "@/lib/utils";

type Props = {
  input: ProposalInput;
  output: ProposalOutput;
  runId: string;
};

function Prose({ children }: { children: React.ReactNode }) {
  return <p className="text-sm text-text2 whitespace-pre-wrap leading-relaxed">{children}</p>;
}

export function ProposalResult({ input, output, runId }: Props) {
  const t = useTranslations("proposal.result");
  const [canPdf, setCanPdf] = useState(false);
  const [canDocx, setCanDocx] = useState(false);
  const [loadingPdf, setLoadingPdf] = useState(false);
  const [pdfError, setPdfError] = useState("");

  useEffect(() => {
    getBillingPlan()
      .then((plan) => {
        setCanPdf(Boolean(plan?.features?.export_pdf));
        setCanDocx(Boolean(plan?.features?.export_docx));
      })
      .catch(() => undefined);
  }, []);

  async function handlePdf() {
    setPdfError("");
    setLoadingPdf(true);
    try {
      const blob = await exportProposalPDF(runId);
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "proposal.pdf";
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      setPdfError(err instanceof Error ? err.message : t("pdfFailed"));
    } finally {
      setLoadingPdf(false);
    }
  }

  const costFormatted = new Intl.NumberFormat("ru-RU").format(input.project_cost_rub) + " ₽";

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-bg p-6">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p className="text-[10px] font-medium uppercase tracking-wider text-text3">{t("document")}</p>
            <h2 className="mt-1 text-base font-medium">{input.solution_name}</h2>
            <p className="mt-1 text-sm text-text2">
              {input.client_company}
              {input.client_contact ? ` · ${input.client_contact}` : ""}
            </p>
            {input.calculator_context && (
              <p className="mt-2 text-xs text-success">
                {t("linkedCalculator", { process: input.calculator_context.process_name })}
              </p>
            )}
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={canPdf ? handlePdf : undefined}
              disabled={!canPdf || loadingPdf}
              className={cn(
                "inline-flex items-center gap-2 rounded-lg border px-3 py-2 text-xs font-medium",
                canPdf ? "border-border2 hover:bg-bg2" : "border-border text-text3 cursor-not-allowed"
              )}
            >
              {!canPdf && <Lock size={14} />}
              {loadingPdf ? t("exporting") : t("exportPdf")}
            </button>
            <button
              type="button"
              disabled
              className="inline-flex items-center gap-2 rounded-lg border border-border px-3 py-2 text-xs font-medium text-text3 cursor-not-allowed"
            >
              {!canDocx && <Lock size={14} />}
              {t("exportDocx")}
            </button>
          </div>
        </div>
        {pdfError && (
          <div className="mt-3 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{pdfError}</div>
        )}
      </div>

      <Section title={t("greeting")}>
        <Prose>{output.greeting}</Prose>
      </Section>

      <Section title={t("taskUnderstanding")}>
        <Prose>{output.task_understanding}</Prose>
      </Section>

      <Section title={t("proposedSolution")}>
        <Prose>{output.proposed_solution}</Prose>
      </Section>

      <div className="grid gap-4 md:grid-cols-2">
        <Section title={t("scopeIncluded")}>
          <ul className="list-disc space-y-1 pl-5 text-sm text-text2">
            {output.scope_included.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </Section>
        <Section title={t("scopeExcluded")}>
          <ul className="list-disc space-y-1 pl-5 text-sm text-text2">
            {output.scope_excluded.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </Section>
      </div>

      {output.timeline.length > 0 && (
        <Section title={t("timeline")}>
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead>
                <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
                  <th className="px-3 py-2">{t("phase")}</th>
                  <th className="px-3 py-2">{t("duration")}</th>
                  <th className="px-3 py-2">{t("description")}</th>
                </tr>
              </thead>
              <tbody>
                {output.timeline.map((ph) => (
                  <tr key={ph.title} className="border-b border-border last:border-0">
                    <td className="px-3 py-2 font-medium text-text">{ph.title}</td>
                    <td className="px-3 py-2 text-text2">{ph.duration_weeks} {t("weeks")}</td>
                    <td className="px-3 py-2 text-text2">{ph.description}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Section>
      )}

      <Section title={t("cost")}>
        <p className="text-lg font-semibold text-success">{costFormatted}</p>
        <div className="mt-2">
          <Prose>{output.cost_summary}</Prose>
        </div>
      </Section>

      <Section title={t("payment")}>
        <Prose>{output.payment_terms}</Prose>
      </Section>

      <Section title={t("whyUs")}>
        <Prose>{output.why_us}</Prose>
      </Section>

      <Section title={t("nextStep")}>
        <Prose>{output.next_step}</Prose>
      </Section>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-border bg-bg p-5">
      <p className="text-[10px] font-medium uppercase tracking-wider text-text3">{title}</p>
      <div className="mt-3">{children}</div>
    </div>
  );
}
