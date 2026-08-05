import { ApiError, apiFetch } from "./api";

export const VIDEO_TRANSCRIPTION_SLUG = "video-transcription";
export const VIDEO_TRANSCRIPTION_MAX_FILE_SIZE = 500 * 1024 * 1024;
export const VIDEO_TRANSCRIPTION_TYPES = [
  "video/mp4",
  "video/webm",
  "video/quicktime",
  "video/x-msvideo",
  "video/x-matroska",
  "video/mpeg",
  "video/ogg",
] as const;

export type VideoTranscriptionFile = {
  id: string;
  original_name: string;
  content_type: string;
  size_bytes: number;
  duration_sec?: number | null;
  has_transcript: boolean;
  created_at: string;
  updated_at: string;
  latest_run_id?: string | null;
  latest_run_status?: string | null;
};

export type VideoTranscriptionOutput = {
  text: string;
  language: string;
  model: string;
  duration_sec: number;
  chunk_count: number;
  preview?: string;
};

export type VideoTranscriptionRun = {
  id: string;
  tool_slug: string;
  status: string;
  output?: VideoTranscriptionOutput;
  artifact_url?: string | null;
  error_msg?: string | null;
  created_at: string;
  completed_at?: string | null;
};

export type VideoTranscriptionFileDetail = {
  runs: VideoTranscriptionRun[];
} & VideoTranscriptionFile;

export type VideoTranscriptionSettings = {
  model: string;
  language_code: string;
  price_rub_per_minute: number;
  text_normalization_enabled: boolean;
  literature_text: boolean;
  profanity_filter: boolean;
};

export type VideoTranscriptionSettingsRecord = {
  settings: VideoTranscriptionSettings;
  updated_at: string;
};

export type VideoTranscriptionAccessUser = {
  id: string;
  email: string;
  role: string;
  has_access: boolean;
};

const apiBase = () =>
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "/api/v1";

export function videoTranscriptionDownloadUrl(fileId: string) {
  return `${apiBase()}/tools/video-transcription/files/${fileId}/download`;
}

export function validateVideoTranscriptionFile(file: File): string | null {
  if (file.size > VIDEO_TRANSCRIPTION_MAX_FILE_SIZE) {
    return "file_too_large";
  }
  const type = file.type.toLowerCase();
  const ok = VIDEO_TRANSCRIPTION_TYPES.some(
    (allowed) => type === allowed || type.startsWith(`${allowed};`)
  );
  if (!ok) {
    const ext = file.name.split(".").pop()?.toLowerCase();
    const extOk = ["mp4", "webm", "mov", "avi", "mkv", "mpeg", "mpg", "ogv"].includes(ext || "");
    if (!extOk) return "invalid_type";
  }
  return null;
}

export function uploadVideoTranscriptionFile(
  file: File,
  onProgress: (percent: number) => void
): { promise: Promise<{ file: VideoTranscriptionFile; run_id: string }>; abort: () => void } {
  const xhr = new XMLHttpRequest();
  const promise = new Promise<{ file: VideoTranscriptionFile; run_id: string }>(
    (resolve, reject) => {
      xhr.open("POST", `${apiBase()}/tools/video-transcription/files`);
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
          // handled below
        }
        if (xhr.status < 200 || xhr.status >= 300) {
          let message =
            typeof data.error === "string" ? data.error : "upload failed";
          if (xhr.status === 502 || xhr.status === 504) {
            message = "server_timeout_or_unavailable";
          } else if (xhr.status === 413) {
            message = "file_too_large";
          }
          reject(
            new ApiError(
              message,
              xhr.status,
              typeof data.code === "string" ? data.code : undefined
            )
          );
          return;
        }
        resolve(data as { file: VideoTranscriptionFile; run_id: string });
      };
      const form = new FormData();
      form.append("file", file);
      xhr.send(form);
    }
  );
  return { promise, abort: () => xhr.abort() };
}

export function listVideoTranscriptionFiles() {
  return apiFetch<{ items: VideoTranscriptionFile[] }>(
    "/tools/video-transcription/files",
    { cache: "no-store" }
  );
}

export function getVideoTranscriptionFile(id: string) {
  return apiFetch<VideoTranscriptionFileDetail>(
    `/tools/video-transcription/files/${id}`,
    { cache: "no-store" }
  );
}

export function deleteVideoTranscriptionFile(id: string) {
  return apiFetch<void>(`/tools/video-transcription/files/${id}`, {
    method: "DELETE",
  });
}

export function transcribeVideoTranscriptionFile(id: string) {
  return apiFetch<{ run_id: string; status?: string }>(
    `/tools/video-transcription/files/${id}/transcribe`,
    { method: "POST" }
  );
}

export function getVideoTranscriptionRun(runId: string) {
  return apiFetch<VideoTranscriptionRun>(`/runs/${runId}`, {
    cache: "no-store",
  });
}

export function fetchAdminVideoTranscriptionSettings() {
  return apiFetch<VideoTranscriptionSettingsRecord>(
    "/admin/video-transcription/settings"
  );
}

export function updateAdminVideoTranscriptionSettings(
  settings: VideoTranscriptionSettings
) {
  return apiFetch<VideoTranscriptionSettingsRecord>(
    "/admin/video-transcription/settings",
    { method: "PUT", body: JSON.stringify(settings) }
  );
}

export function fetchAdminVideoTranscriptionAccess() {
  return apiFetch<{ items: VideoTranscriptionAccessUser[] }>(
    "/admin/video-transcription/access"
  );
}

export function updateAdminVideoTranscriptionAccess(
  userId: string,
  enabled: boolean
) {
  return apiFetch<{ user_id: string; enabled: boolean }>(
    `/admin/video-transcription/access/${userId}`,
    { method: "PUT", body: JSON.stringify({ enabled }) }
  );
}
