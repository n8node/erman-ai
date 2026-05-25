"use client";

import { useTranslations } from "next-intl";
import { Mail, MessageSquare, Sparkles } from "lucide-react";
import {
  normalizeProposalScenario,
  type ProposalInput,
  type ProposalScenario,
} from "@/lib/api-proposal";
import { cn } from "@/lib/utils";

const SCENARIO_ICONS: Record<ProposalScenario, typeof Mail> = {
  after_contact: MessageSquare,
  cold_outreach: Mail,
  proactive_offer: Sparkles,
};

type Props = {
  input: ProposalInput;
  step?: 1 | 2 | 3;
  className?: string;
};

export function ProposalScenarioBar({ input, step, className }: Props) {
  const t = useTranslations("proposal");
  const scenario = normalizeProposalScenario(input.proposal_scenario);
  const Icon = SCENARIO_ICONS[scenario];

  return (
    <div
      className={cn(
        "rounded-xl border border-ai/30 bg-ai-bg px-4 py-3",
        className
      )}
    >
      <div className="flex flex-wrap items-start gap-3">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-bg text-ai">
          <Icon size={18} strokeWidth={2} />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="text-sm font-medium text-text">{t(`scenarios.${scenario}.title`)}</p>
            {step != null && (
              <span className="rounded-md bg-bg px-2 py-0.5 text-[10px] font-medium uppercase tracking-wider text-text3">
                {t(`steps.${step}`)}
              </span>
            )}
            {!input.include_pricing && (
              <span className="rounded-md border border-border2 bg-bg px-2 py-0.5 text-[10px] font-medium text-text2">
                {t("form.includePricingLabelOff")}
              </span>
            )}
          </div>
          <p className="mt-1 text-xs leading-relaxed text-text2">{t(`scenarios.${scenario}.description`)}</p>
        </div>
      </div>
    </div>
  );
}
