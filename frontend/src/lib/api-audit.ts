export type AuditProcessInput = {
  name: string;
  department: string;
  frequency_range: string;
  minutes_per_cycle_range: string;
  fte_range: string;
  hourly_rate_range: string;
  error_rate_range: string;
  error_cost_range: string;
  systems: string[];
  systems_other?: string;
  bottleneck: string;
  standardization: "high" | "medium" | "low";
  integration_complexity: 1 | 2 | 3 | 4 | 5;
  automation_readiness: "high" | "medium" | "low";
  expected_impact: "high" | "medium" | "low";
};

export type AuditInput = {
  company_name: string;
  primary_goal: string;
  priority_criteria: string[];
  company_size: string;
  operational_headcount: string;
  it_systems: string[];
  it_systems_other?: string;
  data_duplication: string;
  budget_range: string;
  decision_timeline: string;
  processes: AuditProcessInput[];
};

export type AuditProcessScore = {
  name: string;
  monthly_hours: number;
  monthly_labor_cost_rub: number;
  monthly_error_cost_rub: number;
  total_monthly_cost_rub: number;
  automation_score: number;
  priority_rank: number;
  quick_win: boolean;
  rationale: string;
};

export type AuditPriorityRow = {
  rank: number;
  process_name: string;
  automation_score: number;
  monthly_savings_est_rub: number;
  payback_months_est: number;
  quick_win: boolean;
  rationale: string;
};

export type AuditRoadmapPhase = {
  phase: string;
  period: string;
  processes: string[];
  deliverables: string[];
};

export type AuditRisk = {
  risk: string;
  mitigation: string;
  severity: "low" | "medium" | "high";
};

export type AuditOutput = {
  executive_summary: string;
  company_context: string;
  process_scores: AuditProcessScore[];
  priority_ranking: AuditPriorityRow[];
  total_monthly_cost_rub: number;
  total_monthly_savings_est_rub: number;
  quick_wins: string[];
  roadmap: AuditRoadmapPhase[];
  risks: AuditRisk[];
  next_steps: string[];
  metrics_to_track: string[];
};

export type AuditRunStart = {
  run_id: string;
  status: string;
};

export const MIN_AUDIT_PROCESSES = 2;
export const MAX_AUDIT_PROCESSES = 7;

export const AUDIT_GOAL_OPTIONS = [
  "reduce_costs",
  "speed",
  "quality",
  "scale",
  "control",
  "other",
] as const;

export const AUDIT_PRIORITY_CRITERIA = [
  "cost_savings",
  "speed",
  "quality",
  "scalability",
  "control",
  "quick_wins",
] as const;

export const AUDIT_COMPANY_SIZE_OPTIONS = ["1-10", "11-50", "51-200", "201-1000", "1000+"] as const;

export const AUDIT_HEADCOUNT_OPTIONS = ["1-10", "11-30", "31-100", "101-500", "500+"] as const;

export const AUDIT_IT_SYSTEMS = [
  "crm",
  "erp",
  "1c",
  "spreadsheets",
  "email",
  "messengers",
  "website",
  "helpdesk",
  "warehouse",
] as const;

export const AUDIT_DATA_DUPLICATION = ["none", "some", "many", "critical"] as const;

export const AUDIT_BUDGET_OPTIONS = ["up_to_500k", "500k_2m", "2m_5m", "5m_plus"] as const;

export const AUDIT_TIMELINE_OPTIONS = ["3months", "6months", "1year", "2years"] as const;

export const AUDIT_DEPARTMENT_OPTIONS = [
  "sales",
  "operations",
  "finance",
  "hr",
  "marketing",
  "logistics",
  "support",
  "it",
  "other",
] as const;

export const AUDIT_PROCESS_TEMPLATES = [
  { key: "incoming_leads", department: "sales" },
  { key: "order_processing", department: "operations" },
  { key: "customer_support", department: "support" },
  { key: "invoicing", department: "finance" },
  { key: "hr_onboarding", department: "hr" },
  { key: "inventory", department: "logistics" },
  { key: "reporting", department: "finance" },
  { key: "custom", department: "" },
] as const;

export const AUDIT_BOTTLENECK_OPTIONS = [
  "manual_data_entry",
  "approvals",
  "integrations",
  "errors_rework",
  "waiting",
  "other",
] as const;

export const AUDIT_SYSTEMS = [
  "crm",
  "erp",
  "1c",
  "spreadsheets",
  "email",
  "messengers",
  "website",
  "helpdesk",
  "other",
] as const;

export const FREQUENCY_RANGES = ["1-10", "11-50", "51-200", "201-500", "500+"] as const;
export const MINUTES_RANGES = ["5-15", "16-30", "31-60", "61-120", "120+"] as const;
export const FTE_RANGES = ["1", "2-3", "4-10", "11+"] as const;
export const HOURLY_RANGES = ["300-500", "501-800", "801-1200", "1201-2000", "2000+"] as const;
export const ERROR_RATE_RANGES = ["0-2", "3-5", "6-10", "11-20", "20+"] as const;
export const ERROR_COST_RANGES = ["500-2000", "2001-5000", "5001-15000", "15001-50000", "50000+"] as const;

export const DEFAULT_AUDIT_PROCESS: AuditProcessInput = {
  name: "",
  department: "",
  frequency_range: "11-50",
  minutes_per_cycle_range: "16-30",
  fte_range: "2-3",
  hourly_rate_range: "501-800",
  error_rate_range: "3-5",
  error_cost_range: "2001-5000",
  systems: [],
  bottleneck: "",
  standardization: "medium",
  integration_complexity: 3,
  automation_readiness: "medium",
  expected_impact: "medium",
};

export const DEFAULT_AUDIT_INPUT: AuditInput = {
  company_name: "",
  primary_goal: "",
  priority_criteria: [],
  company_size: "11-50",
  operational_headcount: "11-30",
  it_systems: [],
  data_duplication: "some",
  budget_range: "500k_2m",
  decision_timeline: "6months",
  processes: [{ ...DEFAULT_AUDIT_PROCESS }, { ...DEFAULT_AUDIT_PROCESS }],
};

export function createProcessFromTemplate(templateKey: string, nameLabel: string, departmentLabel: string): AuditProcessInput {
  const tpl = AUDIT_PROCESS_TEMPLATES.find((t) => t.key === templateKey);
  if (!tpl || templateKey === "custom") {
    return { ...DEFAULT_AUDIT_PROCESS };
  }
  return {
    ...DEFAULT_AUDIT_PROCESS,
    name: nameLabel,
    department: departmentLabel || tpl.department,
  };
}

export function isAuditProcessValid(p: AuditProcessInput): boolean {
  return Boolean(
    p.name.trim() &&
      p.department.trim() &&
      p.frequency_range &&
      p.minutes_per_cycle_range &&
      p.fte_range &&
      p.hourly_rate_range &&
      p.error_rate_range &&
      p.error_cost_range &&
      (p.systems.length > 0 || (p.systems_other ?? "").trim()) &&
      p.bottleneck.trim() &&
      p.standardization &&
      p.automation_readiness &&
      p.expected_impact
  );
}

export function isAuditInputValid(input: AuditInput): boolean {
  if (!input.company_name.trim() || !input.primary_goal.trim()) return false;
  if (input.priority_criteria.length === 0) return false;
  if (input.it_systems.length === 0 && !(input.it_systems_other ?? "").trim()) return false;
  if (input.processes.length < MIN_AUDIT_PROCESSES || input.processes.length > MAX_AUDIT_PROCESSES) return false;
  return input.processes.every(isAuditProcessValid);
}
