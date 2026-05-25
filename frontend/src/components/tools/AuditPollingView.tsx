"use client";

import { useEffect, useRef } from "react";
import { useTranslations } from "next-intl";
import { getRun } from "@/lib/api";

const POLL_MS = 3000;

type Props = {
  runId: string;
  onComplete: () => void;
  onError: (message: string) => void;
};

export function AuditPollingView({ runId, onComplete, onError }: Props) {
  const t = useTranslations("audit.polling");
  const completedRef = useRef(false);
  const onCompleteRef = useRef(onComplete);
  const onErrorRef = useRef(onError);

  useEffect(() => {
    onCompleteRef.current = onComplete;
    onErrorRef.current = onError;
  }, [onComplete, onError]);

  useEffect(() => {
    let cancelled = false;

    async function poll() {
      if (cancelled || completedRef.current) return;
      try {
        const run = await getRun(runId);
        if (run.status === "done") {
          completedRef.current = true;
          onCompleteRef.current();
        } else if (run.status === "error") {
          completedRef.current = true;
          onErrorRef.current(run.error_msg || t("failed"));
        }
      } catch {
        // keep polling
      }
    }

    void poll();
    const timer = setInterval(poll, POLL_MS);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [runId, t]);

  return (
    <div className="rounded-xl border border-border bg-bg p-8 text-center">
      <div className="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-border2 border-t-accent" />
      <p className="mt-4 text-sm font-medium text-text">{t("title")}</p>
      <p className="mt-1 text-xs text-text3">{t("hint")}</p>
    </div>
  );
}
