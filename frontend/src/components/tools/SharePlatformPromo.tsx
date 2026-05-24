"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";

type Props = {
  /** Controlled by API: false for white_label partners */
  visible: boolean;
  registerHref?: string;
};

/**
 * Optional Erman AI promo on public share pages.
 * Kept in a dedicated component so white_label can hide it via show_platform_cta.
 */
export function SharePlatformPromo({
  visible,
  registerHref = "/register",
}: Props) {
  const t = useTranslations("calculator.share.platformPromo");

  if (!visible) return null;

  return (
    <aside
      data-share-promo="platform"
      className="rounded-xl border border-border bg-bg px-5 py-4 text-center"
    >
      <p className="text-sm text-text2">{t("text")}</p>
      <Link
        href={registerHref}
        className="mt-2 inline-block text-sm font-medium text-accent hover:underline"
      >
        {t("link")}
      </Link>
    </aside>
  );
}
