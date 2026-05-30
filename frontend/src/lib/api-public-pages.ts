export type PublicPage = {
  id: string;
  slug: string;
  title: string;
  content_html: string;
  meta_description: string;
  is_published: boolean;
  sort_order: number;
  updated_at: string;
  created_at: string;
};

const serverApiBase = () =>
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/api\/v1\/?$/, "") ||
  process.env.API_INTERNAL_URL ||
  "http://backend:8080";

export async function fetchPublicPage(slug: string): Promise<PublicPage> {
  const res = await fetch(
    `${serverApiBase()}/api/v1/public/pages/${encodeURIComponent(slug)}`,
    { next: { revalidate: 60 } }
  );
  if (!res.ok) {
    throw new Error("page not found");
  }
  return res.json() as Promise<PublicPage>;
}

export async function fetchPublicPageSlugs(): Promise<string[]> {
  const res = await fetch(`${serverApiBase()}/api/v1/public/pages/slugs`, {
    next: { revalidate: 60 },
  });
  if (!res.ok) {
    return ["about", "privacy-policy", "terms", "shop-terms", "refund"];
  }
  const data = (await res.json()) as { slugs?: string[] };
  return data.slugs ?? [];
}
