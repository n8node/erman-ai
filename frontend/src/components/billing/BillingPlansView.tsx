"use client";

import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import {
  ApiError,
  createBillingCheckout,
  fetchBillingPlans,
  fetchMe,
  getBillingPlan,
  switchBillingPlan,
  type BillingPlan,
  type PublicPlan,
  type User,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const FEATURE_KEYS = [
  "export_pdf",
  "export_docx",
  "api_access",
  "share_report",
  "priority_queue",
  "white_label",
] as const;

const TOOL_KEYS = ["calculator", "strategy", "proposal"] as const;

function formatRub(n: number) {
  return new Intl.NumberFormat("ru-RU").format(n) + " ₽";
}

function formatLimit(n: number, t: ReturnType<typeof useTranslations>) {
  if (n === -1) return t("unlimited");
  return t("runsPerMonth", { count: n });
}

export function BillingPlansView() {
  const t = useTranslations("billing");
  const searchParams = useSearchParams();
  const paymentStatus = searchParams.get("payment");

  const [plans, setPlans] = useState<PublicPlan[]>([]);
  const [current, setCurrent] = useState<BillingPlan | null>(null);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const isSuperadmin = user?.role === "superadmin";

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [plansRes, currentPlan, me] = await Promise.all([
        fetchBillingPlans(),
        getBillingPlan(),
        fetchMe(),
      ]);
      setPlans(plansRes.items);
      setCurrent(currentPlan);
      setUser(me);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (paymentStatus === "success") {
      setNotice(t("paymentSuccess"));
      void load();
    } else if (paymentStatus === "failed") {
      setError(t("paymentFailed"));
    }
  }, [paymentStatus, load, t]);

  async function handleAction(plan: PublicPlan) {
    if (current?.plan_id === plan.id) return;

    setBusyId(plan.id);
    setError("");
    setNotice("");

    const isFree = plan.price_monthly_rub === 0;
    const paymentsEnabled = Boolean(current?.payments_enabled);

    try {
      if (isFree || isSuperadmin) {
        await switchBillingPlan(plan.id);
        setNotice(t("activated", { name: plan.name }));
        await load();
        return;
      }

      if (!paymentsEnabled) {
        setError(t("paymentUnavailable"));
        return;
      }

      const checkout = await createBillingCheckout(plan.id);
      window.location.href = checkout.checkout_url;
    } catch (err) {
      if (err instanceof ApiError && err.status === 402 && paymentsEnabled) {
        try {
          const checkout = await createBillingCheckout(plan.id);
          window.location.href = checkout.checkout_url;
          return;
        } catch (checkoutErr) {
          setError(checkoutErr instanceof Error ? checkoutErr.message : t("actionFailed"));
          return;
        }
      }
      setError(err instanceof Error ? err.message : t("actionFailed"));
    } finally {
      setBusyId(null);
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
        {current && (
          <p className="mt-3 text-sm text-text2">
            {t("currentPlan")}: <span className="font-medium text-text">{current.plan_name}</span>
          </p>
        )}
        {isSuperadmin && (
          <p className="mt-2 text-xs text-text3">{t("adminDirectSwitchHint")}</p>
        )}
      </div>

      {current?.usage?.tools && (
        <div className="rounded-xl border border-border bg-bg p-5">
          <p className="text-[10px] font-medium uppercase tracking-wider text-text3">{t("usageTitle")}</p>
          <ul className="mt-3 space-y-2 text-sm">
            {TOOL_KEYS.map((slug) => {
              const row = current.usage.tools?.[slug];
              if (!row) return null;
              const label =
                row.limit === -1
                  ? t("usageUnlimited")
                  : t("usageRemaining", { used: row.used, limit: row.limit });
              return (
                <li key={slug} className="flex justify-between gap-3">
                  <span className="text-text2">{t(`tools.${slug}` as "tools.calculator")}</span>
                  <span className="font-medium text-text">{label}</span>
                </li>
              );
            })}
          </ul>
        </div>
      )}

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}
      {notice && (
        <div className="rounded-md border border-accent bg-accent-bg px-3 py-2 text-sm text-accent">
          {notice}
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {plans.map((plan) => {
          const isCurrent = current?.plan_id === plan.id;
          const isFree = plan.price_monthly_rub === 0;
          const disabled = busyId === plan.id || isCurrent;

          const buttonLabel =
            busyId === plan.id
              ? t("processing")
              : isCurrent
                ? t("currentButton")
                : isFree || isSuperadmin
                  ? t("activate")
                  : t("purchase");

          return (
            <article
              key={plan.id}
              className={cn(
                "flex min-w-0 flex-col rounded-xl border bg-bg p-5 sm:p-6",
                isCurrent ? "border-text shadow-sm" : "border-border"
              )}
            >
              <div className="space-y-2">
                <div className="flex flex-wrap items-center gap-2">
                  <h2 className="text-lg font-medium">{plan.name}</h2>
                  {isCurrent && (
                    <span className="rounded bg-bg2 px-2 py-0.5 text-[10px] font-medium uppercase tracking-wider text-text2">
                      {t("currentBadge")}
                    </span>
                  )}
                </div>
                <div>
                  <p className="text-xl font-medium">
                    {isFree ? t("free") : formatRub(plan.price_monthly_rub)}
                  </p>
                  {!isFree && (
                    <p className="text-xs text-text3">
                      {t("perMonth")}
                      {plan.price_yearly_rub > 0 && (
                        <> · {formatRub(plan.price_yearly_rub)} {t("perYear")}</>
                      )}
                    </p>
                  )}
                </div>
              </div>

              <ul className="mt-5 space-y-2 text-sm text-text2">
                {TOOL_KEYS.map((slug) =>
                  plan.tool_limits[slug] !== undefined ? (
                    <li key={slug} className="flex justify-between gap-3">
                      <span className="min-w-0">{t(`tools.${slug}` as "tools.calculator")}</span>
                      <span className="shrink-0 font-medium text-text">
                        {formatLimit(plan.tool_limits[slug], t)}
                      </span>
                    </li>
                  ) : null
                )}
              </ul>

              <ul className="mt-4 flex-1 space-y-1.5 border-t border-border pt-4 text-sm">
                {FEATURE_KEYS.map((key) => {
                  const enabled = Boolean(plan.features[key]);
                  return (
                    <li
                      key={key}
                      className={cn(
                        "flex items-center gap-2",
                        enabled ? "text-text" : "text-text3 line-through"
                      )}
                    >
                      <span
                        className={cn(
                          "inline-block h-1.5 w-1.5 shrink-0 rounded-full",
                          enabled ? "bg-success" : "bg-border2"
                        )}
                      />
                      {t(`features.${key}` as "features.export_pdf")}
                    </li>
                  );
                })}
              </ul>

              <button
                type="button"
                disabled={disabled}
                onClick={() => handleAction(plan)}
                className={cn(
                  "mt-6 w-full rounded-lg px-4 py-2.5 text-sm font-medium disabled:opacity-50",
                  isCurrent
                    ? "border border-border2 bg-bg2 text-text2"
                    : "bg-text text-white"
                )}
              >
                {buttonLabel}
              </button>
            </article>
          );
        })}
      </div>
    </div>
  );
}
