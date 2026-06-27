"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { useEffect, useState } from "react";
import { intlLocale } from "@/i18n/intl-locale";
import { deleteRun, listRuns, type RunListItem } from "@/lib/api";
import { cn } from "@/lib/utils";

export function LegalScanHistory() {
  const t = useTranslations("legalScan.history");
  const locale = useLocale();
  const router = useRouter();
  const [items, setItems] = useState<RunListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [deletingId, setDeletingId] = useState<string | null>(null);

  useEffect(() => {
    listRuns({ tool_slug: "legal-scan", limit: 50 })
      .then((data) => setItems(data.items))
      .catch((err) => setError(err instanceof Error ? err.message : t("loadFailed")))
      .finally(() => setLoading(false));
  }, [t]);

  async function handleDelete(id: string) {
    if (!confirm(t("deleteConfirm"))) return;
    setDeletingId(id);
    try {
      await deleteRun(id);
      setItems((prev) => prev.filter((r) => r.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("deleteFailed"));
    } finally {
      setDeletingId(null);
    }
  }

  if (loading) {
    return <p className="text-sm text-text2">{t("loading")}</p>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </div>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}

      {items.length === 0 ? (
        <div className="rounded-xl border border-border bg-bg p-8 text-center">
          <p className="text-sm text-text2">{t("empty")}</p>
          <Link
            href="/tools/legal-scan"
            className="mt-4 inline-block text-sm text-accent hover:underline"
          >
            {t("emptyAction")}
          </Link>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-border bg-bg">
          <table className="w-full min-w-[620px] text-left text-sm">
            <thead>
              <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
                <th className="px-4 py-3 font-medium">{t("colSite")}</th>
                <th className="px-4 py-3 font-medium">{t("colStatus")}</th>
                <th className="px-4 py-3 font-medium">{t("colDate")}</th>
                <th className="px-4 py-3 font-medium" />
              </tr>
            </thead>
            <tbody>
              {items.map((row) => (
                <tr key={row.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-3 font-medium text-text">{row.process_name || "—"}</td>
                  <td className="px-4 py-3">
                    <StatusBadge status={row.status} t={t} />
                  </td>
                  <td className="px-4 py-3 text-xs text-text3 whitespace-nowrap">
                    {new Date(row.created_at).toLocaleString(intlLocale(locale), {
                      day: "2-digit",
                      month: "short",
                      year: "numeric",
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </td>
                  <td className="px-4 py-3 text-right whitespace-nowrap">
                    <button
                      type="button"
                      onClick={() => router.push(`/tools/legal-scan?run=${row.id}`)}
                      className="mr-3 text-xs text-accent hover:underline"
                    >
                      {row.status === "done" ? t("open") : t("continue")}
                    </button>
                    <button
                      type="button"
                      disabled={deletingId === row.id}
                      onClick={() => handleDelete(row.id)}
                      className="text-xs text-text3 hover:text-red-600 disabled:opacity-50"
                    >
                      {t("delete")}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function StatusBadge({ status, t }: { status: string; t: ReturnType<typeof useTranslations> }) {
  const colors: Record<string, string> = {
    done: "bg-success-bg text-success",
    error: "bg-error-bg text-error",
    processing: "bg-accent-bg text-accent",
    pending: "bg-bg2 text-text3",
  };
  return (
    <span className={cn("inline-block rounded px-2 py-0.5 text-xs", colors[status] ?? colors.pending)}>
      {t(`status.${status}` as "status.done")}
    </span>
  );
}
