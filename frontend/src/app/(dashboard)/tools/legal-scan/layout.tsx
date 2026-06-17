import { LegalScanSubnav } from "@/components/tools/LegalScanSubnav";

export default function LegalScanLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="space-y-6">
      <LegalScanSubnav />
      {children}
    </div>
  );
}
