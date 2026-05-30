"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { submitProposalRequest } from "@/lib/api";

type Props = {
  runId: string;
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
};

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

export function ProposalRequestModal({ runId, open, onClose, onSuccess }: Props) {
  const t = useTranslations("calculator.proposalModal");
  const [requesterName, setRequesterName] = useState("");
  const [telegram, setTelegram] = useState("");
  const [businessNote, setBusinessNote] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) return;
    setError("");
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose]);

  if (!open) return null;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await submitProposalRequest({
        run_id: runId,
        requester_name: requesterName.trim(),
        telegram: telegram.trim(),
        business_note: businessNote.trim(),
      });
      onSuccess();
      onClose();
    } catch (err) {
      const message = err instanceof Error ? err.message : t("failed");
      if (message === "proposal request already submitted") {
        onSuccess();
        onClose();
        return;
      }
      setError(message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      onClick={onClose}
      role="presentation"
    >
      <div
        className="w-full max-w-lg rounded-xl border border-border bg-bg p-6 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="proposal-modal-title"
      >
        <h2 id="proposal-modal-title" className="text-base font-medium">
          {t("title")}
        </h2>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>

        <form onSubmit={handleSubmit} className="mt-5 space-y-4">
          {error && (
            <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
              {error}
            </div>
          )}
          <div>
            <label className="mb-1.5 block text-xs font-medium text-text2">
              {t("name")} *
            </label>
            <input
              required
              className={fieldClass}
              value={requesterName}
              onChange={(e) => setRequesterName(e.target.value)}
              placeholder={t("namePlaceholder")}
            />
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium text-text2">
              {t("telegram")} *
            </label>
            <input
              required
              className={fieldClass}
              value={telegram}
              onChange={(e) => setTelegram(e.target.value)}
              placeholder={t("telegramPlaceholder")}
            />
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium text-text2">
              {t("businessNote")} *
            </label>
            <textarea
              required
              rows={4}
              className={fieldClass}
              value={businessNote}
              onChange={(e) => setBusinessNote(e.target.value)}
              placeholder={t("businessNotePlaceholder")}
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2 disabled:opacity-50"
            >
              {t("cancel")}
            </button>
            <button
              type="submit"
              disabled={loading}
              className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
            >
              {loading ? t("submitting") : t("submit")}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
