"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { fetchTools, type ToolListItem } from "@/lib/api";
import { ToolLimitBadge } from "./ToolLimitBadge";

const CARD_TOOLS = [
  { slug: "calculator", href: "/tools/calculator" },
  { slug: "strategy", href: "/tools/strategy" },
  { slug: "proposal", href: "/tools/proposal" },
] as const;

export function DashboardToolCards() {
  const t = useTranslations("dashboard");
  const tLimits = useTranslations("toolLimits");
  const [tools, setTools] = useState<ToolListItem[]>([]);

  useEffect(() => {
    fetchTools()
      .then((data) => setTools(data.tools))
      .catch(() => {});
  }, []);

  function toolMeta(slug: string) {
    return tools.find((x) => x.slug === slug);
  }

  return (
    <div className="mt-8 grid gap-4 sm:grid-cols-3">
      {CARD_TOOLS.map(({ slug, href }) => {
        const meta = toolMeta(slug);
        return (
          <Link
            key={slug}
            href={href}
            className="rounded-xl border border-border bg-bg p-5 hover:border-border2 transition-colors"
          >
            <div className="flex items-start justify-between gap-2">
              <h2 className="text-sm font-medium">{t(`tools.${slug}.name`)}</h2>
              {meta && (
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
