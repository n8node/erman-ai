"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { loginPathWithReturn, registerPathWithReturn } from "@/lib/return-url";

type Props = {
  open: boolean;
  onClose: () => void;
  reason?: "generate" | "account";
};

export function GuestLoginModal({ open, onClose, reason = "generate" }: Props) {
  const t = useTranslations("guest");
  const pathname = usePathname();

  if (!open) return null;

  const loginHref = loginPathWithReturn(pathname);
  const registerHref = registerPathWithReturn(pathname);
  const message = reason === "generate" ? t("generateGate") : t("accountGate");

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md rounded-xl border border-border bg-bg p-6 shadow-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-base font-medium">{t("modalTitle")}</h2>
        <p className="mt-2 text-sm text-text2">{message}</p>
        <div className="mt-6 flex flex-wrap gap-2">
          <Link
            href={loginHref}
            className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white hover:opacity-90"
          >
            {t("login")}
          </Link>
          <Link
            href={registerHref}
            className="rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2"
          >
            {t("register")}
          </Link>
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg px-4 py-2 text-sm text-text2 hover:bg-bg2"
          >
            {t("cancel")}
          </button>
        </div>
      </div>
    </div>
  );
}
