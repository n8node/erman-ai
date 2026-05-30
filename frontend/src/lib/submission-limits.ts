const GUEST_DISCUSS_KEY = "erman_project_inquiry_submitted";

export function isGuestDiscussSubmitted(): boolean {
  if (typeof window === "undefined") return false;
  return window.localStorage.getItem(GUEST_DISCUSS_KEY) === "1";
}

export function markGuestDiscussSubmitted(): void {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(GUEST_DISCUSS_KEY, "1");
}
