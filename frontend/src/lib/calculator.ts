import type { CalculatorInput, CalculatorOutput, ProcessStep } from "./api";

export function totalMinutesPerUnit(steps: ProcessStep[]): number {
  return steps.reduce((s, r) => s + (r.minutes_per_unit || 0), 0);
}

export function resolveHoursSaved(input: CalculatorInput): number {
  if (input.hm_mode === "direct") {
    return input.hours_saved_month;
  }
  const totalMin = totalMinutesPerUnit(input.process_steps || []);
  if (totalMin <= 0 || input.units_per_month <= 0 || input.automation_pct <= 0) {
    return input.hours_saved_month;
  }
  return (totalMin * input.units_per_month * input.automation_pct) / 100 / 60;
}

export function previewCalculator(
  input: CalculatorInput,
  locale = "ru"
): CalculatorOutput {
  const Hm = resolveHoursSaved(input);
  const L = input.utilization > 0 && input.utilization <= 1 ? input.utilization : 0.8;
  const Ch = input.hourly_cost_loaded;
  const Sm = input.other_benefit_monthly;
  const Om = input.monthly_solution_cost;
  const I0 = input.capex;
  const n = input.horizon_months > 0 ? input.horizon_months : 12;
  const r = input.discount_rate_annual > 0 ? input.discount_rate_annual : 0.2;

  const Bm = Hm * Ch + Sm - Om;
  const FTE = L > 0 ? Hm / (168 * L) : 0;
  const payback = Bm > 0 ? I0 / Bm : Infinity;
  const roiPct = I0 > 0 ? ((Bm * n - I0) / I0) * 100 : 0;
  const tco = I0 + Om * n;

  const rm = Math.pow(1 + r, 1 / 12) - 1;
  let npv = -I0;
  for (let t = 1; t <= n; t++) {
    npv += Bm / Math.pow(1 + rm, t);
  }

  let rec: CalculatorOutput["recommendation"] = "automate";
  let text =
    locale === "en"
      ? "Automation is recommended"
      : "Рекомендуется автоматизировать";
  if (Bm <= 0 || !isFinite(payback) || payback > 24) {
    rec = "not_recommended";
    text =
      locale === "en"
        ? "Automation is not recommended"
        : "Автоматизация не рекомендуется";
  } else if (payback >= 12 || roiPct <= 0) {
    rec = "consider";
    text =
      locale === "en" ? "Consider automation" : "Рассмотрите автоматизацию";
  }

  return {
    hours_saved_month: Hm,
    minutes_per_unit: totalMinutesPerUnit(input.process_steps || []),
    net_benefit_monthly: Bm,
    fte: FTE,
    payback_months: payback,
    roi_horizon_pct: roiPct,
    tco_horizon: tco,
    npv,
    payroll_savings: Hm * Ch,
    monthly_hours_before: 0,
    monthly_hours_after: 0,
    recommendation: rec,
    recommendation_text: text,
    kpi_rows: [],
  };
}

export const defaultCalculatorInput: CalculatorInput = {
  process_name: "",
  hm_mode: "process",
  process_steps: [
    {
      operation: "Ввод счета в систему",
      executor: "Оператор",
      minutes_per_unit: 8,
      rate_rub_per_min: 17,
    },
    {
      operation: "Проверка и подтверждение",
      executor: "Менеджер",
      minutes_per_unit: 2,
      rate_rub_per_min: 33,
    },
  ],
  units_per_month: 8000,
  automation_pct: 80,
  hours_saved_month: 800,
  error_rate_before_pct: 5,
  throughput_before_per_day: 20,
  hourly_cost_loaded: 2500,
  utilization: 0.8,
  other_benefit_monthly: 15000,
  monthly_solution_cost: 10000,
  capex: 300000,
  horizon_months: 12,
  discount_rate_annual: 0.2,
};
