package i18n

import "strings"

const DefaultLocale = "ru"

var SupportedLocales = []string{"en", "ru", "de", "es", "fr", "zh"}

func NormalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	switch locale {
	case "en", "ru", "de", "es", "fr", "zh":
		return locale
	case "zh-cn", "zh-hans", "zh-hant", "zh-tw":
		return "zh"
	default:
		if strings.HasPrefix(locale, "zh") {
			return "zh"
		}
		return DefaultLocale
	}
}

func IsSupported(locale string) bool {
	n := NormalizeLocale(locale)
	for _, s := range SupportedLocales {
		if s == n {
			return true
		}
	}
	return false
}

func LanguageName(locale string) string {
	switch NormalizeLocale(locale) {
	case "en":
		return "English"
	case "de":
		return "German"
	case "es":
		return "Spanish"
	case "fr":
		return "French"
	case "zh":
		return "Chinese"
	default:
		return "Russian"
	}
}
