import type { LegalScanInput } from "./api-legal-scan";

const DRAFT_KEY = "erman:legal-scan-draft:v1";
const DRAFT_MAX_AGE_MS = 30 * 24 * 60 * 60 * 1000;

export type LegalScanDraft = {
  input: LegalScanInput;
  savedAt: string;
};

export function loadLegalScanDraft(): LegalScanDraft | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(DRAFT_KEY);
    if (!raw) return null;
    const draft = JSON.parse(raw) as LegalScanDraft;
    if (!draft.savedAt || Date.now() - new Date(draft.savedAt).getTime() > DRAFT_MAX_AGE_MS) {
      localStorage.removeItem(DRAFT_KEY);
      return null;
    }
    if (!draft.input?.url && !draft.input?.industry) return null;
    return draft;
  } catch {
    return null;
  }
}

export function saveLegalScanDraft(input: LegalScanInput) {
  if (typeof window === "undefined") return;
  const payload: LegalScanDraft = { input, savedAt: new Date().toISOString() };
  localStorage.setItem(DRAFT_KEY, JSON.stringify(payload));
}

export function clearLegalScanDraft() {
  if (typeof window === "undefined") return;
  localStorage.removeItem(DRAFT_KEY);
}
