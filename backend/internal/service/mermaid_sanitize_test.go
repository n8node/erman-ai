package service

import (
	"strings"
	"testing"
)

func TestSanitizeMermaidStripsFencesAndAddsFlowchart(t *testing.T) {
	got := SanitizeMermaid("```mermaid\nA[Start] --> B[End]\n```")
	want := "flowchart TD\nA[Start] --> B[End]"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSanitizeMermaidKeepsExistingHeader(t *testing.T) {
	src := "flowchart LR\nQ1[Пилот] --> Q2[Масштабирование]"
	if got := SanitizeMermaid(src); got != src {
		t.Fatalf("unexpected change: %q", got)
	}
}

func TestSanitizeMermaidRemovesHTML(t *testing.T) {
	got := SanitizeMermaid("flowchart TD\nA[<b>Start</b>] --> B[End]")
	if strings.Contains(got, "<b>") {
		t.Fatalf("html not removed: %q", got)
	}
}
