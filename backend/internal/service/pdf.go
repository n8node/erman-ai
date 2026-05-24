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
		{"Total current cost /month", fmt.Sprintf("%s RUB", FormatMoney(output.TotalCurrentCost))},
		{"Monthly savings", fmt.Sprintf("%s RUB", FormatMoney(output.MonthlySavings))},
		{"Net monthly savings", fmt.Sprintf("%s RUB", FormatMoney(output.NetMonthlySavings))},
		{"Payback months", fmt.Sprintf("%.1f", output.PaybackMonths)},
		{"Annual savings", fmt.Sprintf("%s RUB", FormatMoney(output.AnnualSavings))},
		{"Recommendation", output.RecommendationText},
	}

	if locale == "ru" {
		rows = []struct{ label, value string }{
			{"Текущие расходы /мес", fmt.Sprintf("%s ₽", FormatMoney(output.TotalCurrentCost))},
			{"Экономия /мес", fmt.Sprintf("%s ₽", FormatMoney(output.MonthlySavings))},
			{"Чистая экономия /мес", fmt.Sprintf("%s ₽", FormatMoney(output.NetMonthlySavings))},
			{"Окупаемость, мес", fmt.Sprintf("%.1f", output.PaybackMonths)},
			{"Годовая экономия", fmt.Sprintf("%s ₽", FormatMoney(output.AnnualSavings))},
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
