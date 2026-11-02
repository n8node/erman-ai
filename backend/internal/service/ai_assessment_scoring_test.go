package service

import (
	"github.com/erman-ai/erman-ai/internal/model"
	"math"
	"testing"
)

func TestCalculateAssessmentFinancial(t *testing.T) {
	in := model.AIAssessmentInput{MonthlyVolume: 600, MinutesPerUnit: 10, HourlyCost: 1200, Capex: 200000, MonthlyCost: 10000, HorizonMonths: 12}
	out := CalculateAssessmentFinancial(in, 50)
	if math.Abs(out.HoursSavedMonthly-50) > 0.001 {
		t.Fatalf("hours saved: got %v", out.HoursSavedMonthly)
	}
	if math.Abs(out.MonthlyBenefit-50000) > 0.001 {
		t.Fatalf("benefit: got %v", out.MonthlyBenefit)
	}
	if math.Abs(out.PaybackMonths-4) > 0.001 {
		t.Fatalf("payback: got %v", out.PaybackMonths)
	}
}

func TestAssessmentAutonomyCriticalRiskCapsMode(t *testing.T) {
	mode, _ := assessmentAutonomy(95, true)
	if mode != "assistant" {
		t.Fatalf("expected assistant, got %s", mode)
	}
}
