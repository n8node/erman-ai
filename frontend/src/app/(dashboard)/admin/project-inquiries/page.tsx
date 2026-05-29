import { redirect } from "next/navigation";
import { AdminProjectInquiriesEditor } from "@/components/admin/AdminProjectInquiriesEditor";
import { getMe } from "@/lib/auth-server";

export default async function AdminProjectInquiriesPage() {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");

  return <AdminProjectInquiriesEditor />;
}
