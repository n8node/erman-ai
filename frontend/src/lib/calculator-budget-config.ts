export type CalculatorBudgetConfig = {
  dev_rate_rub: number;
  audit_base: number;
  training: number;
  om_base: number;
  capex_min: number;
  integration_mult_simple: number;
  integration_mult_standard: number;
  integration_mult_complex: number;
  integration_cost_2_roles: number;
  integration_cost_3_roles: number;
  integration_cost_4plus_roles: number;
  om_units_500: number;
  om_units_2000: number;
  om_steps_3: number;
  om_steps_5: number;
  om_hm_150: number;
  dev_hours_base: number;
  dev_hours_per_step: number;
  dev_hours_units_500: number;
  dev_hours_units_1000: number;
  dev_hours_units_2000: number;
  dev_hours_automation_high: number;
  dev_hours_hm_80: number;
  dev_hours_hm_200: number;
  automation_pct_threshold: number;
  capex_round_step: number;
  om_round_step: number;
  tier_simple_max: number;
  tier_standard_max: number;
  units_threshold_500: number;
  units_threshold_1000: number;
  units_threshold_2000: number;
  hm_threshold_50: number;
  hm_threshold_80: number;
  hm_threshold_150: number;
  hm_threshold_200: number;
};

export const DEFAULT_CALCULATOR_BUDGET_CONFIG: CalculatorBudgetConfig = {
  dev_rate_rub: 4000,
  audit_base: 100_000,
  training: 40_000,
  om_base: 8_000,
  capex_min: 150_000,
  integration_mult_simple: 0.85,
  integration_mult_standard: 1,
  integration_mult_complex: 1.35,
  integration_cost_2_roles: 40_000,
  integration_cost_3_roles: 80_000,
  integration_cost_4plus_roles: 150_000,
  om_units_500: 4_000,
  om_units_2000: 7_000,
  om_steps_3: 3_000,
  om_steps_5: 5_000,
  om_hm_150: 3_000,
  dev_hours_base: 20,
  dev_hours_per_step: 6,
  dev_hours_units_500: 15,
  dev_hours_units_1000: 30,
  dev_hours_units_2000: 40,
  dev_hours_automation_high: 15,
  dev_hours_hm_80: 10,
  dev_hours_hm_200: 20,
  automation_pct_threshold: 70,
  capex_round_step: 50_000,
  om_round_step: 1_000,
  tier_simple_max: 3,
  tier_standard_max: 6,
  units_threshold_500: 500,
  units_threshold_1000: 1000,
  units_threshold_2000: 2000,
  hm_threshold_50: 50,
  hm_threshold_80: 80,
  hm_threshold_150: 150,
  hm_threshold_200: 200,
};

export type CalculatorBudgetConfigRecord = {
  config: CalculatorBudgetConfig;
  updated_at?: string;
};

export type BudgetConfigFieldKey = keyof CalculatorBudgetConfig;

export const BUDGET_CONFIG_SECTIONS: {
  sectionKey: string;
  fields: BudgetConfigFieldKey[];
}[] = [
  {
    sectionKey: "globalRates",
    fields: ["dev_rate_rub", "audit_base", "training", "om_base", "capex_min"],
  },
  {
    sectionKey: "integration",
    fields: [
      "integration_mult_simple",
      "integration_mult_standard",
      "integration_mult_complex",
      "integration_cost_2_roles",
      "integration_cost_3_roles",
      "integration_cost_4plus_roles",
    ],
  },
  {
    sectionKey: "omAddons",
    fields: ["om_units_500", "om_units_2000", "om_steps_3", "om_steps_5", "om_hm_150"],
  },
  {
    sectionKey: "devHours",
    fields: [
      "dev_hours_base",
      "dev_hours_per_step",
      "dev_hours_units_500",
      "dev_hours_units_1000",
      "dev_hours_units_2000",
      "dev_hours_automation_high",
      "dev_hours_hm_80",
      "dev_hours_hm_200",
      "automation_pct_threshold",
    ],
  },
  {
    sectionKey: "thresholds",
    fields: [
      "tier_simple_max",
      "tier_standard_max",
      "units_threshold_500",
      "units_threshold_1000",
      "units_threshold_2000",
      "hm_threshold_50",
      "hm_threshold_80",
      "hm_threshold_150",
      "hm_threshold_200",
    ],
  },
  {
    sectionKey: "rounding",
    fields: ["capex_round_step", "om_round_step"],
  },
];

export function isIntegerBudgetField(key: BudgetConfigFieldKey): boolean {
  return key.startsWith("tier_");
}
