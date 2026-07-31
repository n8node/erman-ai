import {
  ApiError,
  apiFetch,
  type LLMProviderStatus,
  type StrategyLLMTestConnectionResult,
} from "./api";

export const GEOLOGICAL_JOURNAL_SLUG = "geological-journal";
export const GEOLOGICAL_JOURNAL_MAX_FILE_SIZE = 10 * 1024 * 1024;
export const GEOLOGICAL_JOURNAL_IMAGE_TYPES = [
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/gif",
] as const;

export type GeologicalJournalRow = {
  date: string;
  drilling_diameter_mm: number | null;
  depth_from_m: number | null;
  depth_to_m: number | null;
  drilling_run_m: number | null;
  core_recovery_m: number | null;
  core_recovery_pct: number | null;
  rock_description: string;
  sampling_interval: string;
  sample_number: string;
  notes: string;
  uncertainties: string[];
};

export type GeologicalJournalOutput = {
  rows: GeologicalJournalRow[];
};

export type GeologicalJournalRun = {
  id: string;
  status: "pending" | "processing" | "done" | "error" | string;
  output?: GeologicalJournalOutput | null;
  error_msg?: string;
  model_used?: string;
  created_at: string;
  updated_at?: string;
  completed_at?: string;
};

export type GeologicalJournalResultVersion = {
  id: string;
  page_id: string;
  user_id?: string;
  result: GeologicalJournalOutput;
  created_at: string;
};

export type GeologicalJournalPage = {
  id: string;
  original_name: string;
  content_type: string;
  size_bytes: number;
  width: number;
  height: number;
  latest_result?: GeologicalJournalOutput | null;
  current_result?: GeologicalJournalOutput | null;
  created_at: string;
  updated_at: string;
};

export type GeologicalJournalPageDetail = GeologicalJournalPage & {
  runs: GeologicalJournalRun[];
  versions?: GeologicalJournalResultVersion[];
};

export type GeologicalJournalExample = {
  id: string;
  title: string;
  description: string;
  content_type: string;
  size_bytes: number;
  sort_order: number;
  is_published: boolean;
  created_at: string;
  updated_at: string;
};

export type GeologicalJournalSettings = {
  provider: "openrouter" | "deepseek" | "yandex";
  openrouter_model: string;
  deepseek_model: string;
  yandex_model: string;
  ocr_model: string;
  system_prompt: string;
  temperature: number;
  max_tokens: number;
};

export type GeologicalJournalSettingsRecord = {
  settings: GeologicalJournalSettings;
  providers: LLMProviderStatus[];
  updated_at?: string;
};

export type GeologicalJournalAccessUser = {
  id?: string;
  user_id?: string;
  email: string;
  role: string;
  has_access: boolean;
};

export type GeologicalJournalExampleMetadata = Pick<
  GeologicalJournalExample,
  "title" | "description" | "sort_order" | "is_published"
>;

const apiBase = () =>
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "/api/v1";

function parseJson<T>(value: unknown): T | null {
  if (!value) return null;
  if (typeof value === "string") {
    try {
      return JSON.parse(value) as T;
    } catch {
      return null;
    }
  }
  return value as T;
}

export function runOutputRows(
  run: GeologicalJournalRun | null | undefined
): GeologicalJournalRow[] {
  const output = parseJson<GeologicalJournalOutput>(run?.output);
  return output?.rows ?? [];
}

export function currentPageResult(
  page: GeologicalJournalPage | GeologicalJournalPageDetail
): GeologicalJournalOutput | null {
  return (
    parseJson<GeologicalJournalOutput>(page.current_result) ??
    parseJson<GeologicalJournalOutput>(page.latest_result)
  );
}

export function validateGeologicalJournalImage(file: File): "type" | "size" | null {
  if (!(GEOLOGICAL_JOURNAL_IMAGE_TYPES as readonly string[]).includes(file.type)) {
    return "type";
  }
  if (file.size > GEOLOGICAL_JOURNAL_MAX_FILE_SIZE) return "size";
  return null;
}

export function geologicalJournalPageImageUrl(pageId: string) {
  return `${apiBase()}/tools/geological-journal/pages/${pageId}/image`;
}

export function geologicalJournalExampleImageUrl(exampleId: string) {
  return `${apiBase()}/tools/geological-journal/examples/${exampleId}/image`;
}

export function uploadGeologicalJournalPage(
  file: File,
  onProgress: (percent: number) => void
): { promise: Promise<{ page: GeologicalJournalPage; run_id: string }>; abort: () => void } {
  const xhr = new XMLHttpRequest();
  const promise = new Promise<{ page: GeologicalJournalPage; run_id: string }>(
    (resolve, reject) => {
      xhr.open("POST", `${apiBase()}/tools/geological-journal/pages`);
      xhr.withCredentials = true;
      xhr.upload.onprogress = (event) => {
        if (event.lengthComputable) {
          onProgress(Math.round((event.loaded / event.total) * 100));
        }
      };
      xhr.onerror = () => reject(new Error("upload failed"));
      xhr.onabort = () => reject(new DOMException("Upload cancelled", "AbortError"));
      xhr.onload = () => {
        let data: Record<string, unknown> = {};
        try {
          data = JSON.parse(xhr.responseText) as Record<string, unknown>;
        } catch {
          // Error handling below provides the fallback message.
        }
        if (xhr.status < 200 || xhr.status >= 300) {
          reject(
            new ApiError(
              typeof data.error === "string" ? data.error : "upload failed",
              xhr.status,
              typeof data.code === "string" ? data.code : undefined
            )
          );
          return;
        }
        resolve(data as { page: GeologicalJournalPage; run_id: string });
      };
      const form = new FormData();
      form.append("image", file);
      xhr.send(form);
    }
  );
  return { promise, abort: () => xhr.abort() };
}

export function listGeologicalJournalPages() {
  return apiFetch<{ items: GeologicalJournalPage[] }>(
    "/tools/geological-journal/pages"
  );
}

export function getGeologicalJournalPage(id: string) {
  return apiFetch<GeologicalJournalPageDetail>(
    `/tools/geological-journal/pages/${id}`
  );
}

export function analyzeGeologicalJournalPage(id: string) {
  return apiFetch<{ run_id: string; status?: string }>(
    `/tools/geological-journal/pages/${id}/analyze`,
    { method: "POST", body: JSON.stringify({}) }
  );
}

export function saveGeologicalJournalResult(
  id: string,
  rows: GeologicalJournalRow[]
) {
  return apiFetch<GeologicalJournalPageDetail | GeologicalJournalOutput>(
    `/tools/geological-journal/pages/${id}/result`,
    { method: "PUT", body: JSON.stringify({ rows }) }
  );
}

export function deleteGeologicalJournalPage(id: string) {
  return apiFetch<{ status: string }>(
    `/tools/geological-journal/pages/${id}`,
    { method: "DELETE" }
  );
}

export function listGeologicalJournalExamples() {
  return apiFetch<{ items: GeologicalJournalExample[] }>(
    "/tools/geological-journal/examples"
  );
}

export function getGeologicalJournalRun(runId: string) {
  return apiFetch<GeologicalJournalRun>(`/runs/${runId}`);
}

export function fetchAdminGeologicalJournalSettings() {
  return apiFetch<GeologicalJournalSettingsRecord>(
    "/admin/geological-journal/settings"
  );
}

export function updateAdminGeologicalJournalSettings(
  settings: GeologicalJournalSettings
) {
  return apiFetch<GeologicalJournalSettingsRecord>(
    "/admin/geological-journal/settings",
    { method: "PUT", body: JSON.stringify({ settings }) }
  );
}

export function refreshAdminGeologicalJournalModels(
  provider: GeologicalJournalSettings["provider"]
) {
  return apiFetch<StrategyLLMTestConnectionResult>(
    "/admin/geological-journal/models/refresh",
    { method: "POST", body: JSON.stringify({ provider }) }
  );
}

export function testAdminGeologicalJournalOCR() {
  return apiFetch<StrategyLLMTestConnectionResult>(
    "/admin/geological-journal/ocr/test",
    { method: "POST" }
  );
}

export function fetchAdminGeologicalJournalAccess() {
  return apiFetch<{ items: GeologicalJournalAccessUser[] }>(
    "/admin/geological-journal/access"
  );
}

export function updateAdminGeologicalJournalAccess(
  userId: string,
  enabled: boolean
) {
  return apiFetch<GeologicalJournalAccessUser>(
    `/admin/geological-journal/access/${userId}`,
    { method: "PUT", body: JSON.stringify({ enabled }) }
  );
}

export function fetchAdminGeologicalJournalExamples() {
  return apiFetch<{ items: GeologicalJournalExample[] }>(
    "/admin/geological-journal/examples"
  );
}

export async function createAdminGeologicalJournalExample(
  file: File,
  metadata: GeologicalJournalExampleMetadata
) {
  const form = new FormData();
  form.append("image", file);
  form.append("title", metadata.title);
  form.append("description", metadata.description);
  form.append("sort_order", String(metadata.sort_order));
  form.append("is_published", String(metadata.is_published));
  const response = await fetch(`${apiBase()}/admin/geological-journal/examples`, {
    method: "POST",
    credentials: "include",
    body: form,
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new ApiError(
      typeof data.error === "string" ? data.error : "upload failed",
      response.status,
      typeof data.code === "string" ? data.code : undefined
    );
  }
  return data as GeologicalJournalExample;
}

export function updateAdminGeologicalJournalExample(
  id: string,
  metadata: GeologicalJournalExampleMetadata
) {
  return apiFetch<GeologicalJournalExample>(
    `/admin/geological-journal/examples/${id}`,
    { method: "PUT", body: JSON.stringify(metadata) }
  );
}

export function deleteAdminGeologicalJournalExample(id: string) {
  return apiFetch<{ status: string }>(
    `/admin/geological-journal/examples/${id}`,
    { method: "DELETE" }
  );
}
