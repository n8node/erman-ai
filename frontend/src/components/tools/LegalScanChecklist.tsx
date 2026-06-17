"use client";

import { Check, AlertCircle, Loader2 } from "lucide-react";
import { useTranslations } from "next-intl";
import type { LegalScanCheckItem } from "@/lib/api-legal-scan";
import { cn } from "@/lib/utils";

type Props = {
  url: string;
  checklist: LegalScanCheckItem[];
  progress: number;
};

export function LegalScanChecklist({ url, checklist, progress }: Props) {
  const t = useTranslations("legalScan.scan");

  return (
    <div className="rounded-xl border border-border bg-bg p-6 shadow-sm">
      <div className="mb-5 flex items-center gap-3">
        <Loader2 className="h-5 w-5 animate-spin text-text" />
        <p className="text-sm">
          {t("scanning")} <span className="font-semibold">{url}</span>…
        </p>
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
              "flex items-center gap-3 border-b border-border/60 py-2.5 text-[13.5px] last:border-0",
              item.status === "risk" ? "text-text" : item.status === "ok" ? "text-text2" : "text-text3"
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
            <span className="flex-1">{item.label}</span>
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
