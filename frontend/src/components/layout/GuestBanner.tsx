"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { Info } from "lucide-react";
import { useIsGuest } from "@/context/AuthContext";
import { loginPathWithReturn, registerPathWithReturn } from "@/lib/return-url";

export function GuestBanner() {
  const isGuest = useIsGuest();
  const pathname = usePathname();
  const t = useTranslations("guest");

  if (!isGuest) return null;

  const loginHref = loginPathWithReturn(pathname);
  const registerHref = registerPathWithReturn(pathname);

  return (
    <div className="mb-6 flex flex-wrap items-center gap-x-3 gap-y-2 rounded-lg border border-accent bg-accent-bg px-4 py-3 text-sm text-accent">
      <Info size={16} className="shrink-0" />
      <p className="flex-1 min-w-[200px]">{t("banner")}</p>
      <div className="flex flex-wrap gap-2">
        <Link
          href={loginHref}
          className="rounded-md bg-text px-3 py-1.5 text-xs font-medium text-white hover:opacity-90"
        >
          {t("login")}
        </Link>
        <Link
          href={registerHref}
          className="rounded-md border border-border2 bg-bg px-3 py-1.5 text-xs font-medium text-text hover:bg-bg2"
        >
          {t("register")}
        </Link>
      </div>
    </div>
  );
}
