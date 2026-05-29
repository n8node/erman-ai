import { getTranslations } from "next-intl/server";
import { DashboardToolCards } from "@/components/dashboard/DashboardToolCards";
import { RecentCalculatorRuns } from "@/components/dashboard/RecentCalculatorRuns";
import { GuestBanner } from "@/components/layout/GuestBanner";

export default async function DashboardPage() {
  const t = await getTranslations("dashboard");

  return (
    <div>
      <GuestBanner />
      <h1 className="text-base font-medium">{t("title")}</h1>
      <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>

      <DashboardToolCards />

      <RecentCalculatorRuns />
    </div>
  );
}
