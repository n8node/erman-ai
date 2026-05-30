import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { ForgotPasswordForm } from "@/components/auth/ForgotPasswordForm";
import { LanguageSwitcher } from "@/components/layout/LanguageSwitcher";

export default async function ForgotPasswordPage() {
  const t = await getTranslations("auth.forgot");

  return (
    <main className="relative flex min-h-screen items-center justify-center bg-bg3 p-6">
      <div className="absolute right-6 top-6">
        <LanguageSwitcher />
      </div>
      <div className="w-full max-w-md rounded-xl border border-border bg-bg p-8">
        <div className="mb-6 flex items-center gap-2.5">
          <div className="flex h-7 w-7 items-center justify-center rounded-md bg-text text-sm font-semibold text-white">
            E
          </div>
          <span className="text-sm font-medium">Erman AI</span>
        </div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 mb-6 text-sm text-text2">{t("subtitle")}</p>
        <Suspense fallback={null}>
          <ForgotPasswordForm />
        </Suspense>
      </div>
    </main>
  );
}
