package service

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/erman-ai/erman-ai/internal/model"
	"html/template"
)

//go:embed templates/calculator-report.html
var calculatorReportHTML embed.FS

type pdfFinanceField struct {
	Label string
	Value string
}

type pdfProcessStep struct {
	Operation string
	Executor  string
	Minutes   string
	Rate      string
	Cost      string
}

type calculatorPDFData struct {
	Lang     string
	Labels   pdfLabels
	LangAttr string

	ProcessName string

	ModeLabel        string
	ShowProcessTable bool
	Steps            []pdfProcessStep
	TotalMinutes     string
	TotalCostPerUnit string
	UnitsPerMonth    string
	AutomationPct    string
	HmDirect         string
	HmPreview        string

	ErrorRateBefore  string
	ThroughputBefore string

	FinanceFields []pdfFinanceField

	NetBenefit         string
	Payback            string
	FTE                string
	ROI                string
	Recommendation     string
	RecommendationText string
	ShowKPI            bool
	KPIRows            []model.KPIRow

	TCO       string
	NPV       string
	HmDetails string
}

func GenerateCalculatorPDF(input model.CalculatorInput, output model.CalculatorOutput, locale string) ([]byte, error) {
	if locale != "en" {
		locale = "ru"
	}

	html, err := renderCalculatorReportHTML(input, output, locale)
	if err != nil {
		return nil, err
	}
	return htmlToPDF(html)
}

func renderCalculatorReportHTML(input model.CalculatorInput, output model.CalculatorOutput, locale string) (string, error) {
	labels := pdfLabelsForLocale(locale)
	hm := output.HoursSavedMonth
	if hm <= 0 {
		hm = ResolveHoursSaved(input)
	}

	totalMin := TotalMinutesPerUnit(input.ProcessSteps)
	var costPerUnit float64
	for _, s := range input.ProcessSteps {
		costPerUnit += s.MinutesPerUnit * s.RateRubPerMin
	}

	data := calculatorPDFData{
		Lang:             locale,
		LangAttr:         locale,
		Labels:           labels,
		ProcessName:      input.ProcessName,
		ShowProcessTable: input.HmMode == "process",
		HmPreview:        formatHm(hm, labels.HoursMonth),
		ErrorRateBefore:  formatPct(input.ErrorRateBeforePct, locale),
		ThroughputBefore: formatLocaleNum(input.ThroughputBeforePerDay, 0, locale),
		NetBenefit:       formatLocaleMoney(output.NetBenefitMonthly, locale),
		Payback:          formatPaybackPDF(output.PaybackMonths, labels.Months, locale),
		FTE:              formatLocaleNum(output.FTE, 2, locale),
		ROI:              formatLocaleNum(output.ROIHorizonPct, 1, locale) + "%",
		Recommendation:     string(output.Recommendation),
		RecommendationText: output.RecommendationText,
		ShowKPI:            len(output.KPIRows) > 0,
		KPIRows:            output.KPIRows,
		TCO:                formatLocaleMoney(output.TCOHorizon, locale),
		NPV:                formatLocaleMoney(output.NPV, locale),
		HmDetails:          formatHm(hm, labels.HoursMonth),
	}

	if input.HmMode == "process" {
		data.ModeLabel = labels.ModeProcess
		data.TotalMinutes = formatLocaleNum(totalMin, 1, locale)
		data.TotalCostPerUnit = formatLocaleMoney(costPerUnit, locale)
		data.UnitsPerMonth = formatLocaleNum(input.UnitsPerMonth, 0, locale)
		data.AutomationPct = formatPct(input.AutomationPct, locale)
		for _, step := range input.ProcessSteps {
			data.Steps = append(data.Steps, pdfProcessStep{
				Operation: step.Operation,
				Executor:  step.Executor,
				Minutes:   formatLocaleNum(step.MinutesPerUnit, 1, locale),
				Rate:      formatLocaleMoney(step.RateRubPerMin, locale),
				Cost:      formatLocaleMoney(step.MinutesPerUnit*step.RateRubPerMin, locale),
			})
		}
	} else {
		data.ModeLabel = labels.ModeDirect
		data.HmDirect = formatHm(hm, labels.HoursMonth)
	}

	data.FinanceFields = []pdfFinanceField{
		{Label: labels.HmLabel, Value: formatHm(hm, labels.HoursMonth)},
		{Label: labels.Horizon, Value: fmt.Sprintf("%d %s", input.HorizonMonths, labels.Months)},
		{Label: labels.Ch, Value: formatLocaleMoney(input.HourlyCostLoaded, locale)},
		{Label: labels.Om, Value: formatLocaleMoney(input.MonthlySolutionCost, locale)},
		{Label: labels.Capex, Value: formatLocaleMoney(input.Capex, locale)},
		{Label: labels.Sm, Value: formatLocaleMoney(input.OtherBenefitMonthly, locale)},
		{Label: labels.Utilization, Value: formatLocaleNum(input.Utilization, 2, locale)},
		{Label: labels.Discount, Value: formatPct(input.DiscountRateAnnual*100, locale)},
	}

	tmpl, err := template.ParseFS(calculatorReportHTML, "templates/calculator-report.html")
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func htmlToPDF(html string) ([]byte, error) {
	execPath := chromeExecPath()
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	if execPath != "" {
		opts = append(opts, chromedp.ExecPath(execPath))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	ctx, cancelTimeout := context.WithTimeout(ctx, 60*time.Second)
	defer cancelTimeout()

	var pdfBuf []byte
	dataURL := "data:text/html;charset=utf-8;base64," + base64.StdEncoding.EncodeToString([]byte(html))
	if err := chromedp.Run(ctx,
		chromedp.Navigate(dataURL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(800*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				WithMarginTop(0.35).
				WithMarginBottom(0.35).
				WithMarginLeft(0.35).
				WithMarginRight(0.35).
				Do(ctx)
			return err
		}),
	); err != nil {
		return nil, fmt.Errorf("pdf render: %w", err)
	}
	return pdfBuf, nil
}

func chromeExecPath() string {
	if p := strings.TrimSpace(os.Getenv("CHROME_PATH")); p != "" {
		return p
	}
	candidates := []string{
		"/usr/bin/chromium-browser",
		"/usr/bin/chromium",
		"/usr/bin/google-chrome-stable",
		"/usr/bin/google-chrome",
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func formatHm(v float64, unit string) string {
	if !isFinite(v) {
		return "—"
	}
	return formatLocaleNum(v, 1, "ru") + " " + unit
}

func formatPaybackPDF(v float64, unit, locale string) string {
	if !isFinite(v) || v <= 0 || v > 1e6 {
		return "—"
	}
	return formatLocaleNum(v, 1, locale) + " " + unit
}

func formatPct(v float64, locale string) string {
	if !isFinite(v) {
		return "—"
	}
	decimals := 0
	if locale == "ru" && v != math.Trunc(v) {
		decimals = 1
	}
	return formatLocaleNum(v, decimals, locale) + "%"
}

func formatLocaleNum(v float64, decimals int, locale string) string {
	if !isFinite(v) {
		return "—"
	}
	p := math.Pow(10, float64(decimals))
	rounded := math.Round(v*p) / p

	var s string
	if decimals == 0 {
		s = fmt.Sprintf("%.0f", rounded)
	} else {
		s = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.*f", decimals, rounded), "0"), ".")
	}

	if locale == "ru" {
		return addThousandsSep(s)
	}
	parts := strings.Split(s, ".")
	parts[0] = addThousandsSep(parts[0])
	return strings.Join(parts, ".")
}

func formatLocaleMoney(v float64, locale string) string {
	if !isFinite(v) {
		return "—"
	}
	num := formatLocaleNum(v, 0, locale)
	if locale == "en" {
		return num + " RUB"
	}
	return num + " ₽"
}

func addThousandsSep(s string) string {
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	var out []byte
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, byte(c))
	}
	result := string(out)
	if len(parts) == 2 {
		result += "." + parts[1]
	}
	if neg {
		result = "-" + result
	}
	return result
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func ParseCalculatorRun(run *model.ToolRun) (model.CalculatorInput, model.CalculatorOutput, error) {
	var in model.CalculatorInput
	var out model.CalculatorOutput
	if err := json.Unmarshal(run.Input, &in); err != nil {
		return in, out, err
	}
	if err := json.Unmarshal(run.Output, &out); err != nil {
		return in, out, err
	}
	return in, out, nil
}
