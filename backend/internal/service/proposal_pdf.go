package service

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
)

//go:embed templates/proposal-report.html
var proposalReportHTML embed.FS

type proposalPDFLabels struct {
	Title             string
	Client            string
	Timeline          string
	Weeks             string
	Date              string
	Greeting          string
	TaskUnderstanding string
	ProposedSolution  string
	Scope             string
	Included          string
	Excluded          string
	TimelineSection   string
	Phase             string
	Duration          string
	Description       string
	Cost              string
	TotalCost         string
	Payment           string
	WhyUs             string
	NextStep          string
	Disclaimer        string
}

type proposalPDFTimelineRow struct {
	Title         string
	DurationWeeks int
	Description   string
}

type proposalPDFData struct {
	LangAttr string
	Labels   proposalPDFLabels
	ShowPricing bool

	ClientCompany       string
	ClientContact       string
	SolutionName        string
	TimelineWeeks       int
	GeneratedDate       string
	SenderCompany       string
	SenderContact       string
	SenderPhone         string
	SenderEmail         string
	ProjectCostFormatted string

	Greeting          template.HTML
	TaskUnderstanding template.HTML
	ProposedSolution  template.HTML
	CostSummary       template.HTML
	PaymentTerms      template.HTML
	WhyUs             template.HTML
	NextStep          template.HTML

	ScopeIncluded []string
	ScopeExcluded []string
	Timeline      []proposalPDFTimelineRow
}

func GenerateProposalPDF(input model.ProposalInput, output model.ProposalOutput, locale string) ([]byte, error) {
	if locale != "en" {
		locale = "ru"
	}
	html, err := renderProposalReportHTML(input, output, locale)
	if err != nil {
		return nil, err
	}
	return htmlToPDF(html)
}

func renderProposalReportHTML(input model.ProposalInput, output model.ProposalOutput, locale string) (string, error) {
	scenario := model.NormalizeProposalScenario(input.ProposalScenario)
	labels := proposalPDFLabelsForScenario(locale, scenario)
	timeline := make([]proposalPDFTimelineRow, 0, len(output.Timeline))
	for _, ph := range output.Timeline {
		timeline = append(timeline, proposalPDFTimelineRow{
			Title:         ph.Title,
			DurationWeeks: ph.DurationWeeks,
			Description:   ph.Description,
		})
	}

	dateFmt := "02.01.2006"
	if locale == "en" {
		dateFmt = "Jan 2, 2006"
	}

	data := proposalPDFData{
		LangAttr:             locale,
		Labels:               labels,
		ShowPricing:          input.IncludePricing,
		ClientCompany:        input.ClientCompany,
		ClientContact:        input.ClientContact,
		SolutionName:         input.SolutionName,
		TimelineWeeks:        input.TimelineWeeks,
		GeneratedDate:        time.Now().Format(dateFmt),
		SenderCompany:        input.SenderCompany,
		SenderContact:        input.SenderContact,
		SenderPhone:          input.SenderPhone,
		SenderEmail:          input.SenderEmail,
		ProjectCostFormatted: formatRubPDF(input.ProjectCostRub, locale),
		Greeting:             template.HTML(strings.ReplaceAll(template.HTMLEscapeString(output.Greeting), "\n", "<br/>")),
		TaskUnderstanding:    template.HTML(strings.ReplaceAll(template.HTMLEscapeString(output.TaskUnderstanding), "\n", "<br/>")),
		ProposedSolution:     template.HTML(strings.ReplaceAll(template.HTMLEscapeString(output.ProposedSolution), "\n", "<br/>")),
		CostSummary:          template.HTML(strings.ReplaceAll(template.HTMLEscapeString(output.CostSummary), "\n", "<br/>")),
		PaymentTerms:         template.HTML(strings.ReplaceAll(template.HTMLEscapeString(output.PaymentTerms), "\n", "<br/>")),
		WhyUs:                template.HTML(strings.ReplaceAll(template.HTMLEscapeString(output.WhyUs), "\n", "<br/>")),
		NextStep:             template.HTML(strings.ReplaceAll(template.HTMLEscapeString(output.NextStep), "\n", "<br/>")),
		ScopeIncluded:        output.ScopeIncluded,
		ScopeExcluded:        output.ScopeExcluded,
		Timeline:             timeline,
	}

	tmpl, err := template.ParseFS(proposalReportHTML, "templates/proposal-report.html")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatRubPDF(amount float64, locale string) string {
	if locale == "en" {
		return fmt.Sprintf("₽ %s", formatIntPDF(int64(amount)))
	}
	return fmt.Sprintf("%s ₽", formatIntPDF(int64(amount)))
}

func formatIntPDF(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	if s != "" {
		parts = append([]string{s}, parts...)
	}
	return strings.Join(parts, " ")
}

func proposalPDFLabelsForScenario(locale, scenario string) proposalPDFLabels {
	base := proposalPDFLabelsForLocale(locale)
	switch scenario {
	case model.ProposalScenarioColdOutreach:
		if locale == "en" {
			base.Title = "Introduction letter"
			base.Greeting = "Opening"
			base.TaskUnderstanding = "Why we are reaching out"
			base.NextStep = "Suggested next step"
		} else {
			base.Title = "Письмо-знакомство"
			base.Greeting = "Обращение"
			base.TaskUnderstanding = "Почему мы обращаемся"
			base.NextStep = "Предлагаемый следующий шаг"
		}
	case model.ProposalScenarioProactiveOffer:
		if locale == "en" {
			base.Title = "Automation opportunity proposal"
			base.Greeting = "Value proposition"
			base.TaskUnderstanding = "Process analysis & ROI"
			base.NextStep = "Next step"
		} else {
			base.Title = "Предложение по автоматизации"
			base.Greeting = "Ценностное предложение"
			base.TaskUnderstanding = "Анализ процесса и ROI"
			base.NextStep = "Следующий шаг"
		}
	}
	return base
}

func proposalPDFLabelsForLocale(locale string) proposalPDFLabels {
	if locale == "en" {
		return proposalPDFLabels{
			Title: "Commercial Proposal", Client: "Client", Timeline: "Timeline", Weeks: "weeks",
			Date: "Date", Greeting: "Introduction", TaskUnderstanding: "Understanding your needs",
			ProposedSolution: "Proposed solution", Scope: "Scope of work", Included: "Included",
			Excluded: "Not included", TimelineSection: "Timeline & phases", Phase: "Phase",
			Duration: "Duration", Description: "Description", Cost: "Investment", TotalCost: "Project cost",
			Payment: "Payment terms", WhyUs: "Why us", NextStep: "Next step",
			Disclaimer: "Generated with Erman AI — review before sending to client.",
		}
	}
	return proposalPDFLabels{
		Title: "Коммерческое предложение", Client: "Клиент", Timeline: "Срок", Weeks: "нед.",
		Date: "Дата", Greeting: "Обращение", TaskUnderstanding: "Понимание задачи",
		ProposedSolution: "Предлагаемое решение", Scope: "Scope of Work", Included: "Входит в scope",
		Excluded: "Не входит", TimelineSection: "Сроки и этапы", Phase: "Этап",
		Duration: "Длительность", Description: "Описание", Cost: "Стоимость", TotalCost: "Стоимость проекта",
		Payment: "Условия оплаты", WhyUs: "Почему мы", NextStep: "Следующий шаг",
		Disclaimer: "Сгенерировано в Erman AI — проверьте перед отправкой клиенту.",
	}
}
