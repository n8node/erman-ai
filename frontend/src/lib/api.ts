export type User = {
  id: string;
  email: string;
  role: string;
  plan_id: string | null;
  locale: string;
  account_segment: "partner" | "direct_lead";
  onboarding_completed: boolean;
  is_blocked: boolean;
  created_at: string;
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
  input?: CalculatorInput;
  output?: CalculatorOutput;
  created_at: string;
  artifact_url?: string | null;
};

export type BillingPlan = {
  plan_slug: string;
  plan_name: string;
  features: Record<string, unknown>;
  tool_limits: Record<string, number>;
  usage: {
    share_report_used: number;
    share_report_limit: number;
  };
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
  /** false when the sharing partner has white_label — hides Erman AI registration promo */
  show_platform_cta: boolean;
};

const clientBase = () =>
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "/api/v1";

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
      throw new Error(
        typeof data.error === "string" ? data.error : "request failed"
      );
    }
    return res.blob() as Promise<T>;
  }

  const data = await res.json().catch(() => ({}));

  if (!res.ok) {
    throw new Error(
      typeof data.error === "string" ? data.error : "request failed"
    );
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
  return apiFetch<User>("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password, referral }),
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

export async function shareRun(runId: string) {
  return apiFetch<ShareResult>(`/runs/${runId}/share`, { method: "POST" });
}

export async function getBillingPlan() {
  return apiFetch<BillingPlan>("/billing/plan");
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

export async function fetchPublicReport(token: string) {
  return apiFetch<PublicReport>(`/shared/${token}`);
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
