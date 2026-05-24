import { redirect } from "next/navigation";
import { AdminStrategyLLMEditor } from "@/components/admin/AdminStrategyLLMEditor";
import { getMe } from "@/lib/auth-server";

export default async function AdminStrategyLLMPage() {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");

  return <AdminStrategyLLMEditor />;
}
