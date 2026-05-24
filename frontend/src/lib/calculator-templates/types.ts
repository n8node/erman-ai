import type { CalculatorInput, ProcessStep } from "@/lib/api";

export type LocaleKey = "ru" | "en";

export type LocalizedText = Record<LocaleKey, string>;

export type TemplateStep = {
  operation: LocalizedText;
  executor: LocalizedText;
  minutes_per_unit: number;
  rate_rub_per_min: number;
};

export type CalculatorProcessTemplate = {
  id: string;
  name: LocalizedText;
  description: LocalizedText;
  process_name: LocalizedText;
  process_steps: TemplateStep[];
  units_per_month: number;
  automation_pct: number;
  error_rate_before_pct: number;
  throughput_before_per_day: number;
};

export type CalculatorSector = {
  id: string;
  name: LocalizedText;
  templates: CalculatorProcessTemplate[];
};

export function localize(text: LocalizedText, locale: string): string {
  return locale === "en" ? text.en : text.ru;
}

export function applyProcessTemplate(
  template: CalculatorProcessTemplate,
  locale: string
): Pick<
  CalculatorInput,
  | "process_name"
  | "hm_mode"
  | "process_steps"
  | "units_per_month"
  | "automation_pct"
  | "error_rate_before_pct"
  | "throughput_before_per_day"
> {
  const steps: ProcessStep[] = template.process_steps.map((s) => ({
    operation: localize(s.operation, locale),
    executor: localize(s.executor, locale),
    minutes_per_unit: s.minutes_per_unit,
    rate_rub_per_min: s.rate_rub_per_min,
  }));

  return {
    process_name: localize(template.process_name, locale),
    hm_mode: "process",
    process_steps: steps,
    units_per_month: template.units_per_month,
    automation_pct: template.automation_pct,
    error_rate_before_pct: template.error_rate_before_pct,
    throughput_before_per_day: template.throughput_before_per_day,
  };
}
