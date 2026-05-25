import { DEFAULT_LOCALE, LOCALE_META, SUPPORTED_LOCALES, type AppLocale } from "./locales";

function parseAcceptLanguage(header: string | null | undefined): Array<{ tag: string; q: number }> {
  if (!header?.trim()) return [];
  return header
    .split(",")
    .map((part) => {
      const [tagPart, ...params] = part.trim().split(";");
      const tag = tagPart.trim().toLowerCase();
      let q = 1;
      for (const p of params) {
        const m = p.trim().match(/^q=([\d.]+)$/i);
        if (m) q = Number(m[1]);
      }
      return { tag, q };
    })
    .filter((x) => x.tag)
    .sort((a, b) => b.q - a.q);
}

function matchLocale(tag: string): AppLocale | null {
  const normalized = tag.toLowerCase();
  for (const code of SUPPORTED_LOCALES) {
    const meta = LOCALE_META[code];
    if (meta.bcp47.some((b) => normalized === b.toLowerCase() || normalized.startsWith(`${b.toLowerCase()}-`))) {
      return code;
    }
  }
  const primary = normalized.split("-")[0];
  if (SUPPORTED_LOCALES.includes(primary as AppLocale)) {
    return primary as AppLocale;
  }
  return null;
}

/** Pick best locale from Accept-Language (browser / OS preference). */
export function negotiateFromAcceptLanguage(header: string | null | undefined): AppLocale {
  for (const { tag } of parseAcceptLanguage(header)) {
    const matched = matchLocale(tag);
    if (matched) return matched;
  }
  return DEFAULT_LOCALE;
}
