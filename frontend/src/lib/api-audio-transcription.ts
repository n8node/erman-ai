import { ApiError, apiFetch } from "./api";

export const AUDIO_TRANSCRIPTION_SLUG = "audio-transcription";
export const AUDIO_TRANSCRIPTION_MAX_FILE_SIZE = 500 * 1024 * 1024;
export const AUDIO_TRANSCRIPTION_TYPES = [
  "audio/mpeg",
  "audio/mp3",
  "audio/wav",
  "audio/x-wav",
  "audio/ogg",
  "audio/opus",
  "audio/webm",
  "audio/mp4",
  "audio/x-m4a",
  "video/mp4",
  "video/webm",
] as const;

export type AudioTranscriptionFile = {
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

export type AudioTranscriptionOutput = {
  text: string;
  language: string;
  model: string;
  duration_sec: number;
  chunk_count: number;
  preview?: string;
};

export type AudioTranscriptionRun = {
  id: string;
  tool_slug: string;
  status: string;
  output?: AudioTranscriptionOutput;
  artifact_url?: string | null;
  error_msg?: string | null;
  created_at: string;
  completed_at?: string | null;
};

export type AudioTranscriptionFileDetail = {
  runs: AudioTranscriptionRun[];
} & AudioTranscriptionFile;

export type AudioTranscriptionSettings = {
  model: string;
  language_code: string;
  price_rub_per_minute: number;
};

export type AudioTranscriptionSettingsRecord = {
  settings: AudioTranscriptionSettings;
  updated_at: string;
};

export type AudioTranscriptionAccessUser = {
  id: string;
  email: string;
  role: string;
  has_access: boolean;
};

const apiBase = () =>
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "/api/v1";

export function audioTranscriptionDownloadUrl(fileId: string) {
  return `${apiBase()}/tools/audio-transcription/files/${fileId}/download`;
}

export function validateAudioTranscriptionFile(file: File): string | null {
  if (file.size > AUDIO_TRANSCRIPTION_MAX_FILE_SIZE) {
    return "file_too_large";
  }
  const type = file.type.toLowerCase();
  const ok = AUDIO_TRANSCRIPTION_TYPES.some(
    (allowed) => type === allowed || type.startsWith(`${allowed};`)
  );
  if (!ok) {
    const ext = file.name.split(".").pop()?.toLowerCase();
    const extOk = ["mp3", "wav", "ogg", "opus", "webm", "m4a", "mp4"].includes(ext || "");
    if (!extOk) return "invalid_type";
  }
  return null;
}

export function uploadAudioTranscriptionFile(
  file: File,
  onProgress: (percent: number) => void
): { promise: Promise<{ file: AudioTranscriptionFile; run_id: string }>; abort: () => void } {
  const xhr = new XMLHttpRequest();
  const promise = new Promise<{ file: AudioTranscriptionFile; run_id: string }>(
    (resolve, reject) => {
      xhr.open("POST", `${apiBase()}/tools/audio-transcription/files`);
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
          reject(
            new ApiError(
              typeof data.error === "string" ? data.error : "upload failed",
              xhr.status,
              typeof data.code === "string" ? data.code : undefined
            )
          );
          return;
        }
        resolve(data as { file: AudioTranscriptionFile; run_id: string });
      };
      const form = new FormData();
      form.append("file", file);
      xhr.send(form);
    }
  );
  return { promise, abort: () => xhr.abort() };
}

export function listAudioTranscriptionFiles() {
  return apiFetch<{ items: AudioTranscriptionFile[] }>(
    "/tools/audio-transcription/files",
    { cache: "no-store" }
  );
}

export function getAudioTranscriptionFile(id: string) {
  return apiFetch<AudioTranscriptionFileDetail>(
    `/tools/audio-transcription/files/${id}`,
    { cache: "no-store" }
  );
}

export function deleteAudioTranscriptionFile(id: string) {
  return apiFetch<void>(`/tools/audio-transcription/files/${id}`, {
    method: "DELETE",
  });
}

export function getAudioTranscriptionRun(runId: string) {
  return apiFetch<AudioTranscriptionRun>(`/runs/${runId}`, {
    cache: "no-store",
  });
}

export function fetchAdminAudioTranscriptionSettings() {
  return apiFetch<AudioTranscriptionSettingsRecord>(
    "/admin/audio-transcription/settings"
  );
}

export function updateAdminAudioTranscriptionSettings(
  settings: AudioTranscriptionSettings
) {
  return apiFetch<AudioTranscriptionSettingsRecord>(
    "/admin/audio-transcription/settings",
    { method: "PUT", body: JSON.stringify(settings) }
  );
}

export function fetchAdminAudioTranscriptionAccess() {
  return apiFetch<{ items: AudioTranscriptionAccessUser[] }>(
    "/admin/audio-transcription/access"
  );
}

export function updateAdminAudioTranscriptionAccess(
  userId: string,
  enabled: boolean
) {
  return apiFetch<{ user_id: string; enabled: boolean }>(
    `/admin/audio-transcription/access/${userId}`,
    { method: "PUT", body: JSON.stringify({ enabled }) }
  );
}
