import { Suspense } from "react";
import { CalculatorWizard } from "@/components/tools/CalculatorWizard";

export default function CalculatorPage() {
  return (
    <Suspense fallback={null}>
      <CalculatorWizard />
    </Suspense>
  );
}
