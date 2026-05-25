import type { ToolListItem } from "./api";

export function remainingRuns(tool: Pick<ToolListItem, "runs_limit" | "runs_used">): number | null {
  if (tool.runs_limit === -1) return null;
  return Math.max(0, tool.runs_limit - tool.runs_used);
}

export function isLimitReached(tool: Pick<ToolListItem, "runs_limit" | "runs_used">): boolean {
  if (tool.runs_limit === -1) return false;
  return tool.runs_used >= tool.runs_limit;
}

export function formatRunsLimitLabel(
  tool: Pick<ToolListItem, "runs_limit" | "runs_used">,
  t: (key: string, values?: Record<string, string | number>) => string
): string {
  if (tool.runs_limit === -1) return t("unlimited");
  const left = remainingRuns(tool) ?? 0;
  return t("remaining", { left, total: tool.runs_limit });
}
