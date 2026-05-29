"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { LogOut } from "lucide-react";
import type { User } from "@/lib/api";
import { logout } from "@/lib/api";
import { LanguageSwitcher } from "./LanguageSwitcher";
import { loginPathWithReturn, registerPathWithReturn } from "@/lib/return-url";

type Props = {
  user: User | null;
};

export function Topbar({ user }: Props) {
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
    <header className="sticky top-0 z-20 flex h-[52px] items-center justify-between border-b border-border bg-bg px-6">
      <span className="text-sm text-text2">{t("dashboard")}</span>
      <div className="flex items-center gap-3">
        <LanguageSwitcher userEmail={user?.email} />
        {user ? (
          <>
            <span className="text-[13px] text-text2">{user.email}</span>
            <button
              type="button"
              onClick={handleLogout}
              className="inline-flex items-center gap-1.5 rounded-md border border-border2 px-3 py-1.5 text-[13px] text-text hover:bg-bg2"
            >
              <LogOut size={14} />
              {t("logout")}
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
