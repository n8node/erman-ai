import type { PublicPage } from "@/lib/api-public-pages";

type Props = {
  page: PublicPage;
};

export function PublicPageView({ page }: Props) {
  return (
    <article className="rounded-xl border border-border bg-bg p-8">
      <p className="text-[10px] font-medium uppercase tracking-wider text-text3">
        Публичные страницы
      </p>
      <h1 className="mt-2 text-xl font-medium">{page.title}</h1>
      <div
        className="public-page-content mt-8"
        dangerouslySetInnerHTML={{ __html: page.content_html }}
      />
    </article>
  );
}
