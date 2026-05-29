import { redirect } from "next/navigation";
import { AdminProjectInquiryDetailView } from "@/components/admin/AdminProjectInquiryDetailView";
import { getMe } from "@/lib/auth-server";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function AdminProjectInquiryDetailPage({ params }: Props) {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");

  const { id } = await params;
  return <AdminProjectInquiryDetailView id={id} />;
}
