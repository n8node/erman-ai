"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useCallback, useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchTools,
  getRun,
  isToolLimitError,
  runLegalScan,
  type ToolListItem,
} from "@/lib/api";
import {
  DEFAULT_LEGAL_SCAN_INPUT,
  LEGAL_SCAN_FEATURE_CHIPS,
  LEGAL_SCAN_INDUSTRY_OPTIONS,
  LEGAL_SCAN_SIZE_OPTIONS,
  isLegalScanInputValid,
  type LegalScanCheckItem,
  type LegalScanInput,
  type LegalScanOutput,
} from "@/lib/api-legal-scan";
import { isLimitReached } from "@/lib/tool-limits";
import { ToolLimitBadge } from "@/components/dashboard/ToolLimitBadge";
import { GuestBanner } from "@/components/layout/GuestBanner";
import { GuestLoginModal } from "@/components/auth/GuestLoginModal";
import { useAuthUser } from "@/context/AuthContext";
import { ToolLimitExceededAlert } from "./ToolLimitExceededAlert";
import { LegalScanChecklist } from "./LegalScanChecklist";
import { LegalScanResult } from "./LegalScanResult";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const DEFAULT_CHECKLIST: LegalScanCheckItem[] = [
  { key: "ssl", label: "SSL-сертификат и HTTPS", status: "pending" },
  { key: "privacy", label: "Политика обработки ПД", status: "pending" },
  { key: "cookie", label: "Cookie-баннер", status: "pending" },
  { key: "cookiepol", label: "Отдельная политика cookie", status: "pending" },
  { key: "consent", label: "Согласие у форм", status: "pending" },
  { key: "req", label: "Реквизиты (ИНН/ОГРН)", status: "pending" },
  { key: "contacts", label: "Контактные данные", status: "pending" },
  { key: "offer", label: "Публичная оферта", status: "pending" },
  { key: "terms", label: "Пользовательское соглашение", status: "pending" },
  { key: "withdraw", label: "Отзыв согласия / удаление данных", status: "pending" },
  { key: "admark", label: "Маркировка рекламы (erid)", status: "pending" },
  { key: "trackers", label: "Трекеры и аналитика", status: "pending" },
  { key: "formenc", label: "Шифрование форм", status: "pending" },
];

function LegalScanWizardInner() {
  const t = useTranslations("legalScan");
  const tLimits = useTranslations("toolLimits");
  const router = useRouter();
  const searchParams = useSearchParams();
  const runIdParam = searchParams.get("run");

  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [input, setInput] = useState<LegalScanInput>(DEFAULT_LEGAL_SCAN_INPUT);
  const [runId, setRunId] = useState<string | null>(null);
  const [result, setResult] = useState<{
    input: LegalScanInput;
    output: LegalScanOutput;
    runId: string;
  } | null>(null);
  const [checklist, setChecklist] = useState<LegalScanCheckItem[]>(DEFAULT_CHECKLIST);
  const [progress, setProgress] = useState(0);
  const [loading, setLoading] = useState(false);
  const [loadingRun, setLoadingRun] = useState(!!runIdParam);
  const [error, setError] = useState("");
  const [limitExceeded, setLimitExceeded] = useState(false);
  const [scanTool, setScanTool] = useState<ToolListItem | null>(null);
  const [loginModalOpen, setLoginModalOpen] = useState(false);
  const animRef = useRef(0);
  const user = useAuthUser();

  const loadToolLimits = useCallback(() => {
    return fetchTools()
      .then((data) => {
        const tool = data.tools.find((x) => x.slug === "legal-scan") ?? null;
        setScanTool(tool);
        if (tool) setLimitExceeded(isLimitReached(tool));
        return tool;
      })
      .catch(() => null);
  }, []);

  useEffect(() => {
    if (!user) return;
    void loadToolLimits();
  }, [user, loadToolLimits]);

  useEffect(() => {
    if (!runIdParam) {
      setLoadingRun(false);
      return;
    }
    if (!user) {
      setLoadingRun(false);
      return;
    }
    setLoadingRun(true);
    getRun(runIdParam)
      .then((run) => {
        if (run.status === "done" && run.input && run.output) {
          setResult({
            input: run.input as unknown as LegalScanInput,
            output: run.output as unknown as LegalScanOutput,
            runId: run.id,
          });
          setStep(3);
        } else if (run.status === "pending" || run.status === "processing") {
          setRunId(run.id);
          setStep(2);
        } else if (run.status === "error") {
          setError(run.error_msg || t("errors.loadFailed"));
        }
      })
      .catch(() => setError(t("errors.loadFailed")))
      .finally(() => setLoadingRun(false));
  }, [runIdParam, user, t]);

  useEffect(() => {
    if (step !== 2 || !runId) return;
    let cancelled = false;
    animRef.current = 0;

    async function poll() {
      if (cancelled) return;
      try {
        const run = await getRun(runId!);
        const partial = run.output as unknown as LegalScanOutput | null;
        if (partial?.layer1?.checklist?.length) {
          setChecklist(partial.layer1.checklist);
          const done = partial.layer1.checklist.filter(
            (c) => c.status === "ok" || c.status === "risk"
          ).length;
          setProgress(Math.round((done / partial.layer1.checklist.length) * 100));
        } else {
          setChecklist((prev) => {
            const idx = animRef.current % prev.length;
            animRef.current += 1;
            return prev.map((item, i) =>
              i === idx ? { ...item, status: "running" } : item
            );
          });
          setProgress((p) => Math.min(p + 8, 85));
        }

        if (run.status === "done" && run.output) {
          setResult({
            input: run.input as unknown as LegalScanInput,
            output: run.output as unknown as LegalScanOutput,
            runId: run.id,
          });
          setProgress(100);
          setStep(3);
        } else if (run.status === "error") {
          setError(run.error_msg || t("errors.scanFailed"));
          setStep(1);
        }
      } catch {
        // keep polling
      }
    }

    void poll();
    const timer = setInterval(poll, 2000);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [step, runId, t]);

  async function handleScan() {
    if (!user) {
      setLoginModalOpen(true);
      return;
    }
    setError("");
    setLoading(true);
    setChecklist(DEFAULT_CHECKLIST.map((c) => ({ ...c, status: "pending" })));
    setProgress(0);
    try {
      const data = await runLegalScan(input);
      setRunId(data.run_id);
      setStep(2);
      router.replace(`/tools/legal-scan?run=${data.run_id}`);
      void loadToolLimits();
    } catch (err) {
      if (isToolLimitError(err)) {
        setLimitExceeded(true);
      } else {
        setError(err instanceof Error ? err.message : t("errors.scanFailed"));
      }
    } finally {
      setLoading(false);
    }
  }

  function handleRestart() {
    setStep(1);
    setRunId(null);
    setResult(null);
    setError("");
    router.replace("/tools/legal-scan");
  }

  function toggleFeature(key: keyof LegalScanInput["site_features"]) {
    setInput((prev) => ({
      ...prev,
      site_features: {
        ...prev.site_features,
        [key]: !prev.site_features[key],
      },
    }));
  }

  const scanLimitReached = limitExceeded;
  const canExport = !!user;

  if (loadingRun) {
    return (
      <div className="mx-auto max-w-5xl">
        <p className="text-sm text-text2">{t("loading")}</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <GuestBanner />
      {scanTool && user && (
        <ToolLimitBadge
          tool={scanTool}
          className="mb-4"
          t={(key, values) => tLimits(key, values as Record<string, string | number> | undefined)}
        />
      )}
      {limitExceeded && <ToolLimitExceededAlert toolName={t("title")} />}

      <h1 className="text-[21px] font-semibold">{t("title")}</h1>
      <p className="mt-1 mb-5 text-sm text-text2">{t("subtitle")}</p>

      <div className="mb-6 flex flex-wrap items-center gap-0">
        {([1, 2, 3] as const).map((n, i) => (
          <div key={n} className="flex items-center">
            {i > 0 && <span className="mx-3 hidden h-px w-11 bg-border sm:block" />}
            <div
              className={cn(
                "flex items-center gap-2",
                step === n && "font-medium text-text",
                step > n && "text-text2"
              )}
            >
              <span
                className={cn(
                  "flex h-[23px] w-[23px] items-center justify-center rounded-full text-xs font-semibold",
                  step >= n ? "bg-text text-white" : "bg-bg2 text-text3"
                )}
              >
                {n}
              </span>
              <span className={cn("text-[13.5px]", step < n && "text-text3")}>
                {t(`steps.${n}`)}
              </span>
            </div>
          </div>
        ))}
      </div>

      {error && (
        <div className="mb-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}

      {step === 1 && (
        <div className="rounded-xl border border-border bg-bg p-6 shadow-sm">
          <p className="mb-4 text-[10px] font-medium uppercase tracking-wider text-text3">
            {t("form.section")}
          </p>
          <div className="mb-4">
            <label className="mb-1.5 block text-[13px] text-text2">{t("form.url")}</label>
            <input
              type="url"
              className={fieldClass}
              placeholder="https://client.ru"
              value={input.url}
              onChange={(e) => setInput((p) => ({ ...p, url: e.target.value }))}
            />
          </div>
          <div className="mb-4 grid gap-4 sm:grid-cols-2">
            <div>
              <label className="mb-1.5 block text-[13px] text-text2">{t("form.industry")}</label>
              <select
                className={fieldClass}
                value={input.industry}
                onChange={(e) => setInput((p) => ({ ...p, industry: e.target.value }))}
              >
                {LEGAL_SCAN_INDUSTRY_OPTIONS.map((o) => (
                  <option key={o.id} value={o.id}>
                    {t(o.labelKey)}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="mb-1.5 block text-[13px] text-text2">{t("form.size")}</label>
              <select
                className={fieldClass}
                value={input.company_size}
                onChange={(e) => setInput((p) => ({ ...p, company_size: e.target.value }))}
              >
                {LEGAL_SCAN_SIZE_OPTIONS.map((o) => (
                  <option key={o.id} value={o.id}>
                    {t(o.labelKey)}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <div className="mb-2">
            <label className="mb-2 block text-[13px] text-text2">
              {t("form.features")}{" "}
              <span className="text-text3">— {t("form.featuresHint")}</span>
            </label>
            <div className="flex flex-wrap gap-2">
              {LEGAL_SCAN_FEATURE_CHIPS.map(({ key, labelKey }) => {
                const on = input.site_features[key];
                return (
                  <button
                    key={key}
                    type="button"
                    onClick={() => toggleFeature(key)}
                    className={cn(
                      "rounded-full border px-3.5 py-1.5 text-[13px] transition-colors",
                      on
                        ? "border-[#c7c2f5] bg-[#f1f0fe] text-[#473fc4]"
                        : "border-border2 bg-bg text-text2 hover:border-border"
                    )}
                  >
                    {t(labelKey)}
                  </button>
                );
              })}
            </div>
            <p className="mt-2 text-xs text-text3">{t("form.featuresNote")}</p>
          </div>
          <div className="mt-6 flex justify-end">
            <button
              type="button"
              disabled={!isLegalScanInputValid(input) || loading || scanLimitReached}
              onClick={() => void handleScan()}
              className="rounded-lg bg-text px-5 py-2.5 text-sm font-medium text-white disabled:cursor-not-allowed disabled:bg-border2 disabled:text-text3"
            >
              {loading ? t("form.scanning") : t("form.submit")}
            </button>
          </div>
        </div>
      )}

      {step === 2 && (
        <LegalScanChecklist url={input.url} checklist={checklist} progress={progress} />
      )}

      {step === 3 && result && (
        <LegalScanResult
          input={result.input}
          output={result.output}
          runId={result.runId}
          canExport={canExport}
          onRestart={handleRestart}
        />
      )}

      <GuestLoginModal open={loginModalOpen} onClose={() => setLoginModalOpen(false)} />
    </div>
  );
}

export function LegalScanWizard() {
  return (
    <Suspense
      fallback={
        <div className="mx-auto max-w-5xl">
          <p className="text-sm text-text2">…</p>
        </div>
      }
    >
      <LegalScanWizardInner />
    </Suspense>
  );
}
