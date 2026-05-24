import { SettingsForm } from "@/components/settings/SettingsForm";
import { getMe } from "@/lib/auth-server";
import { redirect } from "next/navigation";

export default async function SettingsPage() {
  const user = await getMe();
  if (!user) redirect("/login");

  return <SettingsForm user={user} />;
}
