"use client";

import { useCallback, useState } from "react";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";
import type { User } from "@/lib/api";

type Props = {
  user: User | null;
  children: React.ReactNode;
};

export function DashboardShell({ user, children }: Props) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const closeSidebar = useCallback(() => setSidebarOpen(false), []);
  const openSidebar = useCallback(() => setSidebarOpen(true), []);

  return (
    <div className="min-h-screen bg-bg3 lg:pl-[220px]">
      <Sidebar
        user={user}
        isOpen={sidebarOpen}
        onClose={closeSidebar}
      />
      <div className="min-w-0">
        <Topbar user={user} onMenuClick={openSidebar} />
        <main className="min-w-0 px-4 py-5 sm:px-6 sm:py-6">{children}</main>
      </div>
    </div>
  );
}
