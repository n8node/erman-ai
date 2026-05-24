"use client";

import { useLocale, useTranslations } from "next-intl";
import { useBudgetConfig } from "@/components/tools/BudgetConfigProvider";
import {
  estimateImplementationBudget,
  formatEstimateRub,
  type IntegrationLevel,
} from "@/lib/calculator-budget-estimate";
import type { CalculatorInput } from "@/lib/api";
import { cn } from "@/lib/utils";

type Props = {
  input: CalculatorInput;
  hm: number;
  templateId: string | null;
  integrationLevel: IntegrationLevel;
  onIntegrationLevelChange: (level: IntegrationLevel) => void;
  capexManuallyEdited: boolean;
  omManuallyEdited: boolean;
  onApplyEstimate: (capex: number, om: number) => void;
  showBreakdown: boolean;
  onToggleBreakdown: () => void;
};

export function BudgetEstimatePanel({
  input,
  hm,
  templateId,
  integrationLevel,
  onIntegrationLevelChange,
  capexManuallyEdited,
  omManuallyEdited,
  onApplyEstimate,
  showBreakdown,
  onToggleBreakdown,
}: Props) {
  const t = useTranslations("calculator.budgetEstimate");
  const locale = useLocale();
  const { config } = useBudgetConfig();

  const estimate = estimateImplementationBudget(input, {
    hm,
    integrationLevel,
    templateId,
    config,
  });

  const needsReapply = capexManuallyEdited || omManuallyEdited;

  return (
    <div className="rounded-lg border border-accent bg-accent-bg p-4 space-y-3">
      <div>
        <p className="text-sm font-medium text-accent">{t("title")}</p>
        <p className="mt-0.5 text-xs text-accent/90">{t("subtitle")}</p>
      </div>

      <div className="flex flex-wrap gap-2">
        {(["simple", "standard", "complex"] as IntegrationLevel[]).map((level) => (
          <button
            key={level}
            type="button"
            onClick={() => onIntegrationLevelChange(level)}
            className={cn(
              "rounded-lg border px-3 py-1.5 text-xs",
              integrationLevel === level
                ? "border-accent bg-bg font-medium text-accent"
                : "border-accent/30 text-accent/80 hover:border-accent/60"
            )}
          >
            {t(`integration.${level}`)}
          </button>
        ))}
      </div>

      <div className="grid gap-2 sm:grid-cols-2 text-sm">
        <div className="rounded-lg bg-bg/80 px-3 py-2">
          <p className="text-[10px] uppercase tracking-wider text-text3">{t("recommendedCapex")}</p>
          <p className="mt-1 font-medium text-text">{formatEstimateRub(estimate.capex, locale)}</p>
        </div>
        <div className="rounded-lg bg-bg/80 px-3 py-2">
          <p className="text-[10px] uppercase tracking-wider text-text3">{t("recommendedOm")}</p>
          <p className="mt-1 font-medium text-text">{formatEstimateRub(estimate.om, locale)}</p>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <button
          type="button"
          onClick={() => onApplyEstimate(estimate.capex, estimate.om)}
          className="rounded-lg bg-text px-3 py-1.5 text-xs font-medium text-white hover:opacity-90"
        >
          {needsReapply ? t("reapply") : t("apply")}
        </button>
        <button
          type="button"
          onClick={onToggleBreakdown}
          className="text-xs text-accent underline hover:no-underline"
        >
          {showBreakdown ? t("hideBreakdown") : t("showBreakdown")}
        </button>
        {needsReapply && <span className="text-xs text-text3">{t("manualHint")}</span>}
      </div>

      {showBreakdown && (
        <ul className="space-y-1 border-t border-accent/20 pt-3 text-xs text-text2">
          {estimate.lines.map((row) => (
            <li key={row.key} className="flex justify-between gap-4">
              <span>{locale === "en" ? row.labelEn : row.labelRu}</span>
              <span className="shrink-0 font-medium text-text">
                {formatEstimateRub(Math.round(row.amount), locale)}
              </span>
            </li>
          ))}
          <li className="flex justify-between gap-4 border-t border-accent/20 pt-2 font-medium text-text">
            <span>{t("totalCapex")}</span>
            <span>{formatEstimateRub(estimate.capex, locale)}</span>
          </li>
        </ul>
      )}
    </div>
  );
}
