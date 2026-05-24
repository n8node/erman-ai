"use client";

import { useTranslations } from "next-intl";
import type { CalculatorInput } from "@/lib/api";
import { totalMinutesPerUnit } from "@/lib/calculator";
import { HelpTooltip, LabelWithHelp } from "@/components/ui/HelpTooltip";

type Props = {
  input: CalculatorInput;
};

function formatRub(n: number) {
  if (!isFinite(n)) return "—";
  return new Intl.NumberFormat("ru-RU").format(Math.round(n)) + " ₽";
}

function formatNum(n: number, decimals = 0) {
  if (!isFinite(n)) return "—";
  return new Intl.NumberFormat("ru-RU", {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(n);
}

function ReadOnlyField({
  label,
  tooltipKey,
  value,
  className,
}: {
  label: string;
  tooltipKey?: string;
  value: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={className}>
      <p className="mb-1.5 text-xs font-medium text-text2">
        {tooltipKey ? (
          <LabelWithHelp label={label} tooltipKey={tooltipKey} />
        ) : (
          label
        )}
      </p>
      <div className="rounded-lg border border-border bg-bg2 px-3 py-2 text-sm text-text">
        {value}
      </div>
    </div>
  );
}

function TableHeader({
  label,
  tooltipKey,
}: {
  label: string;
  tooltipKey: string;
}) {
  return (
    <th className="px-3 py-2 normal-case">
      <span className="inline-flex items-center gap-1 whitespace-nowrap text-[10px] font-medium uppercase tracking-wider text-text3">
        {label}
        <HelpTooltip tooltipKey={tooltipKey} />
      </span>
    </th>
  );
}

export function CalculatorInputSummary({ input }: Props) {
  const t = useTranslations("calculator");
  const ts = useTranslations("calculator.share");

  const steps = input.process_steps || [];
  const totalMin = totalMinutesPerUnit(steps);
  const costPerUnit = steps.reduce(
    (s, r) => s + r.minutes_per_unit * r.rate_rub_per_min,
    0
  );
  const hm = input.hours_saved_month;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {ts("inputsTitle")}
        </h2>
        <p className="mt-2 text-base font-medium">{input.process_name}</p>
      </div>

      <section className="space-y-4">
        <h3 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("wizard.step1")}
        </h3>

        <ReadOnlyField
          label={ts("calculationMode")}
          tooltipKey="calculator.wizard.hm_mode_direct"
          value={
            input.hm_mode === "process"
              ? t("wizard.modeProcess")
              : t("wizard.modeDirect")
          }
        />

        {input.hm_mode === "process" ? (
          <>
            {steps.length > 0 && (
              <div className="overflow-x-auto rounded-lg border border-border">
                <table className="w-full text-left text-xs">
                  <thead>
                    <tr className="border-b border-border bg-bg2">
                      <TableHeader
                        label={t("wizard.colOperation")}
                        tooltipKey="calculator.wizard.col_operation"
                      />
                      <TableHeader
                        label={t("wizard.colExecutor")}
                        tooltipKey="calculator.wizard.col_executor"
                      />
                      <TableHeader
                        label={t("wizard.colMinutes")}
                        tooltipKey="calculator.wizard.col_minutes"
                      />
                      <TableHeader
                        label={t("wizard.colRate")}
                        tooltipKey="calculator.wizard.col_rate"
                      />
                      <TableHeader
                        label={t("wizard.colCost")}
                        tooltipKey="calculator.wizard.col_cost"
                      />
                    </tr>
                  </thead>
                  <tbody>
                    {steps.map((row, idx) => (
                      <tr key={idx} className="border-b border-border last:border-0">
                        <td className="px-3 py-2.5 font-medium">{row.operation || "—"}</td>
                        <td className="px-3 py-2.5 text-text2">{row.executor || "—"}</td>
                        <td className="px-3 py-2.5">{formatNum(row.minutes_per_unit, 1)}</td>
                        <td className="px-3 py-2.5">{formatRub(row.rate_rub_per_min)}</td>
                        <td className="px-3 py-2.5">
                          {formatRub(row.minutes_per_unit * row.rate_rub_per_min)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            <div className="rounded-lg bg-bg2 px-4 py-3 text-sm text-text2">
              {t("wizard.totalPerUnit")}: {formatNum(totalMin, 1)} {t("wizard.minShort")} ·{" "}
              {formatRub(costPerUnit)}
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <ReadOnlyField
                label={t("wizard.unitsPerMonth")}
                tooltipKey="calculator.wizard.units_per_month"
                value={formatNum(input.units_per_month)}
              />
              <ReadOnlyField
                label={t("wizard.automationPct")}
                tooltipKey="calculator.wizard.automation_pct"
                value={`${formatNum(input.automation_pct)}%`}
              />
            </div>
          </>
        ) : (
          <ReadOnlyField
            label={t("wizard.hmDirect")}
            tooltipKey="calculator.wizard.hm_mode_direct"
            value={`${formatNum(hm, 1)} ${t("wizard.hoursMonth")}`}
          />
        )}

        <div className="grid gap-4 sm:grid-cols-2 border-t border-border pt-4">
          <ReadOnlyField
            label={t("wizard.errorRateBefore")}
            tooltipKey="calculator.wizard.error_rate_before"
            value={`${formatNum(input.error_rate_before_pct)}%`}
          />
          <ReadOnlyField
            label={t("wizard.throughputBefore")}
            tooltipKey="calculator.wizard.throughput_before"
            value={formatNum(input.throughput_before_per_day)}
          />
        </div>

        <div className="rounded-lg border border-accent bg-accent-bg px-4 py-3 text-sm text-accent">
          <span className="inline-flex items-center gap-1.5 font-medium">
            Hm:
            <HelpTooltip tooltipKey="calculator.wizard.hm_preview" />
          </span>{" "}
          {formatNum(hm, 1)} {t("wizard.hoursMonth")}
        </div>
      </section>

      <section className="space-y-4 border-t border-border pt-6">
        <h3 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("wizard.step2")}
        </h3>

        <div className="grid gap-4 sm:grid-cols-2">
          <ReadOnlyField
            label={t("wizard.hmLabel")}
            tooltipKey="calculator.wizard.hm_label"
            value={`${formatNum(hm, 1)} ${t("wizard.hoursMonth")}`}
          />
          <ReadOnlyField
            label={t("fields.horizon")}
            tooltipKey="calculator.wizard.horizon"
            value={`${input.horizon_months} ${t("metrics.months")}`}
          />
          <ReadOnlyField
            label={t("fields.ch")}
            tooltipKey="calculator.wizard.ch"
            value={formatRub(input.hourly_cost_loaded)}
          />
          <ReadOnlyField
            label={t("fields.om")}
            tooltipKey="calculator.wizard.om"
            value={formatRub(input.monthly_solution_cost)}
          />
          <ReadOnlyField
            label={t("fields.capex")}
            tooltipKey="calculator.wizard.capex"
            value={formatRub(input.capex)}
          />
          <ReadOnlyField
            label={t("fields.sm")}
            tooltipKey="calculator.wizard.sm"
            value={formatRub(input.other_benefit_monthly)}
          />
          <ReadOnlyField
            label={t("fields.utilization")}
            tooltipKey="calculator.wizard.utilization"
            value={formatNum(input.utilization, 2)}
          />
          <ReadOnlyField
            label={t("fields.discount")}
            tooltipKey="calculator.wizard.discount"
            value={`${formatNum(input.discount_rate_annual * 100, 1)}%`}
          />
        </div>
      </section>
    </div>
  );
}
