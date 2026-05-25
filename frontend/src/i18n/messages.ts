import type { AbstractIntlMessages } from "next-intl";
import type { AppLocale } from "./locales";

type MessageTree = AbstractIntlMessages;

function isPlainObject(v: unknown): v is Record<string, AbstractIntlMessages | string> {
  return typeof v === "object" && v !== null && !Array.isArray(v);
}

export function deepMergeMessages(base: MessageTree, overlay: MessageTree): MessageTree {
  const out: MessageTree = { ...base };
  for (const [key, value] of Object.entries(overlay)) {
    if (isPlainObject(value) && isPlainObject(out[key])) {
      out[key] = deepMergeMessages(out[key] as MessageTree, value as MessageTree);
    } else {
      out[key] = value;
    }
  }
  return out;
}

export function flattenMessages(obj: MessageTree, prefix = ""): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [key, value] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key;
    if (typeof value === "string") {
      out[fullKey] = value;
    } else if (isPlainObject(value)) {
      Object.assign(out, flattenMessages(value as MessageTree, fullKey));
    }
  }
  return out;
}

export function unflattenMessages(flat: Record<string, string>): MessageTree {
  const out: MessageTree = {};
  for (const [key, value] of Object.entries(flat)) {
    const parts = key.split(".");
    let cursor: MessageTree = out;
    for (let i = 0; i < parts.length - 1; i++) {
      const part = parts[i];
      if (!isPlainObject(cursor[part])) cursor[part] = {};
      cursor = cursor[part] as MessageTree;
    }
    cursor[parts[parts.length - 1]] = value;
  }
  return out;
}

async function loadJsonLocale(locale: AppLocale): Promise<MessageTree> {
  switch (locale) {
    case "en":
      return (await import("../messages/en.json")).default;
    case "ru":
      return (await import("../messages/ru.json")).default;
    case "de":
      return (await import("../messages/de.json")).default;
    case "es":
      return (await import("../messages/es.json")).default;
    case "fr":
      return (await import("../messages/fr.json")).default;
    case "zh":
      return (await import("../messages/zh.json")).default;
    default:
      return (await import("../messages/en.json")).default;
  }
}

function serverApiBase() {
  return process.env.INTERNAL_API_URL?.replace(/\/$/, "") || "http://backend:8080";
}

async function fetchTranslationOverrides(locale: AppLocale): Promise<Record<string, string>> {
  try {
    const res = await fetch(
      `${serverApiBase()}/api/v1/public/translations?locale=${encodeURIComponent(locale)}`,
      { next: { revalidate: 60 } }
    );
    if (!res.ok) return {};
    const data = (await res.json()) as { translations?: Record<string, string> };
    return data.translations ?? {};
  } catch {
    return {};
  }
}

function isBadTranslation(value: string): boolean {
  return value.includes("MYMEMORY") || value.includes("USAGELIMITS.PHP");
}

function sanitizeMessages(base: MessageTree, fallback: MessageTree): MessageTree {
  const out: MessageTree = {};
  for (const [key, value] of Object.entries(base)) {
    if (typeof value === "string") {
      const fb = typeof fallback[key] === "string" ? (fallback[key] as string) : value;
      out[key] = isBadTranslation(value) ? fb : value;
    } else if (isPlainObject(value)) {
      const fbTree = isPlainObject(fallback[key]) ? (fallback[key] as MessageTree) : {};
      out[key] = sanitizeMessages(value as MessageTree, fbTree);
    } else {
      out[key] = value;
    }
  }
  return out;
}

export async function loadMessagesForLocale(locale: AppLocale): Promise<MessageTree> {
  const en = (await import("../messages/en.json")).default as MessageTree;
  let base: MessageTree;
  try {
    base = await loadJsonLocale(locale);
  } catch {
    base = en;
  }
  if (locale !== "en") {
    base = sanitizeMessages(deepMergeMessages(en, base), en);
  }
  const overrides = await fetchTranslationOverrides(locale);
  if (Object.keys(overrides).length === 0) return base;
  const merged = deepMergeMessages(base, unflattenMessages(overrides));
  return locale === "en" ? merged : sanitizeMessages(merged, en);
}
