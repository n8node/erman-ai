import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";
import type { User } from "@/lib/api";

type Props = {
  user: User;
  children: React.ReactNode;
};

export function DashboardShell({ user, children }: Props) {
  return (
    <div className="min-h-screen bg-bg3">
      <Sidebar user={user} />
      <div className="ml-[220px]">
        <Topbar user={user} />
        <main className="px-6 py-6">{children}</main>
      </div>
    </div>
  );
}
