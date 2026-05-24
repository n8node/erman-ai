package service

import (
	"os"
	"testing"

	"github.com/erman-ai/erman-ai/internal/model"
)

func TestGenerateCalculatorPDF_Russian(t *testing.T) {
	if os.Getenv("CHROME_PATH") == "" && chromeExecPath() == "" {
		t.Skip("chrome not available")
	}

	input := model.CalculatorInput{
		ProcessName:            "Обработка заявок",
		HmMode:                 "direct",
		HoursSavedMonth:        120,
		ErrorRateBeforePct:     5,
		ThroughputBeforePerDay: 40,
		HourlyCostLoaded:       850,
		Utilization:            0.8,
		OtherBenefitMonthly:    0,
		MonthlySolutionCost:    15000,
		Capex:                  350000,
		HorizonMonths:          12,
		DiscountRateAnnual:     0.2,
	}

	output := CalculateExcel(input, "ru")

	pdf, err := GenerateCalculatorPDF(input, output, "ru")
	if err != nil {
		t.Fatalf("GenerateCalculatorPDF: %v", err)
	}
	if len(pdf) < 1000 {
		t.Fatalf("pdf too small: %d bytes", len(pdf))
	}
	if pdf[0] != '%' || pdf[1] != 'P' || pdf[2] != 'D' || pdf[3] != 'F' {
		t.Fatalf("invalid pdf header")
	}
}

func TestRenderCalculatorReportHTML(t *testing.T) {
	input := model.CalculatorInput{
		ProcessName: "Тест",
		HmMode:      "process",
		ProcessSteps: []model.ProcessStep{
			{Operation: "Проверка", Executor: "Менеджер", MinutesPerUnit: 10, RateRubPerMin: 5},
		},
		UnitsPerMonth:          1000,
		AutomationPct:          70,
		ErrorRateBeforePct:     3,
		ThroughputBeforePerDay: 25,
		HourlyCostLoaded:       900,
		HorizonMonths:          12,
	}
	output := CalculateExcel(input, "ru")

	html, err := renderCalculatorReportHTML(input, output, "ru")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(html, "Обработка") && !contains(html, "Тест") {
		t.Fatalf("html missing process name")
	}
	if !contains(html, "Проверка") {
		t.Fatalf("html missing cyrillic step")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
