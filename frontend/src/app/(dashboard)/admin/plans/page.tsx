import { redirect } from "next/navigation";
import { AdminPlansEditor } from "@/components/admin/AdminPlansEditor";
import { getMe } from "@/lib/auth-server";

export default async function AdminPlansPage() {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");

  return <AdminPlansEditor />;
}
