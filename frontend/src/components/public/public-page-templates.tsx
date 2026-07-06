import type { PublicPage } from "@/lib/api-public-pages";
import { CalculatorLandingPage } from "@/components/public/calculator-landing/CalculatorLandingPage";
import { StrategyLandingPage } from "@/components/public/strategy-landing/StrategyLandingPage";
import { ProposalLandingPage } from "@/components/public/proposal-landing/ProposalLandingPage";
import { AuditLandingPage } from "@/components/public/audit-landing/AuditLandingPage";
import { LegalScanLandingPage } from "@/components/public/legal-scan-landing/LegalScanLandingPage";
import { AutoRtkLandingPage } from "@/components/public/auto-rtk-landing/AutoRtkLandingPage";
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
    case "legal-scan-landing":
      return <LegalScanLandingPage page={page} />;
    case "auto-rtk-landing":
      return <AutoRtkLandingPage page={page} />;
    default:
      return <PublicPageView page={page} />;
  }
}

export function isFullBleedPublicPage(template: string) {
  return (
    template === "calculator-landing" ||
    template === "strategy-landing" ||
    template === "proposal-landing" ||
    template === "audit-landing" ||
    template === "legal-scan-landing" ||
    template === "auto-rtk-landing"
  );
}
