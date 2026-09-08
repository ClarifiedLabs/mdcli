package viewer

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/ClarifiedLabs/mdcli/internal/mermaid"
)

func TestRenderMermaidFenceBecomesASCII(t *testing.T) {
	in := "# Title\n\n```mermaid\nflowchart TD\n  A[Start] --> B[End]\n```\n"
	got := Render(in, Options{})
	if strings.Contains(got, "```") {
		t.Fatalf("mermaid fence was not consumed:\n%s", got)
	}
	if !strings.Contains(got, "+-------+") || !strings.Contains(got, "| Start |") {
		t.Fatalf("expected ASCII flowchart boxes, got:\n%s", got)
	}
	if !strings.Contains(got, "# Title") {
		t.Fatalf("surrounding markdown lost:\n%s", got)
	}
}

func TestRenderNonMermaidFencePassesThrough(t *testing.T) {
	in := "```go\nfmt.Println(\"hi\")\n```\n"
	got := Render(in, Options{})
	if !strings.Contains(got, "```go") || !strings.Contains(got, "fmt.Println") {
		t.Fatalf("non-mermaid fence should pass through unchanged, got:\n%s", got)
	}
}

func TestRenderUnrenderableMermaidFallsBackToFence(t *testing.T) {
	in := "```mermaid\nnot valid mermaid at all\n```\n"
	got := Render(in, Options{})
	if !strings.Contains(got, "```mermaid") || !strings.Contains(got, "not valid mermaid at all") {
		t.Fatalf("expected fallback code fence, got:\n%s", got)
	}
}

func TestRenderUnclosedMermaidFenceAtEOF(t *testing.T) {
	in := "before\n\n```mermaid\nflowchart TD\n  A --> B\n"
	got := Render(in, Options{})
	if !strings.Contains(got, "before") {
		t.Fatalf("text before unclosed fence lost:\n%s", got)
	}
	if strings.Contains(got, "```") {
		t.Fatalf("unclosed mermaid fence should still render as a diagram:\n%s", got)
	}
}

func TestRenderMarkdownStillApplied(t *testing.T) {
	got := Render("Use **bold** and *italic*.", Options{})
	want := "Use bold and italic.\n"
	if got != want {
		t.Fatalf("Render = %q, want %q", got, want)
	}
}

func TestRenderMermaidStripsTerminalControls(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		fallback     bool
	}{
		{"label", "flowchart TD\nA[\"\x1b]0;label\a\"]", false},
		{"title", "---\ntitle: \x1b]0;title\a\n---\nflowchart TD\nA", false},
		{"C0 and C1", "flowchart TD\nA[\"a\x00\b\x7f\u009bb\"]", false},
		{"fallback", "invalid \x1b[31mred\x1b[0m\a\u009b", true},
	} {
		for _, ansi := range []bool{false, true} {
			for _, ending := range []string{"\n```\x1b[0m\n", ""} {
				t.Run(fmt.Sprintf("%s/ansi=%v/closed=%v", tt.name, ansi, ending != ""), func(t *testing.T) {
					got := Render("```mermaid\n"+tt.source+ending, Options{ANSI: ansi})
					for _, r := range got {
						if r != '\n' && unicode.IsControl(r) {
							t.Fatalf("terminal control %U survived: %q", r, got)
						}
					}
					if strings.Contains(got, "```mermaid") != tt.fallback {
						t.Fatalf("unexpected fallback state: %q", got)
					}
				})
			}
		}
	}
}

func TestRenderMermaidTabWhitespace(t *testing.T) {
	source := "sequenceDiagram\nparticipant\tAlice\nAlice->>Alice: 日本語\tmessage"
	got := Render("```mermaid\n"+source+"\n```\n", Options{})
	want := Render("```mermaid\n"+strings.ReplaceAll(source, "\t", "    ")+"\n```\n", Options{})
	if got != want || strings.Contains(got, "```") || !strings.Contains(got, "日本語    message") {
		t.Fatalf("tabs must remain whitespace and Unicode must survive: %q; want %q", got, want)
	}
}

func TestRenderMermaidFallbackDoesNotReparseSanitizedFences(t *testing.T) {
	input := "```mermaid\nunsupported\n\x1b```go\n**literal**\n```\n**after**\n"
	want := "  ```mermaid\n  unsupported\n  ```go\n  **literal**\n  ```\nafter\n"
	if got := Render(input, Options{}); got != want {
		t.Fatalf("sanitized source must remain plain code: %q; want %q", got, want)
	}
}

func TestRenderMermaidLimitFallback(t *testing.T) {
	for _, tt := range []struct {
		name, source, resource string
	}{
		{"source", "flowchart TD\nA\n%%" + strings.Repeat("x", mermaid.MaxSourceBytes), "source bytes"},
		{"source before filtering", "flowchart TD\nA\n%%" + strings.Repeat("\x00", mermaid.MaxSourceBytes), "source bytes"},
		{"source after tab expansion", "flowchart TD\nA\n%%" + strings.Repeat("\t", mermaid.MaxSourceBytes/4), "source bytes"},
		{"edges", "flowchart TD\n" + strings.Repeat("A --> B\n", 1025), "edges"},
		{"canvas", "flowchart TD\nA[" + strings.Repeat("x", 4096) + "]", "canvas dimension"},
	} {
		for _, closed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/closed=%v", tt.name, closed), func(t *testing.T) {
				input := "```mermaid\n" + tt.source
				if closed {
					input += "\n```\n**after**\n```mermaid\nflowchart TD\nC[Next]\n```\n"
				}
				got := Render(input, Options{})
				if !strings.Contains(got, "rendering limit exceeded") || !strings.Contains(got, tt.resource) || !strings.Contains(got, "showing source") {
					t.Fatalf("missing limit explanation: %.200q", got)
				}
				if strings.Count(got, "```mermaid") != 1 || !strings.Contains(got, "  flowchart TD\n") || strings.ContainsRune(got, '\x00') {
					t.Fatalf("missing or unsafe source fallback: %.200q", got)
				}
				if closed && (!strings.Contains(got, "\nafter\n") || !strings.Contains(got, "| Next |")) {
					t.Fatal("limit fallback corrupted subsequent Markdown/diagram")
				}
			})
		}
	}
}

func TestRenderMermaidSourceLimitBoundary(t *testing.T) {
	source := "flowchart TD\nA\n%%"
	source += strings.Repeat("x", mermaid.MaxSourceBytes-len(source))
	got := Render("```mermaid\n"+source+"\n```\n", Options{})
	if strings.Contains(got, "```") || !strings.Contains(got, "| A |") {
		t.Fatal("source exactly at the byte limit must render")
	}
}

func TestRenderEmpty(t *testing.T) {
	if got := Render("", Options{}); got != "" {
		t.Fatalf("Render empty = %q, want empty", got)
	}
}
