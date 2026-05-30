import { notFound } from "next/navigation";
import type { Metadata } from "next";
import { PublicPageView } from "@/components/public/PublicPageView";
import { fetchPublicPage } from "@/lib/api-public-pages";

type Props = {
  params: Promise<{ slug: string }>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  try {
    const page = await fetchPublicPage(slug);
    return {
      title: `${page.title} — Erman AI`,
      description: page.meta_description || page.title,
    };
  } catch {
    return { title: "Erman AI" };
  }
}

export default async function PublicPage({ params }: Props) {
  const { slug } = await params;

  let page;
  try {
    page = await fetchPublicPage(slug);
  } catch {
    notFound();
  }

  return <PublicPageView page={page} />;
}
