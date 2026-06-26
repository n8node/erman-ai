import { redirect } from "next/navigation";
import { AdminConsultationsEditor } from "@/components/admin/AdminConsultationsEditor";
import { getMe } from "@/lib/auth-server";

export default async function AdminConsultationsPage() {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");

  return <AdminConsultationsEditor />;
}
