import { AuditSubnav } from "@/components/tools/AuditSubnav";

export default function AuditLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="space-y-6">
      <AuditSubnav />
      {children}
    </div>
  );
}
