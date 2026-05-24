import { Suspense } from "react";
import { StrategyWizard } from "@/components/tools/StrategyWizard";

export default function StrategyPage() {
  return (
    <Suspense fallback={null}>
      <StrategyWizard />
    </Suspense>
  );
}
