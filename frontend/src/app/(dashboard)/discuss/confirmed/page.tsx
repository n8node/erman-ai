"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { verifyProjectInquiry } from "@/lib/api";
import { updateGuestDiscussHistoryStatus } from "@/lib/submission-limits";

function DiscussConfirmedInner() {
  const t = useTranslations("discuss.confirm");
  const searchParams = useSearchParams();
  const token = searchParams.get("token") || "";
  const [state, setState] = useState<"loading" | "ok" | "error">("loading");

  useEffect(() => {
    if (!token) {
      setState("error");
      return;
    }
    verifyProjectInquiry(token)
      .then((res) => {
        if (res.id) {
          updateGuestDiscussHistoryStatus(res.id, res.status || "new");
        }
        setState("ok");
      })
      .catch(() => setState("error"));
  }, [token]);

  if (state === "loading") {
    return <p className="text-sm text-text2">{t("loading")}</p>;
  }

  if (state === "error") {
    return (
      <div className="mx-auto max-w-lg rounded-xl border border-red-200 bg-red-50 p-6 text-sm text-red-900">
        <h1 className="text-base font-medium">{t("errorTitle")}</h1>
        <p className="mt-2">{t("errorBody")}</p>
        <Link href="/discuss" className="mt-4 inline-block text-accent hover:underline">
          {t("back")}
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-lg rounded-xl border border-green-200 bg-green-50 p-6 text-sm text-green-900">
      <h1 className="text-base font-medium">{t("title")}</h1>
      <p className="mt-2">{t("body")}</p>
      <Link href="/" className="mt-4 inline-block text-accent hover:underline">
        {t("dashboard")}
      </Link>
    </div>
  );
}

export default function DiscussConfirmedPage() {
  return (
    <Suspense fallback={<p className="text-sm text-text2">…</p>}>
      <DiscussConfirmedInner />
    </Suspense>
  );
}
