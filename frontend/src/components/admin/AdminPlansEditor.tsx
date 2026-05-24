"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import {
  createAdminPlan,
  fetchAdminPlans,
  fetchAdminPlansMeta,
  updateAdminPlan,
  type AdminPlan,
  type AdminPlanMeta,
  type PlanUpsertInput,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const BOOL_FEATURES = [
  "export_pdf",
  "export_docx",
  "api_access",
  "priority_queue",
  "white_label",
  "share_report",
] as const;

type FormState = {
  slug: string;
  name: string;
  price_monthly_rub: number;
  price_yearly_rub: number;
  tool_limits: Record<string, number>;
  features: Record<string, unknown>;
  support_level: string;
  is_public: boolean;
  is_archived: boolean;
};

function emptyForm(meta: AdminPlanMeta | null): FormState {
  const tool_limits: Record<string, number> = {};
  meta?.tools.forEach((t) => {
    tool_limits[t.slug] = t.slug === "calculator" ? -1 : 0;
  });
  return {
    slug: "",
    name: "",
    price_monthly_rub: 0,
    price_yearly_rub: 0,
    tool_limits,
    features: { ...(meta?.default_features ?? {}) },
    support_level: meta?.support_levels[0] ?? "community",
    is_public: true,
    is_archived: false,
  };
}

function planToForm(plan: AdminPlan, meta: AdminPlanMeta | null): FormState {
  const base = emptyForm(meta);
  return {
    slug: plan.slug,
    name: plan.name,
    price_monthly_rub: plan.price_monthly_rub,
    price_yearly_rub: plan.price_yearly_rub,
    tool_limits: { ...base.tool_limits, ...plan.tool_limits },
    features: { ...base.features, ...plan.features },
    support_level: plan.support_level,
    is_public: plan.is_public,
    is_archived: plan.is_archived,
  };
}

function formToPayload(form: FormState, creating: boolean): PlanUpsertInput & { slug?: string } {
  const features = { ...form.features };
  for (const key of BOOL_FEATURES) {
    features[key] = Boolean(features[key]);
  }
  features.share_report_limit = Number(features.share_report_limit ?? 0);

  const payload: PlanUpsertInput & { slug?: string } = {
    name: form.name.trim(),
    price_monthly_rub: form.price_monthly_rub,
    price_yearly_rub: form.price_yearly_rub,
    tool_limits: form.tool_limits,
    features,
    support_level: form.support_level,
    is_public: form.is_public,
    is_archived: form.is_archived,
  };
  if (creating) {
    payload.slug = form.slug.trim().toLowerCase();
  }
  return payload;
}

function formatRub(n: number) {
  return new Intl.NumberFormat("ru-RU").format(n) + " ₽";
}

export function AdminPlansEditor() {
  const t = useTranslations("admin.plans");
  const [plans, setPlans] = useState<AdminPlan[]>([]);
  const [meta, setMeta] = useState<AdminPlanMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [form, setForm] = useState<FormState>(() => emptyForm(null));

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [plansRes, metaRes] = await Promise.all([
        fetchAdminPlans(),
        fetchAdminPlansMeta(),
      ]);
      setPlans(plansRes.items);
      setMeta(metaRes);
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
    setForm(emptyForm(meta));
    setSuccess(false);
    setError("");
  }

  function startEdit(plan: AdminPlan) {
    setCreating(false);
    setEditingId(plan.id);
    setForm(planToForm(plan, meta));
    setSuccess(false);
    setError("");
  }

  function cancelForm() {
    setCreating(false);
    setEditingId(null);
    setForm(emptyForm(meta));
    setError("");
  }

  function patchForm(partial: Partial<FormState>) {
    setForm((prev) => ({ ...prev, ...partial }));
    setSuccess(false);
  }

  function setToolLimit(slug: string, value: number) {
    setForm((prev) => ({
      ...prev,
      tool_limits: { ...prev.tool_limits, [slug]: value },
    }));
    setSuccess(false);
  }

  function setFeature(key: string, value: unknown) {
    setForm((prev) => ({
      ...prev,
      features: { ...prev.features, [key]: value },
    }));
    setSuccess(false);
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess(false);
    try {
      const payload = formToPayload(form, creating);
      if (creating) {
        if (!payload.slug) {
          setError(t("slugRequired"));
          return;
        }
        await createAdminPlan(payload as PlanUpsertInput & { slug: string });
      } else if (editingId) {
        await updateAdminPlan(editingId, payload);
      }
      await load();
      setSuccess(true);
      if (creating) {
        setCreating(false);
        setEditingId(null);
        setForm(emptyForm(meta));
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  const sortedPlans = useMemo(
    () => [...plans].sort((a, b) => a.price_monthly_rub - b.price_monthly_rub),
    [plans]
  );

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
            className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white"
          >
            {t("create")}
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

      <div className={cn("grid gap-6", showForm && "lg:grid-cols-5")}>
        <div className={cn("space-y-3", showForm && "lg:col-span-2")}>
          {sortedPlans.map((plan) => (
            <button
              key={plan.id}
              type="button"
              onClick={() => startEdit(plan)}
              className={cn(
                "w-full rounded-xl border bg-bg p-4 text-left transition-colors hover:border-border2",
                editingId === plan.id ? "border-text" : "border-border"
              )}
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="font-medium">{plan.name}</p>
                  <p className="mt-0.5 text-xs text-text3">{plan.slug}</p>
                </div>
                <p className="text-sm font-medium whitespace-nowrap">
                  {formatRub(plan.price_monthly_rub)}
                  <span className="text-xs font-normal text-text3"> / {t("month")}</span>
                </p>
              </div>
              <div className="mt-3 flex flex-wrap gap-2 text-[10px] uppercase tracking-wider">
                {plan.is_public && (
                  <span className="rounded bg-accent-bg px-2 py-0.5 text-accent">{t("badgePublic")}</span>
                )}
                {plan.is_archived && (
                  <span className="rounded bg-bg2 px-2 py-0.5 text-text3">{t("badgeArchived")}</span>
                )}
                <span className="rounded bg-bg2 px-2 py-0.5 text-text3">
                  {t("usersCount", { count: plan.user_count ?? 0 })}
                </span>
              </div>
            </button>
          ))}
        </div>

        {showForm && (
          <div className="lg:col-span-3 rounded-xl border border-border bg-bg p-6 space-y-6">
            <div className="flex items-center justify-between gap-3">
              <h2 className="text-sm font-medium">
                {creating ? t("formCreate") : t("formEdit")}
              </h2>
              <button type="button" onClick={cancelForm} className="text-xs text-text2 hover:text-text">
                {t("cancel")}
              </button>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <label className="mb-1.5 block text-xs font-medium">{t("fieldName")}</label>
                <input
                  className={fieldClass}
                  value={form.name}
                  onChange={(e) => patchForm({ name: e.target.value })}
                />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium">{t("fieldSlug")}</label>
                <input
                  className={fieldClass}
                  value={form.slug}
                  disabled={!creating}
                  placeholder={t("slugPlaceholder")}
                  onChange={(e) => patchForm({ slug: e.target.value.toLowerCase() })}
                />
                {creating && (
                  <p className="mt-1 text-[10px] text-text3">{t("slugHint")}</p>
                )}
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium">{t("fieldPriceMonthly")}</label>
                <input
                  type="number"
                  min={0}
                  className={fieldClass}
                  value={form.price_monthly_rub}
                  onChange={(e) => patchForm({ price_monthly_rub: Number(e.target.value) })}
                />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium">{t("fieldPriceYearly")}</label>
                <input
                  type="number"
                  min={0}
                  className={fieldClass}
                  value={form.price_yearly_rub}
                  onChange={(e) => patchForm({ price_yearly_rub: Number(e.target.value) })}
                />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium">{t("fieldSupport")}</label>
                <select
                  className={fieldClass}
                  value={form.support_level}
                  onChange={(e) => patchForm({ support_level: e.target.value })}
                >
                  {(meta?.support_levels ?? ["community", "email", "priority"]).map((level) => (
                    <option key={level} value={level}>
                      {t(`support.${level}` as "support.community")}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div className="flex flex-wrap gap-4">
              <label className="inline-flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={form.is_public}
                  onChange={(e) => patchForm({ is_public: e.target.checked })}
                />
                {t("fieldPublic")}
              </label>
              <label className="inline-flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={form.is_archived}
                  onChange={(e) => patchForm({ is_archived: e.target.checked })}
                />
                {t("fieldArchived")}
              </label>
            </div>

            <section className="space-y-3 border-t border-border pt-4">
              <h3 className="text-[10px] font-medium uppercase tracking-wider text-text3">
                {t("sectionLimits")}
              </h3>
              <p className="text-xs text-text3">{t("limitsHint")}</p>
              <div className="grid gap-3 sm:grid-cols-2">
                {(meta?.tools ?? []).map((tool) => (
                  <div key={tool.slug}>
                    <label className="mb-1.5 block text-xs font-medium">{tool.name}</label>
                    <input
                      type="number"
                      className={fieldClass}
                      value={form.tool_limits[tool.slug] ?? 0}
                      onChange={(e) => setToolLimit(tool.slug, Number(e.target.value))}
                    />
                  </div>
                ))}
              </div>
            </section>

            <section className="space-y-3 border-t border-border pt-4">
              <h3 className="text-[10px] font-medium uppercase tracking-wider text-text3">
                {t("sectionFeatures")}
              </h3>
              <div className="grid gap-2 sm:grid-cols-2">
                {BOOL_FEATURES.map((key) => (
                  <label key={key} className="inline-flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={Boolean(form.features[key])}
                      onChange={(e) => setFeature(key, e.target.checked)}
                    />
                    {t(`features.${key}` as "features.export_pdf")}
                  </label>
                ))}
              </div>
              <div className="max-w-xs">
                <label className="mb-1.5 block text-xs font-medium">
                  {t("features.share_report_limit")}
                </label>
                <input
                  type="number"
                  className={fieldClass}
                  value={Number(form.features.share_report_limit ?? 0)}
                  onChange={(e) => setFeature("share_report_limit", Number(e.target.value))}
                />
                <p className="mt-1 text-[10px] text-text3">{t("limitsHint")}</p>
              </div>
            </section>

            <div className="flex gap-2 pt-2">
              <button
                type="button"
                onClick={handleSave}
                disabled={saving || !form.name.trim() || (creating && !form.slug.trim())}
                className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
              >
                {saving ? t("saving") : t("save")}
              </button>
              <button
                type="button"
                onClick={cancelForm}
                className="rounded-lg border border-border2 px-4 py-2 text-sm"
              >
                {t("cancel")}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
