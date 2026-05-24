import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const authPages = ["/login", "/register"];

function isSharePath(pathname: string) {
  return pathname.startsWith("/share/") || pathname.startsWith("/dashboard/share/");
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get("access_token");

  if (isSharePath(pathname)) {
    return NextResponse.next();
  }

  if (!token && !authPages.includes(pathname)) {
    const url = request.nextUrl.clone();
    url.pathname = "/login";
    return NextResponse.redirect(url);
  }

  if (token && authPages.includes(pathname)) {
    const url = request.nextUrl.clone();
    url.pathname = "/";
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
