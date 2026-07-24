import { ExternalLink } from "lucide-react";
import { getTranslations } from "next-intl/server";
import { redirect } from "next/navigation";
import { getMe } from "@/lib/auth-server";
import {
  WORKSPACE_SSO_PATH,
  WorkspaceFrame,
} from "@/components/tools/WorkspaceFrame";

export default async function WorkspacePage() {
  const [user, t] = await Promise.all([
    getMe(),
    getTranslations("workspace"),
  ]);

  if (!user) {
    redirect("/login?next=/tools/workspace");
  }
  if (user.role !== "superadmin") {
    redirect("/");
  }

  return (
    <section className="-mx-4 -my-5 flex h-[calc(100dvh-52px)] min-h-[560px] flex-col bg-bg sm:-mx-6 sm:-my-6">
      <header className="flex min-h-12 items-center justify-between gap-3 border-b border-border px-4 sm:px-5">
        <div className="min-w-0">
          <h1 className="truncate text-sm font-medium text-text">{t("title")}</h1>
          <p className="hidden truncate text-xs text-text3 sm:block">
            {t("subtitle")}
          </p>
        </div>
        <a
          href={WORKSPACE_SSO_PATH}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex shrink-0 items-center gap-1.5 rounded-md border border-border2 px-2.5 py-1.5 text-xs text-text hover:bg-bg2"
        >
          <ExternalLink size={14} aria-hidden="true" />
          {t("openNewTab")}
        </a>
      </header>

      <WorkspaceFrame title={t("frameTitle")} />
    </section>
  );
}
