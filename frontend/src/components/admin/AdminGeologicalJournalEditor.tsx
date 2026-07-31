"use client";

import {
  AlertCircle,
  Check,
  FileImage,
  LoaderCircle,
  Plus,
  RefreshCw,
  Save,
  Settings2,
  Trash2,
  Upload,
  Users,
} from "lucide-react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { useCallback, useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";
import type { LLMProviderStatus } from "@/lib/api";
import { ModelPicker } from "@/components/admin/ModelPicker";
import {
  createAdminGeologicalJournalExample,
  deleteAdminGeologicalJournalExample,
  fetchAdminGeologicalJournalAccess,
  fetchAdminGeologicalJournalExamples,
  fetchAdminGeologicalJournalSettings,
  geologicalJournalExampleImageUrl,
  refreshAdminGeologicalJournalModels,
  updateAdminGeologicalJournalAccess,
  updateAdminGeologicalJournalExample,
  updateAdminGeologicalJournalSettings,
  validateGeologicalJournalImage,
  type GeologicalJournalAccessUser,
  type GeologicalJournalExample,
  type GeologicalJournalExampleMetadata,
  type GeologicalJournalSettings,
} from "@/lib/api-geological-journal";

type Tab = "settings" | "access" | "examples";

const fieldClass =
  "w-full rounded-lg border border-border2 bg-bg px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const DEFAULT_SETTINGS: GeologicalJournalSettings = {
  provider: "openrouter",
  openrouter_model: "",
  yandex_model: "",
  system_prompt: "",
  temperature: 0.2,
  max_tokens: 4096,
};

const EMPTY_EXAMPLE: GeologicalJournalExampleMetadata = {
  title: "",
  description: "",
  sort_order: 0,
  is_published: false,
};

export function AdminGeologicalJournalEditor() {
  const t = useTranslations("admin.geologicalJournal");
  const [tab, setTab] = useState<Tab>("settings");
  const [settings, setSettings] =
    useState<GeologicalJournalSettings>(DEFAULT_SETTINGS);
  const [providers, setProviders] = useState<LLMProviderStatus[]>([]);
  const [users, setUsers] = useState<GeologicalJournalAccessUser[]>([]);
  const [examples, setExamples] = useState<GeologicalJournalExample[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [refreshingModels, setRefreshingModels] = useState(false);
  const [busyId, setBusyId] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [settingsData, accessData, examplesData] = await Promise.all([
        fetchAdminGeologicalJournalSettings(),
        fetchAdminGeologicalJournalAccess(),
        fetchAdminGeologicalJournalExamples(),
      ]);
      setSettings(settingsData.settings ?? (settingsData as unknown as GeologicalJournalSettings));
      setProviders(settingsData.providers ?? []);
      setUsers(accessData.items ?? []);
      setExamples(examplesData.items ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.load"));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    void load();
  }, [load]);

  function patchSettings(partial: Partial<GeologicalJournalSettings>) {
    setSettings((current) => ({ ...current, ...partial }));
    setSuccess("");
  }

  async function saveSettings() {
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      const response = await updateAdminGeologicalJournalSettings(settings);
      setSettings(response.settings ?? settings);
      setProviders(response.providers ?? providers);
      setSuccess(t("settings.saved"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.save"));
    } finally {
      setSaving(false);
    }
  }

  async function refreshModels() {
    setRefreshingModels(true);
    setError("");
    setSuccess("");
    try {
      const result = await refreshAdminGeologicalJournalModels(settings.provider);
      if (!result.ok) {
        throw new Error(result.message);
      }
      setProviders((current) =>
        current.map((provider) =>
          provider.id === settings.provider
            ? { ...provider, models: result.models ?? [] }
            : provider
        )
      );
      setSuccess(t("settings.modelsLoaded", { count: result.models?.length ?? 0 }));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.models"));
    } finally {
      setRefreshingModels(false);
    }
  }

  async function toggleAccess(user: GeologicalJournalAccessUser) {
    const id = user.user_id || user.id;
    if (!id) return;
    setBusyId(id);
    setError("");
    try {
      const enabled = !user.has_access;
      await updateAdminGeologicalJournalAccess(id, enabled);
      setUsers((current) =>
        current.map((item) =>
          (item.user_id || item.id) === id
            ? { ...item, has_access: enabled }
            : item
        )
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.access"));
    } finally {
      setBusyId("");
    }
  }

  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </header>

      <div className="flex gap-1 overflow-x-auto border-b border-border">
        {(
          [
            ["settings", Settings2],
            ["access", Users],
            ["examples", FileImage],
          ] as const
        ).map(([id, Icon]) => (
          <button
            key={id}
            type="button"
            onClick={() => setTab(id)}
            className={cn(
              "-mb-px inline-flex shrink-0 items-center gap-2 border-b-2 px-3 py-2 text-sm",
              tab === id
                ? "border-text font-medium text-text"
                : "border-transparent text-text2 hover:text-text"
            )}
          >
            <Icon size={15} />
            {t(`tabs.${id}`)}
          </button>
        ))}
      </div>

      {error && (
        <div className="flex items-start gap-2 rounded-lg border border-red-200 bg-error-bg px-3 py-2.5 text-sm text-error">
          <AlertCircle size={16} className="mt-0.5 shrink-0" />
          {error}
        </div>
      )}
      {success && (
        <div className="flex items-center gap-2 rounded-lg border border-green-200 bg-success-bg px-3 py-2.5 text-sm text-success">
          <Check size={16} />
          {success}
        </div>
      )}

      {loading ? (
        <div className="flex items-center gap-2 text-sm text-text2">
          <LoaderCircle size={16} className="animate-spin" />
          {t("loading")}
        </div>
      ) : (
        <>
          {tab === "settings" && (
            <SettingsPanel
              settings={settings}
              providers={providers}
              saving={saving}
              refreshingModels={refreshingModels}
              patch={patchSettings}
              save={saveSettings}
              refreshModels={refreshModels}
              t={t}
            />
          )}
          {tab === "access" && (
            <AccessPanel
              users={users}
              busyId={busyId}
              toggle={toggleAccess}
              t={t}
            />
          )}
          {tab === "examples" && (
            <ExamplesPanel
              examples={examples}
              setExamples={setExamples}
              busyId={busyId}
              setBusyId={setBusyId}
              setError={setError}
              setSuccess={setSuccess}
              t={t}
            />
          )}
        </>
      )}
    </div>
  );
}

function SettingsPanel({
  settings,
  providers,
  saving,
  refreshingModels,
  patch,
  save,
  refreshModels,
  t,
}: {
  settings: GeologicalJournalSettings;
  providers: LLMProviderStatus[];
  saving: boolean;
  refreshingModels: boolean;
  patch: (partial: Partial<GeologicalJournalSettings>) => void;
  save: () => void;
  refreshModels: () => void;
  t: ReturnType<typeof useTranslations>;
}) {
  const activeModel =
    settings.provider === "yandex"
      ? settings.yandex_model
      : settings.openrouter_model;
  const providerModels =
    providers.find((provider) => provider.id === settings.provider)?.models ?? [];
  const modelOptions = Array.from(
    new Set([activeModel, ...providerModels].filter(Boolean))
  );

  return (
    <div className="space-y-5">
      <section className="space-y-4 rounded-xl border border-border bg-bg p-5">
        <div>
          <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
            {t("settings.providerSection")}
          </h2>
          <p className="mt-1 text-xs text-text3">{t("settings.providerHint")}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          {(["openrouter", "yandex"] as const).map((provider) => (
            <button
              key={provider}
              type="button"
              onClick={() => patch({ provider })}
              className={cn(
                "rounded-lg border px-4 py-2 text-sm transition-colors duration-150",
                settings.provider === provider
                  ? "border-ai bg-ai-bg font-medium text-ai"
                  : "border-border2 text-text2 hover:bg-bg2"
              )}
            >
              {t(`settings.providers.${provider}`)}
            </button>
          ))}
        </div>
      </section>

      <section className="space-y-4 rounded-xl border border-border bg-bg p-5">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
              {t("settings.modelsSection")}
            </h2>
            <p className="mt-1 text-xs text-text3">{t("settings.modelsHint")}</p>
          </div>
          <button
            type="button"
            onClick={refreshModels}
            disabled={refreshingModels}
            className="inline-flex items-center gap-2 rounded-lg border border-border2 px-3 py-2 text-xs font-medium text-text2 hover:bg-bg2 disabled:opacity-50"
          >
            <RefreshCw
              size={14}
              className={cn(refreshingModels && "animate-spin")}
            />
            {refreshingModels
              ? t("settings.refreshingModels")
              : t("settings.refreshModels")}
          </button>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <ModelPicker
            label={
              settings.provider === "yandex"
                ? t("settings.yandexModel")
                : t("settings.openrouterModel")
            }
            value={activeModel}
            models={modelOptions}
            onChange={(value) =>
              patch(
                settings.provider === "yandex"
                  ? { yandex_model: value }
                  : { openrouter_model: value }
              )
            }
            placeholder={
              settings.provider === "yandex"
                ? "yandexgpt/latest"
                : "google/gemini-2.5-flash"
            }
          />
          <Field label={t("settings.temperature")}>
            <input
              type="number"
              min={0}
              max={2}
              step={0.1}
              className={fieldClass}
              value={settings.temperature}
              onChange={(event) => patch({ temperature: Number(event.target.value) })}
            />
          </Field>
          <Field label={t("settings.maxTokens")}>
            <input
              type="number"
              min={256}
              max={32768}
              step={256}
              className={fieldClass}
              value={settings.max_tokens}
              onChange={(event) => patch({ max_tokens: Number(event.target.value) })}
            />
          </Field>
        </div>
        <p className="text-xs text-text3">
          {t("settings.activeModel")}:{" "}
          <span className="font-mono text-text">{activeModel}</span>
        </p>
      </section>

      <section className="space-y-3 rounded-xl border border-border bg-bg p-5">
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("settings.promptSection")}
        </h2>
        <textarea
          rows={18}
          className={cn(fieldClass, "font-mono text-xs leading-relaxed")}
          value={settings.system_prompt}
          onChange={(event) => patch({ system_prompt: event.target.value })}
        />
        <p className="text-xs text-text3">{t("settings.promptHint")}</p>
      </section>

      <button
        type="button"
        onClick={save}
        disabled={saving}
        className="inline-flex items-center gap-2 rounded-lg bg-text px-4 py-2.5 text-sm font-medium text-white disabled:opacity-50"
      >
        {saving ? (
          <LoaderCircle size={15} className="animate-spin" />
        ) : (
          <Save size={15} />
        )}
        {saving ? t("settings.saving") : t("settings.save")}
      </button>
    </div>
  );
}

function AccessPanel({
  users,
  busyId,
  toggle,
  t,
}: {
  users: GeologicalJournalAccessUser[];
  busyId: string;
  toggle: (user: GeologicalJournalAccessUser) => void;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <section className="overflow-hidden rounded-xl border border-border bg-bg">
      <div className="border-b border-border px-5 py-4">
        <h2 className="text-sm font-medium">{t("access.title")}</h2>
        <p className="mt-1 text-xs text-text3">{t("access.hint")}</p>
      </div>
      {users.length === 0 ? (
        <p className="px-5 py-8 text-sm text-text3">{t("access.empty")}</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full min-w-[600px] text-left text-sm">
            <thead>
              <tr className="border-b border-border bg-bg2 text-[10px] uppercase tracking-wider text-text3">
                <th className="px-4 py-3 font-medium">{t("access.email")}</th>
                <th className="px-4 py-3 font-medium">{t("access.role")}</th>
                <th className="px-4 py-3 text-right font-medium">
                  {t("access.permission")}
                </th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => {
                const id = user.user_id || user.id || user.email;
                return (
                  <tr key={id} className="border-b border-border last:border-0">
                    <td className="px-4 py-3 font-medium">{user.email}</td>
                    <td className="px-4 py-3 text-xs text-text2">{user.role}</td>
                    <td className="px-4 py-3 text-right">
                      <button
                        type="button"
                        role="switch"
                        aria-checked={user.has_access}
                        disabled={busyId === id}
                        onClick={() => void toggle(user)}
                        className={cn(
                          "relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-150 disabled:opacity-50",
                          user.has_access ? "bg-success" : "bg-border2"
                        )}
                      >
                        <span
                          className={cn(
                            "h-4 w-4 rounded-full bg-white transition-transform duration-150",
                            user.has_access ? "translate-x-6" : "translate-x-1"
                          )}
                        />
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

function ExamplesPanel({
  examples,
  setExamples,
  busyId,
  setBusyId,
  setError,
  setSuccess,
  t,
}: {
  examples: GeologicalJournalExample[];
  setExamples: React.Dispatch<React.SetStateAction<GeologicalJournalExample[]>>;
  busyId: string;
  setBusyId: (id: string) => void;
  setError: (message: string) => void;
  setSuccess: (message: string) => void;
  t: ReturnType<typeof useTranslations>;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [draft, setDraft] =
    useState<GeologicalJournalExampleMetadata>(EMPTY_EXAMPLE);
  const [uploading, setUploading] = useState(false);

  async function create() {
    if (!file || !draft.title.trim()) return;
    setUploading(true);
    setError("");
    setSuccess("");
    try {
      const created = await createAdminGeologicalJournalExample(file, draft);
      setExamples((current) =>
        [...current, created].sort((a, b) => a.sort_order - b.sort_order)
      );
      setFile(null);
      setDraft(EMPTY_EXAMPLE);
      if (inputRef.current) inputRef.current.value = "";
      setSuccess(t("examples.created"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.exampleUpload"));
    } finally {
      setUploading(false);
    }
  }

  async function save(item: GeologicalJournalExample) {
    setBusyId(item.id);
    setError("");
    try {
      const updated = await updateAdminGeologicalJournalExample(item.id, {
        title: item.title,
        description: item.description,
        sort_order: item.sort_order,
        is_published: item.is_published,
      });
      setExamples((current) =>
        current
          .map((example) => (example.id === item.id ? updated : example))
          .sort((a, b) => a.sort_order - b.sort_order)
      );
      setSuccess(t("examples.saved"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.exampleSave"));
    } finally {
      setBusyId("");
    }
  }

  async function remove(item: GeologicalJournalExample) {
    if (!window.confirm(t("examples.deleteConfirm", { title: item.title }))) return;
    setBusyId(item.id);
    setError("");
    try {
      await deleteAdminGeologicalJournalExample(item.id);
      setExamples((current) => current.filter((example) => example.id !== item.id));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("errors.exampleDelete"));
    } finally {
      setBusyId("");
    }
  }

  function patchExample(
    id: string,
    partial: Partial<GeologicalJournalExample>
  ) {
    setExamples((current) =>
      current.map((item) => (item.id === id ? { ...item, ...partial } : item))
    );
  }

  return (
    <div className="space-y-5">
      <section className="rounded-xl border border-border bg-bg p-5">
        <div className="mb-4">
          <h2 className="text-sm font-medium">{t("examples.addTitle")}</h2>
          <p className="mt-1 text-xs text-text3">{t("examples.addHint")}</p>
        </div>
        <div className="grid gap-4 lg:grid-cols-[220px_1fr]">
          <button
            type="button"
            onClick={() => inputRef.current?.click()}
            className="flex min-h-[150px] flex-col items-center justify-center rounded-lg border border-dashed border-border2 bg-bg2 p-4 text-center hover:border-accent"
          >
            <Upload size={20} className="text-text3" />
            <span className="mt-2 text-xs font-medium">
              {file ? file.name : t("examples.chooseImage")}
            </span>
            <span className="mt-1 text-[11px] text-text3">
              {t("examples.formats")}
            </span>
          </button>
          <div className="grid gap-3 sm:grid-cols-2">
            <Field label={t("examples.titleField")}>
              <input
                className={fieldClass}
                value={draft.title}
                onChange={(event) =>
                  setDraft((current) => ({ ...current, title: event.target.value }))
                }
              />
            </Field>
            <Field label={t("examples.orderField")}>
              <input
                type="number"
                className={fieldClass}
                value={draft.sort_order}
                onChange={(event) =>
                  setDraft((current) => ({
                    ...current,
                    sort_order: Number(event.target.value),
                  }))
                }
              />
            </Field>
            <div className="sm:col-span-2">
              <Field label={t("examples.descriptionField")}>
                <textarea
                  rows={3}
                  className={fieldClass}
                  value={draft.description}
                  onChange={(event) =>
                    setDraft((current) => ({
                      ...current,
                      description: event.target.value,
                    }))
                  }
                />
              </Field>
            </div>
            <label className="flex items-center gap-2 text-xs text-text2">
              <input
                type="checkbox"
                checked={draft.is_published}
                onChange={(event) =>
                  setDraft((current) => ({
                    ...current,
                    is_published: event.target.checked,
                  }))
                }
              />
              {t("examples.published")}
            </label>
            <div className="flex justify-end">
              <button
                type="button"
                disabled={!file || !draft.title.trim() || uploading}
                onClick={() => void create()}
                className="inline-flex items-center gap-2 rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-40"
              >
                {uploading ? (
                  <LoaderCircle size={15} className="animate-spin" />
                ) : (
                  <Plus size={15} />
                )}
                {uploading ? t("examples.uploading") : t("examples.add")}
              </button>
            </div>
          </div>
        </div>
        <input
          ref={inputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp,image/gif"
          className="hidden"
          onChange={(event) => {
            const next = event.target.files?.[0] ?? null;
            if (next) {
              const validation = validateGeologicalJournalImage(next);
              if (validation) {
                setError(t(`errors.${validation}`));
                event.target.value = "";
                return;
              }
            }
            setFile(next);
          }}
        />
      </section>

      <div className="space-y-3">
        {examples.length === 0 ? (
          <div className="rounded-xl border border-border bg-bg px-5 py-10 text-center text-sm text-text3">
            {t("examples.empty")}
          </div>
        ) : (
          examples.map((item) => (
            <article
              key={item.id}
              className="grid gap-4 rounded-xl border border-border bg-bg p-4 md:grid-cols-[150px_1fr_auto]"
            >
              <div className="relative aspect-[16/10] overflow-hidden rounded-lg border border-border bg-bg2">
                <Image
                  src={geologicalJournalExampleImageUrl(item.id)}
                  alt={item.title}
                  fill
                  sizes="150px"
                  unoptimized
                  className="h-full w-full object-cover"
                />
              </div>
              <div className="grid gap-3 sm:grid-cols-[1fr_120px]">
                <input
                  aria-label={t("examples.titleField")}
                  className={fieldClass}
                  value={item.title}
                  onChange={(event) =>
                    patchExample(item.id, { title: event.target.value })
                  }
                />
                <input
                  type="number"
                  aria-label={t("examples.orderField")}
                  className={fieldClass}
                  value={item.sort_order}
                  onChange={(event) =>
                    patchExample(item.id, {
                      sort_order: Number(event.target.value),
                    })
                  }
                />
                <textarea
                  rows={2}
                  aria-label={t("examples.descriptionField")}
                  className={cn(fieldClass, "sm:col-span-2")}
                  value={item.description}
                  onChange={(event) =>
                    patchExample(item.id, { description: event.target.value })
                  }
                />
                <label className="flex items-center gap-2 text-xs text-text2 sm:col-span-2">
                  <input
                    type="checkbox"
                    checked={item.is_published}
                    onChange={(event) =>
                      patchExample(item.id, {
                        is_published: event.target.checked,
                      })
                    }
                  />
                  {t("examples.published")}
                </label>
              </div>
              <div className="flex items-start gap-2 md:flex-col">
                <button
                  type="button"
                  disabled={busyId === item.id}
                  onClick={() => void save(item)}
                  className="inline-flex items-center gap-1.5 rounded-lg border border-border2 px-3 py-2 text-xs font-medium hover:bg-bg2 disabled:opacity-50"
                >
                  <Save size={14} />
                  {t("examples.save")}
                </button>
                <button
                  type="button"
                  disabled={busyId === item.id}
                  onClick={() => void remove(item)}
                  className="inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-xs text-error hover:bg-error-bg disabled:opacity-50"
                >
                  <Trash2 size={14} />
                  {t("examples.delete")}
                </button>
              </div>
            </article>
          ))
        )}
      </div>
    </div>
  );
}

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <label className="block">
      <span className="mb-1.5 block text-xs font-medium">{label}</span>
      {children}
    </label>
  );
}
