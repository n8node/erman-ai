"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchAdminProjectInquiries,
  updateAdminProjectInquiryStatus,
  type AdminProjectInquiryRow,
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

export function AdminProjectInquiriesEditor() {
  const t = useTranslations("admin.projectInquiries");
  const [items, setItems] = useState<AdminProjectInquiryRow[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await fetchAdminProjectInquiries({ limit: PAGE_SIZE, offset });
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

  async function handleStatusChange(id: string, status: AdminProjectInquiryRow["status"]) {
    setBusyId(id);
    try {
      await updateAdminProjectInquiryStatus(id, status);
      setItems((prev) => prev.map((x) => (x.id === id ? { ...x, status } : x)));
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
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}

      {loading ? (
        <p className="text-sm text-text2">{t("loading")}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-text2">{t("empty")}</p>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-border bg-bg">
          <table className="w-full min-w-[980px] text-left text-sm">
            <thead>
              <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
                <th className="px-4 py-2 font-medium">{t("colDate")}</th>
                <th className="px-4 py-2 font-medium">{t("colName")}</th>
                <th className="px-4 py-2 font-medium">{t("colEmail")}</th>
                <th className="px-4 py-2 font-medium">{t("colTelegram")}</th>
                <th className="px-4 py-2 font-medium">{t("colProject")}</th>
                <th className="px-4 py-2 font-medium">{t("colRun")}</th>
                <th className="px-4 py-2 font-medium">{t("colStatus")}</th>
                <th className="px-4 py-2 font-medium" />
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-3 text-text2">{formatDate(item.created_at)}</td>
                  <td className="px-4 py-3">{item.name}</td>
                  <td className="px-4 py-3 text-text2">{item.email}</td>
                  <td className="px-4 py-3">{item.telegram}</td>
                  <td className="px-4 py-3 max-w-[160px] truncate">{item.project_title || "—"}</td>
                  <td className="px-4 py-3 text-text2">
                    {item.process_name ? `${item.process_name} · ${formatRub(item.net_benefit_monthly)}` : "—"}
                  </td>
                  <td className="px-4 py-3">
                    <select
                      disabled={busyId === item.id}
                      value={item.status}
                      onChange={(e) =>
                        handleStatusChange(item.id, e.target.value as AdminProjectInquiryRow["status"])
                      }
                      className="rounded border border-border2 px-2 py-1 text-xs"
                    >
                      <option value="pending_email">{t("status.pending_email")}</option>
                      <option value="new">{t("status.new")}</option>
                      <option value="in_progress">{t("status.in_progress")}</option>
                      <option value="done">{t("status.done")}</option>
                      <option value="spam">{t("status.spam")}</option>
                    </select>
                  </td>
                  <td className="px-4 py-3">
                    <Link href={`/admin/project-inquiries/${item.id}`} className="text-accent hover:underline text-xs">
                      {t("open")}
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
          <span>{t("pagination", { page, total: totalPages, count: total })}</span>
          <div className="flex gap-2">
            <button
              type="button"
              disabled={offset === 0}
              onClick={() => setOffset((o) => Math.max(0, o - PAGE_SIZE))}
              className="rounded border border-border2 px-3 py-1 disabled:opacity-50"
            >
              {t("prev")}
            </button>
            <button
              type="button"
              disabled={offset + PAGE_SIZE >= total}
              onClick={() => setOffset((o) => o + PAGE_SIZE)}
              className="rounded border border-border2 px-3 py-1 disabled:opacity-50"
            >
              {t("next")}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
