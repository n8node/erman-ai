"use client";

import { useLocale, useTranslations } from "next-intl";
import { useMemo } from "react";
import { TooltipProvider } from "@/components/ui/HelpTooltip";

const TOOLTIP_KEYS = [
  "scenario",
  "scenario_after_contact",
  "scenario_cold_outreach",
  "scenario_proactive_offer",
  "prior_contact_summary",
  "problem_source",
  "include_pricing",
  "client_company",
  "client_contact",
  "client_industry",
  "client_problem",
  "calculator_run",
  "solution_name",
  "solution_description",
  "deliverables",
  "project_cost",
  "timeline_weeks",
  "payment_schedule",
  "sender_company",
  "sender_contact",
  "sender_phone",
  "sender_email",
] as const;

export function ProposalTooltipProvider({ children }: { children: React.ReactNode }) {
  const locale = useLocale();
  const t = useTranslations("proposal.tooltipsFallback");

  const fallbacks = useMemo(() => {
    const map: Record<string, string> = {};
    for (const key of TOOLTIP_KEYS) {
      map[`proposal.wizard.${key}`] = t(key);
    }
    return map;
  }, [t]);

  return (
    <TooltipProvider prefix="proposal" locale={locale} fallbacks={fallbacks}>
      {children}
    </TooltipProvider>
  );
}
