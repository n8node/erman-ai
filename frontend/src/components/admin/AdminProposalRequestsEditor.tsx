"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchAdminProposalRequests,
  updateAdminProposalRequestStatus,
  type AdminProposalRequestRow,
} from "@/lib/api";

const PAGE_SIZE = 50;

function formatDate(value: string) {
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function formatRub(n?: number | null) {
  if (n == null || !isFinite(n)) return "—";
  return new Intl.NumberFormat("ru-RU").format(Math.round(n)) + " ₽";
}

function formatPayback(n?: number | null) {
  if (n == null || !isFinite(n) || n <= 0 || n > 1e6) return "—";
  return `${n.toFixed(1)} мес`;
}

export function AdminProposalRequestsEditor() {
  const t = useTranslations("admin.proposalRequests");
  const [items, setItems] = useState<AdminProposalRequestRow[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await fetchAdminProposalRequests({
        limit: PAGE_SIZE,
        offset,
      });
      setItems(res.items);
      setTotal(res.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [offset, t]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleStatusChange(
    item: AdminProposalRequestRow,
    status: AdminProposalRequestRow["status"]
  ) {
    if (item.status === status) return;
    setBusyId(item.id);
    setError("");
    setSuccess("");
    try {
      const updated = await updateAdminProposalRequestStatus(item.id, status);
      setItems((prev) =>
        prev.map((row) =>
          row.id === item.id ? { ...row, status: updated.status as AdminProposalRequestRow["status"] } : row
        )
      );
      setSuccess(t("statusUpdated"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setBusyId(null);
    }
  }

  const page = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

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
      {success && (
        <div className="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-800">
          {success}
        </div>
      )}

      {loading ? (
        <p className="text-sm text-text2">{t("loading")}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-text2">{t("empty")}</p>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-border bg-bg">
          <table className="w-full min-w-[1100px] text-left text-sm">
            <thead>
              <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
                <th className="px-4 py-3">{t("colDate")}</th>
                <th className="px-4 py-3">{t("colUser")}</th>
                <th className="px-4 py-3">{t("colRequester")}</th>
                <th className="px-4 py-3">{t("colTelegram")}</th>
                <th className="px-4 py-3">{t("colProcess")}</th>
                <th className="px-4 py-3">{t("colBenefit")}</th>
                <th className="px-4 py-3">{t("colPayback")}</th>
                <th className="px-4 py-3">{t("colStatus")}</th>
                <th className="px-4 py-3">{t("colRun")}</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-3 text-text2">{formatDate(item.created_at)}</td>
                  <td className="px-4 py-3">{item.user_email}</td>
                  <td className="px-4 py-3">{item.requester_name || "—"}</td>
                  <td className="px-4 py-3 text-text2">{item.telegram || "—"}</td>
                  <td className="px-4 py-3 font-medium">{item.process_name || "—"}</td>
                  <td className="px-4 py-3">{formatRub(item.net_benefit_monthly)}</td>
                  <td className="px-4 py-3">{formatPayback(item.payback_months)}</td>
                  <td className="px-4 py-3">
                    <select
                      disabled={busyId === item.id}
                      value={item.status}
                      onChange={(e) =>
                        handleStatusChange(
                          item,
                          e.target.value as AdminProposalRequestRow["status"]
                        )
                      }
                      className="rounded-lg border border-border2 px-2 py-1.5 text-sm outline-none focus:border-accent"
                    >
                      <option value="new">{t("status.new")}</option>
                      <option value="in_progress">{t("status.in_progress")}</option>
                      <option value="done">{t("status.done")}</option>
                    </select>
                  </td>
                  <td className="px-4 py-3">
                    <Link
                      href={`/admin/proposal-requests/${item.id}`}
                      className="text-accent hover:underline"
                    >
                      {t("openRun")}
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {total > PAGE_SIZE && (
        <div className="flex items-center justify-between text-sm text-text2">
          <span>
            {t("pagination", {
              page,
              total: totalPages,
              count: total,
            })}
          </span>
          <div className="flex gap-2">
            <button
              type="button"
              disabled={offset === 0}
              onClick={() => setOffset((v) => Math.max(0, v - PAGE_SIZE))}
              className="rounded-lg border border-border2 px-3 py-1.5 disabled:opacity-50"
            >
              {t("prev")}
            </button>
            <button
              type="button"
              disabled={offset + PAGE_SIZE >= total}
              onClick={() => setOffset((v) => v + PAGE_SIZE)}
              className="rounded-lg border border-border2 px-3 py-1.5 disabled:opacity-50"
            >
              {t("next")}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
