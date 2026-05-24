import { getTranslations } from "next-intl/server";

type Props = { params: Promise<{ slug: string }> };

export default async function ToolStubPage({ params }: Props) {
  const { slug } = await params;
  const t = await getTranslations("stub");

  return (
    <div className="rounded-xl border border-border bg-bg p-6">
      <h1 className="text-base font-medium capitalize">{slug}</h1>
      <p className="mt-2 text-sm text-text2">{t("comingSoon")}</p>
    </div>
  );
}
