import type { AppLocale } from "@/i18n/locales";
import { isAppLocale } from "@/i18n/locales";

const INTL_BY_LOCALE: Record<AppLocale, string> = {
  en: "en-US",
  ru: "ru-RU",
  de: "de-DE",
  es: "es-ES",
  fr: "fr-FR",
  zh: "zh-CN",
};

export function intlLocale(locale: string): string {
  if (isAppLocale(locale)) return INTL_BY_LOCALE[locale];
  return INTL_BY_LOCALE.ru;
}
