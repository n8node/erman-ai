import type { ToolListItem } from "@/lib/api";
import { cn } from "@/lib/utils";
import { formatRunsLimitLabel, isLimitReached } from "@/lib/tool-limits";

type Props = {
  tool: Pick<ToolListItem, "slug" | "runs_limit" | "runs_used">;
  t: (key: string, values?: Record<string, string | number>) => string;
  className?: string;
};

const slugStyles: Record<string, { text: string; bg: string }> = {
  calculator: { text: "text-accent", bg: "bg-accent-bg" },
  strategy: { text: "text-ai", bg: "bg-ai-bg" },
  proposal: { text: "text-success", bg: "bg-success-bg" },
};

export function ToolLimitBadge({ tool, t, className }: Props) {
  const styles = slugStyles[tool.slug] ?? { text: "text-text2", bg: "bg-bg2" };
  const exhausted = isLimitReached(tool);

  return (
    <span
      className={cn(
        "inline-flex rounded-lg px-2.5 py-1 text-[11px] font-medium",
        exhausted ? "bg-amber-50 text-amber-900" : styles.bg,
        !exhausted && styles.text,
        className
      )}
    >
      {exhausted ? t("limitReached") : formatRunsLimitLabel(tool, t)}
    </span>
  );
}
