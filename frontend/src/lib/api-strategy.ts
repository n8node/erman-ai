export type StrategyInput = {
  company_name: string;
  industry: string;
  company_size: "1-10" | "11-50" | "51-200" | "201-1000" | "1000+";
  annual_revenue_range: string;
  current_ai_level: "none" | "exploring" | "piloting" | "scaling";
  main_goals: string[];
  pain_points: string;
  budget_range: string;
  timeline: "3months" | "6months" | "1year" | "2years";
  existing_tools: string;
  calculator_run_id?: string | null;
};

export type StrategySolution = {
  title: string;
  description: string;
  rationale: string;
  priority: number;
};

export type StrategyRoadmapPhase = {
  phase: string;
  quarter: string;
  initiatives: string[];
};

export type StrategyRisk = {
  risk: string;
  mitigation: string;
  severity: "low" | "medium" | "high";
};

export type StrategyMetric = {
  metric: string;
  target: string;
  timeframe: string;
};

export type StrategyOutput = {
  executive_summary: string;
  current_situation: string;
  recommended_solutions: StrategySolution[];
  roadmap: StrategyRoadmapPhase[];
  risks: StrategyRisk[];
  success_metrics: StrategyMetric[];
  next_30_days: string[];
};

export type StrategyRunStart = {
  run_id: string;
  status: string;
};

export const DEFAULT_STRATEGY_INPUT: StrategyInput = {
  company_name: "",
  industry: "",
  company_size: "11-50",
  annual_revenue_range: "до 10M",
  current_ai_level: "exploring",
  main_goals: [],
  pain_points: "",
  budget_range: "до 500K",
  timeline: "1year",
  existing_tools: "",
  calculator_run_id: null,
};

export const STRATEGY_GOAL_OPTIONS = [
  "reduce_costs",
  "speed_processes",
  "improve_quality",
  "grow_sales",
  "analytics",
  "customer_service",
] as const;

export const STRATEGY_INDUSTRY_OPTIONS = [
  "retail",
  "finance",
  "manufacturing",
  "services",
  "it",
  "logistics",
  "healthcare",
  "other",
] as const;
