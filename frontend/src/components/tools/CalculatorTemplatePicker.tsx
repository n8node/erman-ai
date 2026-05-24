"use client";

import { useLocale, useTranslations } from "next-intl";
import { useState } from "react";
import {
  CALCULATOR_SECTORS,
  applyProcessTemplate,
  localize,
  type CalculatorProcessTemplate,
} from "@/lib/calculator-templates";
import type { CalculatorInput } from "@/lib/api";
import { cn } from "@/lib/utils";

type Props = {
  onApply: (patch: Partial<CalculatorInput>, templateId: string) => void;
  activeTemplateId?: string | null;
};

export function CalculatorTemplatePicker({ onApply, activeTemplateId }: Props) {
  const t = useTranslations("calculator.templates");
  const locale = useLocale();
  const [sectorId, setSectorId] = useState(CALCULATOR_SECTORS[0]?.id ?? "");
  const [expanded, setExpanded] = useState(true);

  const sector = CALCULATOR_SECTORS.find((s) => s.id === sectorId) ?? CALCULATOR_SECTORS[0];

  function handleApply(template: CalculatorProcessTemplate) {
    onApply(applyProcessTemplate(template, locale), template.id);
  }

  return (
    <div className="rounded-xl border border-border bg-bg2 p-4 space-y-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="text-sm font-medium">{t("title")}</h3>
          <p className="mt-0.5 text-xs text-text2">{t("subtitle")}</p>
        </div>
        <button
          type="button"
          onClick={() => setExpanded((v) => !v)}
          className="text-xs text-accent hover:underline shrink-0"
        >
          {expanded ? t("collapse") : t("expand")}
        </button>
      </div>

      {expanded && (
        <>
          <div className="flex flex-wrap gap-2">
            {CALCULATOR_SECTORS.map((s) => (
              <button
                key={s.id}
                type="button"
                onClick={() => setSectorId(s.id)}
                className={cn(
                  "rounded-lg border px-3 py-1.5 text-xs transition-colors",
                  sector.id === s.id
                    ? "border-text bg-bg font-medium text-text"
                    : "border-border text-text2 hover:border-border2 hover:text-text"
                )}
              >
                {localize(s.name, locale)}
              </button>
            ))}
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            {sector.templates.map((template) => (
              <button
                key={template.id}
                type="button"
                onClick={() => handleApply(template)}
                className={cn(
                  "rounded-lg border bg-bg p-3 text-left transition-colors hover:border-border2",
                  activeTemplateId === template.id
                    ? "border-accent ring-1 ring-accent/30"
                    : "border-border"
                )}
              >
                <p className="text-sm font-medium">{localize(template.name, locale)}</p>
                <p className="mt-1 text-xs text-text2 line-clamp-2">
                  {localize(template.description, locale)}
                </p>
                <p className="mt-2 text-[10px] text-text3">
                  {template.process_steps.length} {t("steps")} · {template.units_per_month}{" "}
                  {t("unitsPerMonth")} · {template.automation_pct}% {t("automation")}
                </p>
              </button>
            ))}
          </div>

          {activeTemplateId && (
            <p className="text-xs text-accent">{t("applied")}</p>
          )}
        </>
      )}
    </div>
  );
}
