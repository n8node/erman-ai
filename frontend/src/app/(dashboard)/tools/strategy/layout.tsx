import { StrategySubnav } from "@/components/tools/StrategySubnav";

export default function StrategyLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-6">
      <StrategySubnav />
      {children}
    </div>
  );
}
