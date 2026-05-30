import { Suspense } from "react";
import { BillingPlansView } from "@/components/billing/BillingPlansView";

export default function BillingPage() {
  return (
    <Suspense fallback={<p className="text-sm text-text2">…</p>}>
      <BillingPlansView />
    </Suspense>
  );
}
