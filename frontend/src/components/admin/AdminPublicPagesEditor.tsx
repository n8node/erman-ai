"use client";

import { ExternalLink, Plus, Trash2 } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  createAdminPublicPage,
  deleteAdminPublicPage,
  fetchAdminPublicPages,
  updateAdminPublicPage,
  type PublicPage,
  type PublicPageInput,
} from "@/lib/api";
import { cn } from "@/lib/utils";
import { RichTextEditor } from "@/components/admin/RichTextEditor";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const emptyDraft = (): PublicPageInput => ({
  slug: "",
  title: "",
  content_html: "",
  meta_description: "",
  is_published: true,
  sort_order: 0,
});

export function AdminPublicPagesEditor() {
  const t = useTranslations("admin.publicPages");
  const [items, setItems] = useState<PublicPage[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [draft, setDraft] = useState<PublicPageInput>(emptyDraft());
  const [isNew, setIsNew] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  const load = useCallback(async () => {
    const data = await fetchAdminPublicPages();
    setItems(data.items);
    return data.items;
  }, []);

  useEffect(() => {
    load()
      .then((list) => {
        if (list.length > 0) {
          const first = list[0];
          setSelectedId(first.id);
          setDraft(pageToInput(first));
        }
      })
      .catch((err) =>
        setError(err instanceof Error ? err.message : t("loadFailed"))
      )
      .finally(() => setLoading(false));
  }, [load, t]);

  function pageToInput(page: PublicPage): PublicPageInput {
    return {
      slug: page.slug,
      title: page.title,
      content_html: page.content_html,
      meta_description: page.meta_description,
      is_published: page.is_published,
      sort_order: page.sort_order,
    };
  }

  function selectPage(page: PublicPage) {
    setSelectedId(page.id);
    setDraft(pageToInput(page));
    setIsNew(false);
    setError("");
    setSuccess(false);
  }

  function startNew() {
    setSelectedId(null);
    setDraft(emptyDraft());
    setIsNew(true);
    setError("");
    setSuccess(false);
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess(false);
    try {
      if (isNew) {
        const created = await createAdminPublicPage(draft);
        const list = await load();
        setSelectedId(created.id);
        setIsNew(false);
        setDraft(pageToInput(created));
        if (!list.find((p) => p.id === created.id)) {
          setItems((prev) => [...prev, created]);
        }
      } else if (selectedId) {
        const updated = await updateAdminPublicPage(selectedId, draft);
        setItems((prev) =>
          prev.map((p) => (p.id === updated.id ? updated : p))
        );
        setDraft(pageToInput(updated));
      }
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    if (!selectedId || isNew) return;
    if (!window.confirm(t("deleteConfirm"))) return;
    setSaving(true);
    setError("");
    try {
      await deleteAdminPublicPage(selectedId);
      const list = await load();
      if (list.length > 0) {
        selectPage(list[0]);
      } else {
        startNew();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t("deleteFailed"));
    } finally {
      setSaving(false);
    }
  }

  const publicUrl = draft.slug
    ? `/dashboard/${draft.slug.replace(/^\//, "")}`
    : "";

  if (loading) {
    return <p className="text-sm text-text2">{t("loading")}</p>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </div>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}
      {success && (
        <div className="rounded-md border border-green-200 bg-success-bg px-3 py-2 text-sm text-success">
          {t("saved")}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-[240px_1fr]">
        <aside className="space-y-2">
          <button
            type="button"
            onClick={startNew}
            className="flex w-full items-center gap-2 rounded-lg border border-border2 px-3 py-2 text-sm hover:bg-bg2"
          >
            <Plus className="h-4 w-4" />
            {t("newPage")}
          </button>
          <div className="space-y-1">
            {items.map((page) => (
              <button
                key={page.id}
                type="button"
                onClick={() => selectPage(page)}
                className={cn(
                  "w-full rounded-lg px-3 py-2 text-left text-sm",
                  selectedId === page.id && !isNew
                    ? "bg-bg2 font-medium text-text"
                    : "text-text2 hover:bg-bg2 hover:text-text"
                )}
              >
                {page.title}
                {!page.is_published && (
                  <span className="ml-1 text-[10px] text-text3">({t("draft")})</span>
                )}
              </button>
            ))}
          </div>
        </aside>

        <div className="space-y-4 rounded-xl border border-border bg-bg p-5">
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-xs font-medium text-text2">
                {t("titleLabel")}
              </label>
              <input
                className={fieldClass}
                value={draft.title}
                onChange={(e) =>
                  setDraft((d) => ({ ...d, title: e.target.value }))
                }
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-text2">
                {t("slugLabel")}
              </label>
              <input
                className={fieldClass}
                value={draft.slug}
                placeholder="privacy-policy"
                onChange={(e) =>
                  setDraft((d) => ({ ...d, slug: e.target.value }))
                }
              />
              {publicUrl && (
                <a
                  href={publicUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="mt-1 inline-flex items-center gap-1 text-xs text-accent hover:underline"
                >
                  {t("publicUrl")}: {publicUrl}
                  <ExternalLink className="h-3 w-3" />
                </a>
              )}
            </div>
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-text2">
              {t("metaLabel")}
            </label>
            <input
              className={fieldClass}
              value={draft.meta_description}
              onChange={(e) =>
                setDraft((d) => ({ ...d, meta_description: e.target.value }))
              }
            />
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-text2">
              {t("contentLabel")}
            </label>
            <p className="mb-2 text-xs text-text3">{t("contentHint")}</p>
            <RichTextEditor
              key={isNew ? "new-page" : selectedId ?? "empty"}
              editorKey={isNew ? "new-page" : selectedId ?? "empty"}
              value={draft.content_html}
              onChange={(content_html) =>
                setDraft((d) => ({ ...d, content_html }))
              }
            />
          </div>

          <div className="flex flex-wrap items-center gap-4">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={draft.is_published}
                onChange={(e) =>
                  setDraft((d) => ({ ...d, is_published: e.target.checked }))
                }
              />
              {t("published")}
            </label>
            <div className="flex items-center gap-2">
              <label className="text-xs text-text2">{t("sortOrder")}</label>
              <input
                type="number"
                className="w-20 rounded-lg border border-border2 px-2 py-1 text-sm"
                value={draft.sort_order}
                onChange={(e) =>
                  setDraft((d) => ({
                    ...d,
                    sort_order: Number(e.target.value) || 0,
                  }))
                }
              />
            </div>
          </div>

          <div className="flex gap-2 pt-2">
            <button
              type="button"
              onClick={handleSave}
              disabled={saving}
              className="rounded-lg bg-text px-4 py-2 text-sm text-white disabled:opacity-50"
            >
              {saving ? t("saving") : t("save")}
            </button>
            {!isNew && selectedId && (
              <button
                type="button"
                onClick={handleDelete}
                disabled={saving}
                className="inline-flex items-center gap-1 rounded-lg border border-red-200 px-4 py-2 text-sm text-red-800 hover:bg-red-50 disabled:opacity-50"
              >
                <Trash2 className="h-4 w-4" />
                {t("delete")}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
