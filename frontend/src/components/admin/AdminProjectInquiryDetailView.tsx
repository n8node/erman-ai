"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import {
  fetchAdminProjectInquiry,
  updateAdminProjectInquiryStatus,
  type AdminProjectInquiryDetail,
} from "@/lib/api";
import { CalculatorShareReport } from "@/components/tools/CalculatorShareReport";
import { TooltipProvider } from "@/components/ui/HelpTooltip";

type Props = {
  id: string;
};

function formatDate(value: string) {
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function AdminProjectInquiryDetailView({ id }: Props) {
  const t = useTranslations("admin.projectInquiries");
  const locale = useLocale();
  const [detail, setDetail] = useState<AdminProjectInquiryDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setDetail(await fetchAdminProjectInquiry(id));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [id, t]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleStatusChange(status: AdminProjectInquiryDetail["status"]) {
    if (!detail || detail.status === status) return;
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      await updateAdminProjectInquiryStatus(id, status);
      setDetail({ ...detail, status });
      setSuccess(t("statusUpdated"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setBusy(false);
    }
  }

  if (loading) return <p className="text-sm text-text2">{t("loading")}</p>;

  if (error && !detail) {
    return (
      <div className="space-y-4">
        <Link href="/admin/project-inquiries" className="text-sm text-accent hover:underline">
          ← {t("backToList")}
        </Link>
        <p className="text-sm text-red-800">{error}</p>
      </div>
    );
  }

  if (!detail) return null;

  const hasCalc = detail.input && detail.output;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <Link href="/admin/project-inquiries" className="text-sm text-accent hover:underline">
            ← {t("backToList")}
          </Link>
          <h1 className="mt-2 text-base font-medium">{t("detailTitle")}</h1>
          <p className="mt-1 text-sm text-text2">
            {formatDate(detail.created_at)} · {detail.email}
            {detail.user_email ? ` · ${detail.user_email}` : ""}
          </p>
        </div>
        <select
          disabled={busy}
          value={detail.status}
          onChange={(e) => handleStatusChange(e.target.value as AdminProjectInquiryDetail["status"])}
          className="rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent"
        >
          <option value="pending_email">{t("status.pending_email")}</option>
          <option value="new">{t("status.new")}</option>
          <option value="in_progress">{t("status.in_progress")}</option>
          <option value="done">{t("status.done")}</option>
          <option value="spam">{t("status.spam")}</option>
        </select>
      </div>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}
      {success && (
        <div className="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-800">{success}</div>
      )}

      <div className="rounded-xl border border-border bg-bg p-6">
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">{t("requestInfo")}</h2>
        <dl className="mt-4 grid gap-4 sm:grid-cols-2 text-sm">
          <div>
            <dt className="text-xs text-text3">{t("colName")}</dt>
            <dd className="mt-1 font-medium">{detail.name}</dd>
          </div>
          <div>
            <dt className="text-xs text-text3">{t("colTelegram")}</dt>
            <dd className="mt-1 font-medium">{detail.telegram}</dd>
          </div>
          <div className="sm:col-span-2">
            <dt className="text-xs text-text3">{t("colProject")}</dt>
            <dd className="mt-1 font-medium">{detail.project_title || "—"}</dd>
          </div>
          <div className="sm:col-span-2">
            <dt className="text-xs text-text3">{t("colDescription")}</dt>
            <dd className="mt-1 whitespace-pre-wrap text-text2">{detail.project_description}</dd>
          </div>
        </dl>
      </div>

      {hasCalc && (
        <TooltipProvider prefix="calculator" locale={locale}>
          <div className="rounded-xl border border-border bg-bg p-8">
            <p className="text-xs text-text3">{t("reportDisclaimer")}</p>
            <div className="mt-8">
              <CalculatorShareReport
                input={detail.input!}
                output={detail.output!}
                processName={detail.process_name || detail.input!.process_name}
              />
            </div>
          </div>
        </TooltipProvider>
      )}
    </div>
  );
}
