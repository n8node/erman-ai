export type User = {
  id: string;
  email: string;
  role: string;
  plan_id: string | null;
  locale: string;
  account_segment: "partner" | "direct_lead";
  onboarding_completed: boolean;
  email_verified: boolean;
  is_blocked: boolean;
  created_at: string;
};

export type RegisterPendingResponse = {
  status: "verification_required";
  email: string;
  email_verified: boolean;
  onboarding_completed: boolean;
};

import type {
  CalculatorInput,
  CalculatorOutput,
  KPIRow,
  ProcessStep,
} from "./api-calculator";

export type { CalculatorInput, CalculatorOutput, KPIRow, ProcessStep };

export type CalculatorRunResult = {
  run_id: string;
  input: CalculatorInput;
  output: CalculatorOutput;
};

export type RunListItem = {
  id: string;
  tool_slug: string;
  process_name?: string;
  net_benefit_monthly?: number;
  payback_months?: number;
  roi_horizon_pct?: number;
  recommendation?: "automate" | "consider" | "not_recommended" | "";
  status: string;
  created_at: string;
};

export type RunListResponse = {
  items: RunListItem[];
  total: number;
};

export type RunDetail = {
  id: string;
  tool_slug: string;
  status: string;
  input?: CalculatorInput | import("./api-strategy").StrategyInput | import("./api-audit").AuditInput;
  output?: CalculatorOutput | import("./api-strategy").StrategyOutput | import("./api-audit").AuditOutput;
  created_at: string;
  updated_at?: string;
  completed_at?: string;
  error_msg?: string;
  model_used?: string;
  artifact_url?: string | null;
};

export type BillingPlan = {
  plan_id: string;
  plan_slug: string;
  plan_name: string;
  features: Record<string, unknown>;
  tool_limits: Record<string, number>;
  usage: {
    share_report_used: number;
    share_report_limit: number;
    tools?: Record<string, { used: number; limit: number }>;
  };
};

export type PublicPlan = {
  id: string;
  slug: string;
  name: string;
  price_monthly_rub: number;
  price_yearly_rub: number;
  tool_limits: Record<string, number>;
  features: Record<string, unknown>;
  support_level: string;
  is_public: boolean;
  is_archived: boolean;
};

export type ShareResult = {
  token: string;
  expires_at: string;
  url: string;
};

export type PublicReport = {
  process_name: string;
  input: CalculatorInput;
  output: CalculatorOutput;
  created_at: string;
  /** false when the sharing partner's plan has hide_share_promo */
  show_platform_cta: boolean;
};

const clientBase = () =>
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "/api/v1";

export class ApiError extends Error {
  status: number;
  code?: string;

  constructor(message: string, status: number, code?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

export function isToolLimitError(err: unknown): boolean {
  if (err instanceof ApiError) {
    return err.status === 402 || err.code === "tool_limit_exceeded";
  }
  if (err instanceof Error) {
    return err.message.includes("limit exceeded");
  }
  return false;
}

export async function apiFetch<T>(
  path: string,
  init?: RequestInit
): Promise<T> {
  const res = await fetch(`${clientBase()}${path}`, {
    ...init,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });

  if (init?.headers && (init.headers as Record<string, string>)["Accept"] === "application/pdf") {
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      const message = typeof data.error === "string" ? data.error : "request failed";
      const code = typeof data.code === "string" ? data.code : undefined;
      throw new ApiError(message, res.status, code);
    }
    return res.blob() as Promise<T>;
  }

  const data = await res.json().catch(() => ({}));

  if (!res.ok) {
    const message = typeof data.error === "string" ? data.error : "request failed";
    const code = typeof data.code === "string" ? data.code : undefined;
    throw new ApiError(message, res.status, code);
  }

  return data as T;
}

export async function login(email: string, password: string) {
  return apiFetch<User>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export async function register(
  email: string,
  password: string,
  referral?: string
) {
  return apiFetch<RegisterPendingResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password, referral }),
  });
}

export async function verifyEmail(token: string) {
  return apiFetch<User>("/auth/verify-email", {
    method: "POST",
    body: JSON.stringify({ token }),
  });
}

export async function resendVerification(email: string) {
  return apiFetch<{ status: string }>("/auth/resend-verification", {
    method: "POST",
    body: JSON.stringify({ email }),
  });
}

export async function completeOnboarding(segment: "partner" | "direct_lead") {
  return apiFetch<User>("/auth/onboarding", {
    method: "POST",
    body: JSON.stringify({ segment }),
  });
}

export async function logout() {
  return apiFetch<{ status: string }>("/auth/logout", { method: "POST" });
}

export async function fetchMe() {
  return apiFetch<User>("/auth/me");
}

export async function updateMe(email: string, locale: string) {
  return apiFetch<User>("/auth/me", {
    method: "PUT",
    body: JSON.stringify({ email, locale }),
  });
}

export async function changePassword(
  currentPassword: string,
  newPassword: string
) {
  return apiFetch<{ status: string }>("/auth/change-password", {
    method: "POST",
    body: JSON.stringify({
      current_password: currentPassword,
      new_password: newPassword,
    }),
  });
}

export async function runCalculator(input: CalculatorInput) {
  return apiFetch<CalculatorRunResult>("/tools/calculator/run", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function runStrategy(input: import("./api-strategy").StrategyInput) {
  return apiFetch<import("./api-strategy").StrategyRunStart>("/tools/strategy/run", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export type StrategyStreamHandlers = {
  onPhase: (data: import("./api-strategy").StrategyStreamPhase) => void;
  onChunk: (delta: string) => void;
  onDone: () => void;
  /** Terminal run failure from backend */
  onRunError: (message: string) => void;
  /** SSE disconnected — generation may still continue */
  onStreamLost?: () => void;
};

export function subscribeStrategyStream(runId: string, handlers: StrategyStreamHandlers): () => void {
  const url = `${clientBase()}/runs/${runId}/stream`;
  let closed = false;
  let finished = false;
  let es: EventSource | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  function cleanupES() {
    es?.close();
    es = null;
  }

  function connect() {
    if (closed || finished) return;
    cleanupES();
    es = new EventSource(url);

    es.addEventListener("phase", (ev) => {
      try {
        handlers.onPhase(JSON.parse((ev as MessageEvent).data));
      } catch {
        /* ignore */
      }
    });

    es.addEventListener("chunk", (ev) => {
      try {
        const data = JSON.parse((ev as MessageEvent).data) as { delta?: string };
        if (data.delta) handlers.onChunk(data.delta);
      } catch {
        /* ignore */
      }
    });

    es.addEventListener("done", () => {
      finished = true;
      cleanupES();
      handlers.onDone();
    });

    es.addEventListener("run_error", (ev) => {
      finished = true;
      cleanupES();
      try {
        const data = JSON.parse((ev as MessageEvent).data) as { message?: string };
        handlers.onRunError(data.message || "Generation failed");
      } catch {
        handlers.onRunError("Generation failed");
      }
    });

    es.onerror = () => {
      if (closed || finished) return;
      cleanupES();
      handlers.onStreamLost?.();
      reconnectTimer = setTimeout(connect, 3000);
    };
  }

  connect();

  return () => {
    closed = true;
    finished = true;
    if (reconnectTimer) clearTimeout(reconnectTimer);
    cleanupES();
  };
}

export type { StrategyInput, StrategyOutput } from "./api-strategy";
export type { AuditInput, AuditOutput } from "./api-audit";

export async function exportCalculatorPDF(runId: string) {
  const res = await fetch(`${clientBase()}/tools/calculator/export`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ run_id: runId }),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(
      typeof data.error === "string" ? data.error : "export failed"
    );
  }
  return res.blob();
}

export async function exportStrategyPDF(runId: string) {
  const res = await fetch(`${clientBase()}/tools/strategy/export`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ run_id: runId }),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(
      typeof data.error === "string" ? data.error : "export failed"
    );
  }
  return res.blob();
}

export async function runProposal(input: import("./api-proposal").ProposalInput) {
  return apiFetch<import("./api-proposal").ProposalRunStart>("/tools/proposal/run", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function runAudit(input: import("./api-audit").AuditInput) {
  return apiFetch<import("./api-audit").AuditRunStart>("/tools/audit/run", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function exportProposalPDF(runId: string) {
  const res = await fetch(`${clientBase()}/tools/proposal/export`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ run_id: runId }),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(
      typeof data.error === "string" ? data.error : "export failed"
    );
  }
  return res.blob();
}

export async function shareRun(runId: string) {
  return apiFetch<ShareResult>(`/runs/${runId}/share`, { method: "POST" });
}

export async function getBillingPlan() {
  return apiFetch<BillingPlan>("/billing/plan");
}

export async function fetchBillingPlans() {
  return apiFetch<{ items: PublicPlan[] }>("/billing/plans");
}

export async function switchBillingPlan(planId: string) {
  return apiFetch<{ plan_id: string; plan_slug: string; plan_name: string; status: string }>(
    "/billing/switch",
    {
      method: "POST",
      body: JSON.stringify({ plan_id: planId }),
    }
  );
}

export async function submitLead(payload: {
  run_id: string;
  name: string;
  email: string;
  company?: string;
  phone?: string;
  message?: string;
}) {
  return apiFetch<{ status: string; message: string }>("/leads", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function submitProposalRequest(payload: {
  run_id: string;
  requester_name: string;
  telegram: string;
  business_note: string;
}) {
  return apiFetch<{ id: string; status: string; created_at: string }>(
    "/tools/calculator/proposal-request",
    {
      method: "POST",
      body: JSON.stringify(payload),
    }
  );
}

export async function fetchPublicReport(token: string) {
  return apiFetch<PublicReport>(`/shared/${token}`);
}

export type ToolListItem = {
  slug: string;
  name: string;
  description: string;
  runs_used: number;
  runs_limit: number;
};

export async function fetchTools() {
  return apiFetch<{ tools: ToolListItem[] }>("/tools");
}

export async function listRuns(params?: {
  tool_slug?: string;
  limit?: number;
  offset?: number;
}) {
  const q = new URLSearchParams();
  if (params?.tool_slug) q.set("tool_slug", params.tool_slug);
  if (params?.limit) q.set("limit", String(params.limit));
  if (params?.offset) q.set("offset", String(params.offset));
  const qs = q.toString();
  return apiFetch<RunListResponse>(`/runs${qs ? `?${qs}` : ""}`);
}

export async function getRun(id: string) {
  return apiFetch<RunDetail>(`/runs/${id}`);
}

export async function deleteRun(id: string) {
  return apiFetch<{ status: string }>(`/runs/${id}`, { method: "DELETE" });
}

export type AdminTooltip = {
  key: string;
  label: string;
  text_ru: string;
  text_en: string;
  sort_order: number;
  updated_at: string;
};

export async function fetchTooltips(prefix = "calculator") {
  return apiFetch<{ tooltips: Record<string, string> }>(
    `/tooltips?prefix=${encodeURIComponent(prefix)}`
  );
}

export async function fetchPublicTooltips(prefix = "calculator", locale = "ru") {
  return apiFetch<{ tooltips: Record<string, string> }>(
    `/public/tooltips?prefix=${encodeURIComponent(prefix)}&locale=${encodeURIComponent(locale)}`
  );
}

export async function fetchAdminTooltips() {
  return apiFetch<{ items: AdminTooltip[] }>("/admin/tooltips");
}

export async function bulkUpdateAdminTooltips(
  items: Pick<AdminTooltip, "key" | "text_ru" | "text_en">[]
) {
  return apiFetch<{ items: AdminTooltip[] }>("/admin/tooltips", {
    method: "PUT",
    body: JSON.stringify({ items }),
  });
}

export type AdminTranslation = {
  key: string;
  locale: string;
  value: string;
  updated_at?: string;
};

export async function searchAdminTranslations(params?: {
  search?: string;
  locale?: string;
  limit?: number;
  offset?: number;
}) {
  const q = new URLSearchParams();
  if (params?.search) q.set("search", params.search);
  if (params?.locale) q.set("locale", params.locale);
  if (params?.limit != null) q.set("limit", String(params.limit));
  if (params?.offset != null) q.set("offset", String(params.offset));
  const suffix = q.toString() ? `?${q.toString()}` : "";
  return apiFetch<{ items: AdminTranslation[] }>(`/admin/translations${suffix}`);
}

export async function bulkUpdateAdminTranslations(
  items: Pick<AdminTranslation, "key" | "locale" | "value">[]
) {
  return apiFetch<{ items: AdminTranslation[] }>("/admin/translations", {
    method: "PUT",
    body: JSON.stringify({ items }),
  });
}

export type { CalculatorBudgetConfig, CalculatorBudgetConfigRecord } from "./calculator-budget-config";

export async function fetchCalculatorBudgetConfig() {
  return apiFetch<import("./calculator-budget-config").CalculatorBudgetConfigRecord>(
    "/tools/calculator/budget-config"
  );
}

export async function fetchAdminCalculatorBudgetConfig() {
  return apiFetch<import("./calculator-budget-config").CalculatorBudgetConfigRecord>(
    "/admin/calculator-budget"
  );
}

export async function updateAdminCalculatorBudgetConfig(
  config: import("./calculator-budget-config").CalculatorBudgetConfig
) {
  return apiFetch<import("./calculator-budget-config").CalculatorBudgetConfigRecord>(
    "/admin/calculator-budget",
    {
      method: "PUT",
      body: JSON.stringify(config),
    }
  );
}

export type LLMProvider = "openrouter" | "deepseek";

export type StrategyLLMSettings = {
  provider: LLMProvider;
  openrouter_model: string;
  deepseek_model: string;
  system_prompt: string;
  proposal_system_prompt: string;
  temperature: number;
  max_tokens: number;
};

export type LLMProviderStatus = {
  id: LLMProvider;
  configured: boolean;
  key_hint?: string;
  default_model: string;
  models: string[];
};

export type StrategyLLMAdminView = {
  settings: StrategyLLMSettings;
  providers: LLMProviderStatus[];
  default_system_prompt: string;
  default_proposal_system_prompt: string;
  updated_at?: string;
};

export type StrategyLLMAdminUpdateRequest = {
  settings: StrategyLLMSettings;
  openrouter_api_key?: string;
  deepseek_api_key?: string;
};

export type StrategyLLMTestConnectionResult = {
  provider: LLMProvider;
  ok: boolean;
  message: string;
  models: string[];
};

export async function fetchAdminStrategyLLMSettings() {
  return apiFetch<StrategyLLMAdminView>("/admin/strategy-llm");
}

export async function updateAdminStrategyLLMSettings(payload: StrategyLLMAdminUpdateRequest) {
  return apiFetch<StrategyLLMAdminView>("/admin/strategy-llm", {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export async function testStrategyLLMConnection(provider: LLMProvider) {
  return apiFetch<StrategyLLMTestConnectionResult>("/admin/strategy-llm/test-connection", {
    method: "POST",
    body: JSON.stringify({ provider }),
  });
}

export type SMTPEncryption = "none" | "ssl" | "tls";

export type SMTPSettings = {
  enabled: boolean;
  from_email: string;
  from_name: string;
  force_from_email: boolean;
  force_from_name: boolean;
  reply_to_from_email: boolean;
  host: string;
  port: number;
  encryption: SMTPEncryption;
  auto_tls: boolean;
  auth: boolean;
  username: string;
};

export type SMTPAdminView = {
  settings: SMTPSettings;
  password_set: boolean;
  password_hint?: string;
  updated_at?: string;
  yandex_preset_host: string;
  yandex_preset_port: number;
};

export type SMTPAdminUpdateRequest = {
  settings: SMTPSettings;
  password?: string;
};

export type SMTPTestEmailResult = {
  ok: boolean;
  message: string;
};

export async function fetchAdminSMTPSettings() {
  return apiFetch<SMTPAdminView>("/admin/email-smtp");
}

export async function updateAdminSMTPSettings(payload: SMTPAdminUpdateRequest) {
  return apiFetch<SMTPAdminView>("/admin/email-smtp", {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export async function sendAdminSMTPTest(to: string) {
  return apiFetch<SMTPTestEmailResult>("/admin/email-smtp/test", {
    method: "POST",
    body: JSON.stringify({ to }),
  });
}

export type AdminPlan = {
  id: string;
  slug: string;
  name: string;
  price_monthly_rub: number;
  price_yearly_rub: number;
  tool_limits: Record<string, number>;
  features: Record<string, unknown>;
  support_level: string;
  is_public: boolean;
  is_archived: boolean;
  user_count?: number;
};

export type AdminPlanMeta = {
  tools: { slug: string; name: string; description: string; enabled: boolean }[];
  default_features: Record<string, unknown>;
  support_levels: string[];
};

export type PlanUpsertInput = {
  slug?: string;
  name: string;
  price_monthly_rub: number;
  price_yearly_rub: number;
  tool_limits: Record<string, number>;
  features: Record<string, unknown>;
  support_level: string;
  is_public: boolean;
  is_archived: boolean;
};

export async function fetchAdminPlans() {
  return apiFetch<{ items: AdminPlan[] }>("/admin/plans");
}

export async function fetchAdminPlansMeta() {
  return apiFetch<AdminPlanMeta>("/admin/plans/meta");
}

export async function createAdminPlan(input: PlanUpsertInput & { slug: string }) {
  return apiFetch<AdminPlan>("/admin/plans", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateAdminPlan(id: string, input: PlanUpsertInput) {
  return apiFetch<AdminPlan>(`/admin/plans/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export type AdminUserRow = {
  id: string;
  email: string;
  role: string;
  plan_id: string | null;
  plan_slug?: string | null;
  plan_name?: string | null;
  locale: string;
  account_segment: string;
  onboarding_completed: boolean;
  email_verified_at?: string | null;
  is_blocked: boolean;
  created_at: string;
  last_active_at?: string | null;
};

export type AdminUserListResponse = {
  items: AdminUserRow[];
  total: number;
  limit: number;
  offset: number;
};

export async function fetchAdminUsers(params?: {
  q?: string;
  limit?: number;
  offset?: number;
}) {
  const q = new URLSearchParams();
  if (params?.q) q.set("q", params.q);
  if (params?.limit) q.set("limit", String(params.limit));
  if (params?.offset) q.set("offset", String(params.offset));
  const qs = q.toString();
  return apiFetch<AdminUserListResponse>(`/admin/users${qs ? `?${qs}` : ""}`);
}

export async function updateAdminUserPlan(userId: string, planId: string) {
  return apiFetch<AdminUserRow>(`/admin/users/${userId}`, {
    method: "PATCH",
    body: JSON.stringify({ plan_id: planId }),
  });
}

export async function deleteAdminUser(userId: string) {
  return apiFetch<{ status: string }>(`/admin/users/${userId}`, {
    method: "DELETE",
  });
}

export async function impersonateAdminUser(userId: string) {
  return apiFetch<User>(`/admin/users/${userId}/impersonate`, {
    method: "POST",
  });
}

export type AdminProposalRequestRow = {
  id: string;
  user_id: string;
  user_email: string;
  run_id: string;
  requester_name: string;
  telegram: string;
  business_note: string;
  process_name: string;
  net_benefit_monthly?: number | null;
  payback_months?: number | null;
  recommendation?: string | null;
  status: "new" | "in_progress" | "done";
  created_at: string;
  updated_at: string;
};

export type AdminProposalRequestDetail = AdminProposalRequestRow & {
  input: CalculatorInput;
  output: CalculatorOutput;
};

export async function fetchAdminProposalRequests(params?: {
  limit?: number;
  offset?: number;
}) {
  const q = new URLSearchParams();
  if (params?.limit) q.set("limit", String(params.limit));
  if (params?.offset) q.set("offset", String(params.offset));
  const qs = q.toString();
  return apiFetch<{
    items: AdminProposalRequestRow[];
    total: number;
    limit: number;
    offset: number;
  }>(`/admin/proposal-requests${qs ? `?${qs}` : ""}`);
}

export async function fetchAdminProposalRequest(id: string) {
  return apiFetch<AdminProposalRequestDetail>(`/admin/proposal-requests/${id}`);
}

export async function updateAdminProposalRequestStatus(
  id: string,
  status: AdminProposalRequestRow["status"]
) {
  return apiFetch<{ id: string; status: string }>(
    `/admin/proposal-requests/${id}`,
    {
      method: "PATCH",
      body: JSON.stringify({ status }),
    }
  );
}
