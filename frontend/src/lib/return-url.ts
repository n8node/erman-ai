/** Safe internal path for post-login redirect. */
export function sanitizeReturnPath(next: string | null | undefined): string {
  if (!next || !next.startsWith("/") || next.startsWith("//")) return "/";
  if (
    next.startsWith("/login") ||
    next.startsWith("/register") ||
    next.startsWith("/verify-email")
  ) {
    return "/";
  }
  return next;
}

export function loginPathWithReturn(returnPath: string): string {
  return `/login?next=${encodeURIComponent(returnPath)}`;
}

export function registerPathWithReturn(returnPath: string): string {
  return `/register?next=${encodeURIComponent(returnPath)}`;
}
