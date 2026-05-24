"use client";

import { useEffect, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { useParams } from "next/navigation";
import { fetchPublicReport, type PublicReport } from "@/lib/api";
import { CalculatorInputSummary } from "@/components/tools/CalculatorInputSummary";
import { CalculatorReportView } from "@/components/tools/CalculatorReportView";
import { TooltipProvider } from "@/components/ui/HelpTooltip";

export default function PublicSharePage() {
  const t = useTranslations("calculator.share");
  const locale = useLocale();
  const params = useParams();
  const token = params.token as string;
  const [report, setReport] = useState<PublicReport | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    fetchPublicReport(token)
      .then(setReport)
      .catch(() => setError(t("notFound")));
  }, [token, t]);

  if (error) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-bg3 p-6">
        <p className="text-sm text-text2">{error}</p>
      </main>
    );
  }

  if (!report) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-bg3 p-6">
        <p className="text-sm text-text2">{t("loading")}</p>
      </main>
    );
  }

  return (
    <TooltipProvider prefix="calculator" locale={locale}>
      <main className="min-h-screen bg-bg3 p-6">
        <div className="mx-auto max-w-5xl space-y-6">
          <div className="rounded-xl border border-border bg-bg p-8">
            <div className="mb-6 flex items-center gap-2.5">
              <div className="flex h-7 w-7 items-center justify-center rounded-md bg-text text-sm font-semibold text-white">
                E
              </div>
              <span className="text-sm font-medium">Erman AI</span>
            </div>
            <p className="text-xs text-text3">{t("disclaimer")}</p>

            <div className="mt-8 space-y-8">
              <CalculatorInputSummary input={report.input} />

              <div className="border-t border-border pt-8">
                <h2 className="mb-6 text-[10px] font-medium uppercase tracking-wider text-text3">
                  {t("reportTitle")}
                </h2>
                <CalculatorReportView
                  processName={report.process_name}
                  output={report.output}
                  withTooltips
                  showHeader={false}
                />
              </div>
            </div>
          </div>

          <p className="text-center text-xs text-text3">{t("readOnlyNote")}</p>
        </div>
      </main>
    </TooltipProvider>
  );
}
