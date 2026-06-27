import type { PublicPage } from "@/lib/api-public-pages";
import { CalculatorLandingPage } from "@/components/public/calculator-landing/CalculatorLandingPage";
import { PublicPageView } from "@/components/public/PublicPageView";

export function renderPublicPage(page: PublicPage) {
  switch (page.template) {
    case "calculator-landing":
      return <CalculatorLandingPage page={page} />;
    default:
      return <PublicPageView page={page} />;
  }
}

export function isFullBleedPublicPage(template: string) {
  return template === "calculator-landing";
}
