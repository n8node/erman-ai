import type { CalculatorInput } from "@/lib/api";
import { resolveHoursSaved } from "@/lib/calculator";
import { CALCULATOR_SECTORS, type CalculatorProcessTemplate } from "@/lib/calculator-templates";

export type IntegrationLevel = "simple" | "standard" | "complex";

export type BudgetEstimateLine = {
  key: string;
  labelRu: string;
  labelEn: string;
  amount: number;
};

export type BudgetEstimate = {
  capex: number;
  om: number;
  complexityScore: number;
  complexityTier: IntegrationLevel;
  lines: BudgetEstimateLine[];
  devHours: number;
};

const DEV_RATE_RUB = 4000;
const AUDIT_BASE = 100_000;
const TRAINING = 40_000;

export function findTemplateById(id: string): CalculatorProcessTemplate | undefined {
  for (const sector of CALCULATOR_SECTORS) {
    const found = sector.templates.find((t) => t.id === id);
    if (found) return found;
  }
  return undefined;
}

function roundCapex(value: number) {
  return Math.round(value / 50_000) * 50_000;
}

function roundOm(value: number) {
  return Math.round(value / 1_000) * 1_000;
}

function integrationMultiplier(level: IntegrationLevel) {
  if (level === "simple") return 0.85;
  if (level === "complex") return 1.35;
  return 1;
}

function countUniqueExecutors(input: CalculatorInput) {
  const roles = new Set(
    (input.process_steps || [])
      .map((s) => s.executor?.trim())
      .filter(Boolean)
  );
  return roles.size || 1;
}

function complexityScore(input: CalculatorInput, hm: number) {
  const steps = input.process_steps?.length ?? 0;
  const units = input.units_per_month;
  let score = 0;
  if (steps <= 3) score += 1;
  else if (steps <= 6) score += 2;
  else score += 3;
  if (units > 2000) score += 2;
  else if (units > 500) score += 1;
  if (hm > 200) score += 2;
  else if (hm > 50) score += 1;
  if (countUniqueExecutors(input) > 2) score += 1;
  return score;
}

function scoreToTier(score: number): IntegrationLevel {
  if (score <= 3) return "simple";
  if (score <= 6) return "standard";
  return "complex";
}

function integrationCost(executors: number) {
  if (executors > 3) return 150_000;
  if (executors > 2) return 80_000;
  if (executors > 1) return 40_000;
  return 0;
}

function devHoursForInput(input: CalculatorInput, hm: number) {
  const stepCount = input.process_steps?.length ?? 0;
  const units = input.units_per_month;
  const autoPct = input.automation_pct;

  let hours = 20 + stepCount * 6;
  if (units > 2000) hours += 40;
  else if (units > 1000) hours += 30;
  else if (units > 500) hours += 15;
  if (autoPct > 70) hours += 15;
  if (hm > 200) hours += 20;
  else if (hm > 80) hours += 10;
  return hours;
}

export function estimateImplementationBudget(
  input: CalculatorInput,
  options?: {
    hm?: number;
    integrationLevel?: IntegrationLevel;
    templateId?: string | null;
  }
): BudgetEstimate {
  const hm = options?.hm ?? resolveHoursSaved(input);
  const level = options?.integrationLevel ?? scoreToTier(complexityScore(input, hm));
  const mult = integrationMultiplier(level);

  const devHours = devHoursForInput(input, hm);
  const devCost = devHours * DEV_RATE_RUB;
  const integrations = integrationCost(countUniqueExecutors(input));
  const subtotal = AUDIT_BASE + devCost + integrations + TRAINING;
  const capex = roundCapex(Math.max(subtotal * mult, 150_000));

  let om = 8_000;
  const units = input.units_per_month;
  const stepCount = input.process_steps?.length ?? 0;
  if (units > 2000) om += 7_000;
  else if (units > 500) om += 4_000;
  if (stepCount > 5) om += 5_000;
  else if (stepCount > 3) om += 3_000;
  if (hm > 150) om += 3_000;
  om = roundOm(om);

  const score = complexityScore(input, hm);

  return {
    capex,
    om,
    complexityScore: score,
    complexityTier: level,
    devHours,
    lines: [
      line("audit", "Аудит и управление проектом", "Discovery and project management", AUDIT_BASE),
      line(
        "dev",
        `Разработка (~${devHours} ч × ${DEV_RATE_RUB.toLocaleString("ru-RU")} ₽)`,
        `Development (~${devHours} h × ${DEV_RATE_RUB.toLocaleString("en-US")} RUB)`,
        devCost
      ),
      ...(integrations > 0
        ? [line("integrations", "Интеграции и контуры", "Integrations and systems", integrations)]
        : []),
      line("training", "Обучение и запуск", "Training and go-live", TRAINING),
      ...(mult !== 1
        ? [
            line(
              "factor",
              level === "complex" ? "Сложные интеграции (+35%)" : "Простой контур (−15%)",
              level === "complex" ? "Complex integrations (+35%)" : "Simple scope (−15%)",
              subtotal * (mult - 1)
            ),
          ]
        : []),
    ],
  };
}

function line(key: string, labelRu: string, labelEn: string, amount: number): BudgetEstimateLine {
  return { key, labelRu, labelEn, amount };
}

export function formatEstimateRub(value: number, locale: string) {
  return new Intl.NumberFormat(locale === "en" ? "en-US" : "ru-RU").format(value) + " ₽";
}

export function defaultIntegrationLevel(input: CalculatorInput, hm: number): IntegrationLevel {
  return scoreToTier(complexityScore(input, hm));
}
