import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { LOCALE_COOKIE, isAppLocale } from "./i18n/locales";
import { negotiateFromAcceptLanguage } from "./i18n/negotiate";

/** Must match next.config.ts basePath. */
const BASE_PATH = "/dashboard";

/** Pages reachable without a session (app paths, without basePath). */
const publicPages = [
  "/login",
  "/register",
  "/verify-email",
  "/forgot-password",
  "/reset-password",
];

/** API routes that must work without a session. */
const publicApiPaths = ["/api/locale"];

const accountPrefixes = ["/settings", "/api-keys"];
const adminPrefix = "/admin";

const reservedSingleSegments = new Set([
  "login",
  "register",
  "verify-email",
  "forgot-password",
  "reset-password",
  "onboarding",
  "share",
  "tools",
  "admin",
  "billing",
  "settings",
  "api-keys",
  "discuss",
  "api",
]);

const defaultPublicPageSlugs = [
  "about",
  "privacy-policy",
  "terms",
  "shop-terms",
  "refund",
];

let cachedPublicSlugs: string[] = defaultPublicPageSlugs;
let cacheExpiresAt = 0;

/** Normalize pathname — middleware may receive with or without basePath. */
function appPathname(pathname: string): string {
  if (pathname === BASE_PATH) return "/";
  if (pathname.startsWith(`${BASE_PATH}/`)) {
    const rest = pathname.slice(BASE_PATH.length);
    return rest.startsWith("/") ? rest : `/${rest}`;
  }
  return pathname;
}

function isSharePath(pathname: string) {
  const p = appPathname(pathname);
  return p.startsWith("/share/");
}

function isGuestBillingPath(pathname: string) {
  const p = appPathname(pathname);
  return p === "/billing" || p.startsWith("/billing/");
}

function isGuestToolPath(pathname: string) {
  const p = appPathname(pathname);
  return p === "/" || p.startsWith("/tools") || p === "/discuss" || p.startsWith("/discuss/");
}

function isPublicApiPath(pathname: string) {
  const p = appPathname(pathname);
  return publicApiPaths.includes(p);
}

function isAccountPath(pathname: string) {
  const p = appPathname(pathname);
  return accountPrefixes.some(
    (prefix) => p === prefix || p.startsWith(`${prefix}/`)
  );
}

function isAdminPath(pathname: string) {
  const p = appPathname(pathname);
  return p === adminPrefix || p.startsWith(`${adminPrefix}/`);
}

function extractSingleSegmentSlug(pathname: string): string | null {
  const p = appPathname(pathname);
  const match = p.match(/^\/([^/]+)\/?$/);
  if (!match) return null;
  const slug = match[1];
  if (reservedSingleSegments.has(slug)) return null;
  return slug;
}

async function refreshPublicPageSlugs() {
  const now = Date.now();
  if (now < cacheExpiresAt) return cachedPublicSlugs;

  const apiBase =
    process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "/api/v1";
  try {
    const res = await fetch(`${apiBase}/public/pages/slugs`, {
      next: { revalidate: 60 },
    });
    if (res.ok) {
      const data = (await res.json()) as { slugs?: string[] };
      if (Array.isArray(data.slugs) && data.slugs.length > 0) {
        cachedPublicSlugs = data.slugs;
      }
    }
  } catch {
    cachedPublicSlugs = defaultPublicPageSlugs;
  }
  cacheExpiresAt = now + 60_000;
  return cachedPublicSlugs;
}

async function isPublicContentPath(pathname: string): Promise<boolean> {
  const slug = extractSingleSegmentSlug(pathname);
  if (!slug) return false;
  const slugs = await refreshPublicPageSlugs();
  return slugs.includes(slug);
}

function loginRedirect(request: NextRequest, returnPath: string) {
  const url = request.nextUrl.clone();
  url.pathname = "/login";
  url.search = "";
  const next = appPathname(returnPath);
  if (next && next !== "/login") {
    url.searchParams.set("next", next);
  }
  return NextResponse.redirect(url);
}

export async function middleware(request: NextRequest) {
  const pathname = appPathname(request.nextUrl.pathname);
  const token = request.cookies.get("access_token");

  let response: NextResponse;

  if (isSharePath(pathname)) {
    response = NextResponse.next();
  } else if (await isPublicContentPath(pathname)) {
    response = NextResponse.next();
  } else if (!token) {
    if (publicPages.includes(pathname) || isPublicApiPath(pathname)) {
      response = NextResponse.next();
    } else if (isAccountPath(pathname) || isAdminPath(pathname)) {
      response = loginRedirect(request, pathname);
    } else if (isGuestToolPath(pathname) || isGuestBillingPath(pathname)) {
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
