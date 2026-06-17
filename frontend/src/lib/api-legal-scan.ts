export type LegalScanSiteFeatures = {
  forms: boolean;
  traffic_from_ads: boolean;
  online_sales: boolean;
  foreign_services: boolean;
};

export type LegalScanInput = {
  url: string;
  industry: string;
  company_size: string;
  site_features: LegalScanSiteFeatures;
};

export type LegalScanFindings = {
  ssl: boolean;
  privacy_policy: boolean;
  cookie_banner: boolean;
  cookie_policy: boolean;
  form_consent: boolean;
  requisites_inn: boolean;
  contacts: boolean;
  offer: boolean;
  terms: boolean;
  consent_withdrawal: boolean;
  ad_marking: boolean;
  trackers: string[];
  forms_collect_pd: boolean;
  forms_unencrypted: boolean;
  foreign_trackers: boolean;
  has_trackers: boolean;
};

export type LegalScanCheckItem = {
  key: string;
  label: string;
  status: "pending" | "running" | "ok" | "risk";
  evidence?: string;
  page_urls?: string[];
  found_data?: string[];
};

export type LegalScanLayer1 = {
  final_url: string;
  findings: LegalScanFindings;
  checklist: LegalScanCheckItem[];
};

export type LegalScanRiskItem = {
  risk_id: string;
  title: string;
  explanation: string;
  article: string;
  fine_text: string;
  severity: "high" | "medium" | "low";
  how_to_fix: string;
  evidence?: string;
  page_urls?: string[];
  found_data?: string[];
};

export type LegalScanSummary = {
  risks_count: number;
  fine_min_total: number;
  fine_max_total: number;
  turnover_fine_note: string;
};

export type LegalScanOutput = {
  layer1: LegalScanLayer1;
  summary: LegalScanSummary;
  risks: LegalScanRiskItem[];
  industry_note: string;
  disclaimer: string;
};

export const LEGAL_SCAN_INDUSTRY_OPTIONS = [
  { id: "ecommerce", labelKey: "industries.ecommerce" },
  { id: "services", labelKey: "industries.services" },
  { id: "medicine", labelKey: "industries.medicine" },
  { id: "finance", labelKey: "industries.finance" },
  { id: "education", labelKey: "industries.education" },
  { id: "other", labelKey: "industries.other" },
] as const;

export const LEGAL_SCAN_SIZE_OPTIONS = [
  { id: "sole", labelKey: "sizes.sole" },
  { id: "small", labelKey: "sizes.small" },
  { id: "medium", labelKey: "sizes.medium" },
] as const;

export const LEGAL_SCAN_FEATURE_CHIPS = [
  { key: "forms" as const, labelKey: "features.forms" },
  { key: "traffic_from_ads" as const, labelKey: "features.ads" },
  { key: "online_sales" as const, labelKey: "features.sales" },
  { key: "foreign_services" as const, labelKey: "features.foreign" },
];

export const DEFAULT_LEGAL_SCAN_INPUT: LegalScanInput = {
  url: "",
  industry: "services",
  company_size: "small",
  site_features: {
    forms: true,
    traffic_from_ads: false,
    online_sales: false,
    foreign_services: false,
  },
};

export function isLegalScanInputValid(input: LegalScanInput): boolean {
  return input.url.trim().length >= 4 && input.industry.trim().length > 0;
}

export function buildProposalProblemFromLegalScan(
  url: string,
  output: LegalScanOutput
): string {
  const lines = output.risks.map(
    (r, i) => `${i + 1}. ${r.title} — ${r.explanation}`
  );
  const header = `На сайте ${url} выявлено ${output.summary.risks_count} юридических рисков.`;
  return [header, ...lines].join("\n");
}

/** Maps layer1 checklist keys to risk_id entries in the report. */
export const CHECKLIST_TO_RISK_IDS: Record<string, string[]> = {
  ssl: ["no_ssl"],
  privacy: ["no_privacy_policy"],
  cookie: ["no_cookie_banner"],
  consent: ["no_consent"],
  req: ["no_requisites"],
  offer: ["no_offer"],
  admark: ["no_ad_marking"],
};

export function findRiskForCheckItem(
  checkKey: string,
  risks: LegalScanRiskItem[]
): LegalScanRiskItem | undefined {
  const ids = CHECKLIST_TO_RISK_IDS[checkKey];
  if (!ids?.length) return undefined;
  return risks.find((r) => ids.includes(r.risk_id));
}
