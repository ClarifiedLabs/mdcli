package markdown

import (
	"strings"
	"testing"
)

func TestCLIMermaidPendingSourceBudgets(t *testing.T) {
	s := NewStream(Options{Enabled: true, MermaidLimitNotice: true})
	if got := s.Write("```mermaid\nflowchart TD\nA\n%%" + strings.Repeat("\t", 12000) + strings.Repeat("\x00", 20000)); got != "" || !s.HasBufferedBlock() {
		t.Fatal("valid incomplete raw32K/normalized48K source rejected")
	}
	got := s.Write("\n```\n") + s.Flush()
	if strings.Contains(got, "```") || !strings.Contains(got, "| A |") {
		t.Fatal("valid incomplete source did not render")
	}
	for _, body := range []string{strings.Repeat("\x00", 65536), strings.Repeat("\t", 16384)} {
		s := NewStream(Options{Enabled: true, MermaidLimitNotice: true})
		got := s.Write("```mermaid\nflowchart TD\nA\n%%" + body)
		if s.HasBufferedBlock() || !strings.Contains(got, "rendering limit exceeded") || strings.ContainsRune(got, '\x00') {
			t.Fatal("incomplete raw or normalized oversized source retained")
		}
		if got := s.Write("```\n**literal**\n```\n**after**\n"); got != "```\n  **literal**\n  ```\nafter\n" {
			t.Fatalf("overflow continuation closed source: %q", got)
		}
	}
}
