"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense, useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import {
  getRun,
  listRuns,
  submitProjectInquiry,
  submitProjectInquiryPublic,
  type RunListItem,
} from "@/lib/api";
import { useAuthUser } from "@/context/AuthContext";
import { GuestBanner } from "@/components/layout/GuestBanner";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

function formatRub(n?: number | null) {
  if (n == null || !isFinite(n)) return "—";
  return new Intl.NumberFormat("ru-RU").format(Math.round(n)) + " ₽";
}

function formatPayback(n?: number | null, monthsLabel = "мес") {
  if (n == null || !isFinite(n) || n <= 0 || n > 1e6) return "—";
  return `${n.toFixed(1)} ${monthsLabel}`;
}

function DiscussProjectFormInner() {
  const t = useTranslations("discuss");
  const locale = useLocale();
  const user = useAuthUser();
  const searchParams = useSearchParams();
  const runIdParam = searchParams.get("run_id") || searchParams.get("calculator_run_id") || "";
  const projectNameParam = searchParams.get("project_name") || "";

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [telegram, setTelegram] = useState("");
  const [projectTitle, setProjectTitle] = useState(projectNameParam);
  const [projectDescription, setProjectDescription] = useState("");
  const [calculatorRunId, setCalculatorRunId] = useState(runIdParam);
  const [calcRuns, setCalcRuns] = useState<RunListItem[]>([]);
  const [selectedRunPreview, setSelectedRunPreview] = useState<{
    process_name?: string;
    net_benefit_monthly?: number;
    payback_months?: number;
    recommendation?: string;
  } | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [done, setDone] = useState<"auth" | "guest" | null>(null);
  const [honeypot, setHoneypot] = useState("");

  useEffect(() => {
    if (user?.email) {
      setEmail(user.email);
    }
  }, [user]);

  useEffect(() => {
    if (!user) return;
    listRuns({ tool_slug: "calculator", limit: 50 })
      .then((data) => setCalcRuns(data.items.filter((r) => r.status === "done")))
      .catch(() => setCalcRuns([]));
  }, [user]);

  useEffect(() => {
    if (!calculatorRunId || !user) {
      setSelectedRunPreview(null);
      return;
    }
    const fromList = calcRuns.find((r) => r.id === calculatorRunId);
    if (fromList) {
      setSelectedRunPreview({
        process_name: fromList.process_name,
        net_benefit_monthly: fromList.net_benefit_monthly,
        payback_months: fromList.payback_months,
        recommendation: fromList.recommendation,
      });
      if (!projectTitle && fromList.process_name) {
        setProjectTitle(fromList.process_name);
      }
      return;
    }
    getRun(calculatorRunId)
      .then((run) => {
        const out = run.output as { net_benefit_monthly?: number; payback_months?: number; recommendation?: string } | undefined;
        const inp = run.input as { process_name?: string } | undefined;
        setSelectedRunPreview({
          process_name: inp?.process_name,
          net_benefit_monthly: out?.net_benefit_monthly,
          payback_months: out?.payback_months,
          recommendation: out?.recommendation,
        });
      })
      .catch(() => setSelectedRunPreview(null));
  }, [calculatorRunId, user, calcRuns, projectTitle]);

  const steps = useMemo(
    () => [t("steps.calc"), t("steps.describe"), t("steps.talk")],
    [t]
  );

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    const payload = {
      name: name.trim(),
      email: email.trim(),
      telegram: telegram.trim(),
      project_title: projectTitle.trim(),
      project_description: projectDescription.trim(),
      calculator_run_id: calculatorRunId || undefined,
      locale,
      website: honeypot,
    };
    try {
      if (user) {
        await submitProjectInquiry(payload);
        setDone("auth");
      } else {
        await submitProjectInquiryPublic(payload);
        setDone("guest");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.submitFailed"));
    } finally {
      setLoading(false);
    }
  }

  if (done) {
    return (
      <div className="mx-auto max-w-2xl space-y-6">
        <GuestBanner />
        <div className="rounded-xl border border-green-200 bg-green-50 p-6 text-sm text-green-900">
          <h2 className="text-base font-medium">{t("success.title")}</h2>
          <p className="mt-2">{done === "guest" ? t("success.guest") : t("success.auth")}</p>
          {telegram && (
            <p className="mt-2 text-green-800">{t("success.telegramHint", { telegram: telegram.startsWith("@") ? telegram : `@${telegram}` })}</p>
          )}
        </div>
        <Link href="/tools/calculator" className="text-sm text-accent hover:underline">
          {t("success.backToCalculator")}
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <GuestBanner />

      <div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </div>

      <div className="rounded-xl border border-accent bg-accent-bg p-5 text-sm text-accent">
        <p className="font-medium">{t("value.title")}</p>
        <ul className="mt-2 list-disc space-y-1 pl-5 text-accent/90">
          <li>{t("value.item1")}</li>
          <li>{t("value.item2")}</li>
          <li>{t("value.item3")}</li>
        </ul>
      </div>

      <div className="grid gap-3 sm:grid-cols-3">
        {steps.map((label, i) => (
          <div key={label} className="rounded-lg border border-border bg-bg p-4">
            <p className="text-[10px] font-medium uppercase tracking-wider text-text3">
              {t("steps.label", { n: i + 1 })}
            </p>
            <p className="mt-1 text-sm">{label}</p>
          </div>
        ))}
      </div>

      <form onSubmit={handleSubmit} className="rounded-xl border border-border bg-bg p-6 space-y-5">
        {error && (
          <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
        )}

        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("form.name")} *</label>
            <input required className={fieldClass} value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("form.email")} *</label>
            <input
              required
              type="email"
              readOnly={!!user}
              className={cn(fieldClass, user && "bg-bg2 text-text2")}
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
            {user && <p className="mt-1 text-[10px] text-text3">{t("form.emailVerified")}</p>}
            {!user && <p className="mt-1 text-[10px] text-text3">{t("form.emailGuestHint")}</p>}
          </div>
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("form.telegram")} *</label>
          <input
            required
            className={fieldClass}
            value={telegram}
            onChange={(e) => setTelegram(e.target.value)}
            placeholder={t("form.telegramPlaceholder")}
          />
          <p className="mt-1 text-[10px] text-text3">{t("form.telegramHint")}</p>
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("form.projectTitle")}</label>
          <input className={fieldClass} value={projectTitle} onChange={(e) => setProjectTitle(e.target.value)} />
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("form.description")} *</label>
          <textarea
            required
            rows={5}
            className={fieldClass}
            value={projectDescription}
            onChange={(e) => setProjectDescription(e.target.value)}
            placeholder={t("form.descriptionPlaceholder")}
          />
        </div>

        {user ? (
          <div>
            <label className="mb-1.5 block text-xs font-medium">{t("form.calculatorRun")}</label>
            <select
              className={fieldClass}
              value={calculatorRunId}
              onChange={(e) => setCalculatorRunId(e.target.value)}
            >
              <option value="">{t("form.noRun")}</option>
              {calcRuns.map((run) => (
                <option key={run.id} value={run.id}>
                  {run.process_name || t("form.unnamedRun")} · {formatRub(run.net_benefit_monthly)} ·{" "}
                  {new Date(run.created_at).toLocaleDateString(locale === "en" ? "en-GB" : "ru-RU")}
                </option>
              ))}
            </select>
            {selectedRunPreview && (
              <div className="mt-3 rounded-lg border border-border bg-bg2 p-4 text-sm">
                <p className="text-[10px] font-medium uppercase tracking-wider text-text3">{t("form.runPreview")}</p>
                <p className="mt-1 font-medium">{selectedRunPreview.process_name || "—"}</p>
                <p className="mt-1 text-text2">
                  {t("form.runBenefit", { value: formatRub(selectedRunPreview.net_benefit_monthly) })} ·{" "}
                  {t("form.runPayback", { value: formatPayback(selectedRunPreview.payback_months, t("form.months")) })}
                </p>
              </div>
            )}
          </div>
        ) : (
          <div className="rounded-lg border border-border bg-bg2 p-4 text-sm text-text2">
            <p>{t("form.guestRunHint")}</p>
            <Link href="/tools/calculator" className="mt-2 inline-block text-accent hover:underline">
              {t("form.goCalculator")}
            </Link>
          </div>
        )}

        <input
          type="text"
          name="website"
          value={honeypot}
          onChange={(e) => setHoneypot(e.target.value)}
          className="hidden"
          tabIndex={-1}
          autoComplete="off"
          aria-hidden
        />

        <div className="flex justify-end border-t border-border pt-4">
          <button
            type="submit"
            disabled={loading}
            className="rounded-lg bg-text px-5 py-2.5 text-sm font-medium text-white hover:opacity-90 disabled:opacity-60"
          >
            {loading ? t("form.sending") : t("form.submit")}
          </button>
        </div>
      </form>
    </div>
  );
}

export function DiscussProjectForm() {
  return (
    <Suspense fallback={<p className="text-sm text-text2">…</p>}>
      <DiscussProjectFormInner />
    </Suspense>
  );
}
