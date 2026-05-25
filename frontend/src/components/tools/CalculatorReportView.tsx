"use client";

import { useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import type { CalculatorInput, CalculatorOutput } from "@/lib/api";
import { HelpTooltip } from "@/components/ui/HelpTooltip";
import { intlLocale } from "@/i18n/intl-locale";
import { cn } from "@/lib/utils";

type Props = {
  processName: string;
  output: CalculatorOutput;
  withTooltips?: boolean;
  showHeader?: boolean;
};

function formatRub(n: number, locale: string) {
  if (!isFinite(n)) return "—";
  return new Intl.NumberFormat(intlLocale(locale)).format(Math.round(n)) + " ₽";
}

function formatPayback(n: number, t: ReturnType<typeof useTranslations>) {
  if (!isFinite(n) || n <= 0 || n > 1e6) return "—";
  return `${n.toFixed(1)} ${t("metrics.months")}`;
}

export function CalculatorReportView({
  processName,
  output,
  withTooltips = true,
  showHeader = true,
}: Props) {
  const t = useTranslations("calculator");
  const locale = useLocale();
  const [showDetails, setShowDetails] = useState(false);

  const recClass =
    output.recommendation === "automate"
      ? "bg-[#eaf3de] text-[#3b6d11] border-[#c5ddb0]"
      : output.recommendation === "consider"
        ? "bg-[#faeeda] text-[#633806] border-[#e8d5b0]"
        : "bg-[#fcebeb] text-[#a32d2d] border-[#e8b4b4]";

  return (
    <div className="space-y-6">
      {showHeader && (
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("resultTitle")} — {processName}
        </h2>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Metric
          label={t("metrics.netBenefit")}
          value={formatRub(output.net_benefit_monthly, locale)}
          large
          tooltipKey={withTooltips ? "calculator.result.net_benefit" : undefined}
        />
        <Metric
          label={t("metrics.payback")}
          value={formatPayback(output.payback_months, t)}
          large
          tooltipKey={withTooltips ? "calculator.result.payback" : undefined}
        />
        <Metric
          label={t("metrics.fte")}
          value={output.fte.toFixed(2)}
          sub={t("metrics.fteHint")}
          tooltipKey={withTooltips ? "calculator.wizard.fte" : undefined}
        />
        <Metric
          label={t("metrics.roi")}
          value={`${output.roi_horizon_pct.toFixed(1)}%`}
          tooltipKey={withTooltips ? "calculator.result.roi" : undefined}
        />
      </div>

      <div className={cn("rounded-lg border px-4 py-3 text-sm", recClass)}>
        {t(`recommendation.${output.recommendation}`, {
          defaultMessage: output.recommendation_text,
        })}
      </div>

      {output.kpi_rows.length > 0 && (
        <section>
          <h3 className="text-[10px] font-medium uppercase tracking-wider text-text3 mb-3">
            {t("kpi.title")}
          </h3>
          <div className="overflow-x-auto rounded-lg border border-border">
            <table className="w-full text-left text-xs">
              <thead>
                <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
                  <th className="px-3 py-2">KPI</th>
                  <th className="px-3 py-2">{t("kpi.before")}</th>
                  <th className="px-3 py-2">{t("kpi.after")}</th>
                  <th className="px-3 py-2">{t("kpi.change")}</th>
                </tr>
              </thead>
              <tbody>
                {output.kpi_rows.map((row) => (
                  <tr key={row.key} className="border-b border-border last:border-0">
                    <td className="px-3 py-2.5 font-medium">
                      {withTooltips ? (
                        <span className="inline-flex items-center gap-1">
                          {row.label}
                          <HelpTooltip tooltipKey={`calculator.kpi.${row.key}`} />
                        </span>
                      ) : (
                        row.label
                      )}
                    </td>
                    <td className="px-3 py-2.5 text-text2">{row.before}</td>
                    <td className="px-3 py-2.5">{row.after}</td>
                    <td className="px-3 py-2.5">
                      <span className="rounded-md bg-[#eaf3de] px-2 py-0.5 text-[#3b6d11]">
                        {row.change}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}

      <button
        type="button"
        onClick={() => setShowDetails(!showDetails)}
        className="text-xs text-accent hover:underline"
      >
        {showDetails ? t("wizard.hideExpert") : t("result.showDetails")}
      </button>

      {showDetails && (
        <div className="grid gap-2 sm:grid-cols-3 text-sm text-text2 border-t border-border pt-4">
          <div className="inline-flex items-center gap-1">
            TCO: {formatRub(output.tco_horizon, locale)}
            {withTooltips && <HelpTooltip tooltipKey="calculator.result.tco" />}
          </div>
          <div className="inline-flex items-center gap-1">
            NPV: {formatRub(output.npv, locale)}
            {withTooltips && <HelpTooltip tooltipKey="calculator.result.npv" />}
          </div>
          <div className="inline-flex items-center gap-1">
            Hm: {Math.round(output.hours_saved_month)} {t("wizard.hoursMonth")}
            {withTooltips && <HelpTooltip tooltipKey="calculator.wizard.hm_preview" />}
          </div>
        </div>
      )}
    </div>
  );
}

function Metric({
  label,
  value,
  sub,
  large,
  tooltipKey,
}: {
  label: string;
  value: string;
  sub?: string;
  large?: boolean;
  tooltipKey?: string;
}) {
  return (
    <div className="rounded-lg border border-border bg-bg2 p-4">
      <p className="text-[10px] font-medium uppercase tracking-wider text-text3">
        {tooltipKey ? (
          <span className="inline-flex items-center gap-1 normal-case">
            <span className="uppercase tracking-wider">{label}</span>
            <HelpTooltip tooltipKey={tooltipKey} />
          </span>
        ) : (
          label
        )}
      </p>
      <p className={cn("mt-2 font-medium", large ? "text-xl" : "text-lg")}>{value}</p>
      {sub && <p className="mt-1 text-[10px] text-text3">{sub}</p>}
    </div>
  );
}
