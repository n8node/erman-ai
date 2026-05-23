import { getTranslations } from "next-intl/server";

export default async function DashboardPage() {
  const t = await getTranslations("app");

  return (
    <main className="min-h-screen flex flex-col items-center justify-center p-8">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-7 h-7 rounded-md bg-text flex items-center justify-center text-white text-sm font-semibold">
          E
        </div>
        <span className="text-base font-medium">{t("title")}</span>
      </div>
      <h1 className="text-2xl font-medium text-text mb-2">{t("dashboard")}</h1>
      <p className="text-text2 text-sm">{t("tagline")}</p>
    </main>
  );
}
