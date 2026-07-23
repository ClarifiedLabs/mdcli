package viewer

import (
	"strings"
	"testing"
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

func TestRenderEmpty(t *testing.T) {
	if got := Render("", Options{}); got != "" {
		t.Fatalf("Render empty = %q, want empty", got)
	}
}
