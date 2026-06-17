package model

type LegalScanSiteFeatures struct {
	Forms              bool `json:"forms"`
	TrafficFromAds     bool `json:"traffic_from_ads"`
	OnlineSales        bool `json:"online_sales"`
	ForeignServices    bool `json:"foreign_services"`
}

type LegalScanInput struct {
	URL          string                `json:"url"`
	Industry     string                `json:"industry"`
	CompanySize  string                `json:"company_size"`
	SiteFeatures LegalScanSiteFeatures `json:"site_features"`
	Locale       string                `json:"locale,omitempty"`
}

type LegalScanFindings struct {
	SSL               bool     `json:"ssl"`
	PrivacyPolicy     bool     `json:"privacy_policy"`
	CookieBanner      bool     `json:"cookie_banner"`
	CookiePolicy      bool     `json:"cookie_policy"`
	FormConsent       bool     `json:"form_consent"`
	RequisitesINN     bool     `json:"requisites_inn"`
	Contacts          bool     `json:"contacts"`
	Offer             bool     `json:"offer"`
	Terms             bool     `json:"terms"`
	ConsentWithdrawal bool     `json:"consent_withdrawal"`
	AdMarking         bool     `json:"ad_marking"`
	Trackers          []string `json:"trackers"`
	FormsCollectPD    bool     `json:"forms_collect_pd"`
	FormsUnencrypted  bool     `json:"forms_unencrypted"`
	ForeignTrackers   bool     `json:"foreign_trackers"`
	HasTrackers       bool     `json:"has_trackers"`
}

type LegalScanCheckItem struct {
	Key       string   `json:"key"`
	Label     string   `json:"label"`
	Status    string   `json:"status"` // pending, running, ok, risk
	Evidence  string   `json:"evidence,omitempty"`
	PageURLs  []string `json:"page_urls,omitempty"`
	FoundData []string `json:"found_data,omitempty"`
}

type LegalScanLayer1 struct {
	FinalURL  string               `json:"final_url"`
	Findings  LegalScanFindings    `json:"findings"`
	Checklist []LegalScanCheckItem `json:"checklist"`
}

type LegalScanRiskItem struct {
	RiskID      string   `json:"risk_id"`
	Title       string   `json:"title"`
	Explanation string   `json:"explanation"`
	Article     string   `json:"article"`
	FineText    string   `json:"fine_text"`
	Severity    string   `json:"severity"`
	HowToFix    string   `json:"how_to_fix"`
	Evidence    string   `json:"evidence,omitempty"`
	PageURLs    []string `json:"page_urls,omitempty"`
	FoundData   []string `json:"found_data,omitempty"`
}

type LegalScanSummary struct {
	RisksCount       int    `json:"risks_count"`
	FineMinTotal     int    `json:"fine_min_total"`
	FineMaxTotal     int    `json:"fine_max_total"`
	TurnoverFineNote string `json:"turnover_fine_note"`
}

type LegalScanOutput struct {
	Layer1       LegalScanLayer1     `json:"layer1"`
	Summary      LegalScanSummary    `json:"summary"`
	Risks        []LegalScanRiskItem `json:"risks"`
	IndustryNote string              `json:"industry_note"`
	Disclaimer   string              `json:"disclaimer"`
}

type LegalScanLLMOutput struct {
	Summary      LegalScanSummary    `json:"summary"`
	Risks        []LegalScanRiskItem `json:"risks"`
	IndustryNote string              `json:"industry_note"`
	Disclaimer   string              `json:"disclaimer"`
}
