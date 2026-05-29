"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  exportCalculatorPDF,
  getBillingPlan,
  shareRun,
} from "@/lib/api";
import { usePathname } from "next/navigation";
import { loginPathWithReturn } from "@/lib/return-url";
import { LeadForm } from "./LeadForm";
import { CalculatorReportView } from "./CalculatorReportView";
import { ProposalRequestModal } from "./ProposalRequestModal";
import type { CalculatorInput, CalculatorOutput, User } from "@/lib/api";

type Props = {
  runId: string;
  input: CalculatorInput;
  output: CalculatorOutput;
  isGuest?: boolean;
};

export function CalculatorResult({ runId, input, output, isGuest = false }: Props) {
  const t = useTranslations("calculator");
  const tGuest = useTranslations("guest");
  const pathname = usePathname();
  const [user, setUser] = useState<User | null>(null);
  const [plan, setPlan] = useState<Awaited<ReturnType<typeof getBillingPlan>> | null>(null);
  const [shareUrl, setShareUrl] = useState("");
  const [actionError, setActionError] = useState("");
  const [loadingShare, setLoadingShare] = useState(false);
  const [loadingPdf, setLoadingPdf] = useState(false);
  const [proposalModalOpen, setProposalModalOpen] = useState(false);
  const [proposalDone, setProposalDone] = useState(false);

  useEffect(() => {
    if (isGuest) return;
    fetch(`${process.env.NEXT_PUBLIC_API_URL || "/api/v1"}/auth/me`, {
      credentials: "include",
    })
      .then((r) => (r.ok ? r.json() : null))
      .then(setUser)
      .catch(() => undefined);
    getBillingPlan().then(setPlan).catch(() => undefined);
  }, [isGuest]);

  const isPartner = user?.account_segment === "partner";
  const isDirect = user?.account_segment === "direct_lead";
  const canShare = Boolean(plan?.features?.share_report);
  const canPdf = Boolean(plan?.features?.export_pdf);

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
      <CalculatorReportView
        processName={input.process_name}
        output={output}
        withTooltips
      />

      {isGuest && (
        <div className="rounded-lg border border-accent bg-accent-bg px-4 py-3 text-sm text-accent">
          {tGuest("calculatorSaveHint")}{" "}
          <Link href={loginPathWithReturn(pathname)} className="font-medium underline">
            {tGuest("login")}
          </Link>
        </div>
      )}

      {actionError && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{actionError}</div>
      )}

      {!isGuest && (
        <div className="flex flex-wrap gap-2 border-t border-border pt-4">
          <Link
            href={`/tools/proposal?calculator_run_id=${runId}`}
            className="rounded-lg bg-[#534ab7] px-4 py-2 text-sm font-medium text-white hover:opacity-90"
          >
            {t("createProposal")}
          </Link>
          <Link
            href={`/discuss?run_id=${runId}${input.process_name ? `&project_name=${encodeURIComponent(input.process_name)}` : ""}`}
            className="rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2"
          >
            {t("discussProject")}
          </Link>
        </div>
      )}

      {isGuest && (
        <div className="flex flex-wrap gap-2 border-t border-border pt-4">
          <Link
            href={`/discuss${input.process_name ? `?project_name=${encodeURIComponent(input.process_name)}` : ""}`}
            className="rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2"
          >
            {t("discussProject")}
          </Link>
        </div>
      )}

      {isPartner && !isGuest && (
        <div className="flex flex-wrap gap-2 border-t border-border pt-4">
          <button type="button" onClick={handleShare} disabled={loadingShare || !canShare} className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-50">
            {loadingShare ? t("sharing") : t("publicLink")}
          </button>
          <button type="button" onClick={handlePdf} disabled={loadingPdf || !canPdf} className="rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2 disabled:opacity-50">
            {loadingPdf ? t("exporting") : t("exportPdf")}
          </button>
          <button
            type="button"
            onClick={() => setProposalModalOpen(true)}
            disabled={proposalDone}
            className="rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2 disabled:opacity-50"
          >
            {proposalDone ? t("proposalRequestSent") : t("requestProposal")}
          </button>
          {!canShare && <p className="w-full text-xs text-text3">{t("shareProOnly")}</p>}
        </div>
      )}

      {!isGuest && (
        <ProposalRequestModal
          runId={runId}
          open={proposalModalOpen}
          onClose={() => setProposalModalOpen(false)}
          onSuccess={() => setProposalDone(true)}
        />
      )}

      {proposalDone && (
        <div className="rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
          {t("proposalRequestSuccess")}
        </div>
      )}

      {shareUrl && (
        <div className="rounded-lg border border-accent bg-accent-bg px-4 py-3 text-sm">
          <p className="font-medium text-accent">{t("shareReady")}</p>
          <a href={shareUrl} target="_blank" rel="noopener noreferrer" className="mt-1 break-all text-accent underline">{shareUrl}</a>
        </div>
      )}

      {isDirect && !isGuest && <LeadForm runId={runId} defaultEmail={user?.email} />}
    </div>
  );
}
