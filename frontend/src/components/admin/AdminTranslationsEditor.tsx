"use client";

import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import {
  bulkUpdateAdminTranslations,
  searchAdminTranslations,
  type AdminTranslation,
} from "@/lib/api";
import {
  LOCALE_META,
  SUPPORTED_LOCALES,
  type AppLocale,
} from "@/i18n/locales";
import {
  buildTranslationCatalog,
  TRANSLATION_NAMESPACES,
  type CatalogRow,
} from "@/lib/message-catalog";

const PAGE_SIZE = 40;

function mergeRows(catalog: CatalogRow[], overrides: AdminTranslation[]): CatalogRow[] {
  const byKey = new Map<string, CatalogRow>();
  for (const row of catalog) {
    byKey.set(row.key, {
      key: row.key,
      namespace: row.namespace,
      values: { ...row.values },
    });
  }
  for (const item of overrides) {
    const locale = item.locale as AppLocale;
    const row = byKey.get(item.key);
    if (row && SUPPORTED_LOCALES.includes(locale)) {
      row.values[locale] = item.value;
    }
  }
  return Array.from(byKey.values());
}

export function AdminTranslationsEditor() {
  const t = useTranslations("admin.translations");
  const catalog = useMemo(() => buildTranslationCatalog(), []);
  const [overrides, setOverrides] = useState<AdminTranslation[]>([]);
  const [rows, setRows] = useState<CatalogRow[]>(catalog);
  const [search, setSearch] = useState("");
  const [namespace, setNamespace] = useState("");
  const [page, setPage] = useState(0);
  const [editLocale, setEditLocale] = useState<AppLocale>("ru");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [dirty, setDirty] = useState<Record<string, string>>({});

  useEffect(() => {
    searchAdminTranslations({ search: "", limit: 5000 })
      .then((data) => {
        setOverrides(data.items);
        setRows(mergeRows(catalog, data.items));
      })
      .catch((err) => setError(err instanceof Error ? err.message : t("loadFailed")))
      .finally(() => setLoading(false));
  }, [catalog, t]);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return rows.filter((row) => {
      if (namespace && row.namespace !== namespace) return false;
      if (!q) return true;
      if (row.key.toLowerCase().includes(q)) return true;
      return SUPPORTED_LOCALES.some((locale) => row.values[locale].toLowerCase().includes(q));
    });
  }, [rows, search, namespace]);

  const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  const pageRows = filtered.slice(page * PAGE_SIZE, page * PAGE_SIZE + PAGE_SIZE);

  function valueFor(row: CatalogRow, locale: AppLocale) {
    const dirtyKey = `${row.key}::${locale}`;
    if (dirty[dirtyKey] !== undefined) return dirty[dirtyKey];
    return row.values[locale];
  }

  function updateValue(key: string, locale: AppLocale, value: string) {
    setDirty((prev) => ({ ...prev, [`${key}::${locale}`]: value }));
    setRows((prev) =>
      prev.map((row) =>
        row.key === key ? { ...row, values: { ...row.values, [locale]: value } } : row
      )
    );
    setSuccess(false);
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess(false);
    try {
      const items: AdminTranslation[] = Object.entries(dirty).map(([compound, value]) => {
        const [key, locale] = compound.split("::");
        return { key, locale, value };
      });
      if (items.length === 0) return;
      const data = await bulkUpdateAdminTranslations(items);
      setOverrides((prev) => {
        const map = new Map(prev.map((x) => [`${x.key}::${x.locale}`, x]));
        for (const item of data.items) map.set(`${item.key}::${item.locale}`, item);
        return Array.from(map.values());
      });
      setDirty({});
      setSuccess(true);
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
    <div className="space-y-6">
      <div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </div>

      <div className="flex flex-wrap gap-3">
        <input
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
            setPage(0);
          }}
          placeholder={t("searchPlaceholder")}
          className="min-w-[240px] flex-1 rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent"
        />
        <select
          value={namespace}
          onChange={(e) => {
            setNamespace(e.target.value);
            setPage(0);
          }}
          className="rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent"
        >
          <option value="">{t("allNamespaces")}</option>
          {TRANSLATION_NAMESPACES.map((ns) => (
            <option key={ns} value={ns}>
              {ns}
            </option>
          ))}
        </select>
        <select
          value={editLocale}
          onChange={(e) => setEditLocale(e.target.value as AppLocale)}
          className="rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent"
        >
          {SUPPORTED_LOCALES.map((code) => (
            <option key={code} value={code}>
              {LOCALE_META[code].nativeName}
            </option>
          ))}
        </select>
      </div>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}
      {success && (
        <div className="rounded-md border border-green-200 bg-success-bg px-3 py-2 text-sm text-success">
          {t("saved")}
        </div>
      )}

      <div className="overflow-x-auto rounded-xl border border-border bg-bg">
        <table className="w-full min-w-[720px] text-left text-xs">
          <thead>
            <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
              <th className="px-3 py-2 w-[34%]">{t("colKey")}</th>
              <th className="px-3 py-2">{t("colValue")}</th>
            </tr>
          </thead>
          <tbody>
            {pageRows.map((row) => (
              <tr key={row.key} className="border-b border-border align-top last:border-0">
                <td className="px-3 py-2 font-mono text-[11px] text-text2">{row.key}</td>
                <td className="px-3 py-2">
                  <textarea
                    rows={2}
                    value={valueFor(row, editLocale)}
                    onChange={(e) => updateValue(row.key, editLocale, e.target.value)}
                    className="w-full rounded-lg border border-border2 px-2 py-1.5 text-xs outline-none focus:border-accent"
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-xs text-text3">
          {t("pagination", {
            from: filtered.length === 0 ? 0 : page * PAGE_SIZE + 1,
            to: Math.min((page + 1) * PAGE_SIZE, filtered.length),
            total: filtered.length,
          })}
        </p>
        <div className="flex gap-2">
          <button
            type="button"
            disabled={page <= 0}
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            className="rounded-lg border border-border2 px-3 py-1.5 text-xs disabled:opacity-50"
          >
            {t("prev")}
          </button>
          <button
            type="button"
            disabled={page >= totalPages - 1}
            onClick={() => setPage((p) => p + 1)}
            className="rounded-lg border border-border2 px-3 py-1.5 text-xs disabled:opacity-50"
          >
            {t("next")}
          </button>
          <button
            type="button"
            onClick={handleSave}
            disabled={saving || Object.keys(dirty).length === 0}
            className="rounded-lg bg-text px-4 py-1.5 text-xs font-medium text-white disabled:opacity-50"
          >
            {saving ? t("saving") : t("save")}
          </button>
        </div>
      </div>
    </div>
  );
}
