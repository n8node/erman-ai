package service

import "unicode/utf8"

func truncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max])
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}
