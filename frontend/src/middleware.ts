import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { LOCALE_COOKIE, isAppLocale } from "./i18n/locales";
import { negotiateFromAcceptLanguage } from "./i18n/negotiate";

/** Pages reachable without a session (basePath is stripped in middleware). */
const publicPages = ["/login", "/register", "/verify-email"];

const accountPrefixes = ["/billing", "/settings", "/api-keys"];
const adminPrefix = "/admin";

function isSharePath(pathname: string) {
  return pathname.startsWith("/share/") || pathname.startsWith("/dashboard/share/");
}

function isGuestToolPath(pathname: string) {
  return pathname === "/" || pathname.startsWith("/tools");
}

function isAccountPath(pathname: string) {
  return accountPrefixes.some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`)
  );
}

function isAdminPath(pathname: string) {
  return pathname === adminPrefix || pathname.startsWith(`${adminPrefix}/`);
}

function loginRedirect(request: NextRequest, returnPath: string) {
  const url = request.nextUrl.clone();
  url.pathname = "/login";
  url.search = "";
  if (returnPath && returnPath !== "/login") {
    url.searchParams.set("next", returnPath);
  }
  return NextResponse.redirect(url);
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get("access_token");

  let response: NextResponse;

  if (isSharePath(pathname)) {
    response = NextResponse.next();
  } else if (!token) {
    if (publicPages.includes(pathname)) {
      response = NextResponse.next();
    } else if (isAccountPath(pathname) || isAdminPath(pathname)) {
      response = loginRedirect(request, pathname);
    } else if (isGuestToolPath(pathname)) {
      response = NextResponse.next();
    } else {
      response = loginRedirect(request, pathname);
    }
  } else if (token && publicPages.includes(pathname)) {
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
