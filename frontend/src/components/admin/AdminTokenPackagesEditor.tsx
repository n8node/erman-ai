"use client";

import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  createAdminTokenPackage,
  fetchAdminTokenPackages,
  updateAdminTokenPackage,
  type TokenPackage,
  type TokenPackageUpsertInput,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

type FormState = {
  slug: string;
  name: string;
  tokens: number;
  price_rub: number;
  sort_order: number;
  is_public: boolean;
  is_archived: boolean;
};

function emptyForm(): FormState {
  return {
    slug: "",
    name: "",
    tokens: 100000,
    price_rub: 490,
    sort_order: 0,
    is_public: true,
    is_archived: false,
  };
}

function pkgToForm(pkg: TokenPackage): FormState {
  return {
    slug: pkg.slug,
    name: pkg.name,
    tokens: pkg.tokens,
    price_rub: pkg.price_rub,
    sort_order: pkg.sort_order,
    is_public: pkg.is_public,
    is_archived: pkg.is_archived,
  };
}

export function AdminTokenPackagesEditor() {
  const t = useTranslations("admin.tokenPackages");
  const [items, setItems] = useState<TokenPackage[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm());
  const [creating, setCreating] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await fetchAdminTokenPackages();
      setItems(res.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    void load();
  }, [load]);

  function selectPackage(pkg: TokenPackage) {
    setCreating(false);
    setSelectedId(pkg.id);
    setForm(pkgToForm(pkg));
    setSuccess("");
    setError("");
  }

  function startCreate() {
    setCreating(true);
    setSelectedId(null);
    setForm(emptyForm());
    setSuccess("");
    setError("");
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess("");
    const payload: TokenPackageUpsertInput = {
      slug: form.slug.trim(),
      name: form.name.trim(),
      tokens: form.tokens,
      price_rub: form.price_rub,
      sort_order: form.sort_order,
      is_public: form.is_public,
      is_archived: form.is_archived,
    };
    try {
      if (creating) {
        const created = await createAdminTokenPackage(payload);
        setCreating(false);
        setSelectedId(created.id);
        setSuccess(t("created"));
      } else if (selectedId) {
        await updateAdminTokenPackage(selectedId, payload);
        setSuccess(t("saved"));
      }
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return <p className="text-sm text-text2">{t("loading")}</p>;
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[240px_minmax(0,1fr)]">
      <div className="space-y-3">
        <button
          type="button"
          onClick={startCreate}
          className="w-full rounded-lg border border-border2 px-3 py-2 text-sm hover:bg-bg2"
        >
          {t("create")}
        </button>
        <ul className="space-y-1">
          {items.map((pkg) => (
            <li key={pkg.id}>
              <button
                type="button"
                onClick={() => selectPackage(pkg)}
                className={cn(
                  "w-full rounded-lg px-3 py-2 text-left text-sm",
                  selectedId === pkg.id && !creating
                    ? "bg-bg2 font-medium text-text"
                    : "text-text2 hover:bg-bg2"
                )}
              >
                {pkg.name}
              </button>
            </li>
          ))}
        </ul>
      </div>

      <div className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <div>
          <h2 className="text-sm font-medium">
            {creating ? t("createTitle") : t("editTitle")}
          </h2>
          <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
        </div>

        {error && (
          <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
            {error}
          </div>
        )}
        {success && (
          <div className="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-900">
            {success}
          </div>
        )}

        {(creating || selectedId) && (
          <>
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <label className="mb-1.5 block text-xs font-medium">{t("slug")}</label>
                <input
                  type="text"
                  value={form.slug}
                  onChange={(e) => setForm((p) => ({ ...p, slug: e.target.value }))}
                  className={fieldClass}
                  disabled={!creating}
                />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium">{t("name")}</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))}
                  className={fieldClass}
                />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium">{t("tokens")}</label>
                <input
                  type="number"
                  min={1}
                  value={form.tokens}
                  onChange={(e) => setForm((p) => ({ ...p, tokens: Number(e.target.value) }))}
                  className={fieldClass}
                />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium">{t("priceRub")}</label>
                <input
                  type="number"
                  min={1}
                  value={form.price_rub}
                  onChange={(e) => setForm((p) => ({ ...p, price_rub: Number(e.target.value) }))}
                  className={fieldClass}
                />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium">{t("sortOrder")}</label>
                <input
                  type="number"
                  value={form.sort_order}
                  onChange={(e) => setForm((p) => ({ ...p, sort_order: Number(e.target.value) }))}
                  className={fieldClass}
                />
              </div>
            </div>

            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={form.is_public}
                onChange={(e) => setForm((p) => ({ ...p, is_public: e.target.checked }))}
                className="rounded border-border2"
              />
              {t("isPublic")}
            </label>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={form.is_archived}
                onChange={(e) => setForm((p) => ({ ...p, is_archived: e.target.checked }))}
                className="rounded border-border2"
              />
              {t("isArchived")}
            </label>

            <button
              type="button"
              disabled={saving}
              onClick={handleSave}
              className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-60"
            >
              {saving ? t("saving") : t("save")}
            </button>
          </>
        )}
      </div>
    </div>
  );
}
