"use client";

import Link from "next/link";
import { AlertTriangle, Check } from "lucide-react";
import { useTranslations } from "next-intl";
import {
  findRiskForCheckItem,
  type LegalScanInput,
  type LegalScanOutput,
  type LegalScanRiskItem,
} from "@/lib/api-legal-scan";
import { cn } from "@/lib/utils";

type Props = {
  input: LegalScanInput;
  output: LegalScanOutput;
  runId: string;
  canExport?: boolean;
  onRestart: () => void;
};

function formatRub(n: number) {
  return new Intl.NumberFormat("ru-RU").format(n) + " ₽";
}

const severityClass = {
  high: "border-l-[#e0492f]",
  medium: "border-l-[#f0a728]",
  low: "border-l-[#9aa0a6]",
  none: "border-l-success",
};

const badgeClass = {
  high: "bg-[#fef3f2] text-[#b42318] border-[#fbdcd8]",
  medium: "bg-[#fffaeb] text-[#b25e09] border-[#fbe6c2]",
  low: "bg-bg2 text-text2 border-border",
  none: "bg-success-bg text-success border-success/30",
};

function RiskCard({
  title,
  severity,
  explanation,
  article,
  fineText,
  howToFix,
  pageUrls,
  foundData,
  t,
}: {
  title: string;
  severity: LegalScanRiskItem["severity"] | "none";
  explanation: string;
  article?: string;
  fineText?: string;
  howToFix?: string;
  pageUrls?: string[];
  foundData?: string[];
  t: ReturnType<typeof useTranslations<"legalScan.result">>;
}) {
  return (
    <div
      className={cn(
        "rounded-[10px] border border-border border-l-[3px] bg-bg p-4 shadow-sm",
        severityClass[severity]
      )}
    >
      <div className="mb-2 flex flex-wrap items-center gap-2">
        <span className="text-[14.5px] font-semibold">{title}</span>
        <span
          className={cn(
            "rounded-full border px-2 py-0.5 text-[11px] font-semibold",
            badgeClass[severity]
          )}
        >
          {t(`severity.${severity}`)}
        </span>
      </div>
      <p className="mb-2.5 text-[13.5px] text-text2">{explanation}</p>
      {foundData && foundData.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-1.5">
          {foundData.map((d) => (
            <span
              key={d}
              className="rounded-md border border-border bg-bg2 px-2 py-0.5 font-mono text-[11.5px] text-text"
            >
              {d}
            </span>
          ))}
        </div>
      )}
      {pageUrls && pageUrls.length > 0 && (
        <div className="mb-2.5 flex flex-col gap-1">
          {pageUrls.map((url) => (
            <a
              key={url}
              href={url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-[12.5px] text-accent hover:underline"
            >
              {url}
            </a>
          ))}
        </div>
      )}
      {severity !== "none" && article && fineText && (
        <>
          <div className="flex flex-wrap gap-2 text-[12.5px]">
            <span className="inline-flex items-center gap-1.5 rounded-md border border-border bg-bg2 px-2 py-1 text-text2">
              {t("article")}: <b className="text-text">{article}</b>
            </span>
            <span className="inline-flex items-center gap-1.5 rounded-md border border-border bg-bg2 px-2 py-1 text-text2">
              {t("fine")}: <b className="text-error">{fineText}</b>
            </span>
          </div>
          {howToFix && (
            <div className="mt-2.5 flex items-start gap-2 text-[13px] text-text">
              <Check className="mt-0.5 h-4 w-4 shrink-0 text-success" strokeWidth={2.4} />
              <span>{howToFix}</span>
            </div>
          )}
        </>
      )}
    </div>
  );
}

export function LegalScanResult({ input, output, runId, canExport, onRestart }: Props) {
  const t = useTranslations("legalScan.result");
  const checklist = output.layer1?.checklist ?? [];
  const shownRiskIds = new Set<string>();
  checklist.forEach((item) => {
    if (item.status === "risk") {
      const matched = findRiskForCheckItem(item.key, output.risks);
      if (matched) shownRiskIds.add(matched.risk_id);
    }
  });
  const extraRisks = output.risks.filter((r) => !shownRiskIds.has(r.risk_id));
  const fineDisplay =
    output.summary.fine_max_total > 0
      ? t("fineUpTo", { amount: formatRub(output.summary.fine_max_total) })
      : t("fineNone");

  const showLeakNote =
    input.site_features.forms &&
    (output.summary.turnover_fine_note || output.risks.some((r) => r.risk_id === "data_leak_exposure"));

  return (
    <div className="space-y-5">
      <div className="overflow-hidden rounded-xl border border-border bg-bg shadow-sm">
        <div className="h-1 bg-gradient-to-r from-[#e0492f] to-[#f0a728]" />
        <div className="flex flex-wrap items-center gap-6 p-6">
          <div>
            <p className="text-[34px] font-bold leading-none">{output.summary.risks_count}</p>
            <p className="mt-1.5 text-sm text-text3">{t("risksFound")}</p>
          </div>
          <div className="hidden h-12 w-px bg-border sm:block" />
          <div>
            <p className="text-[34px] font-bold leading-none text-error">{fineDisplay}</p>
            <p className="mt-1.5 text-sm text-text3">{t("fixedFines")}</p>
          </div>
          <div className="ml-auto text-right text-xs text-text3">
            {t("scanned")}
            <br />
            <span className="font-semibold text-text2">{output.layer1.final_url || input.url}</span>
            {output.layer1.crawl && output.layer1.crawl.pages_fetched > 1 && (
              <>
                <br />
                <span className="text-text3">
                  {t("pagesCrawled", {
                    count: output.layer1.crawl.pages_fetched,
                  })}
                </span>
              </>
            )}
          </div>
        </div>
      </div>

      {output.layer1.crawl && output.layer1.crawl.fetched_urls.length > 1 && (
        <details className="rounded-lg border border-border bg-bg2 px-4 py-3 text-xs text-text2">
          <summary className="cursor-pointer font-medium text-text">
            {t("crawledPages", { count: output.layer1.crawl.pages_fetched })}
          </summary>
          <ul className="mt-2 space-y-1">
            {output.layer1.crawl.fetched_urls.map((u) => (
              <li key={u}>
                <a href={u} target="_blank" rel="noopener noreferrer" className="text-accent hover:underline">
                  {u}
                </a>
              </li>
            ))}
          </ul>
        </details>
      )}

      {showLeakNote && output.summary.turnover_fine_note && (
        <div className="flex gap-3 rounded-xl border border-[#fbdcd8] bg-[#fef3f2] px-4 py-3 text-sm text-[#b42318]">
          <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
          <p>
            <b>{t("leakTitle")}</b> {output.summary.turnover_fine_note} {t("leakSeparate")}
          </p>
        </div>
      )}

      <div className="flex flex-wrap items-baseline justify-between gap-3 px-0.5">
        <h3 className="text-[15px] font-semibold">{t("checksTitle")}</h3>
        <div className="flex flex-wrap gap-3 text-xs text-text3">
          <span className="inline-flex items-center gap-1.5">
            <i className="h-2 w-2 rounded-full bg-success" />
            {t("severity.none")}
          </span>
          <span className="inline-flex items-center gap-1.5">
            <i className="h-2 w-2 rounded-full bg-[#e0492f]" />
            {t("severity.high")}
          </span>
          <span className="inline-flex items-center gap-1.5">
            <i className="h-2 w-2 rounded-full bg-[#f0a728]" />
            {t("severity.medium")}
          </span>
          <span className="inline-flex items-center gap-1.5">
            <i className="h-2 w-2 rounded-full bg-[#9aa0a6]" />
            {t("severity.low")}
          </span>
        </div>
      </div>

      {checklist.length === 0 ? (
        <div className="rounded-xl border border-border bg-bg p-6 text-sm text-text2">{t("noRisks")}</div>
      ) : (
        checklist.map((item) => {
          if (item.status === "ok") {
            return (
              <RiskCard
                key={item.key}
                title={item.label}
                severity="none"
                explanation={item.evidence || t("noRiskExplanation")}
                pageUrls={item.page_urls}
                foundData={item.found_data}
                t={t}
              />
            );
          }

          const risk = findRiskForCheckItem(item.key, output.risks);
          return (
            <RiskCard
              key={item.key}
              title={risk?.title ?? item.label}
              severity={risk?.severity ?? "medium"}
              explanation={risk?.explanation ?? item.evidence ?? t("riskDetectedGeneric")}
              article={risk?.article}
              fineText={risk?.fine_text}
              howToFix={risk?.how_to_fix}
              pageUrls={risk?.page_urls ?? item.page_urls}
              foundData={risk?.found_data ?? item.found_data}
              t={t}
            />
          );
        })
      )}

      {extraRisks.map((risk) => (
        <RiskCard
          key={risk.risk_id}
          title={risk.title}
          severity={risk.severity}
          explanation={risk.explanation}
          article={risk.article}
          fineText={risk.fine_text}
          howToFix={risk.how_to_fix}
          pageUrls={risk.page_urls}
          foundData={risk.found_data}
          t={t}
        />
      ))}

      {output.industry_note && (
        <div className="rounded-[10px] border border-border bg-bg2 px-4 py-3.5 text-[13px] text-text2">
          {output.industry_note}
        </div>
      )}

      <p className="px-0.5 text-xs leading-relaxed text-text3">{output.disclaimer}</p>

      <div className="flex flex-wrap items-center gap-3 border-t border-border pt-5">
        <Link
          href={`/tools/proposal?legal_scan_run_id=${runId}`}
          className="rounded-lg bg-text px-5 py-2.5 text-sm font-medium text-white hover:opacity-90"
        >
          {t("toProposal")}
        </Link>
        {canExport ? (
          <button
            type="button"
            disabled
            className="rounded-lg border border-border2 px-5 py-2.5 text-sm font-medium text-text2 opacity-60"
            title={t("exportSoon")}
          >
            {t("exportPdf")}
          </button>
        ) : (
          <span className="text-xs text-text3">{t("exportLocked")}</span>
        )}
        <button
          type="button"
          onClick={onRestart}
          className="ml-auto text-sm text-text3 hover:text-text"
        >
          {t("restart")}
        </button>
      </div>
    </div>
  );
}
