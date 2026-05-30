"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { deleteRun, listRuns, type RunListItem } from "@/lib/api";

export function ProposalHistory() {
  const t = useTranslations("proposal.history");
  const router = useRouter();
  const [items, setItems] = useState<RunListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    listRuns({ tool_slug: "proposal", limit: 50 })
      .then((data) => setItems(data.items))
      .catch((err) => setError(err instanceof Error ? err.message : t("loadFailed")))
      .finally(() => setLoading(false));
  }, [t]);

  async function handleDelete(id: string) {
    if (!confirm(t("deleteConfirm"))) return;
    try {
      await deleteRun(id);
      setItems((prev) => prev.filter((r) => r.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("deleteFailed"));
    }
  }

  if (loading) return <p className="text-sm text-text2">{t("loading")}</p>;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </div>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}

      {items.length === 0 ? (
        <div className="rounded-xl border border-border bg-bg p-8 text-center">
          <p className="text-sm text-text2">{t("empty")}</p>
          <Link href="/tools/proposal" className="mt-4 inline-block text-sm text-accent hover:underline">
            {t("emptyAction")}
          </Link>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-border bg-bg">
          <table className="w-full text-left text-xs">
            <thead>
              <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
                <th className="px-4 py-3">{t("colClient")}</th>
                <th className="px-4 py-3">{t("colStatus")}</th>
                <th className="px-4 py-3">{t("colDate")}</th>
                <th className="px-4 py-3">{t("colActions")}</th>
              </tr>
            </thead>
            <tbody>
              {items.map((row) => (
                <tr key={row.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-3 text-sm text-text">{row.process_name || "—"}</td>
                  <td className="px-4 py-3 text-text2">{row.status}</td>
                  <td className="px-4 py-3 text-text2">{new Date(row.created_at).toLocaleString()}</td>
                  <td className="px-4 py-3 space-x-2">
                    <button
                      type="button"
                      onClick={() => router.push(`/tools/proposal?run=${row.id}`)}
                      className="text-accent hover:underline"
                    >
                      {t("open")}
                    </button>
                    <button type="button" onClick={() => handleDelete(row.id)} className="text-red-700 hover:underline">
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
