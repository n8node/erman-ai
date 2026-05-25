import type { CalculatorInput } from "@/lib/api";
import { resolveHoursSaved } from "@/lib/calculator";
import { CALCULATOR_SECTORS, type CalculatorProcessTemplate } from "@/lib/calculator-templates";
import {
  DEFAULT_CALCULATOR_BUDGET_CONFIG,
  type CalculatorBudgetConfig,
} from "@/lib/calculator-budget-config";
import { intlLocale } from "@/i18n/intl-locale";

export type IntegrationLevel = "simple" | "standard" | "complex";

export type BudgetEstimateLine = {
  key: string;
  amount: number;
  params?: Record<string, string | number>;
};

export type BudgetEstimate = {
  capex: number;
  om: number;
  complexityScore: number;
  complexityTier: IntegrationLevel;
  lines: BudgetEstimateLine[];
  devHours: number;
};

export function findTemplateById(id: string): CalculatorProcessTemplate | undefined {
  for (const sector of CALCULATOR_SECTORS) {
    const found = sector.templates.find((t) => t.id === id);
    if (found) return found;
  }
  return undefined;
}

function roundCapex(value: number, step: number) {
  return Math.round(value / step) * step;
}

function roundOm(value: number, step: number) {
  return Math.round(value / step) * step;
}

function integrationMultiplier(level: IntegrationLevel, cfg: CalculatorBudgetConfig) {
  if (level === "simple") return cfg.integration_mult_simple;
  if (level === "complex") return cfg.integration_mult_complex;
  return cfg.integration_mult_standard;
}

function countUniqueExecutors(input: CalculatorInput) {
  const roles = new Set(
    (input.process_steps || [])
      .map((s) => s.executor?.trim())
      .filter(Boolean)
  );
  return roles.size || 1;
}

function complexityScore(input: CalculatorInput, hm: number, cfg: CalculatorBudgetConfig) {
  const steps = input.process_steps?.length ?? 0;
  const units = input.units_per_month;
  let score = 0;
  if (steps <= 3) score += 1;
  else if (steps <= 6) score += 2;
  else score += 3;
  if (units > cfg.units_threshold_2000) score += 2;
  else if (units > cfg.units_threshold_500) score += 1;
  if (hm > cfg.hm_threshold_200) score += 2;
  else if (hm > cfg.hm_threshold_50) score += 1;
  if (countUniqueExecutors(input) > 2) score += 1;
  return score;
}

function scoreToTier(score: number, cfg: CalculatorBudgetConfig): IntegrationLevel {
  if (score <= cfg.tier_simple_max) return "simple";
  if (score <= cfg.tier_standard_max) return "standard";
  return "complex";
}

function integrationCost(executors: number, cfg: CalculatorBudgetConfig) {
  if (executors > 3) return cfg.integration_cost_4plus_roles;
  if (executors > 2) return cfg.integration_cost_3_roles;
  if (executors > 1) return cfg.integration_cost_2_roles;
  return 0;
}

function devHoursForInput(input: CalculatorInput, hm: number, cfg: CalculatorBudgetConfig) {
  const stepCount = input.process_steps?.length ?? 0;
  const units = input.units_per_month;
  const autoPct = input.automation_pct;

  let hours = cfg.dev_hours_base + stepCount * cfg.dev_hours_per_step;
  if (units > cfg.units_threshold_2000) hours += cfg.dev_hours_units_2000;
  else if (units > cfg.units_threshold_1000) hours += cfg.dev_hours_units_1000;
  else if (units > cfg.units_threshold_500) hours += cfg.dev_hours_units_500;
  if (autoPct > cfg.automation_pct_threshold) hours += cfg.dev_hours_automation_high;
  if (hm > cfg.hm_threshold_200) hours += cfg.dev_hours_hm_200;
  else if (hm > cfg.hm_threshold_80) hours += cfg.dev_hours_hm_80;
  return hours;
}

export function estimateImplementationBudget(
  input: CalculatorInput,
  options?: {
    hm?: number;
    integrationLevel?: IntegrationLevel;
    templateId?: string | null;
    config?: CalculatorBudgetConfig;
  }
): BudgetEstimate {
  const cfg = options?.config ?? DEFAULT_CALCULATOR_BUDGET_CONFIG;
  const hm = options?.hm ?? resolveHoursSaved(input);
  const level = options?.integrationLevel ?? scoreToTier(complexityScore(input, hm, cfg), cfg);
  const mult = integrationMultiplier(level, cfg);

  const devHours = devHoursForInput(input, hm, cfg);
  const devCost = devHours * cfg.dev_rate_rub;
  const integrations = integrationCost(countUniqueExecutors(input), cfg);
  const subtotal = cfg.audit_base + devCost + integrations + cfg.training;
  const capex = roundCapex(Math.max(subtotal * mult, cfg.capex_min), cfg.capex_round_step);

  let om = cfg.om_base;
  const units = input.units_per_month;
  const stepCount = input.process_steps?.length ?? 0;
  if (units > cfg.units_threshold_2000) om += cfg.om_units_2000;
  else if (units > cfg.units_threshold_500) om += cfg.om_units_500;
  if (stepCount > 5) om += cfg.om_steps_5;
  else if (stepCount > 3) om += cfg.om_steps_3;
  if (hm > cfg.hm_threshold_150) om += cfg.om_hm_150;
  om = roundOm(om, cfg.om_round_step);

  const score = complexityScore(input, hm, cfg);
  const pct = Math.round(Math.abs(mult - 1) * 100);

  return {
    capex,
    om,
    complexityScore: score,
    complexityTier: level,
    devHours,
    lines: [
      line("audit", cfg.audit_base),
      line("dev", devCost, {
        hours: devHours,
        rate: cfg.dev_rate_rub,
      }),
      ...(integrations > 0 ? [line("integrations", integrations)] : []),
      line("training", cfg.training),
      ...(mult !== 1
        ? [
            line(
              level === "complex" ? "factorComplex" : "factorSimple",
              subtotal * (mult - 1),
              { pct }
            ),
          ]
        : []),
    ],
  };
}

function line(
  key: string,
  amount: number,
  params?: Record<string, string | number>
): BudgetEstimateLine {
  return { key, amount, params };
}

export function formatEstimateRub(value: number, locale: string) {
  return new Intl.NumberFormat(intlLocale(locale)).format(value) + " ₽";
}

export function defaultIntegrationLevel(
  input: CalculatorInput,
  hm: number,
  config?: CalculatorBudgetConfig
): IntegrationLevel {
  const cfg = config ?? DEFAULT_CALCULATOR_BUDGET_CONFIG;
  return scoreToTier(complexityScore(input, hm, cfg), cfg);
}
