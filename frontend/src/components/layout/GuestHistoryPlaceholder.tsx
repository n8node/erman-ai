"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { useIsGuest } from "@/context/AuthContext";
import { loginPathWithReturn } from "@/lib/return-url";

type Props = {
  children: React.ReactNode;
};

export function GuestHistoryPlaceholder({ children }: Props) {
  const isGuest = useIsGuest();
  const pathname = usePathname();
  const t = useTranslations("guest");

  if (!isGuest) return <>{children}</>;

  return (
    <div className="rounded-xl border border-border bg-bg p-8 text-center">
      <h2 className="text-base font-medium">{t("historyTitle")}</h2>
      <p className="mt-2 text-sm text-text2">{t("historyMessage")}</p>
      <Link
        href={loginPathWithReturn(pathname)}
        className="mt-4 inline-block rounded-lg bg-text px-4 py-2 text-sm font-medium text-white hover:opacity-90"
      >
        {t("login")}
      </Link>
    </div>
  );
}
