"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { PasswordField, isPasswordValid, checkPasswordRules } from "@/components/auth/PasswordField";
import { resetPassword } from "@/lib/api";

export function ResetPasswordForm() {
  const t = useTranslations("auth.reset");
  const tAuth = useTranslations("auth");
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token") || "";

  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const rules = useMemo(() => checkPasswordRules(password), [password]);
  const passwordOk = isPasswordValid(rules);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!token) {
      setError(t("missingToken"));
      return;
    }
    if (!passwordOk) {
      setError(tAuth("passwordPolicy.invalid"));
      return;
    }
    if (password !== confirm) {
      setError(tAuth("passwordMismatch"));
      return;
    }
    setLoading(true);
    try {
      await resetPassword(token, password);
      router.push("/login?reset=ok");
    } catch (err) {
      setError(err instanceof Error ? err.message : t("failed"));
    } finally {
      setLoading(false);
    }
  }

  if (!token) {
    return (
      <div className="space-y-4">
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {t("missingToken")}
        </div>
        <p className="text-center text-sm text-text2">
          <Link href="/forgot-password" className="text-accent hover:underline">
            {t("requestNewLink")}
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
      <PasswordField
        id="new-password"
        label={t("newPassword")}
        value={password}
        onChange={setPassword}
        autoComplete="new-password"
        showStrength
        showRequirements
      />
      <PasswordField
        id="confirm-password"
        label={tAuth("confirmPassword")}
        value={confirm}
        onChange={setConfirm}
        autoComplete="new-password"
      />
      <button
        type="submit"
        disabled={loading || !passwordOk}
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
