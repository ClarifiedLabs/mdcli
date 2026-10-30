package viewer

import (
	"strings"
	"testing"
)

func TestMermaidOverflowDoesNotReparseFilteredCloser(t *testing.T) {
	for _, control := range []string{"\x1b", "\x00", "\u009b", "\u202e"} {
		body := "flowchart TD\n%%" + strings.Repeat("x", 65536) + "\n"
		input := "```mermaid\n" + body + control + "```\n**literal**\x1b[31m\n```\n**after**\n"
		got := Render(input, Options{})
		if !strings.Contains(got, "rendering limit exceeded") || !strings.HasSuffix(got, "  ```\n  **literal**[31m\n  ```\nafter\n") || strings.ContainsRune(got, '\x1b') {
			t.Fatalf("filtered closer changed parsing or leaked controls (%q)", control)
		}
	}
}

func TestMermaidRawAndNormalizedSourceBudgets(t *testing.T) {
	for _, closed := range []bool{false, true} {
		for _, tt := range []struct {
			name, body string
			limited    bool
		}{
			{"raw32K-normalized48K", strings.Repeat("\t", 12000) + strings.Repeat("\x00", 20000), false},
			{"raw-over-limit-filtered-small", strings.Repeat("\x00", 65536), true},
			{"normalized-over-limit", strings.Repeat("\t", 16384), true},
		} {
			t.Run(tt.name+map[bool]string{false: "/open", true: "/closed"}[closed], func(t *testing.T) {
				input := "```mermaid\nflowchart TD\nA\n%%" + tt.body
				if closed {
					input += "\n```\n**after**\n```mermaid\nflowchart TD\nB\n```\n"
				}
				got := Render(input, Options{})
				if strings.Contains(got, "rendering limit exceeded") != tt.limited || strings.Contains(got, "```mermaid") != tt.limited || strings.ContainsRune(got, '\x00') {
					t.Fatal("raw and normalized budgets were not independently enforced")
				}
				if !tt.limited && !strings.Contains(got, "| A |") {
					t.Fatal("valid source was not rendered")
				}
				if closed && (!strings.Contains(got, "\nafter\n") || !strings.Contains(got, "| B |")) {
					t.Fatal("source budget leaked into subsequent Markdown or diagram")
				}
			})
		}
	}
}
