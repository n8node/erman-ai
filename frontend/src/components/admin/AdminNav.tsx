"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";

export function AdminNav() {
  const t = useTranslations("admin.nav");
  const pathname = usePathname();

  const links = [
    { href: "/admin/users", label: t("users") },
    { href: "/admin/plans", label: t("plans") },
    { href: "/admin/tooltips", label: t("tooltips") },
  ];

  return (
    <nav className="flex gap-1 border-b border-border">
      {links.map(({ href, label }) => (
        <Link
          key={href}
          href={href}
          className={cn(
            "px-3 py-2 text-sm border-b-2 -mb-px",
            pathname === href || pathname.startsWith(`${href}/`)
              ? "border-text font-medium text-text"
              : "border-transparent text-text2 hover:text-text"
          )}
        >
          {label}
        </Link>
      ))}
    </nav>
  );
}
