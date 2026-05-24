"use client";

import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { fetchCalculatorBudgetConfig } from "@/lib/api";
import {
  DEFAULT_CALCULATOR_BUDGET_CONFIG,
  type CalculatorBudgetConfig,
} from "@/lib/calculator-budget-config";

type BudgetConfigContextValue = {
  config: CalculatorBudgetConfig;
  loading: boolean;
};

const BudgetConfigContext = createContext<BudgetConfigContextValue>({
  config: DEFAULT_CALCULATOR_BUDGET_CONFIG,
  loading: true,
});

export function BudgetConfigProvider({ children }: { children: ReactNode }) {
  const [config, setConfig] = useState<CalculatorBudgetConfig>(DEFAULT_CALCULATOR_BUDGET_CONFIG);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchCalculatorBudgetConfig()
      .then((data) => setConfig(data.config))
      .catch(() => setConfig(DEFAULT_CALCULATOR_BUDGET_CONFIG))
      .finally(() => setLoading(false));
  }, []);

  return (
    <BudgetConfigContext.Provider value={{ config, loading }}>
      {children}
    </BudgetConfigContext.Provider>
  );
}

export function useBudgetConfig() {
  return useContext(BudgetConfigContext);
}
