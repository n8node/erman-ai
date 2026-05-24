export type ProcessStep = {
  operation: string;
  executor?: string;
  minutes_per_unit: number;
  rate_rub_per_min: number;
};

export type CalculatorInput = {
  process_name: string;
  hm_mode: "process" | "direct";
  process_steps?: ProcessStep[];
  units_per_month: number;
  automation_pct: number;
  hours_saved_month: number;
  error_rate_before_pct: number;
  throughput_before_per_day: number;
  hourly_cost_loaded: number;
  utilization: number;
  other_benefit_monthly: number;
  monthly_solution_cost: number;
  capex: number;
  horizon_months: number;
  discount_rate_annual: number;
};

export type KPIRow = {
  key: string;
  label: string;
  before: string;
  after: string;
  change: string;
};

export type CalculatorOutput = {
  hours_saved_month: number;
  minutes_per_unit: number;
  net_benefit_monthly: number;
  fte: number;
  payback_months: number;
  roi_horizon_pct: number;
  tco_horizon: number;
  npv: number;
  payroll_savings: number;
  monthly_hours_before: number;
  monthly_hours_after: number;
  recommendation: "automate" | "consider" | "not_recommended";
  recommendation_text: string;
  kpi_rows: KPIRow[];
};
