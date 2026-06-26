import { Suspense } from "react";
import { ConsultationBookingView } from "@/components/consultations/ConsultationBookingView";

export default function ConsultationsPage() {
  return (
    <Suspense fallback={<p className="text-sm text-text2">…</p>}>
      <ConsultationBookingView />
    </Suspense>
  );
}
