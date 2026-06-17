"use client";

import { Check, AlertCircle, Loader2 } from "lucide-react";
import { useTranslations } from "next-intl";
import type { LegalScanCheckItem, LegalScanCrawlMeta } from "@/lib/api-legal-scan";
import { LEGAL_SCAN_CHECKLIST_KEYS } from "@/lib/api-legal-scan";
import { cn } from "@/lib/utils";

type Props = {
  url: string;
  checklist: LegalScanCheckItem[];
  progress: number;
  crawl?: LegalScanCrawlMeta | null;
};

export function LegalScanChecklist({ url, checklist, progress, crawl }: Props) {
  const t = useTranslations("legalScan.scan");
  const tChecklist = useTranslations("legalScan.checklist");

  const checklistLabel = (item: LegalScanCheckItem) =>
    (LEGAL_SCAN_CHECKLIST_KEYS as readonly string[]).includes(item.key)
      ? tChecklist(item.key)
      : item.label;

  return (
    <div className="rounded-xl border border-border bg-bg p-6 shadow-sm">
      <div className="mb-5 flex items-center gap-3">
        <Loader2 className="h-5 w-5 animate-spin text-text" />
        <div className="text-sm">
          <p>
            {t("scanning")} <span className="font-semibold">{url}</span>…
          </p>
          {crawl && crawl.pages_requested > 1 && (
            <p className="mt-1 text-xs text-text3">
              {t("pagesProgress", {
                done: crawl.pages_fetched,
                total: crawl.pages_requested,
              })}
            </p>
          )}
        </div>
      </div>
      <div className="mb-5 h-1.5 overflow-hidden rounded-full bg-bg2">
        <div
          className="h-full rounded-full bg-text transition-all duration-300"
          style={{ width: `${progress}%` }}
        />
      </div>
      <div className="space-y-0.5">
        {checklist.map((item) => (
          <div
            key={item.key}
            className={cn(
              "flex items-center gap-3 border-b py-2.5 text-[13.5px] last:border-0",
              item.status === "ok" && "border-success/20 bg-success-bg/40 text-success -mx-2 px-2 rounded-md",
              item.status === "risk" && "border-border/60 text-text",
              item.status !== "ok" && item.status !== "risk" && "border-border/60 text-text3"
            )}
          >
            <span className="flex h-[18px] w-[18px] shrink-0 items-center justify-center">
              {item.status === "running" && (
                <Loader2 className="h-3.5 w-3.5 animate-spin text-text3" />
              )}
              {item.status === "ok" && <Check className="h-4 w-4 text-success" strokeWidth={2.5} />}
              {item.status === "risk" && (
                <AlertCircle className="h-4 w-4 text-error" strokeWidth={2.4} />
              )}
              {item.status === "pending" && (
                <span className="h-1.5 w-1.5 rounded-full bg-border2" />
              )}
            </span>
            <span className={cn("flex-1", item.status === "ok" && "font-medium")}>
              {checklistLabel(item)}
            </span>
            {item.status === "ok" && (
              <span className="text-xs font-semibold text-success">{t("ok")}</span>
            )}
            {item.status === "risk" && (
              <span className="text-xs font-semibold text-error">{t("risk")}</span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
