import type { CalculatorRunContext } from "./api-strategy";

export type ProposalInput = {
  client_company: string;
  client_contact: string;
  client_industry: string;
  client_problem: string;
  solution_name: string;
  solution_description: string;
  deliverables: string[];
  project_cost_rub: number;
  timeline_weeks: number;
  payment_schedule: string;
  sender_company: string;
  sender_contact: string;
  sender_phone: string;
  sender_email: string;
  calculator_run_id?: string;
  calculator_context?: CalculatorRunContext;
};

export type ProposalTimelinePhase = {
  title: string;
  duration_weeks: number;
  description: string;
};

export type ProposalOutput = {
  greeting: string;
  task_understanding: string;
  proposed_solution: string;
  scope_included: string[];
  scope_excluded: string[];
  timeline: ProposalTimelinePhase[];
  cost_summary: string;
  payment_terms: string;
  why_us: string;
  next_step: string;
};

export type ProposalRunStart = {
  run_id: string;
  status: string;
};

export const DEFAULT_PROPOSAL_INPUT: ProposalInput = {
  client_company: "",
  client_contact: "",
  client_industry: "",
  client_problem: "",
  solution_name: "",
  solution_description: "",
  deliverables: [""],
  project_cost_rub: 0,
  timeline_weeks: 8,
  payment_schedule: "50% предоплата, 50% по завершению",
  sender_company: "",
  sender_contact: "",
  sender_phone: "",
  sender_email: "",
};

export const PAYMENT_SCHEDULE_OPTIONS = [
  "50% предоплата, 50% по завершению",
  "100% предоплата",
  "30% предоплата, 70% по завершению",
  "Поэтапная оплата по milestone",
] as const;

export const PROPOSAL_INDUSTRY_OPTIONS = [
  "retail",
  "finance",
  "manufacturing",
  "services",
  "it",
  "logistics",
  "healthcare",
  "other",
] as const;
