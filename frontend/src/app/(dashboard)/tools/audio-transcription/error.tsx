"use client";

import { AlertCircle } from "lucide-react";
import { useTranslations } from "next-intl";

export default function AudioTranscriptionError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  const t = useTranslations("audioTranscription");

  return (
    <div className="space-y-4 rounded-xl border border-border bg-bg p-6">
      <div className="flex items-start gap-2 text-sm text-[#a32d2d]">
        <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
        <div>
          <p className="font-medium">{t("loadFailed")}</p>
          <p className="mt-1 text-text2">{error.message}</p>
        </div>
      </div>
      <button
        type="button"
        onClick={reset}
        className="rounded-lg bg-text px-4 py-2 text-sm text-white"
      >
        Retry
      </button>
    </div>
  );
}
