import { ProposalSubnav } from "@/components/tools/ProposalSubnav";

export default function ProposalLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-6">
      <ProposalSubnav />
      {children}
    </div>
  );
}
