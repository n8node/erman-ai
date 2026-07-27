package service

import (
	"regexp"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

// SanitizeMermaid normalizes LLM-generated Mermaid before client/PDF rendering.
func SanitizeMermaid(source string) string {
	s := strings.TrimSpace(source)
	if s == "" {
		return s
	}

	s = strings.TrimPrefix(s, "```mermaid")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = htmlTagPattern.ReplaceAllString(s, "")

	lines := make([]string, 0, strings.Count(s, "\n")+1)
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, " \t")
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return ""
	}

	first := strings.ToLower(strings.TrimSpace(lines[0]))
	if !strings.HasPrefix(first, "flowchart") && !strings.HasPrefix(first, "graph") {
		lines = append([]string{"flowchart TD"}, lines...)
	}

	return strings.Join(lines, "\n")
}

func normalizeStrategyDiagrams(o *model.StrategyOutput) {
	if o == nil || len(o.Diagrams) == 0 {
		return
	}
	for i := range o.Diagrams {
		o.Diagrams[i].Mermaid = SanitizeMermaid(o.Diagrams[i].Mermaid)
	}
}
