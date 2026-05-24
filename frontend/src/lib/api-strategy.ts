export type CalculatorRunContext = {
  run_id: string;
  process_name: string;
  net_benefit_monthly_rub: number;
  payback_months: number;
  roi_horizon_pct: number;
  npv_rub: number;
  capex_rub: number;
  monthly_support_rub: number;
  recommendation: string;
  recommendation_text: string;
};

export type StrategyInput = {
  company_name: string;
  business_description: string;
  industry: string;
  company_size: "1-10" | "11-50" | "51-200" | "201-1000" | "1000+";
  annual_revenue_range: string;
  market_position: string;
  current_ai_level: "none" | "exploring" | "piloting" | "scaling";
  main_goals: string[];
  key_processes: string[];
  pain_points: string;
  data_maturity: "none" | "basic" | "team";
  change_readiness: "low" | "medium" | "high";
  budget_range: string;
  timeline: "3months" | "6months" | "1year" | "2years";
  existing_tools: string;
  calculator_run_ids?: string[];
  calculator_contexts?: CalculatorRunContext[];
  report_mode?: "standard" | "consulting";
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

export type StrategyProcessAnalysis = {
  name: string;
  current_state: string;
  pain_points: string;
  ai_potential: string;
};

export type StrategyUseCase = {
  title: string;
  description: string;
  priority: number;
  impact: string;
};

export type StrategyImplementationPhase = {
  title: string;
  duration: string;
  deliverables: string[];
};

export type StrategyBudgetLine = {
  category: string;
  amount_range: string;
  notes: string;
};

export type StrategyROILine = {
  run_id: string;
  process_name: string;
  net_benefit_monthly_rub: number;
  payback_months: number;
  roi_horizon_pct: number;
  npv_rub: number;
  capex_rub: number;
  recommendation: string;
};

export type StrategyROISummary = {
  lines: StrategyROILine[];
  total_monthly_benefit_rub: number;
  total_npv_rub: number;
};

export type StrategyMaturityRow = {
  criterion: string;
  current_level: string;
  target_level: string;
  gap: string;
};

export type StrategyPriorityRow = {
  use_case: string;
  impact: number;
  effort: number;
  score: number;
  priority: number;
};

export type StrategyTechStackRow = {
  layer: string;
  tool: string;
  role: string;
  status: string;
};

export type StrategyStakeholderRow = {
  role: string;
  responsibility: string;
  involvement: string;
};

export type StrategyBudgetPhase = {
  phase: string;
  capex_rub: string;
  opex_monthly_rub: string;
  cumulative_rub: string;
};

export type StrategyDiagram = {
  type: string;
  title: string;
  mermaid: string;
};

export type StrategyOutput = {
  executive_summary: string;
  current_situation: string;
  goals_and_rationale: string;
  process_analysis: StrategyProcessAnalysis[];
  data_and_infrastructure: string;
  ai_use_cases: StrategyUseCase[];
  recommended_solutions: StrategySolution[];
  implementation_plan: StrategyImplementationPhase[];
  team_and_training: string;
  architecture_overview: string;
  data_governance: string;
  ethics_and_compliance: string;
  budget_overview: {
    summary: string;
    lines: StrategyBudgetLine[];
  };
  success_metrics: StrategyMetric[];
  strategy_adjustment_plan: string;
  risks: StrategyRisk[];
  roadmap: StrategyRoadmapPhase[];
  next_30_days: string[];
  roi_summary?: StrategyROISummary;
  maturity_matrix?: StrategyMaturityRow[];
  priority_matrix?: StrategyPriorityRow[];
  tech_stack?: StrategyTechStackRow[];
  stakeholder_plan?: StrategyStakeholderRow[];
  budget_phases?: StrategyBudgetPhase[];
  diagrams?: StrategyDiagram[];
};

export type StrategyRunStart = {
  run_id: string;
  status: string;
};

export type StrategyStreamPhase = {
  id: string;
  status: "active" | "done";
};

export const MAX_CALCULATOR_LINKS = 10;

export const DEFAULT_STRATEGY_INPUT: StrategyInput = {
  company_name: "",
  business_description: "",
  industry: "",
  company_size: "11-50",
  annual_revenue_range: "до 10M",
  market_position: "regional",
  current_ai_level: "exploring",
  main_goals: [],
  key_processes: [],
  pain_points: "",
  data_maturity: "basic",
  change_readiness: "medium",
  budget_range: "до 500K",
  timeline: "1year",
  existing_tools: "",
  calculator_run_ids: [],
  report_mode: "consulting",
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

export const STRATEGY_KEY_PROCESS_OPTIONS = [
  "sales",
  "customer_support",
  "operations",
  "procurement",
  "finance_accounting",
  "hr",
  "marketing",
  "logistics",
  "other",
] as const;

export const STRATEGY_MARKET_OPTIONS = [
  "local",
  "regional",
  "national",
  "international",
] as const;
