"use client";

import { useTranslations } from "next-intl";
import type { CalculatorInput, CalculatorOutput } from "@/lib/api";
import { CalculatorInputSummary } from "./CalculatorInputSummary";
import { CalculatorReportView } from "./CalculatorReportView";

type Props = {
  input: CalculatorInput;
  output: CalculatorOutput;
  processName: string;
};

export function CalculatorShareReport({ input, output, processName }: Props) {
  const t = useTranslations("calculator.share");

  return (
    <div className="space-y-8">
      <CalculatorInputSummary input={input} />
      <div className="border-t border-border pt-8">
        <h2 className="mb-6 text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("reportTitle")}
        </h2>
        <CalculatorReportView
          processName={processName}
          output={output}
          withTooltips
          showHeader={false}
        />
      </div>
    </div>
  );
}
