"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchAdminLegalRisks,
  updateAdminLegalRisk,
  type LegalRisk,
  type LegalRiskUpdateRequest,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

export function AdminLegalRisksEditor() {
  const t = useTranslations("admin.legalRisks");
  const [items, setItems] = useState<LegalRisk[]>([]);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [draft, setDraft] = useState<LegalRiskUpdateRequest | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    fetchAdminLegalRisks()
      .then((data) => setItems(data.items))
      .catch((err) => setError(err instanceof Error ? err.message : t("loadFailed")))
      .finally(() => setLoading(false));
  }, [t]);

  function startEdit(item: LegalRisk) {
    setEditingId(item.risk_id);
    setDraft({
      title_ru: item.title_ru,
      title_en: item.title_en,
      what_ru: item.what_ru,
      article: item.article,
      fine_text_ru: item.fine_text_ru,
      fine_min: item.fine_min,
      fine_max: item.fine_max,
      severity: item.severity,
      how_to_fix_ru: item.how_to_fix_ru,
      how_to_fix_en: item.how_to_fix_en,
      trigger_findings: item.trigger_findings,
      trigger_flags: item.trigger_flags,
      is_turnover_fine: item.is_turnover_fine,
      is_context_only: item.is_context_only,
      is_active: item.is_active,
      sort_order: item.sort_order,
    });
    setSuccess("");
  }

  async function handleSave(riskId: string) {
    if (!draft) return;
    setSaving(true);
    setError("");
    try {
      const updated = await updateAdminLegalRisk(riskId, draft);
      setItems((prev) => prev.map((x) => (x.risk_id === riskId ? updated : x)));
      setEditingId(null);
      setDraft(null);
      setSuccess(t("saved"));
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

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}
      {success && (
        <div className="rounded-md border border-green-200 bg-success-bg px-3 py-2 text-sm text-success">
          {success}
        </div>
      )}

      <div className="space-y-3">
        {items.map((item) => {
          const isEditing = editingId === item.risk_id && draft;
          return (
            <div key={item.risk_id} className="rounded-xl border border-border bg-bg p-4">
              <div className="flex flex-wrap items-start justify-between gap-2">
                <div>
                  <p className="font-mono text-xs text-text3">{item.risk_id}</p>
                  <p className="text-sm font-medium">{item.title_ru}</p>
                  <p className="mt-1 text-xs text-text2">
                    {item.article} · {item.fine_text_ru}
                  </p>
                </div>
                <div className="flex gap-2">
                  {!isEditing ? (
                    <button
                      type="button"
                      onClick={() => startEdit(item)}
                      className="rounded-lg border border-border2 px-3 py-1.5 text-xs hover:bg-bg2"
                    >
                      {t("edit")}
                    </button>
                  ) : (
                    <>
                      <button
                        type="button"
                        disabled={saving}
                        onClick={() => void handleSave(item.risk_id)}
                        className="rounded-lg bg-text px-3 py-1.5 text-xs text-white disabled:opacity-60"
                      >
                        {saving ? t("saving") : t("save")}
                      </button>
                      <button
                        type="button"
                        onClick={() => {
                          setEditingId(null);
                          setDraft(null);
                        }}
                        className="rounded-lg border border-border2 px-3 py-1.5 text-xs"
                      >
                        {t("cancel")}
                      </button>
                    </>
                  )}
                </div>
              </div>

              {isEditing && draft && (
                <div className="mt-4 grid gap-3 sm:grid-cols-2">
                  <Field label={t("fields.title")} value={draft.title_ru} onChange={(v) => setDraft({ ...draft, title_ru: v })} />
                  <Field label={t("fields.article")} value={draft.article} onChange={(v) => setDraft({ ...draft, article: v })} />
                  <Field label={t("fields.fineText")} value={draft.fine_text_ru} onChange={(v) => setDraft({ ...draft, fine_text_ru: v })} className="sm:col-span-2" />
                  <Field label={t("fields.what")} value={draft.what_ru} onChange={(v) => setDraft({ ...draft, what_ru: v })} className="sm:col-span-2" />
                  <Field label={t("fields.howToFix")} value={draft.how_to_fix_ru} onChange={(v) => setDraft({ ...draft, how_to_fix_ru: v })} className="sm:col-span-2" />
                  <div>
                    <label className="mb-1 block text-xs font-medium">{t("fields.severity")}</label>
                    <select
                      className={fieldClass}
                      value={draft.severity}
                      onChange={(e) =>
                        setDraft({
                          ...draft,
                          severity: e.target.value as LegalRiskUpdateRequest["severity"],
                        })
                      }
                    >
                      <option value="high">{t("severity.high")}</option>
                      <option value="medium">{t("severity.medium")}</option>
                      <option value="low">{t("severity.low")}</option>
                    </select>
                  </div>
                  <div>
                    <label className="mb-1 block text-xs font-medium">{t("fields.sortOrder")}</label>
                    <input
                      type="number"
                      className={fieldClass}
                      value={draft.sort_order}
                      onChange={(e) => setDraft({ ...draft, sort_order: Number(e.target.value) })}
                    />
                  </div>
                  <div>
                    <label className="mb-1 block text-xs font-medium">{t("fields.fineMin")}</label>
                    <input
                      type="number"
                      className={fieldClass}
                      value={draft.fine_min ?? ""}
                      onChange={(e) =>
                        setDraft({
                          ...draft,
                          fine_min: e.target.value === "" ? null : Number(e.target.value),
                        })
                      }
                    />
                  </div>
                  <div>
                    <label className="mb-1 block text-xs font-medium">{t("fields.fineMax")}</label>
                    <input
                      type="number"
                      className={fieldClass}
                      value={draft.fine_max ?? ""}
                      onChange={(e) =>
                        setDraft({
                          ...draft,
                          fine_max: e.target.value === "" ? null : Number(e.target.value),
                        })
                      }
                    />
                  </div>
                  <JsonField
                    label={t("fields.triggerFindings")}
                    value={draft.trigger_findings}
                    onChange={(v) => setDraft({ ...draft, trigger_findings: v })}
                  />
                  <JsonField
                    label={t("fields.triggerFlags")}
                    value={draft.trigger_flags}
                    onChange={(v) => setDraft({ ...draft, trigger_flags: v })}
                  />
                  <label className="flex items-center gap-2 text-sm sm:col-span-2">
                    <input
                      type="checkbox"
                      checked={draft.is_active}
                      onChange={(e) => setDraft({ ...draft, is_active: e.target.checked })}
                    />
                    {t("fields.active")}
                  </label>
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={draft.is_turnover_fine}
                      onChange={(e) => setDraft({ ...draft, is_turnover_fine: e.target.checked })}
                    />
                    {t("fields.turnoverFine")}
                  </label>
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={draft.is_context_only}
                      onChange={(e) => setDraft({ ...draft, is_context_only: e.target.checked })}
                    />
                    {t("fields.contextOnly")}
                  </label>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

function Field({
  label,
  value,
  onChange,
  className,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  className?: string;
}) {
  return (
    <div className={className}>
      <label className="mb-1 block text-xs font-medium">{label}</label>
      <input className={fieldClass} value={value} onChange={(e) => onChange(e.target.value)} />
    </div>
  );
}

function JsonField({
  label,
  value,
  onChange,
}: {
  label: string;
  value: Record<string, boolean>;
  onChange: (v: Record<string, boolean>) => void;
}) {
  const [text, setText] = useState(JSON.stringify(value, null, 2));
  const [invalid, setInvalid] = useState(false);

  useEffect(() => {
    setText(JSON.stringify(value, null, 2));
  }, [value]);

  return (
    <div className="sm:col-span-2">
      <label className="mb-1 block text-xs font-medium">{label}</label>
      <textarea
        rows={3}
        className={cn(fieldClass, "font-mono text-xs", invalid && "border-red-300")}
        value={text}
        onChange={(e) => {
          setText(e.target.value);
          try {
            const parsed = JSON.parse(e.target.value) as Record<string, boolean>;
            onChange(parsed);
            setInvalid(false);
          } catch {
            setInvalid(true);
          }
        }}
      />
    </div>
  );
}
