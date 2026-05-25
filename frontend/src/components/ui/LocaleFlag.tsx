import CN from "country-flag-icons/react/3x2/CN";
import DE from "country-flag-icons/react/3x2/DE";
import ES from "country-flag-icons/react/3x2/ES";
import FR from "country-flag-icons/react/3x2/FR";
import GB from "country-flag-icons/react/3x2/GB";
import RU from "country-flag-icons/react/3x2/RU";
import type { AppLocale } from "@/i18n/locales";
import { cn } from "@/lib/utils";

type FlagComponent = typeof GB;

const FLAG_BY_LOCALE: Record<AppLocale, FlagComponent> = {
  en: GB,
  ru: RU,
  de: DE,
  es: ES,
  fr: FR,
  zh: CN,
};

type Props = {
  locale: AppLocale;
  className?: string;
};

export function LocaleFlag({ locale, className }: Props) {
  const Flag = FLAG_BY_LOCALE[locale] ?? RU;
  return (
    <Flag
      aria-hidden
      className={cn("h-3.5 w-[1.375rem] shrink-0 rounded-[2px] object-cover shadow-sm", className)}
    />
  );
}
