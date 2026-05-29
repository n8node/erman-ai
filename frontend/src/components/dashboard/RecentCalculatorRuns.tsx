"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { listRuns, type RunListItem } from "@/lib/api";
import { ProcessNameWithStatus } from "@/components/tools/RecommendationStatus";
import { useIsGuest } from "@/context/AuthContext";

export function RecentCalculatorRuns() {
  const t = useTranslations("dashboard");
  const tGuest = useTranslations("guest");
  const tc = useTranslations("calculator");
  const router = useRouter();
  const isGuest = useIsGuest();
  const [items, setItems] = useState<RunListItem[]>([]);

  useEffect(() => {
    if (isGuest) return;
    listRuns({ tool_slug: "calculator", limit: 5 })
      .then((data) => setItems(data.items))
      .catch(() => setItems([]));
  }, [isGuest]);

  if (isGuest) {
    return (
      <div className="mt-10">
        <h2 className="text-sm font-medium">{t("recentRuns.title")}</h2>
        <p className="mt-2 text-sm text-text2">{tGuest("historyMessage")}</p>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className="mt-10">
        <h2 className="text-sm font-medium">{t("recentRuns.title")}</h2>
        <p className="mt-2 text-sm text-text2">{t("recentRuns.empty")}</p>
      </div>
    );
  }

  return (
    <div className="mt-10">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-medium">{t("recentRuns.title")}</h2>
        <Link
          href="/tools/calculator/history"
          className="text-xs text-accent hover:underline"
        >
          {t("recentRuns.viewAll")}
        </Link>
      </div>
      <div className="mt-3 overflow-x-auto rounded-xl border border-border bg-bg">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
              <th className="px-4 py-2 font-medium">{tc("history.colProcess")}</th>
              <th className="px-4 py-2 font-medium">{tc("history.colBenefit")}</th>
              <th className="px-4 py-2 font-medium">{tc("history.colPayback")}</th>
              <th className="px-4 py-2 font-medium">{tc("history.colDate")}</th>
            </tr>
          </thead>
          <tbody>
            {items.map((row) => (
              <tr
                key={row.id}
                className="border-b border-border last:border-0 cursor-pointer hover:bg-bg2"
                onClick={() => router.push(`/tools/calculator?run=${row.id}`)}
              >
                <td className="px-4 py-2 font-medium">
                  <ProcessNameWithStatus
                    name={row.process_name || "—"}
                    recommendation={row.recommendation}
                    statusLabel={
                      row.recommendation
                        ? tc(`history.rec.${row.recommendation}` as "history.rec.automate")
                        : undefined
                    }
                  />
                </td>
                <td className="px-4 py-2 text-text2">
                  {formatRub(row.net_benefit_monthly)}
                </td>
                <td className="px-4 py-2 text-text2">
                  {row.payback_months != null && isFinite(row.payback_months)
                    ? `${row.payback_months.toFixed(1)} ${tc("metrics.months")}`
                    : "—"}
                </td>
                <td className="px-4 py-2 text-text3 text-xs whitespace-nowrap">
                  {formatDate(row.created_at)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function formatRub(n?: number) {
  if (n == null || !isFinite(n)) return "—";
  return new Intl.NumberFormat("ru-RU").format(Math.round(n)) + " ₽";
}

function formatDate(iso: string) {
  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).format(new Date(iso));
}
