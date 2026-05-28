import { redirect } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { AdminPaymentEditor } from "@/components/admin/AdminPaymentEditor";
import { getMe } from "@/lib/auth-server";

export default async function AdminPaymentsPage() {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");
  const t = await getTranslations("admin.payments");

  return (
    <div className="max-w-3xl space-y-6">
      <div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </div>
      <AdminPaymentEditor />
    </div>
  );
}
