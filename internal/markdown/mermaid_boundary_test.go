package markdown

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ClarifiedLabs/mdcli/internal/mermaid"
)

func TestMermaidExactCapClosingFence(t *testing.T) {
	for _, cli := range []bool{false, true} {
		for _, marker := range []string{"```", "~~~"} {
			t.Run(fmt.Sprintf("cli=%v/%s", cli, marker), func(t *testing.T) {
				opts := Options{Enabled: true, MermaidLimitNotice: cli}
				base := "flowchart TD\nA\n%%"
				budget := mermaid.MaxSourceBytes
				if !cli {
					budget-- // Public streams count the final body newline.
				}
				input := marker + "mermaid\n" + base + strings.Repeat("x", budget-len(base)) + "\n"
				for _, ending := range []string{marker, marker + "\n"} {
					got := Render(input+ending, opts)
					if strings.Contains(got, marker) || !strings.Contains(got, "| A |") {
						t.Fatalf("Render ending %q rejected exact-cap diagram", ending)
					}
				}
				for split := 1; split <= len(marker); split++ {
					s := NewStream(opts)
					if got := s.Write(input); got != "" {
						t.Fatal("exact-cap body was not buffered")
					}
					if got := s.Write(marker[:split]); got != "" || !s.HasBufferedBlock() {
						t.Fatalf("split %d charged closing candidate as body", split)
					}
					got := s.Write(marker[split:]+"\n**after**\n") + s.Flush()
					if strings.Contains(got, marker) || !strings.Contains(got, "| A |") || !strings.HasSuffix(got, "after\n") {
						t.Fatalf("split %d lost diagram or following Markdown", split)
					}
				}
				s := NewStream(opts)
				s.Write(input)
				for _, c := range marker {
					if got := s.Write(string(c)); got != "" {
						t.Fatal("character-split closer emitted body")
					}
				}
				if got := s.Flush(); strings.Contains(got, marker) || !strings.Contains(got, "| A |") || !s.LineOpen() {
					t.Fatal("EOF closer did not render without a final newline")
				}
				if got := s.Write("\n**after**\n"); got != "\nafter\n" {
					t.Fatalf("EOF closer left following Markdown in code: %q", got)
				}
			})
		}
	}
}

func TestMermaidPendingFenceCandidatesBounded(t *testing.T) {
	for _, cli := range []bool{false, true} {
		for _, prefix := range []string{" ", "\t", "\u2003", "```"} {
			s := NewStream(Options{Enabled: true, MermaidLimitNotice: cli})
			s.Write("```mermaid\nflowchart TD\nA\n")
			for i := 0; i < 80; i++ {
				s.Write(strings.Repeat(prefix, 1024))
				if len(s.pending) > mermaid.MaxSourceBytes {
					t.Fatalf("cli=%v prefix=%q retained unbounded candidate", cli, prefix)
				}
			}
			if s.HasBufferedBlock() {
				t.Fatalf("cli=%v prefix=%q retained oversized fake candidate", cli, prefix)
			}
		}
	}
}

func TestMermaidCRLFSourceBoundary(t *testing.T) {
	for _, size := range []int{mermaid.MaxSourceBytes - 1, mermaid.MaxSourceBytes, mermaid.MaxSourceBytes + 1} {
		for _, partial := range []bool{false, true} {
			for _, cli := range []bool{false, true} {
				t.Run(fmt.Sprintf("bytes=%d/partial=%v/cli=%v", size, partial, cli), func(t *testing.T) {
					base := "flowchart TD\r\nA\r\n%%"
					// Include a final CR to exercise normalization of both complete
					// lines in Write and an incomplete line finalized by Flush.
					source := base + strings.Repeat("x", size-len(base)-1) + "\r"
					input := "```mermaid\r\n" + source
					if !partial {
						input += "\n```\r\n**after**\n"
					}
					s := NewStream(Options{Enabled: true, MermaidLimitNotice: cli})
					got := s.Write(input) + s.Flush()
					overflow := cli && size > mermaid.MaxSourceBytes
					if strings.Contains(got, "rendering limit exceeded") != overflow || strings.Contains(got, "```mermaid") != overflow {
						t.Fatal("raw CRLF budget or public normalized budget changed")
					}
					if !overflow && !strings.Contains(got, "| A |") {
						t.Fatal("within-budget CRLF diagram did not render")
					}
					if !partial && !strings.HasSuffix(got, "after\n") {
						t.Fatal("CRLF fallback lost following Markdown")
					}
				})
			}
		}
	}
}
