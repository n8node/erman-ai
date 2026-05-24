import { getTranslations } from "next-intl/server";

export default async function BillingPage() {
  const t = await getTranslations("stub");

  return (
    <div className="rounded-xl border border-border bg-bg p-6">
      <h1 className="text-base font-medium">Billing</h1>
      <p className="mt-2 text-sm text-text2">{t("comingSoon")}</p>
    </div>
  );
}
