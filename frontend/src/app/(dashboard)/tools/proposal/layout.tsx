import { ProposalSubnav } from "@/components/tools/ProposalSubnav";
import { TooltipProvider } from "@/components/ui/HelpTooltip";

export default function ProposalLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <TooltipProvider prefix="proposal">
      <div className="space-y-6">
        <ProposalSubnav />
        {children}
      </div>
    </TooltipProvider>
  );
}
