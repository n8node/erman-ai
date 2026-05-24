import { CalculatorSubnav } from "@/components/tools/CalculatorSubnav";
import { TooltipProvider } from "@/components/ui/HelpTooltip";

export default function CalculatorLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <TooltipProvider prefix="calculator">
      <div className="space-y-6">
        <CalculatorSubnav />
        {children}
      </div>
    </TooltipProvider>
  );
}
