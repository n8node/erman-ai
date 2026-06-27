"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { LogOut, Menu } from "lucide-react";
import type { User } from "@/lib/api";
import { logout } from "@/lib/api";
import { LanguageSwitcher } from "./LanguageSwitcher";
import { loginPathWithReturn, registerPathWithReturn } from "@/lib/return-url";

type Props = {
  user: User | null;
  onMenuClick: () => void;
};

export function Topbar({ user, onMenuClick }: Props) {
  const t = useTranslations("nav");
  const tGuest = useTranslations("guest");
  const router = useRouter();
  const pathname = usePathname();

  async function handleLogout() {
    await logout().catch(() => undefined);
    router.push("/login");
    router.refresh();
  }

  return (
    <header className="sticky top-0 z-20 flex min-h-[52px] items-center justify-between gap-3 border-b border-border bg-bg px-4 py-2 sm:px-6">
      <div className="flex min-w-0 items-center gap-3">
        <button
          type="button"
          aria-label="Открыть меню"
          onClick={onMenuClick}
          className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-border2 text-text hover:bg-bg2 lg:hidden"
        >
          <Menu size={18} />
        </button>
        <span className="truncate text-sm text-text2">{t("dashboard")}</span>
      </div>
      <div className="flex min-w-0 items-center gap-2 sm:gap-3">
        <LanguageSwitcher userEmail={user?.email} />
        {user ? (
          <>
            <span className="hidden max-w-[220px] truncate text-[13px] text-text2 sm:inline">
              {user.email}
            </span>
            <button
              type="button"
              onClick={handleLogout}
              className="inline-flex items-center gap-1.5 rounded-md border border-border2 px-2.5 py-1.5 text-[13px] text-text hover:bg-bg2 sm:px-3"
            >
              <LogOut size={14} />
              <span className="hidden sm:inline">{t("logout")}</span>
            </button>
          </>
        ) : (
          <>
            <Link
              href={loginPathWithReturn(pathname)}
              className="rounded-md border border-border2 px-3 py-1.5 text-[13px] text-text hover:bg-bg2"
            >
              {tGuest("login")}
            </Link>
            <Link
              href={registerPathWithReturn(pathname)}
              className="rounded-md bg-text px-3 py-1.5 text-[13px] font-medium text-white hover:opacity-90"
            >
              {tGuest("register")}
            </Link>
          </>
        )}
      </div>
    </header>
  );
}
