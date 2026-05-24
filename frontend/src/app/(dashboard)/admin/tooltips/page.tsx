import { redirect } from "next/navigation";
import { AdminTooltipsEditor } from "@/components/admin/AdminTooltipsEditor";
import { getMe } from "@/lib/auth-server";

export default async function AdminTooltipsPage() {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");

  return <AdminTooltipsEditor />;
}
