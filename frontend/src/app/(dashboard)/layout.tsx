import { redirect } from "next/navigation";
import { DashboardShell } from "@/components/layout/DashboardShell";
import { getMe } from "@/lib/auth-server";

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const user = await getMe();
  if (!user) {
    redirect("/login");
  }
  if (!user.email_verified) {
    redirect(`/verify-email?email=${encodeURIComponent(user.email)}`);
  }
  if (!user.onboarding_completed) {
    redirect("/onboarding");
  }

  return <DashboardShell user={user}>{children}</DashboardShell>;
}
