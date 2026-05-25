import { ProposalSubnav } from "@/components/tools/ProposalSubnav";
import { TooltipProvider } from "@/components/ui/HelpTooltip";
import { PROPOSAL_TOOLTIP_FALLBACKS_RU } from "@/lib/proposal-tooltip-fallbacks";

export default function ProposalLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <TooltipProvider prefix="proposal" fallbacks={PROPOSAL_TOOLTIP_FALLBACKS_RU}>
      <div className="space-y-6">
        <ProposalSubnav />
        {children}
      </div>
    </TooltipProvider>
  );
}
