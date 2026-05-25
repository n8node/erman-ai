"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";

type Props = {
  toolName: string;
};

export function ToolLimitExceededAlert({ toolName }: Props) {
  const t = useTranslations("toolLimits");

  return (
    <div className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-950">
      <p className="font-medium">{t("exceededTitle", { tool: toolName })}</p>
      <p className="mt-1 text-xs text-amber-900/90">{t("exceededBody")}</p>
      <Link
        href="/billing"
        className="mt-3 inline-flex rounded-lg bg-text px-3 py-1.5 text-xs font-medium text-white hover:opacity-90"
      >
        {t("upgradeCta")}
      </Link>
    </div>
  );
}
