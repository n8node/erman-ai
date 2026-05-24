import { redirect } from "next/navigation";
import { AdminCalculatorBudgetEditor } from "@/components/admin/AdminCalculatorBudgetEditor";
import { getMe } from "@/lib/auth-server";

export default async function AdminCalculatorBudgetPage() {
  const user = await getMe();
  if (!user) redirect("/login");
  if (user.role !== "superadmin") redirect("/");

  return <AdminCalculatorBudgetEditor />;
}
