"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import type { CalculatorInput, CalculatorOutput, User } from "@/lib/api";
import {
  exportCalculatorPDF,
  getBillingPlan,
  shareRun,
} from "@/lib/api";
import { LeadForm } from "./LeadForm";
import { cn } from "@/lib/utils";

type Props = {
  runId: string;
  input: CalculatorInput;
  output: CalculatorOutput;
};

function formatRub(n: number) {
  if (!isFinite(n)) return "—";
  return new Intl.NumberFormat("ru-RU").format(Math.round(n)) + " ₽";
}

export function CalculatorResult({ runId, input, output }: Props) {
  const t = useTranslations("calculator");
  const [user, setUser] = useState<User | null>(null);
  const [plan, setPlan] = useState<Awaited<ReturnType<typeof getBillingPlan>> | null>(null);
  const [shareUrl, setShareUrl] = useState("");
  const [actionError, setActionError] = useState("");
  const [loadingShare, setLoadingShare] = useState(false);
  const [loadingPdf, setLoadingPdf] = useState(false);
  const [showDetails, setShowDetails] = useState(false);

  useEffect(() => {
    fetch(`${process.env.NEXT_PUBLIC_API_URL || "/api/v1"}/auth/me`, {
      credentials: "include",
    })
      .then((r) => (r.ok ? r.json() : null))
      .then(setUser)
      .catch(() => undefined);
    getBillingPlan().then(setPlan).catch(() => undefined);
  }, []);

  const isPartner = user?.account_segment === "partner";
  const isDirect = user?.account_segment === "direct_lead";
  const canShare = Boolean(plan?.features?.share_report);
  const canPdf = Boolean(plan?.features?.export_pdf);

  const recClass =
    output.recommendation === "automate"
      ? "bg-[#eaf3de] text-[#3b6d11] border-[#c5ddb0]"
      : output.recommendation === "consider"
        ? "bg-[#faeeda] text-[#633806] border-[#e8d5b0]"
        : "bg-[#fcebeb] text-[#a32d2d] border-[#e8b4b4]";

  async function handleShare() {
    setActionError("");
    setLoadingShare(true);
    try {
      const res = await shareRun(runId);
      setShareUrl(res.url);
    } catch (err) {
      setActionError(err instanceof Error ? err.message : t("shareFailed"));
    } finally {
      setLoadingShare(false);
    }
  }

  async function handlePdf() {
    setActionError("");
    setLoadingPdf(true);
    try {
      const blob = await exportCalculatorPDF(runId);
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "calculator-report.pdf";
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      setActionError(err instanceof Error ? err.message : t("pdfFailed"));
    } finally {
      setLoadingPdf(false);
    }
  }

  return (
    <div className="rounded-xl border border-border bg-bg p-6 space-y-6">
      <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
        {t("resultTitle")} — {input.process_name}
      </h2>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Metric label={t("metrics.netBenefit")} value={formatRub(output.net_benefit_monthly)} large />
        <Metric label={t("metrics.payback")} value={formatPayback(output.payback_months, t)} large />
        <Metric label={t("metrics.fte")} value={output.fte.toFixed(2)} sub={t("metrics.fteHint")} />
        <Metric label={t("metrics.roi")} value={`${output.roi_horizon_pct.toFixed(1)}%`} />
      </div>

      <div className={cn("rounded-lg border px-4 py-3 text-sm", recClass)}>
        {output.recommendation_text}
      </div>

      {output.kpi_rows.length > 0 && (
        <section>
          <h3 className="text-[10px] font-medium uppercase tracking-wider text-text3 mb-3">
            {t("kpi.title")}
          </h3>
          <div className="overflow-x-auto rounded-lg border border-border">
            <table className="w-full text-left text-xs">
              <thead>
                <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
                  <th className="px-3 py-2">KPI</th>
                  <th className="px-3 py-2">{t("kpi.before")}</th>
                  <th className="px-3 py-2">{t("kpi.after")}</th>
                  <th className="px-3 py-2">{t("kpi.change")}</th>
                </tr>
              </thead>
              <tbody>
                {output.kpi_rows.map((row) => (
                  <tr key={row.key} className="border-b border-border last:border-0">
                    <td className="px-3 py-2.5 font-medium">{row.label}</td>
                    <td className="px-3 py-2.5 text-text2">{row.before}</td>
                    <td className="px-3 py-2.5">{row.after}</td>
                    <td className="px-3 py-2.5">
                      <span className="rounded-md bg-[#eaf3de] px-2 py-0.5 text-[#3b6d11]">{row.change}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}

      <button type="button" onClick={() => setShowDetails(!showDetails)} className="text-xs text-accent hover:underline">
        {showDetails ? t("wizard.hideExpert") : t("result.showDetails")}
      </button>

      {showDetails && (
        <div className="grid gap-2 sm:grid-cols-3 text-sm text-text2 border-t border-border pt-4">
          <div>TCO: {formatRub(output.tco_horizon)}</div>
          <div>NPV: {formatRub(output.npv)}</div>
          <div>Hm: {Math.round(output.hours_saved_month)} {t("wizard.hoursMonth")}</div>
        </div>
      )}

      {actionError && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{actionError}</div>
      )}

      {isPartner && (
        <div className="flex flex-wrap gap-2 border-t border-border pt-4">
          <button type="button" onClick={handleShare} disabled={loadingShare || !canShare} className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-50">
            {loadingShare ? t("sharing") : t("shareWithClient")}
          </button>
          <button type="button" onClick={handlePdf} disabled={loadingPdf || !canPdf} className="rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2 disabled:opacity-50">
            {loadingPdf ? t("exporting") : t("exportPdf")}
          </button>
          <Link href={`/tools/proposal?calculator_run_id=${runId}`} className="rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2">
            {t("createProposal")}
          </Link>
          {!canShare && <p className="w-full text-xs text-text3">{t("shareProOnly")}</p>}
        </div>
      )}

      {shareUrl && (
        <div className="rounded-lg border border-accent bg-accent-bg px-4 py-3 text-sm">
          <p className="font-medium text-accent">{t("shareReady")}</p>
          <a href={shareUrl} target="_blank" rel="noopener noreferrer" className="mt-1 break-all text-accent underline">{shareUrl}</a>
        </div>
      )}

      {isDirect && <LeadForm runId={runId} defaultEmail={user?.email} />}
    </div>
  );
}

function Metric({ label, value, sub, large }: { label: string; value: string; sub?: string; large?: boolean }) {
  return (
    <div className="rounded-lg border border-border bg-bg2 p-4">
      <p className="text-[10px] uppercase tracking-wider text-text3">{label}</p>
      <p className={cn("mt-2 font-medium", large ? "text-xl" : "text-lg")}>{value}</p>
      {sub && <p className="mt-1 text-[10px] text-text3">{sub}</p>}
    </div>
  );
}

function formatPayback(n: number, t: ReturnType<typeof useTranslations>) {
  if (!isFinite(n) || n <= 0 || n > 1e6) return "—";
  return `${n.toFixed(1)} ${t("metrics.months")}`;
}
