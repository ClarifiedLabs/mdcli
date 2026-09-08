package viewer

import (
	"fmt"
	"os"
	"path/filepath"
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

func TestRenderDefaultWidth(t *testing.T) {
	data, err := os.ReadFile("../mermaid/testdata/responsive/publishing.mmd")
	if err != nil {
		t.Fatal(err)
	}
	input := strings.Repeat("wrapped paragraph ", 20) + "\n\n```mermaid\n" + string(data) + "```\n"
	for _, ansi := range []bool{false, true} {
		want := Render(input, Options{Width: 80, ANSI: ansi})
		for _, width := range []int{0, -1, -100} {
			if got := Render(input, Options{Width: width, ANSI: ansi}); got != want {
				t.Fatalf("nonpositive width %d should use 80 columns:\n%s", width, got)
			}
		}
		if strings.Contains(want, "```mermaid") || !strings.Contains(want, "merges") {
			t.Fatalf("default-width diagram failed to adapt:\n%s", want)
		}
	}
}

func TestRenderResponsiveMermaidFixtures(t *testing.T) {
	for _, name := range []string{"roadmap", "publishing", "processing"} {
		data, err := os.ReadFile(filepath.Join("..", "mermaid", "testdata", "responsive", name+".mmd"))
		if err != nil {
			t.Fatal(err)
		}
		source := strings.TrimSuffix(string(data), "\n")
		for _, width := range []int{120, 100, 80, 60, 40} {
			t.Run(fmt.Sprintf("%s/width=%d", name, width), func(t *testing.T) {
				want, err := mermaid.Render(source, mermaid.Options{Width: width})
				if err != nil {
					t.Fatal(err)
				}
				if name == "roadmap" && width == 120 {
					natural, err := mermaid.Render(source, mermaid.Options{Width: 1000})
					if err != nil {
						t.Fatal(err)
					}
					if want != natural {
						t.Fatalf("fitting roadmap diagram changed from natural layout:\ngot:\n%s\nwant:\n%s", want, natural)
					}
				}
				for _, ansi := range []bool{false, true} {
					for _, closed := range []bool{false, true} {
						t.Run(fmt.Sprintf("ansi=%v/closed=%v", ansi, closed), func(t *testing.T) {
							input := "```mermaid\n" + source + "\n"
							if closed {
								input += "```\n"
							}
							got := Render(input, Options{Width: width, ANSI: ansi})
							if got != want {
								t.Fatalf("viewer differs from width-aware API:\ngot:\n%s\nwant:\n%s", got, want)
							}
							if strings.Contains(got, "```") {
								t.Fatalf("responsive diagram fell back to source:\n%s", got)
							}
							if width == 120 && (strings.Contains(got, "Nodes:") || strings.Contains(got, "Connections:")) {
								t.Fatalf("fixture should fit as a diagram at width 120:\n%s", got)
							}
						})
					}
				}
			})
		}
	}
}

func TestRenderResponsiveMermaidNormalizesSource(t *testing.T) {
	source := "flowchart LR\n\tA[\"日本語\t**literal**\x1b\a\x00\u009b\"] --> B[Done]"
	normalized := "flowchart LR\n    A[\"日本語    **literal**\"] --> B[Done]"
	want, err := mermaid.Render(normalized, mermaid.Options{Width: 40})
	if err != nil {
		t.Fatal(err)
	}
	for _, ansi := range []bool{false, true} {
		for _, ending := range []string{"\n```\n", ""} {
			t.Run(fmt.Sprintf("ansi=%v/closed=%v", ansi, ending != ""), func(t *testing.T) {
				got := Render("```mermaid\n"+source+ending, Options{Width: 40, ANSI: ansi})
				if got != want {
					t.Fatalf("width-aware rendering must normalize tabs and controls: %q; want %q", got, want)
				}
				for _, r := range got {
					if r != '\n' && unicode.IsControl(r) {
						t.Fatalf("terminal control %U survived: %q", r, got)
					}
				}
			})
		}
	}
}

func TestRenderResponsiveMermaidListDoesNotReparseMarkdown(t *testing.T) {
	// A wide fan-out cannot fit as a TD diagram even after label wrapping.
	var source strings.Builder
	source.WriteString("flowchart TD\nA[\"**literal**\"]\n")
	for i := 0; i < 16; i++ {
		fmt.Fprintf(&source, "A --> N%d[\"use `code` here\"]\n", i)
	}
	want, err := mermaid.Render(source.String(), mermaid.Options{Width: 40})
	if err != nil {
		t.Fatal(err)
	}
	for _, literal := range []string{"Nodes:", "Connections:", "**literal**", "`code`"} {
		if !strings.Contains(want, literal) {
			t.Fatalf("expected semantic list with literal %q:\n%s", literal, want)
		}
	}
	if !strings.Contains(strings.Join(strings.Fields(want), " "), "Flowchart shown as a list to fit the display.") {
		t.Fatalf("missing wrapped list fallback explanation:\n%s", want)
	}
	for _, ansi := range []bool{false, true} {
		t.Run(fmt.Sprintf("ansi=%v", ansi), func(t *testing.T) {
			opts := Options{Width: 40, ANSI: ansi}
			input := "```mermaid\n" + source.String() + "```\n**after**\n"
			got := Render(input, opts)
			// Only the following Markdown is styled; list labels remain literal.
			if expected := want + Render("**after**\n", opts); got != expected {
				t.Fatalf("semantic list or following Markdown was reparsed incorrectly:\ngot:\n%q\nwant:\n%q", got, expected)
			}
		})
	}
}

func TestRenderWidthLeavesOtherDiagramsAndCodeUnchanged(t *testing.T) {
	for _, source := range []string{
		"sequenceDiagram\nparticipant A as A very long participant name\nA->>A: A long message that must not be reflowed",
		"stateDiagram-v2\nFirstState --> SecondState: A long state transition label",
		"classDiagram\nclass ExampleClass {\n+String aLongMethodName()\n}",
	} {
		want, err := mermaid.Render(source, mermaid.Options{})
		if err != nil {
			t.Fatal(err)
		}
		for _, width := range []int{120, 100, 80, 60, 40} {
			for _, ansi := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/width=%d/ansi=%v", strings.SplitN(source, "\n", 2)[0], width, ansi), func(t *testing.T) {
					got := Render("```mermaid\n"+source+"\n```\n", Options{Width: width, ANSI: ansi})
					if got != want {
						t.Fatalf("non-flowchart layout changed:\ngot:\n%s\nwant:\n%s", got, want)
					}
				})
			}
		}
	}
	for _, ansi := range []bool{false, true} {
		input := "```go\nfmt.Println(\"" + strings.Repeat("wide code ", 20) + "\")\n```\n"
		want := Render(input, Options{ANSI: ansi})
		for _, width := range []int{120, 100, 80, 60, 40} {
			if got := Render(input, Options{Width: width, ANSI: ansi}); got != want {
				t.Fatalf("ordinary code changed at width %d, ANSI %v: %q; want %q", width, ansi, got, want)
			}
		}
	}
}

func TestRenderResponsiveMermaidLimitFallback(t *testing.T) {
	for _, tt := range []struct {
		name, source, resource string
	}{
		{"source", "flowchart TD\nA\n%%" + strings.Repeat("x", mermaid.MaxSourceBytes), "source bytes"},
		{"source before filtering", "flowchart TD\nA\n%%" + strings.Repeat("\x00", mermaid.MaxSourceBytes), "source bytes"},
		{"source after tab expansion", "flowchart TD\nA\n%%" + strings.Repeat("\t", mermaid.MaxSourceBytes/4), "source bytes"},
		{"edges", "flowchart TD\n" + strings.Repeat("A --> B\n", 1025), "edges"},
		{"natural flowchart canvas", "flowchart TD\nA[" + strings.Repeat("x", 4096) + "]", "canvas dimension"},
		// Other diagram kinds still enforce canvas limits rather than adapting.
		{"canvas", "sequenceDiagram\nparticipant A as " + strings.Repeat("x", 4096), "canvas dimension"},
	} {
		for _, ansi := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/ansi=%v", tt.name, ansi), func(t *testing.T) {
				opts := Options{Width: 40, ANSI: ansi}
				input := "```mermaid\n" + tt.source + "\n```\n**after**\n"
				got := Render(input, opts)
				if !strings.Contains(got, "rendering limit exceeded") || !strings.Contains(got, tt.resource) || !strings.Contains(got, "showing source") {
					t.Fatalf("missing width-aware limit explanation: %.200q", got)
				}
				if strings.Count(got, "```mermaid") != 1 || !strings.Contains(got, "  ```mermaid\n") || strings.ContainsRune(got, '\x00') || strings.ContainsRune(got, '\t') {
					t.Fatalf("missing or unsafe width-aware source fallback: %.200q", got)
				}
				if !strings.HasSuffix(got, Render("**after**\n", opts)) {
					t.Fatal("width-aware limit fallback corrupted following Markdown")
				}
			})
		}
	}
}

func TestRenderEmpty(t *testing.T) {
	if got := Render("", Options{}); got != "" {
		t.Fatalf("Render empty = %q, want empty", got)
	}
}
