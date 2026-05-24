"use client";

import { useEffect, useRef } from "react";
import { useTranslations } from "next-intl";
import { getRun, type RunDetail } from "@/lib/api";
import type { StrategyOutput } from "@/lib/api-strategy";
import { Loader2 } from "lucide-react";

type Props = {
  runId: string;
  onComplete: (run: RunDetail & { output?: StrategyOutput }) => void;
  onError: (message: string) => void;
};

const POLL_MS = 3000;
const TERMINAL = new Set(["done", "error"]);

export function RunStatusPoller({ runId, onComplete, onError }: Props) {
  const t = useTranslations("strategy.poller");
  const onCompleteRef = useRef(onComplete);
  const onErrorRef = useRef(onError);

  useEffect(() => {
    onCompleteRef.current = onComplete;
    onErrorRef.current = onError;
  }, [onComplete, onError]);

  useEffect(() => {
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | null = null;

    async function poll() {
      try {
        const run = await getRun(runId);
        if (cancelled) return;

        if (run.status === "error") {
          onErrorRef.current(run.error_msg || t("failed"));
          return;
        }
        if (run.status === "done") {
          onCompleteRef.current(run as RunDetail & { output?: StrategyOutput });
          return;
        }

        timer = setTimeout(poll, POLL_MS);
      } catch (err) {
        if (!cancelled) {
          onErrorRef.current(err instanceof Error ? err.message : t("failed"));
        }
      }
    }

    poll();
    return () => {
      cancelled = true;
      if (timer) clearTimeout(timer);
    };
  }, [runId, t]);

  return (
    <div className="rounded-xl border border-border bg-bg p-8 text-center space-y-3">
      <Loader2 className="mx-auto h-8 w-8 animate-spin text-ai" />
      <p className="text-sm font-medium text-text">{t("title")}</p>
      <p className="text-xs text-text2">{t("hint")}</p>
    </div>
  );
}
