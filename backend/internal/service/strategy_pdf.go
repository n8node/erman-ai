package service

import (
	"bytes"
	"embed"
	"html/template"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
)

//go:embed templates/strategy-report.html
var strategyReportHTML embed.FS

type strategyPDFLabels struct {
	Title              string
	Company            string
	ExecutiveSummary   string
	CurrentSituation   string
	Solutions          string
	Roadmap            string
	Next30             string
	ProcessAnalysis    string
	UseCases           string
	Metrics            string
	Risks              string
	Budget             string
	BudgetPhases       string
	ROI                string
	Process            string
	NetBenefit         string
	Payback            string
	NPV                string
	Recommendation     string
	Maturity           string
	Criterion          string
	Current            string
	Target             string
	Gap                string
	PriorityMatrix     string
	UseCase            string
	Impact             string
	Effort             string
	Score              string
	TechStack          string
	Layer              string
	Tool               string
	Role               string
	Status             string
	Stakeholders       string
	Involvement        string
	Diagrams           string
	TotalMonthly       string
	TotalNPV           string
	Disclaimer         string
}

type strategyPDFROIRow struct {
	Process        string
	NetBenefit     string
	Payback        string
	NPV            string
	Recommendation string
	BarPct         int
}

type strategyPDFData struct {
	LangAttr string
	Labels   strategyPDFLabels

	CompanyName         string
	BusinessDescription string
	ExecutiveSummary    string
	CurrentSituation    string

	Solutions []model.StrategySolution
	Roadmap   []model.StrategyRoadmapPhase
	Next30    []string

	ShowROI    bool
	ROIRows    []strategyPDFROIRow
	TotalBenefit string
	TotalNPV     string

	ShowMaturity bool
	Maturity     []model.StrategyMaturityRow
	ShowPriority bool
	Priority     []model.StrategyPriorityRow
	ShowBudgetPhases bool
	BudgetPhases     []model.StrategyBudgetPhase
	BudgetLines      []model.StrategyBudgetLine
	BudgetSummary    string

	Metrics []model.StrategyMetric
	Risks   []model.StrategyRisk

	ShowDiagrams bool
	Diagrams     []strategyPDFDiagramView
}

type strategyPDFDiagramView struct {
	Title   string
	Type    string
	Mermaid template.HTML
}

func GenerateStrategyPDF(input model.StrategyInput, output model.StrategyOutput, locale string) ([]byte, error) {
	if locale != "en" {
		locale = "ru"
	}
	html, hasDiagrams, err := renderStrategyReportHTML(input, output, locale)
	if err != nil {
		return nil, err
	}
	opts := pdfRenderOptions{}
	if hasDiagrams {
		opts = pdfRenderOptions{
			WaitForBodyAttr: "data-pdf-ready",
			WaitForValue:    "true",
			RenderTimeout:   90 * time.Second,
		}
	}
	return htmlToPDFWithOptions(html, opts)
}

func renderStrategyReportHTML(input model.StrategyInput, output model.StrategyOutput, locale string) (string, bool, error) {
	labels := strategyPDFLabelsForLocale(locale)
	data := strategyPDFData{
		LangAttr:            locale,
		Labels:              labels,
		CompanyName:         input.CompanyName,
		BusinessDescription: input.BusinessDescription,
		ExecutiveSummary:    output.ExecutiveSummary,
		CurrentSituation:    output.CurrentSituation,
		Solutions:           output.RecommendedSolutions,
		Roadmap:             output.Roadmap,
		Next30:              output.Next30Days,
		BudgetSummary:       output.BudgetOverview.Summary,
		BudgetLines:         output.BudgetOverview.Lines,
		BudgetPhases:        output.BudgetPhases,
		Metrics:             output.SuccessMetrics,
		Risks:               output.Risks,
		Maturity:            output.MaturityMatrix,
		Priority:            output.PriorityMatrix,
		ShowMaturity:        len(output.MaturityMatrix) > 0,
		ShowPriority:        len(output.PriorityMatrix) > 0,
		ShowBudgetPhases:    len(output.BudgetPhases) > 0,
		ShowDiagrams:        len(output.Diagrams) > 0,
	}
	for _, d := range output.Diagrams {
		if strings.TrimSpace(d.Mermaid) == "" {
			continue
		}
		data.Diagrams = append(data.Diagrams, strategyPDFDiagramView{
			Title:   d.Title,
			Type:    d.Type,
			Mermaid: template.HTML(d.Mermaid),
		})
	}
	data.ShowDiagrams = len(data.Diagrams) > 0

	if output.ROISummary != nil && len(output.ROISummary.Lines) > 0 {
		data.ShowROI = true
		data.TotalBenefit = formatLocaleMoney(output.ROISummary.TotalMonthlyBenefit, locale)
		data.TotalNPV = formatLocaleMoney(output.ROISummary.TotalNPV, locale)
		maxBenefit := 0.0
		for _, l := range output.ROISummary.Lines {
			if l.NetBenefitMonthly > maxBenefit {
				maxBenefit = l.NetBenefitMonthly
			}
		}
		for _, l := range output.ROISummary.Lines {
			pct := 40
			if maxBenefit > 0 {
				pct = int(l.NetBenefitMonthly / maxBenefit * 100)
				if pct < 8 {
					pct = 8
				}
			}
			monthlySuffix := "/мес"
			if locale == "en" {
				monthlySuffix = "/mo"
			}
			data.ROIRows = append(data.ROIRows, strategyPDFROIRow{
				Process:        l.ProcessName,
				NetBenefit:     formatLocaleMoney(l.NetBenefitMonthly, locale) + monthlySuffix,
				Payback:        formatPaybackPDF(l.PaybackMonths, "мес", locale),
				NPV:            formatLocaleMoney(l.NPV, locale),
				Recommendation: l.Recommendation,
				BarPct:         pct,
			})
		}
	}

	tmpl, err := template.ParseFS(strategyReportHTML, "templates/strategy-report.html")
	if err != nil {
		return "", false, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", false, err
	}
	return buf.String(), data.ShowDiagrams, nil
}

func strategyPDFLabelsForLocale(locale string) strategyPDFLabels {
	if locale == "en" {
		return strategyPDFLabels{
			Title: "AI Adoption Strategy", Company: "Company", ExecutiveSummary: "Executive summary",
			CurrentSituation: "Current situation", Solutions: "Recommended solutions", Roadmap: "Roadmap",
			Next30: "Next 30 days", ProcessAnalysis: "Process analysis", UseCases: "AI use cases",
			Metrics: "Success metrics", Risks: "Risks", Budget: "Budget overview", BudgetPhases: "Budget by phase",
			ROI: "ROI from calculator", Process: "Process", NetBenefit: "Net benefit", Payback: "Payback",
			NPV: "NPV", Recommendation: "Recommendation", Maturity: "Maturity matrix", Criterion: "Criterion",
			Current: "Current", Target: "Target", Gap: "Gap", PriorityMatrix: "Priority matrix",
			UseCase: "Use case", Impact: "Impact", Effort: "Effort", Score: "Score", TechStack: "Technology stack",
			Layer: "Layer", Tool: "Tool", Role: "Role", Status: "Status", Stakeholders: "Stakeholders",
			Involvement: "Involvement", Diagrams: "Diagrams", TotalMonthly: "Total monthly benefit",
			TotalNPV: "Total NPV", Disclaimer: "Generated by Erman AI — for internal planning purposes.",
		}
	}
	return strategyPDFLabels{
		Title: "Стратегия внедрения AI", Company: "Компания", ExecutiveSummary: "Executive summary",
		CurrentSituation: "Текущая ситуация", Solutions: "Рекомендуемые решения", Roadmap: "Roadmap",
		Next30: "Следующие 30 дней", ProcessAnalysis: "Анализ процессов", UseCases: "AI use cases",
		Metrics: "Метрики успеха", Risks: "Риски", Budget: "Бюджет", BudgetPhases: "Бюджет по фазам",
		ROI: "ROI из калькулятора", Process: "Процесс", NetBenefit: "Чистая выгода", Payback: "мес",
		NPV: "NPV", Recommendation: "Рекомендация", Maturity: "Матрица зрелости", Criterion: "Критерий",
		Current: "Текущий", Target: "Целевой", Gap: "Gap", PriorityMatrix: "Матрица приоритетов",
		UseCase: "Use case", Impact: "Impact", Effort: "Effort", Score: "Score", TechStack: "IT-стек",
		Layer: "Слой", Tool: "Инструмент", Role: "Роль", Status: "Статус", Stakeholders: "Стейкхолдеры",
		Involvement: "Вовлечённость", Diagrams: "Диаграммы", TotalMonthly: "Суммарная выгода в месяц",
		TotalNPV: "Суммарный NPV", Disclaimer: "Сгенерировано Erman AI — для внутреннего планирования.",
	}
}
