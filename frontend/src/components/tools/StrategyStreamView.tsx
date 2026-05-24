"use client";

import { useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { getRun, subscribeStrategyStream } from "@/lib/api";
import type { StrategyStreamPhase } from "@/lib/api-strategy";
import { cn } from "@/lib/utils";

const PREP_PHASES = ["profile", "processes", "data", "priorities", "roadmap"] as const;
const LLM_PASSES = ["context", "analysis", "plan", "governance", "pass_expand"] as const;
const POLL_MS = 3000;

type PhaseState = "pending" | "active" | "done";

type Props = {
  runId: string;
  onComplete: () => void;
  onError: (message: string) => void;
};

function formatElapsed(seconds: number) {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return m > 0 ? `${m}:${s.toString().padStart(2, "0")}` : `${s}s`;
}

export function StrategyStreamView({ runId, onComplete, onError }: Props) {
  const t = useTranslations("strategy.stream");
  const [phases, setPhases] = useState<Record<string, PhaseState>>(
    Object.fromEntries(PREP_PHASES.map((p) => [p, "pending"]))
  );
  const [llmPasses, setLlmPasses] = useState<Record<string, PhaseState>>(
    Object.fromEntries(LLM_PASSES.map((p) => [p, "pending"]))
  );
  const [generating, setGenerating] = useState(false);
  const [streamText, setStreamText] = useState("");
  const [streamLost, setStreamLost] = useState(false);
  const [elapsedSec, setElapsedSec] = useState(0);
  const streamRef = useRef<HTMLDivElement>(null);
  const onCompleteRef = useRef(onComplete);
  const onErrorRef = useRef(onError);
  const completedRef = useRef(false);

  useEffect(() => {
    onCompleteRef.current = onComplete;
    onErrorRef.current = onError;
  }, [onComplete, onError]);

  useEffect(() => {
    const tick = setInterval(() => setElapsedSec((s) => s + 1), 1000);
    return () => clearInterval(tick);
  }, []);

  useEffect(() => {
    let cancelled = false;
    let pollTimer: ReturnType<typeof setInterval> | null = null;

    function markComplete() {
      if (completedRef.current) return;
      completedRef.current = true;
      if (pollTimer) clearInterval(pollTimer);
      onCompleteRef.current();
    }

    function applyPhase(data: StrategyStreamPhase) {
      if (data.id === "generating") {
        setGenerating(data.status === "active" || data.status === "done");
        return;
      }
      if (LLM_PASSES.includes(data.id as (typeof LLM_PASSES)[number])) {
        setGenerating(true);
        setLlmPasses((prev) => {
          const next = { ...prev };
          const idx = LLM_PASSES.indexOf(data.id as (typeof LLM_PASSES)[number]);
          for (let i = 0; i < idx; i++) next[LLM_PASSES[i]] = "done";
          next[data.id] = data.status === "done" ? "done" : "active";
          return next;
        });
        return;
      }
      if (!PREP_PHASES.includes(data.id as (typeof PREP_PHASES)[number])) return;
      const idx = PREP_PHASES.indexOf(data.id as (typeof PREP_PHASES)[number]);
      setPhases((prev) => {
        const next = { ...prev };
        for (let i = 0; i < idx; i++) next[PREP_PHASES[i]] = "done";
        next[data.id] = data.status === "done" ? "done" : "active";
        return next;
      });
    }

    async function pollStatus() {
      try {
        const run = await getRun(runId);
        if (cancelled || completedRef.current) return;
        if (run.status === "processing") {
          setGenerating(true);
          setPhases((prev) => {
            const next = { ...prev };
            for (const p of PREP_PHASES) next[p] = "done";
            return next;
          });
        }
        if (run.status === "done") {
          markComplete();
        } else if (run.status === "error") {
          if (pollTimer) clearInterval(pollTimer);
          onErrorRef.current(run.error_msg || t("pollFailed"));
        }
      } catch {
        /* retry */
      }
    }

    pollTimer = setInterval(pollStatus, POLL_MS);
    void pollStatus();

    const unsub = subscribeStrategyStream(runId, {
      onPhase: applyPhase,
      onChunk: (delta) => {
        setGenerating(true);
        setStreamLost(false);
        setStreamText((prev) => prev + delta);
      },
      onDone: markComplete,
      onRunError: (msg) => {
        if (pollTimer) clearInterval(pollTimer);
        onErrorRef.current(msg);
      },
      onStreamLost: () => setStreamLost(true),
    });

    return () => {
      cancelled = true;
      unsub();
      if (pollTimer) clearInterval(pollTimer);
    };
  }, [runId, t]);

  useEffect(() => {
    streamRef.current?.scrollTo({ top: streamRef.current.scrollHeight, behavior: "smooth" });
  }, [streamText]);

  const activePass = LLM_PASSES.find((id) => llmPasses[id] === "active");
  const donePasses = LLM_PASSES.filter((id) => llmPasses[id] === "done").length;

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-bg p-6 space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <p className="text-sm font-medium text-text">{t("prepTitle")}</p>
          <span className="font-mono text-xs text-text3">{formatElapsed(elapsedSec)}</span>
        </div>

        {streamLost && (
          <p className="text-xs text-text2">{t("fallbackHint")}</p>
        )}
        {generating && elapsedSec > 30 && (
          <p className="text-xs text-text2">{t("longRunningHint")}</p>
        )}

        <div className="space-y-3">
          {PREP_PHASES.map((id) => (
            <PrepStep key={id} label={t(`phases.${id}`)} state={phases[id] ?? "pending"} />
          ))}
          <PrepStep
            label={t("phases.generating")}
            state={generating ? "active" : "pending"}
            accent="ai"
            pulse={generating}
          />
        </div>

        {generating && (
          <div className="border-t border-border pt-4 space-y-2">
            <p className="text-[10px] font-medium uppercase tracking-wider text-text3">
              {t("llmPassesTitle")} {donePasses > 0 ? `(${donePasses}/${LLM_PASSES.length})` : ""}
            </p>
            {LLM_PASSES.map((id) => (
              <PrepStep
                key={id}
                label={t(`passes.${id}`)}
                state={llmPasses[id] ?? "pending"}
                accent="ai"
                compact
              />
            ))}
            {activePass && (
              <p className="text-xs text-ai pt-1">{t("activePass", { pass: t(`passes.${activePass}`) })}</p>
            )}
          </div>
        )}
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
  pulse = false,
  compact = false,
}: {
  label: string;
  state: PhaseState;
  accent?: "accent" | "ai";
  pulse?: boolean;
  compact?: boolean;
}) {
  const barColor =
    state === "done"
      ? "bg-green-600"
      : state === "active"
        ? accent === "ai"
          ? "bg-ai"
          : "bg-accent"
        : "bg-border";

  const pct = state === "done" ? 100 : state === "active" ? (pulse ? 75 : 65) : 8;

  return (
    <div className={compact ? "opacity-90" : undefined}>
      <div className="mb-1 flex items-center justify-between gap-2">
        <span
          className={cn(
            compact ? "text-[11px]" : "text-xs",
            state === "done" ? "text-green-800 font-medium" : state === "active" ? "text-text font-medium" : "text-text3"
          )}
        >
          {label}
        </span>
        {state === "done" && <span className="text-[10px] text-green-700">✓</span>}
        {state === "active" && pulse && (
          <span className="text-[10px] text-ai animate-pulse">…</span>
        )}
      </div>
      <div className={cn("overflow-hidden rounded-full bg-bg2", compact ? "h-1" : "h-1.5")}>
        <div
          className={cn(
            "h-full rounded-full transition-all duration-700 ease-out",
            barColor,
            state === "active" && pulse && "animate-pulse"
          )}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}
