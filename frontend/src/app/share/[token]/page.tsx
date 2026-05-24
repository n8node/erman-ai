"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { useParams } from "next/navigation";
import { fetchPublicReport, type PublicReport } from "@/lib/api";

function formatRub(n: number) {
  return new Intl.NumberFormat("ru-RU").format(Math.round(n)) + " ₽";
}

export default function PublicSharePage() {
  const t = useTranslations("calculator.share");
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

  const out = report.output;

  return (
    <main className="min-h-screen bg-bg3 p-6">
      <div className="mx-auto max-w-2xl rounded-xl border border-border bg-bg p-8">
        <div className="mb-6 flex items-center gap-2.5">
          <div className="flex h-7 w-7 items-center justify-center rounded-md bg-text text-sm font-semibold text-white">
            E
          </div>
          <span className="text-sm font-medium">Erman AI</span>
        </div>
        <h1 className="text-base font-medium">{report.process_name}</h1>
        <p className="mt-1 text-xs text-text3">{t("disclaimer")}</p>

        <div className="mt-8 grid gap-4 sm:grid-cols-3">
          <div className="rounded-lg border border-border bg-bg2 p-4">
            <p className="text-[10px] uppercase text-text3">{t("currentCost")}</p>
            <p className="mt-2 text-lg font-medium">{formatRub(out.total_current_cost)}</p>
          </div>
          <div className="rounded-lg border border-border bg-bg2 p-4">
            <p className="text-[10px] uppercase text-text3">{t("savings")}</p>
            <p className="mt-2 text-lg font-medium">{formatRub(out.net_monthly_savings)}</p>
          </div>
          <div className="rounded-lg border border-border bg-bg2 p-4">
            <p className="text-[10px] uppercase text-text3">{t("payback")}</p>
            <p className="mt-2 text-lg font-medium">
              {out.payback_months > 0 && out.payback_months < 1e6
                ? `${out.payback_months.toFixed(1)} ${t("months")}`
                : "—"}
            </p>
          </div>
        </div>

        <div className="mt-6 rounded-lg border border-border bg-bg2 px-4 py-3 text-sm">
          {out.recommendation_text}
        </div>
      </div>
    </main>
  );
}
