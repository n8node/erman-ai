package service

type pdfLabels struct {
	Disclaimer       string
	InputsTitle      string
	ReportTitle      string
	ReadOnlyNote     string
	Step1            string
	Step2            string
	CalculationMode  string
	ModeProcess      string
	ModeDirect       string
	ColOperation     string
	ColExecutor      string
	ColMinutes       string
	ColRate          string
	ColCost          string
	TotalPerUnit     string
	MinShort         string
	UnitsPerMonth    string
	AutomationPct    string
	HmDirect         string
	ErrorRateBefore  string
	ThroughputBefore string
	HmLabel          string
	HoursMonth       string
	Ch               string
	Om               string
	Capex            string
	Horizon          string
	Sm               string
	Utilization      string
	Discount         string
	NetBenefit       string
	Payback          string
	FTE              string
	FTEHint          string
	ROI              string
	KPITitle         string
	Before           string
	After            string
	Change           string
	Months           string
}

func pdfLabelsForLocale(locale string) pdfLabels {
	if locale == "en" {
		return pdfLabels{
			Disclaimer:       "Process automation ROI estimate",
			InputsTitle:      "Calculation inputs",
			ReportTitle:      "Calculation result",
			ReadOnlyNote:     "Read-only report. Calculation parameters cannot be edited.",
			Step1:            "Process",
			Step2:            "Finance",
			CalculationMode:  "Hours savings method",
			ModeProcess:      "Break down process",
			ModeDirect:       "Hm is known",
			ColOperation:     "Operation",
			ColExecutor:      "Role",
			ColMinutes:       "Min/unit",
			ColRate:          "RUB/min",
			ColCost:          "RUB/unit",
			TotalPerUnit:     "Total per unit",
			MinShort:         "min",
			UnitsPerMonth:    "Units per month",
			AutomationPct:    "Automation % (hours removed)",
			HmDirect:         "Hm — hours saved per month",
			ErrorRateBefore:  "Error rate before automation, %",
			ThroughputBefore: "Throughput per employee / day (before)",
			HmLabel:          "Hm — hours saved / month",
			HoursMonth:       "h/mo",
			Ch:               "Ch — fully loaded hourly cost, RUB",
			Om:               "Om — solution support / month, RUB",
			Capex:            "I₀ — implementation budget, RUB",
			Horizon:          "Evaluation horizon, months",
			Sm:               "Sm — other monthly benefit, RUB",
			Utilization:      "L — utilization factor (0–1)",
			Discount:         "r — discount rate (annual)",
			NetBenefit:       "Net benefit / month",
			Payback:          "Payback",
			FTE:              "FTE",
			FTEHint:          "full-time equivalents",
			ROI:              "ROI over horizon",
			KPITitle:         "KPI before / after",
			Before:           "Before",
			After:            "After",
			Change:           "Change",
			Months:           "mo",
		}
	}
	return pdfLabels{
		Disclaimer:       "Расчёт ROI автоматизации процесса",
		InputsTitle:      "Исходные параметры расчёта",
		ReportTitle:      "Результат расчёта",
		ReadOnlyNote:     "Отчёт только для просмотра. Параметры расчёта изменить нельзя.",
		Step1:            "Процесс",
		Step2:            "Финансы",
		CalculationMode:  "Способ расчёта экономии часов",
		ModeProcess:      "Разобрать процесс",
		ModeDirect:       "Hm уже известен",
		ColOperation:     "Операция",
		ColExecutor:      "Исполнитель",
		ColMinutes:       "Мин/ед",
		ColRate:          "₽/мин",
		ColCost:          "₽/ед",
		TotalPerUnit:     "Итого на 1 единицу",
		MinShort:         "мин",
		UnitsPerMonth:    "Единиц в месяц",
		AutomationPct:    "% автоматизации (снимаемых часов)",
		HmDirect:         "Hm — экономия часов в месяц",
		ErrorRateBefore:  "% ошибок до автоматизации",
		ThroughputBefore: "Объём на сотрудника / день (до)",
		HmLabel:          "Hm — экономия часов/мес",
		HoursMonth:       "ч/мес",
		Ch:               "Ch — полная стоимость часа, ₽",
		Om:               "Om — поддержка решения / мес, ₽",
		Capex:            "I₀ — бюджет внедрения, ₽",
		Horizon:          "Горизонт оценки, мес",
		Sm:               "Sm — прочая выгода / мес, ₽",
		Utilization:      "L — коэфф. загрузки (0–1)",
		Discount:         "r — ставка дисконтирования (год)",
		NetBenefit:       "Чистая выгода / мес",
		Payback:          "Окупаемость",
		FTE:              "FTE",
		FTEHint:          "эквивалент полных ставок",
		ROI:              "ROI за горизонт",
		KPITitle:         "KPI до / после",
		Before:           "До",
		After:            "После",
		Change:           "Изменение",
		Months:           "мес",
	}
}
