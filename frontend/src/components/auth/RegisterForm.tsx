"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { register } from "@/lib/api";
import { checkPasswordRules, isPasswordValid } from "@/lib/password-policy";
import { PasswordField } from "@/components/auth/PasswordField";

export function RegisterForm() {
  const t = useTranslations("auth");
  const tPolicy = useTranslations("auth.passwordPolicy");
  const router = useRouter();
  const searchParams = useSearchParams();
  const referral = searchParams.get("ref") || undefined;
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const rules = useMemo(() => checkPasswordRules(password), [password]);
  const passwordOk = isPasswordValid(rules);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!passwordOk) {
      setError(tPolicy("invalid"));
      return;
    }
    if (password !== confirm) {
      setError(t("passwordMismatch"));
      return;
    }
    setLoading(true);
    try {
      const result = await register(email, password, referral);
      router.push(`/verify-email?email=${encodeURIComponent(result.email)}`);
      router.refresh();
    } catch (err) {
      const msg = err instanceof Error ? err.message : t("registerFailed");
      if (msg === "invalid input") {
        setError(tPolicy("invalid"));
      } else {
        setError(msg);
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {referral === "erman" && (
        <div className="rounded-md border border-accent bg-accent-bg px-3 py-2 text-sm text-accent">
          {t("ermanReferral")}
        </div>
      )}
      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
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
      <PasswordField
        id="password"
        label={t("password")}
        value={password}
        onChange={setPassword}
        autoComplete="new-password"
        showStrength
        showRequirements
      />
      <PasswordField
        id="confirm"
        label={t("confirmPassword")}
        value={confirm}
        onChange={setConfirm}
        autoComplete="new-password"
      />
      <button
        type="submit"
        disabled={loading || !passwordOk}
        className="w-full rounded-lg bg-text px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-60"
      >
        {loading ? t("loading") : t("register")}
      </button>
      <p className="text-center text-sm text-text2">
        {t("hasAccount")}{" "}
        <Link href="/login" className="text-accent hover:underline">
          {t("login")}
        </Link>
      </p>
    </form>
  );
}
