"use client";

import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  deleteAdminUser,
  fetchAdminPlans,
  fetchAdminUsers,
  fetchMe,
  impersonateAdminUser,
  updateAdminUserPlan,
  type AdminPlan,
  type AdminUserRow,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 50;

function formatDate(value?: string | null) {
  if (!value) return "—";
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function userStatus(user: AdminUserRow): "blocked" | "verified" | "pending" {
  if (user.is_blocked) return "blocked";
  if (user.email_verified_at) return "verified";
  return "pending";
}

export function AdminUsersEditor() {
  const t = useTranslations("admin.users");
  const [users, setUsers] = useState<AdminUserRow[]>([]);
  const [plans, setPlans] = useState<AdminPlan[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [search, setSearch] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [usersRes, plansRes, meRes] = await Promise.all([
        fetchAdminUsers({ q: query, limit: PAGE_SIZE, offset }),
        fetchAdminPlans(),
        fetchMe().catch(() => null),
      ]);
      setUsers(usersRes.items);
      setTotal(usersRes.total);
      setPlans(plansRes.items.filter((p) => !p.is_archived));
      setCurrentUserId(meRes?.id ?? null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [query, offset, t]);

  useEffect(() => {
    load();
  }, [load]);

  function handleSearchSubmit(e: React.FormEvent) {
    e.preventDefault();
    setOffset(0);
    setQuery(search.trim());
  }

  async function handlePlanChange(user: AdminUserRow, planId: string) {
    if (user.plan_id === planId) return;
    setBusyId(user.id);
    setError("");
    setSuccess("");
    try {
      const updated = await updateAdminUserPlan(user.id, planId);
      setUsers((prev) => prev.map((u) => (u.id === updated.id ? updated : u)));
      setSuccess(t("planUpdated"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setBusyId(null);
    }
  }

  async function handleDelete(user: AdminUserRow) {
    if (user.role === "superadmin") return;
    if (!window.confirm(t("deleteConfirm", { email: user.email }))) return;
    setBusyId(user.id);
    setError("");
    setSuccess("");
    try {
      await deleteAdminUser(user.id);
      setSuccess(t("deleted"));
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("deleteFailed"));
    } finally {
      setBusyId(null);
    }
  }

  async function handleImpersonate(user: AdminUserRow) {
    if (user.role === "superadmin" || user.is_blocked) return;
    if (!window.confirm(t("impersonateConfirm", { email: user.email }))) return;
    setBusyId(user.id);
    setError("");
    setSuccess("");
    try {
      await impersonateAdminUser(user.id);
      window.location.href = "/";
    } catch (err) {
      setError(err instanceof Error ? err.message : t("impersonateFailed"));
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

      <form onSubmit={handleSearchSubmit} className="flex flex-wrap gap-2">
        <input
          type="search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder={t("searchPlaceholder")}
          className="min-w-[240px] flex-1 rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
        />
        <button
          type="submit"
          className="rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2"
        >
          {t("search")}
        </button>
      </form>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}
      {success && (
        <div className="rounded-md border border-green-200 bg-success-bg px-3 py-2 text-sm text-success">
          {success}
        </div>
      )}

      {loading ? (
        <p className="text-sm text-text2">{t("loading")}</p>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-border bg-bg">
          <table className="w-full min-w-[920px] text-left text-sm">
            <thead>
              <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
                <th className="px-4 py-3">{t("colEmail")}</th>
                <th className="px-4 py-3">{t("colStatus")}</th>
                <th className="px-4 py-3">{t("colPlan")}</th>
                <th className="px-4 py-3">{t("colRole")}</th>
                <th className="px-4 py-3">{t("colSegment")}</th>
                <th className="px-4 py-3">{t("colCreated")}</th>
                <th className="px-4 py-3">{t("colActive")}</th>
                <th className="px-4 py-3">{t("colActions")}</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => {
                const isSuperadmin = user.role === "superadmin";
                const isSelf = currentUserId === user.id;
                const canEditPlan = !isSuperadmin || isSelf;
                const disabled = busyId === user.id;
                return (
                  <tr key={user.id} className="border-b border-border last:border-0">
                    <td className="px-4 py-3">
                      <div className="font-medium">{user.email}</div>
                    </td>
                    <td className="px-4 py-3">
                      {(() => {
                        const status = userStatus(user);
                        return (
                          <span
                            className={cn(
                              "inline-block rounded px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide",
                              status === "blocked" && "bg-red-50 text-red-700",
                              status === "verified" && "bg-success-bg text-success",
                              status === "pending" && "bg-warning-bg text-warning"
                            )}
                          >
                            {t(`status.${status}`)}
                          </span>
                        );
                      })()}
                    </td>
                    <td className="px-4 py-3">
                      {canEditPlan ? (
                        <select
                          disabled={disabled || plans.length === 0}
                          value={user.plan_id ?? ""}
                          onChange={(e) => handlePlanChange(user, e.target.value)}
                          className="rounded-lg border border-border2 px-2 py-1.5 text-sm outline-none focus:border-accent"
                        >
                          <option value="" disabled>
                            {t("selectPlan")}
                          </option>
                          {plans.map((plan) => (
                            <option key={plan.id} value={plan.id}>
                              {plan.name}
                            </option>
                          ))}
                        </select>
                      ) : (
                        <span className="text-text2">{user.plan_name ?? "—"}</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-text2">{user.role}</td>
                    <td className="px-4 py-3 text-text2">{user.account_segment}</td>
                    <td className="px-4 py-3 text-text2 whitespace-nowrap">
                      {formatDate(user.created_at)}
                    </td>
                    <td className="px-4 py-3 text-text2 whitespace-nowrap">
                      {formatDate(user.last_active_at)}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-2">
                        <button
                          type="button"
                          disabled={disabled || isSuperadmin || user.is_blocked}
                          onClick={() => handleImpersonate(user)}
                          className={cn(
                            "rounded-lg border border-border2 px-2.5 py-1.5 text-xs hover:bg-bg2 disabled:opacity-50"
                          )}
                        >
                          {t("impersonate")}
                        </button>
                        <button
                          type="button"
                          disabled={disabled || isSuperadmin}
                          onClick={() => handleDelete(user)}
                          className="rounded-lg border border-red-200 px-2.5 py-1.5 text-xs text-red-700 hover:bg-red-50 disabled:opacity-50"
                        >
                          {t("delete")}
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {!loading && totalPages > 1 && (
        <div className="flex items-center justify-between text-sm text-text2">
          <span>{t("pagination", { page, total: totalPages, count: total })}</span>
          <div className="flex gap-2">
            <button
              type="button"
              disabled={offset === 0}
              onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
              className="rounded-lg border border-border2 px-3 py-1.5 disabled:opacity-50"
            >
              {t("prev")}
            </button>
            <button
              type="button"
              disabled={offset + PAGE_SIZE >= total}
              onClick={() => setOffset(offset + PAGE_SIZE)}
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
