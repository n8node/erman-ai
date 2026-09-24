import {
  ApiError,
  apiFetch,
  type LLMProviderStatus,
  type StrategyLLMTestConnectionResult,
} from "./api";
import type { GeologicalJournalIssueCode } from "./geological-journal-validate";

export type { GeologicalJournalIssueCode } from "./geological-journal-validate";

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
  field_issues?: Partial<Record<string, GeologicalJournalIssueCode[]>>;
};

export type GeologicalJournalOutput = {
  rows: GeologicalJournalRow[];
};

export type GeologicalJournalDocumentTableResult = {
  columns: string[];
  rows: Array<Record<string, unknown>>;
};

export type GeologicalJournalPreprocessing = {
  applied: boolean;
  used_for_ocr: boolean;
  has_preprocessed_image: boolean;
  perspective_corrected: boolean;
  deskew_angle: number;
  scale: number;
  width: number;
  height: number;
  fallback_reason?: string;
};

export type GeologicalJournalLayoutMode = "auto" | "spread" | "single";

export type GeologicalJournalOCRDiagnostics = {
  char_count: number;
  word_count: number;
  estimated_depth_rows: number;
  structured_row_bands: number;
  geometry_used: boolean;
  spread_split: boolean;
  layout_mode: GeologicalJournalLayoutMode | string;
  structuring_chunks: number;
  llm_prompt_tokens?: number;
  llm_completion_tokens?: number;
  ocr_preview?: string;
};

export type GeologicalJournalRunInput = {
  page_id: string;
  phase?: "preprocessing" | "ocr" | "structuring";
  layout_mode?: GeologicalJournalLayoutMode;
  preprocessing?: GeologicalJournalPreprocessing;
  ocr?: GeologicalJournalOCRDiagnostics;
};

export type GeologicalJournalRun = {
  id: string;
  status: "pending" | "processing" | "done" | "error" | string;
  input?: GeologicalJournalRunInput;
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
  has_preprocessed_image?: boolean;
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
  vision_model: string;
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

export type GeologicalJournalDocument = {
  id: string;
  original_name: string;
  content_type: string;
  size_bytes: number;
  page_count: number;
  status: string;
  is_shared: boolean;
  error_msg?: string;
  created_at: string;
  updated_at: string;
  analysis_started_at?: string;
  analysis_completed_at?: string;
};

export type GeologicalJournalDocumentPage = {
  id: string;
  document_id: string;
  page_number: number;
  status: string;
  ocr_text?: string;
  table_result?: GeologicalJournalDocumentTableResult | null;
  analysis?: Record<string, unknown>;
  content_type?: string;
  orientation_degrees: number;
  orientation_confidence: number;
  table_count: number;
  text_char_count: number;
  error_msg?: string;
};

export type GeologicalJournalDocumentLLMResult = {
  id: string;
  document_id: string;
  page_id: string;
  mode: string;
  result: Record<string, unknown> | string | null;
  model_used?: string;
  created_at: string;
};

export type GeologicalJournalDocumentDetail = {
  document: GeologicalJournalDocument;
  pages: GeologicalJournalDocumentPage[];
};

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

/** Pick the run that drives the Recognition tab UI.
 * Prefer in-flight runs; if the latest run failed but the page already has a
 * saved table (Library badge), prefer the last successful run so the tabs match.
 */
export function pickWorkspaceRun(
  runs: GeologicalJournalRun[] | null | undefined,
  pageResult: GeologicalJournalOutput | null
): GeologicalJournalRun | null {
  const list = runs ?? [];
  const inFlight = list.find((item) =>
    ["pending", "processing"].includes(item.status)
  );
  if (inFlight) return inFlight;

  const latest = list[0] ?? null;
  if (!latest) return null;
  if (latest.status === "done") return latest;

  if ((pageResult?.rows?.length ?? 0) > 0) {
    const lastGood =
      list.find(
        (item) => item.status === "done" && runOutputRows(item).length > 0
      ) ?? list.find((item) => item.status === "done");
    if (lastGood) return lastGood;
    // Page already has a saved table (Library), but no successful run to bind.
    // Present as done so Recognition matches Library instead of a stale error.
    return {
      id: latest.id,
      status: "done",
      created_at: latest.created_at,
      output: pageResult,
    };
  }

  return latest;
}

export function validateGeologicalJournalImage(file: File): "type" | "size" | null {
  if (!(GEOLOGICAL_JOURNAL_IMAGE_TYPES as readonly string[]).includes(file.type)) {
    return "type";
  }
  if (file.size > GEOLOGICAL_JOURNAL_MAX_FILE_SIZE) return "size";
  return null;
}

export function runJournalInput(
  run: GeologicalJournalRun | null | undefined
): GeologicalJournalRunInput | null {
  return parseJson<GeologicalJournalRunInput>(run?.input);
}

export function geologicalJournalPageImageUrl(pageId: string) {
  return `${apiBase()}/tools/geological-journal/pages/${pageId}/image`;
}

export function geologicalJournalPreprocessedImageUrl(
  pageId: string,
  cacheKey?: string
) {
  const url = `${apiBase()}/tools/geological-journal/pages/${pageId}/preprocessed-image`;
  return cacheKey ? `${url}?v=${encodeURIComponent(cacheKey)}` : url;
}

export function geologicalJournalExampleImageUrl(exampleId: string) {
  return `${apiBase()}/tools/geological-journal/examples/${exampleId}/image`;
}

export function uploadGeologicalJournalPage(
  file: File,
  onProgress: (percent: number) => void,
  layoutMode: GeologicalJournalLayoutMode = "auto"
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
      form.append("layout_mode", layoutMode);
      xhr.send(form);
    }
  );
  return { promise, abort: () => xhr.abort() };
}

export function listGeologicalJournalPages() {
  return apiFetch<{ items: GeologicalJournalPage[] }>(
    "/tools/geological-journal/pages",
    { cache: "no-store" }
  );
}

export function uploadGeologicalJournalDocument(
  file: File,
  onProgress: (percent: number) => void
): { promise: Promise<GeologicalJournalDocument>; abort: () => void } {
  const xhr = new XMLHttpRequest();
  const promise = new Promise<GeologicalJournalDocument>((resolve, reject) => {
    xhr.open("POST", `${apiBase()}/tools/geological-journal/documents`);
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
      try { data = JSON.parse(xhr.responseText) as Record<string, unknown>; } catch { /* fallback below */ }
      if (xhr.status < 200 || xhr.status >= 300) {
        reject(new ApiError(typeof data.error === "string" ? data.error : "upload failed", xhr.status));
        return;
      }
      resolve(data as unknown as GeologicalJournalDocument);
    };
    const form = new FormData();
    form.append("file", file);
    xhr.send(form);
  });
  return { promise, abort: () => xhr.abort() };
}

export function deleteGeologicalJournalDocument(id: string) {
  return apiFetch<void>(`/tools/geological-journal/documents/${id}`, { method: "DELETE" });
}

export function listGeologicalJournalDocuments() {
  return apiFetch<{ items: GeologicalJournalDocument[] }>("/tools/geological-journal/documents", { cache: "no-store" });
}

export function getGeologicalJournalDocument(id: string) {
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), 10000);
  return apiFetch<GeologicalJournalDocumentDetail>(`/tools/geological-journal/documents/${id}`, {
    cache: "no-store",
    signal: controller.signal,
  }).finally(() => window.clearTimeout(timeout)).then((data) => ({
    document: data?.document ?? ({} as GeologicalJournalDocument),
    pages: Array.isArray(data?.pages) ? data.pages : [],
  }));
}

export function startGeologicalJournalDocumentAnalysis(id: string) {
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), 15000);
  return apiFetch<{ status: string }>(`/tools/geological-journal/documents/${id}/analyze`, {
    method: "POST",
    signal: controller.signal,
  }).finally(() => window.clearTimeout(timeout));
}

export function setGeologicalJournalDocumentSharing(id: string, shared: boolean) {
  return apiFetch<{ shared: boolean }>(`/tools/geological-journal/documents/${id}/sharing`, {
    method: "PUT",
    body: JSON.stringify({ shared }),
  });
}

export function processGeologicalJournalDocumentLLM(id: string, pageNumbers: number[], mode = "summary", includeNeighbors = false) {
  return apiFetch<{ items: GeologicalJournalDocumentLLMResult[] }>(`/tools/geological-journal/documents/${id}/llm`, {
    method: "POST",
    body: JSON.stringify({ page_numbers: pageNumbers, mode, include_neighbors: includeNeighbors }),
  });
}

export function saveGeologicalJournalDocumentPageResult(
  documentId: string,
  pageId: string,
  rows: GeologicalJournalRow[],
  ocrText: string,
  columns: string[] = []
) {
  return apiFetch<GeologicalJournalOutput>(
    `/tools/geological-journal/documents/${documentId}/pages/${pageId}/result`,
    { method: "PUT", body: JSON.stringify({ rows, ocr_text: ocrText, columns }) }
  );
}

export function analyzeGeologicalJournalDocumentPageLocally(documentId: string, pageId: string) {
  return apiFetch<{ status: string }>(`/tools/geological-journal/documents/${documentId}/pages/${pageId}/local-analyze`, { method: "POST" });
}

export function chatGeologicalJournalDocument(id: string, pageNumbers: number[], message: string, sessionId?: string, includeNeighbors = false) {
  return apiFetch<{ content: string; sources: unknown; confidence?: string }>(`/tools/geological-journal/documents/${id}/chat`, {
    method: "POST",
    body: JSON.stringify({ page_numbers: pageNumbers, message, session_id: sessionId, include_neighbors: includeNeighbors }),
  });
}

export function geologicalJournalDocumentPageAssetUrl(documentId: string, pageId: string, kind: "original" | "oriented" | "preprocessed") {
  return `${apiBase()}/tools/geological-journal/documents/${documentId}/pages/${pageId}/assets/${kind}`;
}

export function getGeologicalJournalPage(id: string) {
  return apiFetch<GeologicalJournalPageDetail>(
    `/tools/geological-journal/pages/${id}`,
    { cache: "no-store" }
  );
}

export function analyzeGeologicalJournalPage(
  id: string,
  layoutMode: GeologicalJournalLayoutMode = "auto"
) {
  return apiFetch<{ run_id: string; status?: string }>(
    `/tools/geological-journal/pages/${id}/analyze`,
    { method: "POST", body: JSON.stringify({ layout_mode: layoutMode }) }
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
  return apiFetch<GeologicalJournalRun>(`/runs/${runId}`, {
    cache: "no-store",
  });
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
