"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  deleteRun,
  listRuns,
  type RunListItem,
} from "@/lib/api";
import { cn } from "@/lib/utils";
import {
  ProcessNameWithStatus,
  recommendationBadgeStyles,
} from "./RecommendationStatus";

const recStyles = recommendationBadgeStyles;

export function CalculatorHistory() {
  const t = useTranslations("calculator");
  const router = useRouter();
  const [items, setItems] = useState<RunListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [deletingId, setDeletingId] = useState<string | null>(null);

  useEffect(() => {
    listRuns({ tool_slug: "calculator", limit: 50 })
      .then((data) => {
        setItems(data.items);
        setTotal(data.total);
      })
      .catch((err) =>
        setError(err instanceof Error ? err.message : t("history.loadFailed"))
      )
      .finally(() => setLoading(false));
  }, [t]);

  async function handleDelete(id: string) {
    if (!confirm(t("history.deleteConfirm"))) return;
    setDeletingId(id);
    try {
      await deleteRun(id);
      setItems((prev) => prev.filter((r) => r.id !== id));
      setTotal((n) => Math.max(0, n - 1));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("history.deleteFailed"));
    } finally {
      setDeletingId(null);
    }
  }

  if (loading) {
    return <p className="text-sm text-text2">{t("history.loading")}</p>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-base font-medium">{t("history.title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("history.subtitle")}</p>
      </div>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}

      {items.length === 0 ? (
        <div className="rounded-xl border border-border bg-bg p-8 text-center">
          <p className="text-sm text-text2">{t("history.empty")}</p>
          <Link
            href="/tools/calculator"
            className="mt-4 inline-block text-sm text-accent hover:underline"
          >
            {t("history.emptyAction")}
          </Link>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-border bg-bg">
          <table className="w-full min-w-[820px] text-left text-sm">
            <thead>
              <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
                <th className="px-4 py-3 font-medium">{t("history.colProcess")}</th>
                <th className="px-4 py-3 font-medium">{t("history.colBenefit")}</th>
                <th className="px-4 py-3 font-medium">{t("history.colPayback")}</th>
                <th className="px-4 py-3 font-medium">{t("history.colRoi")}</th>
                <th className="px-4 py-3 font-medium">{t("history.colRec")}</th>
                <th className="px-4 py-3 font-medium">{t("history.colDate")}</th>
                <th className="px-4 py-3 font-medium" />
              </tr>
            </thead>
            <tbody>
              {items.map((row) => (
                <tr key={row.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-3 font-medium">
                    <ProcessNameWithStatus
                      name={row.process_name || "—"}
                      recommendation={row.recommendation}
                      statusLabel={
                        row.recommendation
                          ? t(`history.rec.${row.recommendation}` as "history.rec.automate")
                          : undefined
                      }
                    />
                  </td>
                  <td className="px-4 py-3 text-text2">
                    {formatRub(row.net_benefit_monthly)}
                  </td>
                  <td className="px-4 py-3 text-text2">
                    {formatPayback(row.payback_months, t)}
                  </td>
                  <td className="px-4 py-3 text-text2">
                    {row.roi_horizon_pct != null && isFinite(row.roi_horizon_pct)
                      ? `${row.roi_horizon_pct.toFixed(1)}%`
                      : "—"}
                  </td>
                  <td className="px-4 py-3">
                    {row.recommendation ? (
                      <span
                        className={cn(
                          "inline-block rounded px-2 py-0.5 text-xs",
                          recStyles[row.recommendation] || "bg-bg2 text-text2"
                        )}
                      >
                        {t(`history.rec.${row.recommendation}` as "history.rec.automate")}
                      </span>
                    ) : (
                      "—"
                    )}
                  </td>
                  <td className="px-4 py-3 text-text3 text-xs whitespace-nowrap">
                    {formatDate(row.created_at)}
                  </td>
                  <td className="px-4 py-3 text-right whitespace-nowrap">
                    <button
                      type="button"
                      onClick={() => router.push(`/tools/calculator?run=${row.id}`)}
                      className="text-xs text-accent hover:underline mr-3"
                    >
                      {t("history.open")}
                    </button>
                    <button
                      type="button"
                      disabled={deletingId === row.id}
                      onClick={() => handleDelete(row.id)}
                      className="text-xs text-text3 hover:text-red-600 disabled:opacity-50"
                    >
                      {t("history.delete")}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {total > items.length && (
            <p className="border-t border-border px-4 py-2 text-xs text-text3">
              {t("history.showing", { count: items.length, total })}
            </p>
          )}
        </div>
      )}
    </div>
  );
}

function formatRub(n?: number) {
  if (n == null || !isFinite(n)) return "—";
  return new Intl.NumberFormat("ru-RU").format(Math.round(n)) + " ₽";
}

function formatPayback(n: number | undefined, t: ReturnType<typeof useTranslations>) {
  if (n == null || !isFinite(n) || n > 1e6) return "—";
  return `${n.toFixed(1)} ${t("metrics.months")}`;
}

function formatDate(iso: string) {
  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(iso));
}
