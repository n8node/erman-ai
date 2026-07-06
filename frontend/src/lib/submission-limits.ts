export type GuestDiscussHistoryItem = {
  id?: string;
  project_title: string;
  project_description?: string;
  status: string;
  created_at: string;
};

const GUEST_DISCUSS_HISTORY_KEY = "erman_project_inquiry_history";

function readHistory(): GuestDiscussHistoryItem[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = window.localStorage.getItem(GUEST_DISCUSS_HISTORY_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(
      (item): item is GuestDiscussHistoryItem =>
        !!item &&
        typeof item === "object" &&
        typeof (item as GuestDiscussHistoryItem).status === "string" &&
        typeof (item as GuestDiscussHistoryItem).created_at === "string",
    );
  } catch {
    return [];
  }
}

function writeHistory(items: GuestDiscussHistoryItem[]) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(GUEST_DISCUSS_HISTORY_KEY, JSON.stringify(items.slice(0, 50)));
}

export function loadGuestDiscussHistory(): GuestDiscussHistoryItem[] {
  return readHistory();
}

export function appendGuestDiscussHistory(item: GuestDiscussHistoryItem) {
  writeHistory([item, ...readHistory()]);
}

export function updateGuestDiscussHistoryStatus(id: string, status: string) {
  const items = readHistory().map((row) => (row.id === id ? { ...row, status } : row));
  writeHistory(items);
}
