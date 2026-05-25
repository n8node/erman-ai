import type { AbstractIntlMessages } from "next-intl";
import en from "@/messages/en.json";
import ru from "@/messages/ru.json";
import de from "@/messages/de.json";
import es from "@/messages/es.json";
import fr from "@/messages/fr.json";
import zh from "@/messages/zh.json";
import { SUPPORTED_LOCALES, type AppLocale } from "@/i18n/locales";
import { flattenMessages } from "@/i18n/messages";

const FILE_MESSAGES: Record<AppLocale, AbstractIntlMessages> = {
  en,
  ru,
  de,
  es,
  fr,
  zh,
};

export type CatalogRow = {
  key: string;
  namespace: string;
  values: Record<AppLocale, string>;
};

export function buildTranslationCatalog(): CatalogRow[] {
  const enFlat = flattenMessages(en);
  const keys = Object.keys(enFlat).sort();

  return keys.map((key) => {
    const namespace = key.includes(".") ? key.split(".")[0] : key;
    const values = {} as Record<AppLocale, string>;
    for (const locale of SUPPORTED_LOCALES) {
      const flat = flattenMessages(FILE_MESSAGES[locale]);
      values[locale] = flat[key] ?? enFlat[key] ?? "";
    }
    return { key, namespace, values };
  });
}

export const TRANSLATION_NAMESPACES = Array.from(
  new Set(buildTranslationCatalog().map((row) => row.namespace))
).sort();
