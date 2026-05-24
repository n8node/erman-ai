import { CalculatorSubnav } from "@/components/tools/CalculatorSubnav";

export default function CalculatorLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-6">
      <CalculatorSubnav />
      {children}
    </div>
  );
}
