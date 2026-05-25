import { cookies, headers } from "next/headers";
import type { AbstractIntlMessages } from "next-intl";
import { getRequestConfig } from "next-intl/server";
import { getMe } from "@/lib/auth-server";
import { LOCALE_COOKIE } from "./locales";
import { loadMessagesForLocale } from "./messages";
import { resolveLocale } from "./resolve-locale";

export default getRequestConfig(async () => {
  const cookieStore = await cookies();
  const headerStore = await headers();
  const user = await getMe();

  const locale = resolveLocale({
    cookie: cookieStore.get(LOCALE_COOKIE)?.value,
    userLocale: user?.locale,
    acceptLanguage: headerStore.get("accept-language"),
  });

  const messages = await loadMessagesForLocale(locale);

  return {
    locale,
    messages: messages as AbstractIntlMessages,
  };
});
