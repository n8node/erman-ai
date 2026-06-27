import type { PublicPage } from "@/lib/api-public-pages";
import { CalculatorLandingPage } from "@/components/public/calculator-landing/CalculatorLandingPage";
import { StrategyLandingPage } from "@/components/public/strategy-landing/StrategyLandingPage";
import { ProposalLandingPage } from "@/components/public/proposal-landing/ProposalLandingPage";
import { AuditLandingPage } from "@/components/public/audit-landing/AuditLandingPage";
import { PublicPageView } from "@/components/public/PublicPageView";

export function renderPublicPage(page: PublicPage) {
  switch (page.template) {
    case "calculator-landing":
      return <CalculatorLandingPage page={page} />;
    case "strategy-landing":
      return <StrategyLandingPage page={page} />;
    case "proposal-landing":
      return <ProposalLandingPage page={page} />;
    case "audit-landing":
      return <AuditLandingPage page={page} />;
    default:
      return <PublicPageView page={page} />;
  }
}

export function isFullBleedPublicPage(template: string) {
  return (
    template === "calculator-landing" ||
    template === "strategy-landing" ||
    template === "proposal-landing" ||
    template === "audit-landing"
  );
}
