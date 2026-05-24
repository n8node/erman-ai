"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { submitLead } from "@/lib/api";

type Props = {
  runId: string;
  defaultEmail?: string;
};

export function LeadForm({ runId, defaultEmail }: Props) {
  const t = useTranslations("calculator.lead");
  const [name, setName] = useState("");
  const [email, setEmail] = useState(defaultEmail || "");
  const [phone, setPhone] = useState("");
  const [company, setCompany] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await submitLead({
        run_id: runId,
        name,
        email,
        phone: phone || undefined,
        company: company || undefined,
        message: message || undefined,
      });
      setDone(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("failed"));
    } finally {
      setLoading(false);
    }
  }

  if (done) {
    return (
      <div className="rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
        {t("success")}
      </div>
    );
  }

  const fieldClass =
    "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

  return (
    <div className="border-t border-border pt-4">
      <h3 className="text-sm font-medium">{t("title")}</h3>
      <p className="mt-1 text-xs text-text2">{t("subtitle")}</p>
      <form onSubmit={handleSubmit} className="mt-4 grid gap-3 sm:grid-cols-2">
        {error && (
          <div className="sm:col-span-2 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
            {error}
          </div>
        )}
        <div>
          <label className="mb-1 block text-xs font-medium">{t("name")}</label>
          <input required className={fieldClass} value={name} onChange={(e) => setName(e.target.value)} />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium">{t("email")}</label>
          <input required type="email" className={fieldClass} value={email} onChange={(e) => setEmail(e.target.value)} />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium">{t("phone")}</label>
          <input className={fieldClass} value={phone} onChange={(e) => setPhone(e.target.value)} />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium">{t("company")}</label>
          <input className={fieldClass} value={company} onChange={(e) => setCompany(e.target.value)} />
        </div>
        <div className="sm:col-span-2">
          <label className="mb-1 block text-xs font-medium">{t("message")}</label>
          <textarea className={fieldClass} rows={3} value={message} onChange={(e) => setMessage(e.target.value)} />
        </div>
        <div className="sm:col-span-2">
          <button
            type="submit"
            disabled={loading}
            className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-60"
          >
            {loading ? t("sending") : t("submit")}
          </button>
        </div>
      </form>
    </div>
  );
}
