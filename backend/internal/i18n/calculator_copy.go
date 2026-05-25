package i18n

type CalculatorKPI struct {
	TimeKey, VolumeKey, ErrorsKey, HoursKey, PayrollKey, PaybackKey string
	Time, Volume, Errors, Hours, Payroll, Payback                   string
	ErrorsAfter, ErrorsChange                                 string
	MinSuffix, HourSuffix, MonthSuffix, MonthForecastSuffix   string
	ThroughputSuffix, ErrBeforeSuffix, RubSuffix              string
}

type CalculatorRecommendation struct {
	Automate       string
	Consider       string
	NotRecommended string
}

var calculatorKPI = map[string]CalculatorKPI{
	"en": {
		TimeKey: "processing_time", VolumeKey: "throughput", ErrorsKey: "errors",
		HoursKey: "monthly_hours", PayrollKey: "payroll_savings", PaybackKey: "payback",
		Time: "Order processing time", Volume: "Volume per employee",
		Errors: "Data entry errors", Hours: "Monthly hours spent",
		Payroll: "Payroll savings (RUB/mo)", Payback: "Project payback",
		ErrorsAfter: "0% (automated)", ErrorsChange: "Errors eliminated",
		MinSuffix: " min", HourSuffix: " h", MonthSuffix: " mo", MonthForecastSuffix: " mo (forecast)",
		ThroughputSuffix: "/day", ErrBeforeSuffix: "% with errors", RubSuffix: " RUB",
	},
	"ru": {
		TimeKey: "processing_time", VolumeKey: "throughput", ErrorsKey: "errors",
		HoursKey: "monthly_hours", PayrollKey: "payroll_savings", PaybackKey: "payback",
		Time: "Время обработки единицы", Volume: "Объём на сотрудника",
		Errors: "Ошибки при обработке", Hours: "Ежемесячно затраченное время",
		Payroll: "Экономия ФОТ (₽/мес)", Payback: "Окупаемость проекта",
		ErrorsAfter: "0% (автоматически)", ErrorsChange: "Ошибки устранены",
		MinSuffix: " мин", HourSuffix: " ч", MonthSuffix: " мес", MonthForecastSuffix: " мес (прогноз)",
		ThroughputSuffix: "/день", ErrBeforeSuffix: "% с ошибками", RubSuffix: " ₽",
	},
	"de": {
		TimeKey: "processing_time", VolumeKey: "throughput", ErrorsKey: "errors",
		HoursKey: "monthly_hours", PayrollKey: "payroll_savings", PaybackKey: "payback",
		Time: "Bearbeitungszeit pro Einheit", Volume: "Volumen pro Mitarbeiter",
		Errors: "Eingabefehler", Hours: "Monatlich aufgewendete Stunden",
		Payroll: "Personaleinsparung (₽/Monat)", Payback: "Amortisation des Projekts",
		ErrorsAfter: "0% (automatisiert)", ErrorsChange: "Fehler beseitigt",
		MinSuffix: " Min.", HourSuffix: " Std.", MonthSuffix: " Mon.", MonthForecastSuffix: " Mon. (Prognose)",
		ThroughputSuffix: "/Tag", ErrBeforeSuffix: "% mit Fehlern", RubSuffix: " ₽",
	},
	"es": {
		TimeKey: "processing_time", VolumeKey: "throughput", ErrorsKey: "errors",
		HoursKey: "monthly_hours", PayrollKey: "payroll_savings", PaybackKey: "payback",
		Time: "Tiempo de procesamiento por unidad", Volume: "Volumen por empleado",
		Errors: "Errores de entrada de datos", Hours: "Horas mensuales dedicadas",
		Payroll: "Ahorro en nómina (₽/mes)", Payback: "Recuperación del proyecto",
		ErrorsAfter: "0% (automatizado)", ErrorsChange: "Errores eliminados",
		MinSuffix: " min", HourSuffix: " h", MonthSuffix: " mes", MonthForecastSuffix: " mes (previsión)",
		ThroughputSuffix: "/día", ErrBeforeSuffix: "% con errores", RubSuffix: " ₽",
	},
	"fr": {
		TimeKey: "processing_time", VolumeKey: "throughput", ErrorsKey: "errors",
		HoursKey: "monthly_hours", PayrollKey: "payroll_savings", PaybackKey: "payback",
		Time: "Temps de traitement par unité", Volume: "Volume par employé",
		Errors: "Erreurs de saisie", Hours: "Heures mensuelles consacrées",
		Payroll: "Économies sur masse salariale (₽/mois)", Payback: "Retour sur investissement",
		ErrorsAfter: "0% (automatisé)", ErrorsChange: "Erreurs éliminées",
		MinSuffix: " min", HourSuffix: " h", MonthSuffix: " mois", MonthForecastSuffix: " mois (prévision)",
		ThroughputSuffix: "/jour", ErrBeforeSuffix: "% avec erreurs", RubSuffix: " ₽",
	},
	"zh": {
		TimeKey: "processing_time", VolumeKey: "throughput", ErrorsKey: "errors",
		HoursKey: "monthly_hours", PayrollKey: "payroll_savings", PaybackKey: "payback",
		Time: "单件处理时间", Volume: "人均处理量",
		Errors: "数据录入错误", Hours: "每月投入工时",
		Payroll: "人力成本节省（₽/月）", Payback: "项目回本周期",
		ErrorsAfter: "0%（已自动化）", ErrorsChange: "错误已消除",
		MinSuffix: " 分钟", HourSuffix: " 小时", MonthSuffix: " 月", MonthForecastSuffix: " 月（预测）",
		ThroughputSuffix: "/天", ErrBeforeSuffix: "% 存在错误", RubSuffix: " ₽",
	},
}

var calculatorRecommendations = map[string]CalculatorRecommendation{
	"en": {
		Automate:       "Automation is recommended — payback under 12 months and positive ROI",
		Consider:       "Consider automation — moderate payback or ROI",
		NotRecommended: "Automation is not recommended at current parameters",
	},
	"ru": {
		Automate:       "Рекомендуется автоматизировать — окупаемость менее 12 месяцев и положительный ROI",
		Consider:       "Рассмотрите автоматизацию — умеренная окупаемость или ROI",
		NotRecommended: "Автоматизация не рекомендуется при текущих параметрах",
	},
	"de": {
		Automate:       "Automatisierung empfohlen — Amortisation unter 12 Monaten und positiver ROI",
		Consider:       "Automatisierung erwägen — moderate Amortisation oder ROI",
		NotRecommended: "Automatisierung bei den aktuellen Parametern nicht empfohlen",
	},
	"es": {
		Automate:       "Se recomienda la automatización — recuperación en menos de 12 meses y ROI positivo",
		Consider:       "Considere la automatización — recuperación o ROI moderados",
		NotRecommended: "La automatización no se recomienda con los parámetros actuales",
	},
	"fr": {
		Automate:       "Automatisation recommandée — retour sur investissement en moins de 12 mois et ROI positif",
		Consider:       "Envisagez l'automatisation — retour sur investissement ou ROI modérés",
		NotRecommended: "L'automatisation n'est pas recommandée avec les paramètres actuels",
	},
	"zh": {
		Automate:       "建议自动化——回本周期少于12个月且ROI为正",
		Consider:       "可考虑自动化——回本周期或ROI适中",
		NotRecommended: "当前参数下不建议自动化",
	},
}

func CalculatorKPIFor(locale string) CalculatorKPI {
	if kpi, ok := calculatorKPI[NormalizeLocale(locale)]; ok {
		return kpi
	}
	return calculatorKPI[DefaultLocale]
}

func CalculatorRecommendationFor(locale string) CalculatorRecommendation {
	if rec, ok := calculatorRecommendations[NormalizeLocale(locale)]; ok {
		return rec
	}
	return calculatorRecommendations[DefaultLocale]
}
