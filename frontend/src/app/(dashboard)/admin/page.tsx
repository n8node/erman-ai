import { getTranslations } from "next-intl/server";
import { redirect } from "next/navigation";
import { getMe } from "@/lib/auth-server";

export default async function AdminPage() {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");

  const t = await getTranslations("admin");

  return (
    <div className="rounded-xl border border-border bg-bg p-6">
      <h1 className="text-base font-medium">{t("title")}</h1>
      <p className="mt-2 text-sm text-text2">{t("subtitle")}</p>
      <p className="mt-4 text-xs text-text3">{t("phaseNote")}</p>
    </div>
  );
}
