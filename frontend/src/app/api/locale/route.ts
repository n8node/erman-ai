import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { LOCALE_COOKIE, isAppLocale } from "@/i18n/locales";

export async function POST(request: Request) {
  const body = (await request.json()) as { locale?: string };
  if (!isAppLocale(body.locale)) {
    return NextResponse.json({ error: "invalid locale" }, { status: 400 });
  }

  const cookieStore = await cookies();
  cookieStore.set(LOCALE_COOKIE, body.locale, {
    path: "/",
    maxAge: 60 * 60 * 24 * 365,
    sameSite: "lax",
  });

  return NextResponse.json({ locale: body.locale });
}
