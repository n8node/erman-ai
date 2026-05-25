import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { LOCALE_COOKIE, isAppLocale } from "./i18n/locales";
import { negotiateFromAcceptLanguage } from "./i18n/negotiate";

const authPages = ["/login", "/register"];

function isSharePath(pathname: string) {
  return pathname.startsWith("/share/") || pathname.startsWith("/dashboard/share/");
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get("access_token");

  let response: NextResponse;

  if (isSharePath(pathname)) {
    response = NextResponse.next();
  } else if (!token && !authPages.includes(pathname)) {
    const url = request.nextUrl.clone();
    url.pathname = "/login";
    response = NextResponse.redirect(url);
  } else if (token && authPages.includes(pathname)) {
    const url = request.nextUrl.clone();
    url.pathname = "/";
    response = NextResponse.redirect(url);
  } else {
    response = NextResponse.next();
  }

  const cookieLocale = request.cookies.get(LOCALE_COOKIE)?.value;
  if (!isAppLocale(cookieLocale)) {
    const negotiated = negotiateFromAcceptLanguage(request.headers.get("accept-language"));
    response.cookies.set(LOCALE_COOKIE, negotiated, {
      path: "/",
      maxAge: 60 * 60 * 24 * 365,
      sameSite: "lax",
    });
  }

  return response;
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
