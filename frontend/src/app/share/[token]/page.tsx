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
      <div className="mx-auto max-w-3xl space-y-6">
        <div className="rounded-xl border border-border bg-bg p-8">
          <div className="mb-6 flex items-center gap-2.5">
            <div className="flex h-7 w-7 items-center justify-center rounded-md bg-text text-sm font-semibold text-white">E</div>
            <span className="text-sm font-medium">Erman AI</span>
          </div>
          <h1 className="text-base font-medium">{report.process_name}</h1>
          <p className="mt-1 text-xs text-text3">{t("disclaimer")}</p>

          <div className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Stat label={t("netBenefit")} value={formatRub(out.net_benefit_monthly)} />
            <Stat label={t("payback")} value={out.payback_months < 1e6 ? `${out.payback_months.toFixed(1)} ${t("months")}` : "—"} />
            <Stat label={t("fte")} value={out.fte.toFixed(2)} />
            <Stat label={t("roi")} value={`${out.roi_horizon_pct.toFixed(1)}%`} />
          </div>

          <div className="mt-6 rounded-lg border border-border bg-bg2 px-4 py-3 text-sm">{out.recommendation_text}</div>
        </div>

        {out.kpi_rows?.length > 0 && (
          <div className="rounded-xl border border-border bg-bg p-6 overflow-x-auto">
            <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3 mb-3">{t("kpiTitle")}</h2>
            <table className="w-full text-left text-xs">
              <thead>
                <tr className="border-b border-border text-[10px] uppercase text-text3">
                  <th className="pb-2 pr-3">KPI</th>
                  <th className="pb-2 pr-3">{t("before")}</th>
                  <th className="pb-2 pr-3">{t("after")}</th>
                  <th className="pb-2">{t("change")}</th>
                </tr>
              </thead>
              <tbody>
                {out.kpi_rows.map((row) => (
                  <tr key={row.key} className="border-b border-border">
                    <td className="py-2 pr-3 font-medium">{row.label}</td>
                    <td className="py-2 pr-3 text-text2">{row.before}</td>
                    <td className="py-2 pr-3">{row.after}</td>
                    <td className="py-2"><span className="rounded bg-[#eaf3de] px-1.5 py-0.5 text-[#3b6d11]">{row.change}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </main>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-bg2 p-4">
      <p className="text-[10px] uppercase text-text3">{label}</p>
      <p className="mt-2 text-lg font-medium">{value}</p>
    </div>
  );
}
