import { redirect } from "next/navigation";
import { AdminProposalRequestsEditor } from "@/components/admin/AdminProposalRequestsEditor";
import { getMe } from "@/lib/auth-server";

export default async function AdminProposalRequestsPage() {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");

  return <AdminProposalRequestsEditor />;
}
