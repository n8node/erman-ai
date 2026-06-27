import { getTranslations } from "next-intl/server";
import { redirect } from "next/navigation";
import { OnboardingForm } from "@/components/onboarding/OnboardingForm";
import { getMe } from "@/lib/auth-server";

export default async function OnboardingPage({
  searchParams,
}: {
  searchParams?: Promise<{ next?: string }>;
}) {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.onboarding_completed) redirect("/");

  const params = await searchParams;
  const t = await getTranslations("onboarding");

  return (
    <main className="flex min-h-screen items-center justify-center bg-bg3 px-4 py-8 sm:p-6">
      <div className="w-full max-w-lg rounded-xl border border-border bg-bg p-6 sm:p-8">
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 mb-6 text-sm text-text2">{t("subtitle")}</p>
        <OnboardingForm next={params?.next} />
      </div>
    </main>
  );
}
