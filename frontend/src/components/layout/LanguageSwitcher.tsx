"use client";

import { useEffect, useRef, useState } from "react";
import { useLocale } from "next-intl";
import { useRouter } from "next/navigation";
import { ChevronDown } from "lucide-react";
import { LOCALE_META, SUPPORTED_LOCALES, type AppLocale } from "@/i18n/locales";
import { updateMe } from "@/lib/api";
import { cn } from "@/lib/utils";

type Props = {
  userEmail?: string;
  className?: string;
};

export function LanguageSwitcher({ userEmail, className }: Props) {
  const locale = useLocale() as AppLocale;
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function onPointerDown(e: MouseEvent) {
      if (!rootRef.current?.contains(e.target as Node)) setOpen(false);
    }
    document.addEventListener("mousedown", onPointerDown);
    return () => document.removeEventListener("mousedown", onPointerDown);
  }, []);

  async function select(next: AppLocale) {
    if (next === locale || saving) return;
    setSaving(true);
    setOpen(false);
    try {
      await fetch("/dashboard/api/locale", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ locale: next }),
      });
      if (userEmail) {
        await updateMe(userEmail, next).catch(() => undefined);
      }
      router.refresh();
    } finally {
      setSaving(false);
    }
  }

  const current = LOCALE_META[locale] ?? LOCALE_META.ru;

  return (
    <div ref={rootRef} className={cn("relative", className)}>
      <button
        type="button"
        disabled={saving}
        onClick={() => setOpen((v) => !v)}
        className="inline-flex items-center gap-2 rounded-md border border-border2 bg-bg px-2.5 py-1.5 text-[13px] text-text hover:bg-bg2 disabled:opacity-60"
        aria-expanded={open}
        aria-haspopup="listbox"
      >
        <span aria-hidden>{current.flag}</span>
        <span>{current.nativeName}</span>
        <ChevronDown size={14} className="text-text3" />
      </button>
      {open && (
        <ul
          role="listbox"
          className="absolute right-0 z-50 mt-1 min-w-[180px] overflow-hidden rounded-lg border border-border bg-bg py-1 shadow-lg"
        >
          {SUPPORTED_LOCALES.map((code) => {
            const meta = LOCALE_META[code];
            return (
              <li key={code}>
                <button
                  type="button"
                  role="option"
                  aria-selected={code === locale}
                  onClick={() => void select(code)}
                  className={cn(
                    "flex w-full items-center gap-2 px-3 py-2 text-left text-[13px] hover:bg-bg2",
                    code === locale && "bg-bg2 font-medium"
                  )}
                >
                  <span aria-hidden>{meta.flag}</span>
                  <span>{meta.nativeName}</span>
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
