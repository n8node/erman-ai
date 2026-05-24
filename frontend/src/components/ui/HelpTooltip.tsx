"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { createPortal } from "react-dom";
import { HelpCircle } from "lucide-react";
import { fetchTooltips } from "@/lib/api";
import { cn } from "@/lib/utils";

type TooltipMap = Record<string, string>;

const PANEL_WIDTH = 256;
const GAP = 8;
const VIEWPORT_PAD = 12;

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

type PanelPosition = {
  top: number;
  left: number;
  arrowLeft: number;
};

function computePosition(trigger: HTMLElement): PanelPosition {
  const rect = trigger.getBoundingClientRect();
  const centerX = rect.left + rect.width / 2;

  let left = centerX - PANEL_WIDTH / 2;
  left = Math.max(
    VIEWPORT_PAD,
    Math.min(left, window.innerWidth - PANEL_WIDTH - VIEWPORT_PAD)
  );

  const top = rect.bottom + GAP;
  const arrowLeft = Math.min(
    Math.max(centerX - left, 16),
    PANEL_WIDTH - 16
  );

  return { top, left, arrowLeft };
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
  const [mounted, setMounted] = useState(false);
  const [position, setPosition] = useState<PanelPosition | null>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);

  const close = useCallback(() => setOpen(false), []);

  useEffect(() => {
    setMounted(true);
  }, []);

  const updatePosition = useCallback(() => {
    if (!triggerRef.current) return;
    setPosition(computePosition(triggerRef.current));
  }, []);

  useLayoutEffect(() => {
    if (!open) return;
    updatePosition();
    window.addEventListener("resize", updatePosition);
    window.addEventListener("scroll", updatePosition, true);
    return () => {
      window.removeEventListener("resize", updatePosition);
      window.removeEventListener("scroll", updatePosition, true);
    };
  }, [open, updatePosition]);

  useEffect(() => {
    if (!open) return;
    function onPointerDown(e: MouseEvent) {
      const target = e.target as Node;
      if (triggerRef.current?.contains(target)) return;
      if (panelRef.current?.contains(target)) return;
      close();
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

  const panel =
    open && position && mounted ? (
      createPortal(
        <div
          ref={panelRef}
          role="tooltip"
          style={{
            position: "fixed",
            top: position.top,
            left: position.left,
            width: PANEL_WIDTH,
            zIndex: 9999,
          }}
          className="rounded-lg border border-border bg-bg px-3 py-2.5 text-xs font-normal normal-case leading-relaxed tracking-normal text-text shadow-lg"
        >
          <div
            className="absolute -top-1.5 h-2.5 w-2.5 rotate-45 border-l border-t border-border bg-bg"
            style={{ left: position.arrowLeft - 5 }}
          />
          {text}
        </div>,
        document.body
      )
    ) : null;

  return (
    <>
      <span className={cn("inline-flex shrink-0 align-middle", className)}>
        <button
          ref={triggerRef}
          type="button"
          aria-label="Help"
          aria-expanded={open}
          onClick={(e) => {
            e.stopPropagation();
            setOpen((v) => {
              const next = !v;
              if (next && triggerRef.current) {
                setPosition(computePosition(triggerRef.current));
              }
              return next;
            });
          }}
          className={cn(
            "inline-flex h-4 w-4 items-center justify-center rounded-full border border-border2 text-text3",
            "transition-colors hover:border-accent hover:text-accent focus:outline-none focus:ring-2 focus:ring-accent/30",
            open && "border-accent text-accent bg-accent-bg"
          )}
        >
          <HelpCircle size={10} strokeWidth={2.25} />
        </button>
      </span>
      {panel}
    </>
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
