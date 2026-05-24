import Link from "next/link";
import { getTranslations } from "next-intl/server";

export default async function DashboardPage() {
  const t = await getTranslations("dashboard");

  const tools = [
    { slug: "calculator", href: "/tools/calculator" },
    { slug: "strategy", href: "/tools/strategy" },
    { slug: "proposal", href: "/tools/proposal" },
  ] as const;

  return (
    <div>
      <h1 className="text-base font-medium">{t("title")}</h1>
      <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>

      <div className="mt-8 grid gap-4 sm:grid-cols-3">
        {tools.map(({ slug, href }) => (
          <Link
            key={slug}
            href={href}
            className="rounded-xl border border-border bg-bg p-5 hover:border-border2"
          >
            <h2 className="text-sm font-medium">{t(`tools.${slug}.name`)}</h2>
            <p className="mt-1 text-xs text-text2">
              {t(`tools.${slug}.description`)}
            </p>
          </Link>
        ))}
      </div>
    </div>
  );
}
