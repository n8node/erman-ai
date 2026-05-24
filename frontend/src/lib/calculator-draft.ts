import type { CalculatorInput } from "./api-calculator";

const DRAFT_KEY = "erman:calculator-draft:v1";
const DRAFT_MAX_AGE_MS = 30 * 24 * 60 * 60 * 1000;

export type CalculatorDraft = {
  input: CalculatorInput;
  step: 1 | 2;
  showExpert: boolean;
  savedAt: string;
};

export function loadCalculatorDraft(): CalculatorDraft | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(DRAFT_KEY);
    if (!raw) return null;
    const draft = JSON.parse(raw) as CalculatorDraft;
    if (!draft.savedAt || Date.now() - new Date(draft.savedAt).getTime() > DRAFT_MAX_AGE_MS) {
      localStorage.removeItem(DRAFT_KEY);
      return null;
    }
    if (draft.step !== 1 && draft.step !== 2) return null;
    return draft;
  } catch {
    return null;
  }
}

export function saveCalculatorDraft(draft: Omit<CalculatorDraft, "savedAt">) {
  if (typeof window === "undefined") return;
  const payload: CalculatorDraft = { ...draft, savedAt: new Date().toISOString() };
  localStorage.setItem(DRAFT_KEY, JSON.stringify(payload));
}

export function clearCalculatorDraft() {
  if (typeof window === "undefined") return;
  localStorage.removeItem(DRAFT_KEY);
}
