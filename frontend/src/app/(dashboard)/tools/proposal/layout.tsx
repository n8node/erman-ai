import { ProposalSubnav } from "@/components/tools/ProposalSubnav";
import { ProposalTooltipProvider } from "@/components/tools/ProposalTooltipProvider";

export default function ProposalLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <ProposalTooltipProvider>
      <div className="space-y-6">
        <ProposalSubnav />
        {children}
      </div>
    </ProposalTooltipProvider>
  );
}
