"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { fetchTools, type ToolListItem } from "@/lib/api";
import { useIsGuest } from "@/context/AuthContext";
import { ToolLimitBadge } from "./ToolLimitBadge";

const CARD_TOOLS = [
  { slug: "calculator", href: "/tools/calculator" },
  { slug: "strategy", href: "/tools/strategy" },
  { slug: "proposal", href: "/tools/proposal" },
  { slug: "audit", href: "/tools/audit" },
  { slug: "legal-scan", href: "/tools/legal-scan" },
] as const;

const CONDITIONAL_TOOLS = [
  { slug: "geological-journal", href: "/tools/geological-journal" },
] as const;

export function DashboardToolCards() {
  const t = useTranslations("dashboard");
  const tLimits = useTranslations("toolLimits");
  const isGuest = useIsGuest();
  const [tools, setTools] = useState<ToolListItem[]>([]);

  useEffect(() => {
    if (isGuest) return;
    fetchTools()
      .then((data) => setTools(data.tools))
      .catch(() => {});
  }, [isGuest]);

  function toolMeta(slug: string) {
    return tools.find((x) => x.slug === slug);
  }

  const visibleTools = [
    ...CARD_TOOLS,
    ...CONDITIONAL_TOOLS.filter(({ slug }) => Boolean(toolMeta(slug))),
  ];

  return (
    <div className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {visibleTools.map(({ slug, href }) => {
        const meta = toolMeta(slug);
        return (
          <Link
            key={slug}
            href={href}
            className="rounded-xl border border-border bg-bg p-5 hover:border-border2 transition-colors"
          >
            <div className="flex items-start justify-between gap-2">
              <h2 className="text-sm font-medium">{t(`tools.${slug}.name`)}</h2>
              {!isGuest && meta && (
                <ToolLimitBadge
                  tool={meta}
                  t={(key, values) => tLimits(key, values as Record<string, string | number> | undefined)}
                />
              )}
            </div>
            <p className="mt-1 text-xs text-text2">{t(`tools.${slug}.description`)}</p>
          </Link>
        );
      })}
    </div>
  );
}
