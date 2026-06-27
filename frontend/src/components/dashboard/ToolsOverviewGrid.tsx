"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import {
  Brain,
  Calculator,
  ClipboardList,
  FileText,
  ShieldCheck,
  type LucideIcon,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { fetchTools, type ToolListItem } from "@/lib/api";
import { useIsGuest } from "@/context/AuthContext";
import { cn } from "@/lib/utils";
import { ToolLimitBadge } from "./ToolLimitBadge";

const TOOLS: {
  slug: string;
  href: string;
  icon: LucideIcon;
  badge: { text: string; bg: string };
}[] = [
  {
    slug: "calculator",
    href: "/tools/calculator",
    icon: Calculator,
    badge: { text: "text-accent", bg: "bg-accent-bg" },
  },
  {
    slug: "strategy",
    href: "/tools/strategy",
    icon: Brain,
    badge: { text: "text-ai", bg: "bg-ai-bg" },
  },
  {
    slug: "proposal",
    href: "/tools/proposal",
    icon: FileText,
    badge: { text: "text-success", bg: "bg-success-bg" },
  },
  {
    slug: "audit",
    href: "/tools/audit",
    icon: ClipboardList,
    badge: { text: "text-warning", bg: "bg-warning-bg" },
  },
  {
    slug: "legal-scan",
    href: "/tools/legal-scan",
    icon: ShieldCheck,
    badge: { text: "text-text2", bg: "bg-bg2" },
  },
];

export function ToolsOverviewGrid() {
  const t = useTranslations("toolsPage");
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

  return (
    <div className="mt-8 grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
      {TOOLS.map(({ slug, href, icon: Icon, badge }) => {
        const meta = toolMeta(slug);
        return (
          <article
            key={slug}
            className={cn(
              "group flex flex-col rounded-xl border border-border bg-bg p-6",
              "transition-all duration-200 ease-out",
              "hover:-translate-y-0.5 hover:border-border2 hover:shadow-[0_8px_24px_rgba(26,26,26,0.08)]"
            )}
          >
            <div className="flex items-start justify-between gap-3">
              <div
                className={cn(
                  "flex h-10 w-10 shrink-0 items-center justify-center rounded-lg",
                  badge.bg,
                  badge.text
                )}
              >
                <Icon size={20} strokeWidth={1.75} />
              </div>
              {!isGuest && meta && (
                <ToolLimitBadge
                  tool={meta}
                  t={(key, values) =>
                    tLimits(key, values as Record<string, string | number> | undefined)
                  }
                />
              )}
            </div>

            <h2 className="mt-4 text-[15px] font-medium leading-snug">
              {t(`tools.${slug}.name`)}
            </h2>

            <p className="mt-2 flex-1 text-sm leading-relaxed text-text2">
              {t(`tools.${slug}.description`)}
            </p>

            <p className="mt-4 text-xs leading-relaxed text-text3">
              <span className="font-medium uppercase tracking-wider text-text3">
                {t("forWhom")}
              </span>
              <span className="mt-1 block normal-case tracking-normal text-text2">
                {t(`tools.${slug}.audience`)}
              </span>
            </p>

            <Link
              href={href}
              className={cn(
                "mt-5 inline-flex w-full items-center justify-center rounded-lg border border-border2",
                "bg-bg px-4 py-2.5 text-sm font-medium text-text",
                "transition-colors group-hover:border-text group-hover:bg-text group-hover:text-white"
              )}
            >
              {t("openTool")}
            </Link>
          </article>
        );
      })}
    </div>
  );
}
