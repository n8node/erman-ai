"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";

export function ProposalSubnav() {
  const t = useTranslations("proposal");
  const pathname = usePathname();
  const isHistory = pathname.endsWith("/history");

  return (
    <div className="flex gap-1 border-b border-border">
      <Link
        href="/tools/proposal"
        className={cn(
          "px-3 py-2 text-sm border-b-2 -mb-px",
          !isHistory
            ? "border-text font-medium text-text"
            : "border-transparent text-text2 hover:text-text"
        )}
      >
        {t("subnav.new")}
      </Link>
      <Link
        href="/tools/proposal/history"
        className={cn(
          "px-3 py-2 text-sm border-b-2 -mb-px",
          isHistory
            ? "border-text font-medium text-text"
            : "border-transparent text-text2 hover:text-text"
        )}
      >
        {t("subnav.history")}
      </Link>
    </div>
  );
}
