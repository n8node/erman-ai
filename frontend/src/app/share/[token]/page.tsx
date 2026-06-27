"use client";

import { useEffect, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { useParams } from "next/navigation";
import { fetchPublicReport, type PublicReport } from "@/lib/api";
import { CalculatorShareReport } from "@/components/tools/CalculatorShareReport";
import { SharePlatformPromo } from "@/components/tools/SharePlatformPromo";
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
      <main className="flex min-h-screen items-center justify-center bg-bg3 px-4 py-8 sm:p-6">
        <p className="text-sm text-text2">{error}</p>
      </main>
    );
  }

  if (!report) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-bg3 px-4 py-8 sm:p-6">
        <p className="text-sm text-text2">{t("loading")}</p>
      </main>
    );
  }

  return (
    <TooltipProvider prefix="calculator" locale={locale}>
      <main className="min-h-screen bg-bg3 px-4 py-6 sm:p-6">
        <div className="mx-auto max-w-5xl space-y-6">
          <div className="rounded-xl border border-border bg-bg p-5 sm:p-8">
            <div className="mb-6 flex items-center gap-2.5">
              <div className="flex h-7 w-7 items-center justify-center rounded-md bg-text text-sm font-semibold text-white">
                E
              </div>
              <span className="text-sm font-medium">Erman AI</span>
            </div>
            <p className="text-xs text-text3">{t("disclaimer")}</p>

            <div className="mt-8">
              <CalculatorShareReport
                input={report.input}
                output={report.output}
                processName={report.process_name}
              />
            </div>
          </div>

          <div className="space-y-4 text-center">
            <p className="text-xs text-text3">{t("readOnlyNote")}</p>
            <SharePlatformPromo visible={report.show_platform_cta !== false} />
          </div>
        </div>
      </main>
    </TooltipProvider>
  );
}
