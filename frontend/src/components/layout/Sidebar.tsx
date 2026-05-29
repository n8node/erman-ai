"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { useCallback, useEffect, useState } from "react";
import {
  Calculator,
  Brain,
  FileText,
  ClipboardList,
  CreditCard,
  Settings,
  Shield,
  ExternalLink,
  MessageCircle,
} from "lucide-react";
import {
  fetchAdminExternalProjects,
  fetchExternalProjects,
  type ExternalProject,
} from "@/lib/api";
import type { User } from "@/lib/api";
import { cn } from "@/lib/utils";

type Props = {
  user: User | null;
};

export function Sidebar({ user }: Props) {
  const t = useTranslations("nav");
  const pathname = usePathname();
  const [projects, setProjects] = useState<ExternalProject[]>([]);
  const isSuperAdmin = user?.role === "superadmin";

  const loadProjects = useCallback(() => {
    if (!user) {
      setProjects([]);
      return;
    }
    const fetcher = isSuperAdmin
      ? fetchAdminExternalProjects
      : fetchExternalProjects;
    fetcher()
      .then((data) => setProjects(data.items))
      .catch(() => setProjects([]));
  }, [isSuperAdmin, user]);

  useEffect(() => {
    loadProjects();
  }, [loadProjects, pathname]);

  useEffect(() => {
    const onProjectsUpdated = () => loadProjects();
    window.addEventListener("external-projects-updated", onProjectsUpdated);
    return () =>
      window.removeEventListener("external-projects-updated", onProjectsUpdated);
  }, [loadProjects]);

  const toolLinks = [
    { href: "/tools/calculator", label: t("calculator"), icon: Calculator },
    { href: "/tools/strategy", label: t("strategy"), icon: Brain },
    { href: "/tools/proposal", label: t("proposal"), icon: FileText },
    { href: "/tools/audit", label: t("audit"), icon: ClipboardList },
  ];

  const discussLink = { href: "/discuss", label: t("discussProject"), icon: MessageCircle };

  const accountLinks = [
    { href: "/billing", label: t("billing"), icon: CreditCard },
    { href: "/settings", label: t("settings"), icon: Settings },
  ];

  const isActive = (href: string) =>
    pathname === href || pathname.startsWith(`${href}/`);

  const sidebarProjects = isSuperAdmin
    ? projects
    : projects.filter((project) => project.is_enabled);

  return (
    <aside className="fixed left-0 top-0 z-30 flex h-screen w-[220px] flex-col border-r border-border bg-bg">
      <div className="flex items-center gap-2.5 px-5 py-5">
        <div className="flex h-7 w-7 items-center justify-center rounded-md bg-text text-sm font-semibold text-white">
          E
        </div>
        <span className="text-sm font-medium">Erman AI</span>
      </div>
      <div className="mx-5 border-b border-border" />

      <nav className="flex-1 overflow-y-auto px-3 py-4">
        <p className="mb-2 px-2 text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("tools")}
        </p>
        <ul className="space-y-0.5">
          {toolLinks.map(({ href, label, icon: Icon }) => (
            <li key={href}>
              <Link
                href={href}
                className={cn(
                  "flex items-center gap-2 rounded-md px-2 py-2 text-[13px] text-text2 hover:bg-bg2",
                  isActive(href) && "bg-bg2 font-medium text-text"
                )}
              >
                <Icon size={16} />
                {label}
              </Link>
            </li>
          ))}
        </ul>

        <p className="mb-2 mt-6 px-2 text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("project")}
        </p>
        <ul className="space-y-0.5">
          <li>
            <Link
              href={discussLink.href}
              className={cn(
                "flex items-center gap-2 rounded-md px-2 py-2 text-[13px] text-text2 hover:bg-bg2",
                isActive(discussLink.href) && "bg-bg2 font-medium text-text"
              )}
            >
              <discussLink.icon size={16} />
              {discussLink.label}
            </Link>
          </li>
        </ul>

        {sidebarProjects.length > 0 && (
          <>
            <p className="mb-2 mt-6 px-2 text-[10px] font-medium uppercase tracking-wider text-text3">
              {t("projects")}
            </p>
            <ul className="space-y-0.5">
              {sidebarProjects.map((project) => (
                <li key={project.id}>
                  <a
                    href={project.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    title={
                      isSuperAdmin && !project.is_enabled
                        ? t("projectHidden")
                        : undefined
                    }
                    className={cn(
                      "flex items-center gap-2 rounded-md px-2 py-2 text-[13px] text-text2 hover:bg-bg2",
                      isSuperAdmin && !project.is_enabled && "opacity-50"
                    )}
                  >
                    <ExternalLink size={16} className="shrink-0" />
                    <span className="truncate">{project.title}</span>
                  </a>
                </li>
              ))}
            </ul>
          </>
        )}

        {user && (
          <>
            <p className="mb-2 mt-6 px-2 text-[10px] font-medium uppercase tracking-wider text-text3">
              {t("account")}
            </p>
            <ul className="space-y-0.5">
              {accountLinks.map(({ href, label, icon: Icon }) => (
                <li key={href}>
                  <Link
                    href={href}
                    className={cn(
                      "flex items-center gap-2 rounded-md px-2 py-2 text-[13px] text-text2 hover:bg-bg2",
                      isActive(href) && "bg-bg2 font-medium text-text"
                    )}
                  >
                    <Icon size={16} />
                    {label}
                  </Link>
                </li>
              ))}
              {user.role === "superadmin" && (
                <li>
                  <Link
                    href="/admin"
                    className={cn(
                      "flex items-center gap-2 rounded-md px-2 py-2 text-[13px] text-text2 hover:bg-bg2",
                      isActive("/admin") && "bg-bg2 font-medium text-text"
                    )}
                  >
                    <Shield size={16} />
                    {t("admin")}
                  </Link>
                </li>
              )}
            </ul>
          </>
        )}
      </nav>
    </aside>
  );
}
