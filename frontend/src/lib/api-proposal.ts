import type { CalculatorRunContext } from "./api-strategy";

export type ProposalScenario = "after_contact" | "cold_outreach" | "proactive_offer";

export type ProposalInput = {
  proposal_scenario: ProposalScenario;
  prior_contact_summary?: string;
  problem_source?: string;
  include_pricing: boolean;
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

export const PROPOSAL_SCENARIOS: ProposalScenario[] = [
  "after_contact",
  "cold_outreach",
  "proactive_offer",
];

export type PaymentScheduleId = "50_50" | "100_prepaid" | "30_70" | "milestone";

export const PAYMENT_SCHEDULE_IDS: PaymentScheduleId[] = [
  "50_50",
  "100_prepaid",
  "30_70",
  "milestone",
];

/** Maps stored payment_schedule text (any locale) to a stable id. */
export const LEGACY_PAYMENT_SCHEDULE: Record<string, PaymentScheduleId> = {
  "50% предоплата, 50% по завершению": "50_50",
  "50% prepayment, 50% on completion": "50_50",
  "50 % Vorauszahlung, 50 % bei Abschluss": "50_50",
  "50 % prepago, 50 % al finalizar": "50_50",
  "50 % d'acompte, 50 % à la fin": "50_50",
  "预付50%，完成时付50%": "50_50",
  "100% предоплата": "100_prepaid",
  "100% prepayment": "100_prepaid",
  "100 % Vorauszahlung": "100_prepaid",
  "100 % prepago": "100_prepaid",
  "100 % d'acompte": "100_prepaid",
  "100% 预付": "100_prepaid",
  "30% предоплата, 70% по завершению": "30_70",
  "30% prepayment, 70% on completion": "30_70",
  "30 % Vorauszahlung, 70 % bei Abschluss": "30_70",
  "30 % prepago, 70 % al finalizar": "30_70",
  "30 % d'acompte, 70 % à la fin": "30_70",
  "预付30%，完成时付70%": "30_70",
  "Поэтапная оплата по milestone": "milestone",
  "Milestone-based payments": "milestone",
  "Zahlung nach Meilensteinen": "milestone",
  "Pago por hitos": "milestone",
  "Paiement par jalons": "milestone",
  "按里程碑付款": "milestone",
};

export function resolvePaymentScheduleId(value: string): PaymentScheduleId {
  if (PAYMENT_SCHEDULE_IDS.includes(value as PaymentScheduleId)) {
    return value as PaymentScheduleId;
  }
  return LEGACY_PAYMENT_SCHEDULE[value] ?? "50_50";
}

export const DEFAULT_PROPOSAL_INPUT: ProposalInput = {
  proposal_scenario: "after_contact",
  prior_contact_summary: "",
  problem_source: "",
  include_pricing: true,
  client_company: "",
  client_contact: "",
  client_industry: "",
  client_problem: "",
  solution_name: "",
  solution_description: "",
  deliverables: [""],
  project_cost_rub: 0,
  timeline_weeks: 8,
  payment_schedule: "50_50",
  sender_company: "",
  sender_contact: "",
  sender_phone: "",
  sender_email: "",
};

export function normalizeProposalScenario(s?: string): ProposalScenario {
  if (s === "cold_outreach" || s === "proactive_offer") return s;
  return "after_contact";
}

/** @deprecated Use PAYMENT_SCHEDULE_IDS with i18n labels */
export const PAYMENT_SCHEDULE_OPTIONS = PAYMENT_SCHEDULE_IDS;

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
