import { cn } from "@/lib/utils";

const dotStyles: Record<string, string> = {
  automate: "bg-[#3b6d11]",
  consider: "bg-[#ba7517]",
  not_recommended: "bg-[#a32d2d]",
};

export const recommendationBadgeStyles: Record<string, string> = {
  automate: "bg-[#eaf3de] text-[#3b6d11]",
  consider: "bg-[#faeeda] text-[#633806]",
  not_recommended: "bg-[#fcebeb] text-[#a32d2d]",
};

export function RecommendationDot({
  recommendation,
  title,
  className,
}: {
  recommendation?: string;
  title?: string;
  className?: string;
}) {
  const color = recommendation ? dotStyles[recommendation] : "bg-border2";

  return (
    <span
      className={cn("inline-block h-2 w-2 shrink-0 rounded-full", color, className)}
      title={title}
      aria-hidden={!title}
      role={title ? "img" : undefined}
      aria-label={title}
    />
  );
}

export function ProcessNameWithStatus({
  name,
  recommendation,
  statusLabel,
}: {
  name: string;
  recommendation?: string;
  statusLabel?: string;
}) {
  return (
    <span className="inline-flex items-center gap-2.5 min-w-0">
      <RecommendationDot recommendation={recommendation} title={statusLabel} />
      <span className="truncate">{name}</span>
    </span>
  );
}
