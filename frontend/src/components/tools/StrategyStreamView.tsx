"use client";

import { useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { subscribeStrategyStream } from "@/lib/api";
import type { StrategyStreamPhase } from "@/lib/api-strategy";
import { cn } from "@/lib/utils";

const PREP_PHASES = ["profile", "processes", "data", "priorities", "roadmap"] as const;

type PhaseState = "pending" | "active" | "done";

type Props = {
  runId: string;
  onComplete: () => void;
  onError: (message: string) => void;
};

export function StrategyStreamView({ runId, onComplete, onError }: Props) {
  const t = useTranslations("strategy.stream");
  const [phases, setPhases] = useState<Record<string, PhaseState>>(
    Object.fromEntries(PREP_PHASES.map((p) => [p, "pending"]))
  );
  const [generating, setGenerating] = useState(false);
  const [streamText, setStreamText] = useState("");
  const streamRef = useRef<HTMLDivElement>(null);
  const onCompleteRef = useRef(onComplete);
  const onErrorRef = useRef(onError);

  useEffect(() => {
    onCompleteRef.current = onComplete;
    onErrorRef.current = onError;
  }, [onComplete, onError]);

  useEffect(() => {
    const unsub = subscribeStrategyStream(runId, {
      onPhase: (data: StrategyStreamPhase) => {
        if (data.id === "generating") {
          setGenerating(data.status === "active");
          return;
        }
        if (!PREP_PHASES.includes(data.id as (typeof PREP_PHASES)[number])) return;
        setPhases((prev) => ({
          ...prev,
          [data.id]: data.status === "done" ? "done" : "active",
        }));
      },
      onChunk: (delta) => {
        setGenerating(true);
        setStreamText((prev) => prev + delta);
      },
      onDone: () => onCompleteRef.current(),
      onError: (msg) => onErrorRef.current(msg),
    });
    return unsub;
  }, [runId]);

  useEffect(() => {
    streamRef.current?.scrollTo({ top: streamRef.current.scrollHeight, behavior: "smooth" });
  }, [streamText]);

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-bg p-6 space-y-4">
        <p className="text-sm font-medium text-text">{t("prepTitle")}</p>
        <div className="space-y-3">
          {PREP_PHASES.map((id) => (
            <PrepStep
              key={id}
              label={t(`phases.${id}`)}
              state={phases[id] ?? "pending"}
            />
          ))}
          <PrepStep
            label={t("phases.generating")}
            state={generating ? (streamText ? "done" : "active") : "pending"}
            accent="ai"
          />
        </div>
      </div>

      {(generating || streamText) && (
        <div className="rounded-xl border border-border bg-bg p-4">
          <p className="mb-2 text-[10px] font-medium uppercase tracking-wider text-text3">
            {t("liveOutput")}
          </p>
          <div
            ref={streamRef}
            className="max-h-80 overflow-y-auto rounded-lg bg-bg3 p-4 font-mono text-xs leading-relaxed text-text2 whitespace-pre-wrap"
          >
            {streamText || t("waitingTokens")}
          </div>
        </div>
      )}
    </div>
  );
}

function PrepStep({
  label,
  state,
  accent = "accent",
}: {
  label: string;
  state: PhaseState;
  accent?: "accent" | "ai";
}) {
  const barColor =
    state === "done"
      ? "bg-green-600"
      : state === "active"
        ? accent === "ai"
          ? "bg-ai"
          : "bg-accent"
        : "bg-border";

  const pct = state === "done" ? 100 : state === "active" ? 65 : 8;

  return (
    <div>
      <div className="mb-1 flex items-center justify-between gap-2">
        <span
          className={cn(
            "text-xs",
            state === "done" ? "text-green-800 font-medium" : state === "active" ? "text-text font-medium" : "text-text3"
          )}
        >
          {label}
        </span>
        {state === "done" && <span className="text-[10px] text-green-700">✓</span>}
      </div>
      <div className="h-1.5 overflow-hidden rounded-full bg-bg2">
        <div
          className={cn("h-full rounded-full transition-all duration-700 ease-out", barColor)}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}
