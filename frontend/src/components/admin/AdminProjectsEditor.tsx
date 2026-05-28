"use client";

import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  createAdminExternalProject,
  deleteAdminExternalProject,
  fetchAdminExternalProjects,
  updateAdminExternalProject,
  type ExternalProject,
  type ExternalProjectInput,
} from "@/lib/api";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

type FormState = {
  title: string;
  url: string;
  sort_order: number;
  is_enabled: boolean;
};

function emptyForm(): FormState {
  return {
    title: "",
    url: "",
    sort_order: 0,
    is_enabled: true,
  };
}

function projectToForm(project: ExternalProject): FormState {
  return {
    title: project.title,
    url: project.url,
    sort_order: project.sort_order,
    is_enabled: project.is_enabled,
  };
}

function formToPayload(form: FormState): ExternalProjectInput {
  return {
    title: form.title.trim(),
    url: form.url.trim(),
    sort_order: form.sort_order,
    is_enabled: form.is_enabled,
  };
}

export function AdminProjectsEditor() {
  const t = useTranslations("admin.projects");
  const [items, setItems] = useState<ExternalProject[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [form, setForm] = useState<FormState>(() => emptyForm());

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await fetchAdminExternalProjects();
      setItems(data.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    load();
  }, [load]);

  const showForm = creating || editingId !== null;

  function startCreate() {
    setCreating(true);
    setEditingId(null);
    setForm(emptyForm());
    setSuccess(false);
    setError("");
  }

  function startEdit(project: ExternalProject) {
    setCreating(false);
    setEditingId(project.id);
    setForm(projectToForm(project));
    setSuccess(false);
    setError("");
  }

  function cancelForm() {
    setCreating(false);
    setEditingId(null);
    setForm(emptyForm());
    setError("");
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess(false);
    try {
      const payload = formToPayload(form);
      if (creating) {
        const created = await createAdminExternalProject(payload);
        setItems((prev) => [...prev, created].sort((a, b) => a.sort_order - b.sort_order));
        cancelForm();
      } else if (editingId) {
        const updated = await updateAdminExternalProject(editingId, payload);
        setItems((prev) =>
          prev
            .map((item) => (item.id === editingId ? updated : item))
            .sort((a, b) => a.sort_order - b.sort_order)
        );
        cancelForm();
      }
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(project: ExternalProject) {
    if (!window.confirm(t("deleteConfirm", { title: project.title }))) {
      return;
    }
    setError("");
    setSuccess(false);
    try {
      await deleteAdminExternalProject(project.id);
      setItems((prev) => prev.filter((item) => item.id !== project.id));
      if (editingId === project.id) {
        cancelForm();
      }
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("deleteFailed"));
    }
  }

  if (loading) {
    return <p className="text-sm text-text2">{t("loading")}</p>;
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-base font-medium">{t("title")}</h1>
          <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
        </div>
        {!showForm && (
          <button
            type="button"
            onClick={startCreate}
            className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white hover:opacity-90"
          >
            {t("add")}
          </button>
        )}
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

      {showForm && (
        <div className="rounded-xl border border-border bg-bg p-5 space-y-4">
          <h2 className="text-sm font-medium">
            {creating ? t("createTitle") : t("editTitle")}
          </h2>
          <div className="grid gap-4 lg:grid-cols-2">
            <div>
              <label className="mb-1 block text-xs font-medium text-text2">
                {t("fieldTitle")}
              </label>
              <input
                type="text"
                className={fieldClass}
                value={form.title}
                onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))}
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-text2">
                {t("fieldUrl")}
              </label>
              <input
                type="url"
                className={fieldClass}
                placeholder="https://"
                value={form.url}
                onChange={(e) => setForm((prev) => ({ ...prev, url: e.target.value }))}
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-text2">
                {t("fieldSortOrder")}
              </label>
              <input
                type="number"
                className={fieldClass}
                value={form.sort_order}
                onChange={(e) =>
                  setForm((prev) => ({
                    ...prev,
                    sort_order: Number(e.target.value) || 0,
                  }))
                }
              />
            </div>
            <div className="flex items-end">
              <label className="flex items-center gap-2 text-sm text-text2">
                <input
                  type="checkbox"
                  checked={form.is_enabled}
                  onChange={(e) =>
                    setForm((prev) => ({ ...prev, is_enabled: e.target.checked }))
                  }
                />
                {t("fieldEnabled")}
              </label>
            </div>
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={handleSave}
              disabled={saving}
              className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-50"
            >
              {saving ? t("saving") : t("save")}
            </button>
            <button
              type="button"
              onClick={cancelForm}
              className="rounded-lg border border-border2 px-4 py-2 text-sm text-text2 hover:bg-bg2"
            >
              {t("cancel")}
            </button>
          </div>
        </div>
      )}

      <div className="overflow-hidden rounded-xl border border-border bg-bg">
        <table className="w-full text-sm">
          <thead className="bg-bg2 text-left text-[10px] font-medium uppercase tracking-wider text-text3">
            <tr>
              <th className="px-4 py-3">{t("colTitle")}</th>
              <th className="px-4 py-3">{t("colUrl")}</th>
              <th className="px-4 py-3">{t("colSortOrder")}</th>
              <th className="px-4 py-3">{t("colEnabled")}</th>
              <th className="px-4 py-3">{t("colActions")}</th>
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-text2">
                  {t("empty")}
                </td>
              </tr>
            ) : (
              items.map((project) => (
                <tr key={project.id} className="border-t border-border">
                  <td className="px-4 py-3 font-medium">{project.title}</td>
                  <td className="px-4 py-3">
                    <a
                      href={project.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-accent hover:underline break-all"
                    >
                      {project.url}
                    </a>
                  </td>
                  <td className="px-4 py-3">{project.sort_order}</td>
                  <td className="px-4 py-3">
                    {project.is_enabled ? t("enabledYes") : t("enabledNo")}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button
                        type="button"
                        onClick={() => startEdit(project)}
                        className="text-accent hover:underline"
                      >
                        {t("edit")}
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDelete(project)}
                        className="text-error hover:underline"
                      >
                        {t("delete")}
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
