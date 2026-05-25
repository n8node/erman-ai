export const LOCALE_COOKIE = "NEXT_LOCALE";

export type AppLocale = "en" | "ru" | "de" | "es" | "fr" | "zh";

export const DEFAULT_LOCALE: AppLocale = "ru";

export const SUPPORTED_LOCALES: AppLocale[] = ["en", "ru", "de", "es", "fr", "zh"];

export type LocaleMeta = {
  code: AppLocale;
  /** Native language name shown in the switcher */
  nativeName: string;
  /** BCP-47 tag for Accept-Language matching */
  bcp47: string[];
};

export const LOCALE_META: Record<AppLocale, LocaleMeta> = {
  en: { code: "en", nativeName: "English", bcp47: ["en", "en-US", "en-GB"] },
  ru: { code: "ru", nativeName: "Русский", bcp47: ["ru", "ru-RU"] },
  de: { code: "de", nativeName: "Deutsch", bcp47: ["de", "de-DE", "de-AT", "de-CH"] },
  es: { code: "es", nativeName: "Español", bcp47: ["es", "es-ES", "es-MX", "es-AR"] },
  fr: { code: "fr", nativeName: "Français", bcp47: ["fr", "fr-FR", "fr-CA", "fr-BE"] },
  zh: { code: "zh", nativeName: "中文", bcp47: ["zh", "zh-CN", "zh-Hans", "zh-Hant", "zh-TW"] },
};

export function isAppLocale(value: string | undefined | null): value is AppLocale {
  return Boolean(value && SUPPORTED_LOCALES.includes(value as AppLocale));
}

export function localeLanguageName(locale: AppLocale): string {
  return LOCALE_META[locale].nativeName;
}
