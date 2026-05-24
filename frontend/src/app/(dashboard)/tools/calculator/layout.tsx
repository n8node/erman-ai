import { CalculatorSubnav } from "@/components/tools/CalculatorSubnav";
import { BudgetConfigProvider } from "@/components/tools/BudgetConfigProvider";
import { TooltipProvider } from "@/components/ui/HelpTooltip";

export default function CalculatorLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <TooltipProvider prefix="calculator">
      <BudgetConfigProvider>
        <div className="space-y-6">
          <CalculatorSubnav />
          {children}
        </div>
      </BudgetConfigProvider>
    </TooltipProvider>
  );
}
