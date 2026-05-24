package service

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/go-pdf/fpdf"
)

func GenerateCalculatorPDF(input model.CalculatorInput, output model.CalculatorOutput, locale string) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)

	title := "Automation Calculator Report"
	if locale == "ru" {
		title = "Отчёт Automation Calculator"
	}
	pdf.CellFormat(0, 10, title, "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 11)
	pdf.Ln(4)
	pdf.Cell(0, 8, fmt.Sprintf("Process: %s", input.ProcessName))
	pdf.Ln(8)

	rows := []struct{ label, value string }{
		{"Net benefit /month", fmt.Sprintf("%s RUB", FormatMoney(output.NetBenefitMonthly))},
		{"Hours saved /month", fmt.Sprintf("%.0f h", output.HoursSavedMonth)},
		{"FTE freed", fmt.Sprintf("%.2f", output.FTE)},
		{"Payback", fmt.Sprintf("%.1f mo", output.PaybackMonths)},
		{"ROI (horizon)", fmt.Sprintf("%.1f%%", output.ROIHorizonPct)},
		{"TCO (horizon)", fmt.Sprintf("%s RUB", FormatMoney(output.TCOHorizon))},
		{"NPV", fmt.Sprintf("%s RUB", FormatMoney(output.NPV))},
		{"Recommendation", output.RecommendationText},
	}

	if locale == "ru" {
		rows = []struct{ label, value string }{
			{"Чистая выгода /мес", fmt.Sprintf("%s ₽", FormatMoney(output.NetBenefitMonthly))},
			{"Экономия часов /мес", fmt.Sprintf("%.0f ч", output.HoursSavedMonth)},
			{"Освобождено FTE", fmt.Sprintf("%.2f", output.FTE)},
			{"Окупаемость", fmt.Sprintf("%.1f мес", output.PaybackMonths)},
			{"ROI за горизонт", fmt.Sprintf("%.1f%%", output.ROIHorizonPct)},
			{"TCO за горизонт", fmt.Sprintf("%s ₽", FormatMoney(output.TCOHorizon))},
			{"NPV", fmt.Sprintf("%s ₽", FormatMoney(output.NPV))},
			{"Рекомендация", output.RecommendationText},
		}
	}

	for _, row := range rows {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.Cell(80, 7, row.label)
		pdf.SetFont("Helvetica", "", 10)
		pdf.Cell(0, 7, row.value)
		pdf.Ln(7)
	}

	if len(output.KPIRows) > 0 {
		pdf.Ln(4)
		kpiTitle := "KPI before / after"
		if locale == "ru" {
			kpiTitle = "KPI до / после"
		}
		pdf.SetFont("Helvetica", "B", 12)
		pdf.Cell(0, 8, kpiTitle)
		pdf.Ln(6)
		pdf.SetFont("Helvetica", "", 9)
		for _, k := range output.KPIRows {
			line := fmt.Sprintf("%s: %s -> %s (%s)", k.Label, k.Before, k.After, k.Change)
			pdf.MultiCell(0, 5, line, "", "L", false)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
