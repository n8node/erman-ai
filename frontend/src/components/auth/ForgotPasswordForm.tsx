"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useState } from "react";
import { useTranslations } from "next-intl";
import { ApiError, requestPasswordReset } from "@/lib/api";

export function ForgotPasswordForm() {
  const t = useTranslations("auth.forgot");
  const searchParams = useSearchParams();
  const emailParam = searchParams.get("email") || "";

  const [email, setEmail] = useState(emailParam);
  const [error, setError] = useState("");
  const [sent, setSent] = useState(false);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!email.trim()) return;
    setLoading(true);
    try {
      await requestPasswordReset(email.trim());
      setSent(true);
    } catch (err) {
      if (err instanceof ApiError && err.status === 429) {
        setError(t("tooSoon"));
      } else {
        setError(err instanceof Error ? err.message : t("failed"));
      }
    } finally {
      setLoading(false);
    }
  }

  if (sent) {
    return (
      <div className="space-y-4">
        <div className="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-900">
          {t("sent")}
        </div>
        <p className="text-center text-sm text-text2">
          <Link href="/login" className="text-accent hover:underline">
            {t("backToLogin")}
          </Link>
        </p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}
      <div>
        <label htmlFor="forgot-email" className="mb-1.5 block text-xs font-medium">
          {t("emailLabel")}
        </label>
        <input
          id="forgot-email"
          type="email"
          autoComplete="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
        />
      </div>
      <button
        type="submit"
        disabled={loading}
        className="w-full rounded-lg bg-text px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-60"
      >
        {loading ? t("loading") : t("submit")}
      </button>
      <p className="text-center text-sm text-text2">
        <Link href="/login" className="text-accent hover:underline">
          {t("backToLogin")}
        </Link>
      </p>
    </form>
  );
}
