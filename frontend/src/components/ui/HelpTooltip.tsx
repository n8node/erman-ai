"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { HelpCircle } from "lucide-react";
import { fetchTooltips } from "@/lib/api";
import { cn } from "@/lib/utils";

type TooltipMap = Record<string, string>;

const TooltipContext = createContext<TooltipMap>({});

export function TooltipProvider({
  prefix = "calculator",
  children,
}: {
  prefix?: string;
  children: ReactNode;
}) {
  const [tooltips, setTooltips] = useState<TooltipMap>({});

  useEffect(() => {
    fetchTooltips(prefix)
      .then((data) => setTooltips(data.tooltips))
      .catch(() => setTooltips({}));
  }, [prefix]);

  return (
    <TooltipContext.Provider value={tooltips}>{children}</TooltipContext.Provider>
  );
}

export function HelpTooltip({
  tooltipKey,
  className,
}: {
  tooltipKey: string;
  className?: string;
}) {
  const tooltips = useContext(TooltipContext);
  const text = tooltips[tooltipKey];
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  const close = useCallback(() => setOpen(false), []);

  useEffect(() => {
    if (!open) return;
    function onPointerDown(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        close();
      }
    }
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") close();
    }
    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open, close]);

  if (!text) return null;

  return (
    <div ref={ref} className={cn("relative inline-flex", className)}>
      <button
        type="button"
        aria-label="Help"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
        className={cn(
          "inline-flex h-4 w-4 items-center justify-center rounded-full border border-border2 text-text3",
          "transition-colors hover:border-accent hover:text-accent focus:outline-none focus:ring-2 focus:ring-accent/30",
          open && "border-accent text-accent bg-accent-bg"
        )}
      >
        <HelpCircle size={10} strokeWidth={2.25} />
      </button>
      {open && (
        <div
          role="tooltip"
          className="absolute left-1/2 top-full z-50 mt-2 w-64 -translate-x-1/2 rounded-lg border border-border bg-bg px-3 py-2.5 text-xs leading-relaxed text-text shadow-lg"
        >
          <div className="absolute -top-1.5 left-1/2 h-2.5 w-2.5 -translate-x-1/2 rotate-45 border-l border-t border-border bg-bg" />
          {text}
        </div>
      )}
    </div>
  );
}

export function LabelWithHelp({
  label,
  tooltipKey,
  className,
}: {
  label: string;
  tooltipKey: string;
  className?: string;
}) {
  return (
    <span className={cn("inline-flex items-center gap-1.5", className)}>
      {label}
      <HelpTooltip tooltipKey={tooltipKey} />
    </span>
  );
}
