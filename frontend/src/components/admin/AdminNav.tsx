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
    { href: "/admin/proposal-requests", label: t("proposalRequests") },
    { href: "/admin/project-inquiries", label: t("projectInquiries") },
    { href: "/admin/consultations", label: t("consultations") },
    { href: "/admin/calculator-budget", label: t("calculatorBudget") },
    { href: "/admin/legal-risks", label: t("legalRisks") },
    { href: "/admin/legal-scan-llm", label: t("legalScanLlm") },
    { href: "/admin/strategy-llm", label: t("strategyLlm") },
    { href: "/admin/email", label: t("emailSmtp") },
    { href: "/admin/payments", label: t("payments") },
    { href: "/admin/telegram", label: t("telegram") },
    { href: "/admin/max", label: t("max") },
    { href: "/admin/yandex-metrika", label: t("yandexMetrika") },
    { href: "/admin/tooltips", label: t("tooltips") },
    { href: "/admin/translations", label: t("translations") },
    { href: "/admin/projects", label: t("projects") },
    { href: "/admin/public-pages", label: t("publicPages") },
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
