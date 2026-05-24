package model

import "time"

type CalculatorBudgetConfig struct {
	DevRateRub              float64 `json:"dev_rate_rub"`
	AuditBase               float64 `json:"audit_base"`
	Training                float64 `json:"training"`
	OmBase                  float64 `json:"om_base"`
	CapexMin                float64 `json:"capex_min"`
	IntegrationMultSimple   float64 `json:"integration_mult_simple"`
	IntegrationMultStandard float64 `json:"integration_mult_standard"`
	IntegrationMultComplex  float64 `json:"integration_mult_complex"`
	IntegrationCost2Roles   float64 `json:"integration_cost_2_roles"`
	IntegrationCost3Roles   float64 `json:"integration_cost_3_roles"`
	IntegrationCost4Plus    float64 `json:"integration_cost_4plus_roles"`
	OmUnits500              float64 `json:"om_units_500"`
	OmUnits2000             float64 `json:"om_units_2000"`
	OmSteps3                float64 `json:"om_steps_3"`
	OmSteps5                float64 `json:"om_steps_5"`
	OmHm150                 float64 `json:"om_hm_150"`
	DevHoursBase            float64 `json:"dev_hours_base"`
	DevHoursPerStep         float64 `json:"dev_hours_per_step"`
	DevHoursUnits500        float64 `json:"dev_hours_units_500"`
	DevHoursUnits1000       float64 `json:"dev_hours_units_1000"`
	DevHoursUnits2000       float64 `json:"dev_hours_units_2000"`
	DevHoursAutomationHigh  float64 `json:"dev_hours_automation_high"`
	DevHoursHm80            float64 `json:"dev_hours_hm_80"`
	DevHoursHm200           float64 `json:"dev_hours_hm_200"`
	AutomationPctThreshold  float64 `json:"automation_pct_threshold"`
	CapexRoundStep          float64 `json:"capex_round_step"`
	OmRoundStep             float64 `json:"om_round_step"`
	TierSimpleMax           int     `json:"tier_simple_max"`
	TierStandardMax         int     `json:"tier_standard_max"`
	UnitsThreshold500       float64 `json:"units_threshold_500"`
	UnitsThreshold1000      float64 `json:"units_threshold_1000"`
	UnitsThreshold2000      float64 `json:"units_threshold_2000"`
	HmThreshold50           float64 `json:"hm_threshold_50"`
	HmThreshold80           float64 `json:"hm_threshold_80"`
	HmThreshold150          float64 `json:"hm_threshold_150"`
	HmThreshold200          float64 `json:"hm_threshold_200"`
}

type CalculatorBudgetConfigRecord struct {
	Config    CalculatorBudgetConfig `json:"config"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func DefaultCalculatorBudgetConfig() CalculatorBudgetConfig {
	return CalculatorBudgetConfig{
		DevRateRub:              4000,
		AuditBase:               100_000,
		Training:                40_000,
		OmBase:                  8_000,
		CapexMin:                150_000,
		IntegrationMultSimple:   0.85,
		IntegrationMultStandard: 1,
		IntegrationMultComplex:  1.35,
		IntegrationCost2Roles:   40_000,
		IntegrationCost3Roles:   80_000,
		IntegrationCost4Plus:    150_000,
		OmUnits500:              4_000,
		OmUnits2000:             7_000,
		OmSteps3:                3_000,
		OmSteps5:                5_000,
		OmHm150:                 3_000,
		DevHoursBase:            20,
		DevHoursPerStep:         6,
		DevHoursUnits500:        15,
		DevHoursUnits1000:       30,
		DevHoursUnits2000:       40,
		DevHoursAutomationHigh:  15,
		DevHoursHm80:            10,
		DevHoursHm200:           20,
		AutomationPctThreshold:  70,
		CapexRoundStep:          50_000,
		OmRoundStep:             1_000,
		TierSimpleMax:           3,
		TierStandardMax:         6,
		UnitsThreshold500:       500,
		UnitsThreshold1000:      1000,
		UnitsThreshold2000:      2000,
		HmThreshold50:           50,
		HmThreshold80:           80,
		HmThreshold150:          150,
		HmThreshold200:          200,
	}
}
