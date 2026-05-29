import { AuditHistory } from "@/components/tools/AuditHistory";
import { GuestHistoryPlaceholder } from "@/components/layout/GuestHistoryPlaceholder";

export default function AuditHistoryPage() {
  return (
    <GuestHistoryPlaceholder>
      <div className="mx-auto max-w-5xl space-y-4">
        <AuditHistory />
      </div>
    </GuestHistoryPlaceholder>
  );
}
