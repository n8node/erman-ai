import { DEFAULT_LOCALE, isAppLocale, type AppLocale } from "./locales";
import { negotiateFromAcceptLanguage } from "./negotiate";

type ResolveInput = {
  cookie?: string | null;
  userLocale?: string | null;
  acceptLanguage?: string | null;
};

/**
 * Priority: explicit cookie (user switcher) → saved profile → OS/browser → default.
 */
export function resolveLocale(input: ResolveInput): AppLocale {
  if (isAppLocale(input.cookie)) return input.cookie;
  if (isAppLocale(input.userLocale)) return input.userLocale;
  return negotiateFromAcceptLanguage(input.acceptLanguage);
}

export function resolveLocaleOrDefault(value: string | undefined | null): AppLocale {
  return isAppLocale(value) ? value : DEFAULT_LOCALE;
}
