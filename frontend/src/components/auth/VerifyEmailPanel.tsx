"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { ApiError, resendVerification, verifyEmail } from "@/lib/api";

export function VerifyEmailPanel() {
  const t = useTranslations("auth.verify");
  const router = useRouter();
  const searchParams = useSearchParams();
  const emailParam = searchParams.get("email") || "";
  const tokenParam = searchParams.get("token") || "";

  const [email, setEmail] = useState(emailParam);
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");
  const [loading, setLoading] = useState(false);
  const [verifying, setVerifying] = useState(!!tokenParam);

  useEffect(() => {
    if (!tokenParam) return;
    let cancelled = false;
    (async () => {
      setVerifying(true);
      setError("");
      try {
        const user = await verifyEmail(tokenParam);
        if (cancelled) return;
        router.replace(user.onboarding_completed ? "/" : "/onboarding");
        router.refresh();
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : t("verifyFailed"));
        }
      } finally {
        if (!cancelled) setVerifying(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [tokenParam, router, t]);

  async function handleResend(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setInfo("");
    if (!email.trim()) return;
    setLoading(true);
    try {
      await resendVerification(email.trim());
      setInfo(t("resent"));
    } catch (err) {
      if (err instanceof ApiError && err.status === 429) {
        setError(t("tooSoon"));
      } else {
        setError(err instanceof Error ? err.message : t("resendFailed"));
      }
    } finally {
      setLoading(false);
    }
  }

  if (verifying) {
    return (
      <div className="text-center text-sm text-text2">
        {t("confirming")}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {tokenParam && error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}
      {!tokenParam && (
        <p className="text-sm text-text2">
          {emailParam ? t("sentTo", { email: emailParam }) : t("checkInbox")}
        </p>
      )}
      {info && (
        <div className="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-900">
          {info}
        </div>
      )}
      {!tokenParam && error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}
      <form onSubmit={handleResend} className="space-y-3">
        <div>
          <label htmlFor="resend-email" className="mb-1.5 block text-xs font-medium">
            {t("emailLabel")}
          </label>
          <input
            id="resend-email"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
          />
        </div>
        <button
          type="submit"
          disabled={loading}
          className="w-full rounded-lg border border-border2 bg-bg px-4 py-2 text-sm font-medium hover:bg-bg2 disabled:opacity-60"
        >
          {loading ? t("loading") : t("resend")}
        </button>
      </form>
      <p className="text-center text-sm text-text2">
        <Link href="/login" className="text-accent hover:underline">
          {t("backToLogin")}
        </Link>
      </p>
    </div>
  );
}
