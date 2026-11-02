import type { RunDetail } from "./api";

export type AIAssessmentInput = {
  description: string;
  company_context?: string;
  industry?: string;
  monthly_volume?: number;
  minutes_per_unit?: number;
  hourly_cost?: number;
  automation_pct?: number;
  capex?: number;
  monthly_cost?: number;
  horizon_months?: number;
  locale?: string;
};

export type AssessmentStep = { name: string; executor?: string; minutes_per_unit: number; automation_pct: number };
export type AssessmentScenario = { name: string; automation_pct: number; net_benefit_monthly: number; payback_months: number; roi_pct: number };
export type AIAssessmentOutput = {
  process_name: string; summary: string; steps: AssessmentStep[]; unknowns: string[];
  criteria: { name: string; score: number; weight: number; rationale: string }[];
  readiness_score: number; autonomy: "autonomous" | "human_in_loop" | "assistant" | "not_ready";
  autonomy_text: string; risk_level: string; risks: string[];
  financial: { hours_saved_monthly: number; monthly_benefit: number; payback_months: number; roi_pct: number; tco: number; npv: number };
  scenarios: AssessmentScenario[]; pilot_plan: string[]; kpis: string[]; assumptions: string[];
};

export type AIAssessmentRunStart = { run_id: string; status: string };
export type AIAssessmentRun = RunDetail & { input?: AIAssessmentInput; output?: AIAssessmentOutput };

export async function runAIAssessment(input: AIAssessmentInput) {
  const res = await fetch("/api/v1/tools/ai-assessment/run", { method: "POST", credentials: "include", headers: { "Content-Type": "application/json" }, body: JSON.stringify(input) });
  if (!res.ok) throw new Error("Не удалось запустить анализ готовности к AI");
  return (await res.json()) as AIAssessmentRunStart;
}

export async function getAIAssessmentRun(id: string) {
  const res = await fetch(`/api/v1/runs/${id}`, { credentials: "include", cache: "no-store" });
  if (!res.ok) throw new Error("Анализ не найден");
  return (await res.json()) as AIAssessmentRun;
}