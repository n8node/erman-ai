import { redirect } from "next/navigation";
import { AdminUsersEditor } from "@/components/admin/AdminUsersEditor";
import { getMe } from "@/lib/auth-server";

export default async function AdminUsersPage() {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");

  return <AdminUsersEditor />;
}
