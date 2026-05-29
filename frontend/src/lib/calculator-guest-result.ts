import type { CalculatorInput, CalculatorOutput } from "./api";

const GUEST_RESULT_KEY = "erman:calculator-guest-result:v1";
const MAX_AGE_MS = 7 * 24 * 60 * 60 * 1000;

export type GuestCalculatorResult = {
  input: CalculatorInput;
  output: CalculatorOutput;
  savedAt: string;
};

export function saveGuestCalculatorResult(input: CalculatorInput, output: CalculatorOutput) {
  if (typeof window === "undefined") return;
  const payload: GuestCalculatorResult = {
    input,
    output,
    savedAt: new Date().toISOString(),
  };
  localStorage.setItem(GUEST_RESULT_KEY, JSON.stringify(payload));
}

export function loadGuestCalculatorResult(): GuestCalculatorResult | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(GUEST_RESULT_KEY);
    if (!raw) return null;
    const data = JSON.parse(raw) as GuestCalculatorResult;
    if (!data.savedAt || Date.now() - new Date(data.savedAt).getTime() > MAX_AGE_MS) {
      localStorage.removeItem(GUEST_RESULT_KEY);
      return null;
    }
    if (!data.input || !data.output) return null;
    return data;
  } catch {
    return null;
  }
}

export function clearGuestCalculatorResult() {
  if (typeof window === "undefined") return;
  localStorage.removeItem(GUEST_RESULT_KEY);
}
