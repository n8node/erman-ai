"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { ApiError, login } from "@/lib/api";
import { sanitizeReturnPath } from "@/lib/return-url";

export function LoginForm() {
  const t = useTranslations("auth");
  const router = useRouter();
  const searchParams = useSearchParams();
  const nextPath = sanitizeReturnPath(searchParams.get("next"));
  const emailParam = searchParams.get("email") || "";
  const resetOk = searchParams.get("reset") === "ok";
  const [email, setEmail] = useState(emailParam);
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (resetOk) {
      setSuccess(t("resetPasswordSuccess"));
    }
  }, [resetOk, t]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const user = await login(email, password);
      if (!user.onboarding_completed) {
        router.push(`/onboarding?next=${encodeURIComponent(nextPath)}`);
      } else {
        router.push(nextPath);
      }
      router.refresh();
    } catch (err) {
      if (err instanceof ApiError && err.code === "email_not_verified") {
        setError(t("emailNotVerified"));
        return;
      }
      setError(err instanceof Error ? err.message : t("loginFailed"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {success && (
        <div className="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-900">
          {success}
        </div>
      )}
      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
          {error === t("emailNotVerified") && email.trim() && (
            <p className="mt-2">
              <Link
                href={`/verify-email?email=${encodeURIComponent(email.trim())}&next=${encodeURIComponent(nextPath)}`}
                className="text-accent hover:underline"
              >
                {t("verify.resendLink")}
              </Link>
            </p>
          )}
        </div>
      )}
      <div>
        <label htmlFor="email" className="mb-1.5 block text-xs font-medium">
          {t("email")}
        </label>
        <input
          id="email"
          type="email"
          autoComplete="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
        />
      </div>
      <div>
        <label htmlFor="password" className="mb-1.5 block text-xs font-medium">
          {t("password")}
        </label>
        <input
          id="password"
          type="password"
          autoComplete="current-password"
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
        />
        <div className="mt-2 text-right">
          <Link
            href={
              email.trim()
                ? `/forgot-password?email=${encodeURIComponent(email.trim())}`
                : "/forgot-password"
            }
            className="text-xs text-accent hover:underline"
          >
            {t("forgotPassword")}
          </Link>
        </div>
      </div>
      <button
        type="submit"
        disabled={loading}
        className="w-full rounded-lg bg-text px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-60"
      >
        {loading ? t("loading") : t("login")}
      </button>
      <p className="text-center text-sm text-text2">
        {t("noAccount")}{" "}
        <Link href={`/register?next=${encodeURIComponent(nextPath)}${email.trim() ? `&email=${encodeURIComponent(email.trim())}` : ""}`} className="text-accent hover:underline">
          {t("register")}
        </Link>
      </p>
    </form>
  );
}
