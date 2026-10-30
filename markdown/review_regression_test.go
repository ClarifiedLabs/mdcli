package markdown_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ClarifiedLabs/mdcli/markdown"
)

func TestMermaidFallbackUsesSourceCloser(t *testing.T) {
	for _, control := range []string{"\x1b", "\x00", "\u009b", "\u202e"} {
		for _, body := range []string{"unsupported\n", "flowchart TD\n%%" + strings.Repeat("x", 65536) + "\n"} {
			input := "```mermaid\n" + body + control + "```\n**literal**\x1b[31m\n```\n**after**\n"
			want := "  ```mermaid\n"
			for _, line := range strings.Split(strings.TrimSuffix(body, "\n"), "\n") {
				want += "  " + line + "\n"
			}
			want += "  ```\n  **literal**[31m\n  ```\nafter\n"
			for _, size := range []int{len(input), 1024} {
				t.Run(fmt.Sprintf("%q/body=%d/chunk=%d", control, len(body), size), func(t *testing.T) {
					s := markdown.NewStream(markdown.Options{Enabled: true})
					var out strings.Builder
					for pos := 0; pos < len(input); pos += size {
						out.WriteString(s.Write(input[pos:min(pos+size, len(input))]))
					}
					out.WriteString(s.Flush())
					if out.String() != want || s.HasBufferedBlock() || !s.AtLineBoundary() {
						t.Fatal("filtered delimiter changed parsing, source, or final stream state")
					}
				})
			}
		}
	}
}

func TestMermaidPartialOverflowCannotCloseSourceLine(t *testing.T) {
	s := markdown.NewStream(markdown.Options{Enabled: true})
	body := "%%" + strings.Repeat("x", 65536)
	got := s.Write("```mermaid\nflowchart TD\n" + body)
	got += s.Write("\x1b```\n**literal**\x1b[31m\n```\n**after**\n") + s.Flush()
	want := "  ```mermaid\n  flowchart TD\n  " + body + "```\n  **literal**[31m\n  ```\nafter\n"
	if got != want || !s.AtLineBoundary() || s.HasBufferedBlock() {
		t.Fatal("continuation of an overflowed source line closed the fence")
	}
}

func TestMermaidVirtualBudgetStateAndClass(t *testing.T) {
	for _, header := range []string{"stateDiagram-v2", "classDiagram"} {
		// A -> C skips B's rank, creating exactly one virtual node.
		source := "```mermaid\n" + header + "\nA --> B\nB --> C\nA --> C\n```\n"
		for _, budget := range []int{1, 2} {
			got := markdown.Render(source, markdown.Options{Enabled: true, MermaidMaxVirtualNodes: budget})
			if strings.Contains(got, "```mermaid") || !strings.Contains(got, "A") {
				t.Fatalf("%s: valid budget %d rejected", header, budget)
			}
		}
		// Add another rank and a second long edge: now the one-node cap fails.
		source = strings.Replace(source, "```\n", "C --> D\nA --> D\n```\n", 1)
		for _, budget := range []int{1, 3} {
			got := markdown.Render(source, markdown.Options{Enabled: true, MermaidMaxVirtualNodes: budget})
			if strings.Contains(got, "```mermaid") != (budget == 1) {
				t.Fatalf("%s: budget %d not forwarded", header, budget)
			}
		}
	}
}

func TestPublicMermaidSourceAccountingUnchanged(t *testing.T) {
	// Public streams retain expanded, unfiltered accounting, unlike the CLI.
	input := "```mermaid\nflowchart TD\nA\n%%" + strings.Repeat("\t", 12000) + strings.Repeat("\x00", 20000) + "\n```\n"
	got := markdown.Render(input, markdown.Options{Enabled: true})
	if !strings.Contains(got, "```mermaid") || strings.ContainsRune(got, '\x00') {
		t.Fatal("public stream budget or fallback sanitization changed")
	}
}
