import { getTranslations } from "next-intl/server";
import { ToolsOverviewGrid } from "@/components/dashboard/ToolsOverviewGrid";
import { GuestBanner } from "@/components/layout/GuestBanner";

export default async function ToolsPage() {
  const t = await getTranslations("toolsPage");

  return (
    <div>
      <GuestBanner />
      <h1 className="text-base font-medium">{t("title")}</h1>
      <p className="mt-1 max-w-2xl text-sm text-text2">{t("subtitle")}</p>
      <ToolsOverviewGrid />
    </div>
  );
}
