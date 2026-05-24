import { redirect } from "next/navigation";
import { AdminProposalRequestDetailView } from "@/components/admin/AdminProposalRequestDetailView";
import { getMe } from "@/lib/auth-server";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function AdminProposalRequestDetailPage({ params }: Props) {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");

  const { id } = await params;
  return <AdminProposalRequestDetailView id={id} />;
}
