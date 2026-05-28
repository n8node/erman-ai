import { redirect } from "next/navigation";
import { AdminProjectsEditor } from "@/components/admin/AdminProjectsEditor";
import { getMe } from "@/lib/auth-server";

export default async function AdminProjectsPage() {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");

  return <AdminProjectsEditor />;
}
