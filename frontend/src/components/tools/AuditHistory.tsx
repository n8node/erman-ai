"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useLocale, useTranslations } from "next-intl";
import { intlLocale } from "@/i18n/intl-locale";
import { listRuns, type RunListItem } from "@/lib/api";

export function AuditHistory() {
  const t = useTranslations("audit.history");
  const locale = useLocale();
  const [items, setItems] = useState<RunListItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    listRuns({ tool_slug: "audit", limit: 50 })
      .then((data) => setItems(data.items))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <p className="text-sm text-text2">{t("loading")}</p>;
  }

  if (items.length === 0) {
    return (
      <div className="rounded-xl border border-border bg-bg p-8 text-center">
        <p className="text-sm text-text2">{t("empty")}</p>
        <Link href="/tools/audit" className="mt-3 inline-block text-sm text-accent hover:underline">
          {t("createFirst")}
        </Link>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto rounded-xl border border-border bg-bg">
      <table className="w-full min-w-[620px] text-left text-sm">
        <thead>
          <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
            <th className="px-4 py-3 font-medium">{t("columns.date")}</th>
            <th className="px-4 py-3 font-medium">{t("columns.company")}</th>
            <th className="px-4 py-3 font-medium">{t("columns.status")}</th>
            <th className="px-4 py-3 font-medium">{t("columns.action")}</th>
          </tr>
        </thead>
        <tbody>
          {items.map((run) => {
            const company = run.process_name || "—";
            return (
              <tr key={run.id} className="border-b border-border last:border-0">
                <td className="px-4 py-3 text-text2">
                  {new Date(run.created_at).toLocaleString(intlLocale(locale), {
                    day: "2-digit",
                    month: "short",
                    year: "numeric",
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                </td>
                <td className="px-4 py-3">{company}</td>
                <td className="px-4 py-3">
                  <StatusBadge status={run.status} t={t} />
                </td>
                <td className="px-4 py-3">
                  {run.status === "done" ? (
                    <Link href={`/tools/audit?run=${run.id}`} className="text-accent hover:underline">
                      {t("view")}
                    </Link>
                  ) : run.status === "pending" || run.status === "processing" ? (
                    <Link href={`/tools/audit?run=${run.id}`} className="text-accent hover:underline">
                      {t("continue")}
                    </Link>
                  ) : (
                    <span className="text-text3">—</span>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function StatusBadge({ status, t }: { status: string; t: ReturnType<typeof useTranslations> }) {
  const colors: Record<string, string> = {
    done: "text-success bg-success/10",
    error: "text-error bg-error/10",
    processing: "text-accent bg-accent/10",
    pending: "text-text3 bg-bg2",
  };
  return (
    <span className={`inline-block rounded px-2 py-0.5 text-xs ${colors[status] ?? colors.pending}`}>
      {t(`status.${status}` as "status.done")}
    </span>
  );
}
