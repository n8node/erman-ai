package service

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	reMailto      = regexp.MustCompile(`(?i)mailto:([a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,})`)
	reEmailStrict = regexp.MustCompile(`(?i)\b([a-zA-Z0-9][a-zA-Z0-9._%+\-]{0,62}@[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)+)\b`)
	rePhoneRU     = regexp.MustCompile(`(?:\+7|8|7)[\s\-]?\(?\d{3}\)?[\s\-]?\d{3}[\s\-]?\d{2}[\s\-]?\d{2}|\b9\d{9}\b|\b8\d{10}\b`)
	reINNLabel    = regexp.MustCompile(`(?i)(?:инн|inn)[\s:№#-]*(\d{10}|\d{12})`)
	reOGRNLabel   = regexp.MustCompile(`(?i)(?:огрн(?:ип)?|ogrn)[\s:№#-]*(\d{13}|\d{15})`)
	reDigits10    = regexp.MustCompile(`\b(\d{10})\b`)
	reDigits12    = regexp.MustCompile(`\b(\d{12})\b`)
	reDigits13    = regexp.MustCompile(`\b(\d{13})\b`)
	reDigits15    = regexp.MustCompile(`\b(\d{15})\b`)
	reScriptBlock = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	reStyleBlock  = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	reHTMLTag     = regexp.MustCompile(`(?s)<[^>]+>`)
)

var blockedEmailSuffixes = []string{
	".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".ico", ".bmp", ".avif",
}

type labeledValue struct {
	Label string
	Value string
}

func stripHTMLForTextExtraction(html string) string {
	html = reScriptBlock.ReplaceAllString(html, " ")
	html = reStyleBlock.ReplaceAllString(html, " ")
	html = reHTMLTag.ReplaceAllString(html, " ")
	return html
}

func isValidContactEmail(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(email) > 254 {
		return false
	}
	if strings.Contains(email, "@2x") || strings.Contains(email, "scaled") {
		return false
	}
	for _, suf := range blockedEmailSuffixes {
		if strings.HasSuffix(email, suf) {
			return false
		}
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	local, domain := parts[0], parts[1]
	if local == "" || domain == "" || !strings.Contains(domain, ".") {
		return false
	}
	tld := domain[strings.LastIndex(domain, ".")+1:]
	if len(tld) < 2 || len(tld) > 24 {
		return false
	}
	for _, r := range tld {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}
	if strings.Contains(domain, "wp-content") || strings.Contains(domain, "uploads") {
		return false
	}
	return true
}

func extractContactEmails(html string) []string {
	var out []string
	for _, m := range reMailto.FindAllStringSubmatch(html, -1) {
		if isValidContactEmail(m[1]) {
			out = append(out, strings.ToLower(m[1]))
		}
	}
	text := stripHTMLForTextExtraction(html)
	for _, m := range reEmailStrict.FindAllStringSubmatch(text, -1) {
		if isValidContactEmail(m[1]) {
			out = append(out, strings.ToLower(m[1]))
		}
	}
	return uniqueStrings(out)
}

func normalizePhone(raw string) string {
	raw = strings.TrimSpace(raw)
	digits := make([]rune, 0, len(raw))
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}
	if len(digits) == 11 && digits[0] == '8' {
		digits[0] = '7'
	}
	if len(digits) == 10 && digits[0] == '9' {
		digits = append([]rune{'7'}, digits...)
	}
	if len(digits) != 11 || digits[0] != '7' {
		return ""
	}
	return "+7 " + string(digits[1:4]) + " " + string(digits[4:7]) + "-" + string(digits[7:9]) + "-" + string(digits[9:11])
}

func extractContactPhones(html string) []string {
	text := stripHTMLForTextExtraction(html)
	var out []string
	for _, m := range rePhoneRU.FindAllString(text, -1) {
		if norm := normalizePhone(m); norm != "" {
			out = append(out, norm)
		}
	}
	return uniqueStrings(out)
}

func validateINN10(inn string) bool {
	if len(inn) != 10 {
		return false
	}
	coeffs := []int{2, 4, 10, 3, 5, 9, 4, 6, 8}
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(inn[i]-'0') * coeffs[i]
	}
	check := sum % 11
	if check > 9 {
		check %= 10
	}
	return check == int(inn[9]-'0')
}

func validateINN12(inn string) bool {
	if len(inn) != 12 {
		return false
	}
	c1 := []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
	c2 := []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
	sum1, sum2 := 0, 0
	for i := 0; i < 10; i++ {
		sum1 += int(inn[i]-'0') * c1[i]
	}
	for i := 0; i < 11; i++ {
		sum2 += int(inn[i]-'0') * c2[i]
	}
	check1 := sum1 % 11
	if check1 > 9 {
		check1 %= 10
	}
	check2 := sum2 % 11
	if check2 > 9 {
		check2 %= 10
	}
	return check1 == int(inn[10]-'0') && check2 == int(inn[11]-'0')
}

func validateOGRN13(ogrn string) bool {
	if len(ogrn) != 13 {
		return false
	}
	n, err := strconv.ParseInt(ogrn[:12], 10, 64)
	if err != nil {
		return false
	}
	check := int(n % 11 % 10)
	return check == int(ogrn[12]-'0')
}

func validateOGRN15(ogrn string) bool {
	if len(ogrn) != 15 {
		return false
	}
	n, err := strconv.ParseInt(ogrn[:14], 10, 64)
	if err != nil {
		return false
	}
	check := int(n % 13 % 10)
	return check == int(ogrn[14]-'0')
}

func extractValidatedRequisites(html string) []labeledValue {
	text := stripHTMLForTextExtraction(html)
	seen := map[string]struct{}{}
	var out []labeledValue

	add := func(label, value string) {
		key := label + ":" + value
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, labeledValue{Label: label, Value: value})
	}

	for _, m := range reINNLabel.FindAllStringSubmatch(text, -1) {
		v := m[1]
		if len(v) == 10 && validateINN10(v) {
			add("ИНН", v)
		} else if len(v) == 12 && validateINN12(v) {
			add("ИНН", v)
		}
	}
	for _, m := range reOGRNLabel.FindAllStringSubmatch(text, -1) {
		v := m[1]
		if len(v) == 13 && validateOGRN13(v) {
			add("ОГРН", v)
		} else if len(v) == 15 && validateOGRN15(v) {
			add("ОГРНИП", v)
		}
	}

	for _, m := range reDigits10.FindAllStringSubmatch(text, -1) {
		if validateINN10(m[1]) {
			add("ИНН", m[1])
		}
	}
	for _, m := range reDigits12.FindAllStringSubmatch(text, -1) {
		if validateINN12(m[1]) {
			add("ИНН", m[1])
		}
	}
	for _, m := range reDigits13.FindAllStringSubmatch(text, -1) {
		if validateOGRN13(m[1]) {
			add("ОГРН", m[1])
		}
	}
	for _, m := range reDigits15.FindAllStringSubmatch(text, -1) {
		if validateOGRN15(m[1]) {
			add("ОГРНИП", m[1])
		}
	}

	return out
}

func formatLabeledValues(items []labeledValue) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Label+": "+item.Value)
	}
	return out
}

func mergeLabeledRequisites(items ...[]labeledValue) []labeledValue {
	seen := map[string]struct{}{}
	var out []labeledValue
	for _, group := range items {
		for _, item := range group {
			key := item.Label + ":" + item.Value
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, item)
		}
	}
	return out
}
