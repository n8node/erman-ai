"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { useTranslations } from "next-intl";
import { completeOnboarding } from "@/lib/api";

export function OnboardingForm() {
  const t = useTranslations("onboarding");
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function choose(segment: "partner" | "direct_lead") {
    setError("");
    setLoading(true);
    try {
      await completeOnboarding(segment);
      router.push(segment === "partner" ? "/tools/calculator" : "/tools/calculator");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("failed"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-4">
      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}
      <button
        type="button"
        disabled={loading}
        onClick={() => choose("partner")}
        className="w-full rounded-xl border border-border bg-bg p-5 text-left hover:border-border2 disabled:opacity-60"
      >
        <p className="text-sm font-medium">{t("partnerTitle")}</p>
        <p className="mt-1 text-xs text-text2">{t("partnerDesc")}</p>
      </button>
      <button
        type="button"
        disabled={loading}
        onClick={() => choose("direct_lead")}
        className="w-full rounded-xl border border-border bg-bg p-5 text-left hover:border-border2 disabled:opacity-60"
      >
        <p className="text-sm font-medium">{t("directTitle")}</p>
        <p className="mt-1 text-xs text-text2">{t("directDesc")}</p>
      </button>
    </div>
  );
}
