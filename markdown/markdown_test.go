package markdown_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ClarifiedLabs/mdcli/highlight"
	"github.com/ClarifiedLabs/mdcli/markdown"
)

func ExampleRender() {
	fmt.Println(markdown.Render("Use **bold** and `code`.", markdown.Options{Enabled: true}))
	// Output: Use bold and code.
}

func ExampleNewStream() {
	stream := markdown.NewStream(markdown.Options{Enabled: true})
	fmt.Print(stream.Write("Use **bo"))
	fmt.Print(stream.Write("ld**.\n"))
	fmt.Print(stream.Flush())
	// Output: Use bold.
}

func TestPublicStreamAndTheme(t *testing.T) {
	opts := markdown.Options{Enabled: true, ANSI: true, ColorTheme: highlight.ThemeLight, Width: 40}
	input := "```go\nfunc main() {}\n```\n```mermaid\nflowchart TD\nA --> B\n```"
	s := markdown.NewStream(opts)
	var out strings.Builder
	for _, c := range input {
		out.WriteString(s.Write(string(c)))
	}
	out.WriteString(s.Flush())
	if out.String() != markdown.Render(input, opts) || !s.LineOpen() || s.HasBufferedBlock() {
		t.Fatal("stream parity/state lost")
	}
	s.CloseLine()
	if !s.AtLineBoundary() || s.LineOpen() || s.Flush() != "" {
		t.Fatal("external newline state lost")
	}
	s = markdown.NewStream(opts)
	if s.Write("| A |\n| --- |\n") != "" || !s.HasBufferedBlock() || !s.AtLineBoundary() {
		t.Fatal("table must stay buffered at a safe line boundary")
	}
	code := markdown.Render("```go\nfunc main() {}\n```", opts)
	if !strings.Contains(code, highlight.NewWithTheme("go", highlight.ThemeLight).Line("func main() {}")) {
		t.Fatal("public theme not forwarded")
	}
	if markdown.DefaultWidth != 80 || markdown.HorizontalRule == "" {
		t.Fatal("missing constants")
	}
}

func TestMermaidBudgetIndependentConcurrentRenders(t *testing.T) {
	var source strings.Builder
	source.WriteString("```mermaid\nflowchart TD\n")
	for i := 0; i < 46; i++ {
		fmt.Fprintf(&source, "N%d --> N%d\n", i, i+1)
	}
	for i := 2; i < 47; i++ {
		fmt.Fprintf(&source, "N0 --> N%d\n", i)
	}
	source.WriteString("```\n") // 1035 virtual nodes.
	for _, budget := range []int{-1, 0, 1024, 1035, 8192, 9000} {
		t.Run(fmt.Sprint(budget), func(t *testing.T) {
			t.Parallel()
			opts := markdown.Options{Enabled: true, Width: 4096, MermaidMaxVirtualNodes: budget}
			got := markdown.Render(source.String(), opts)
			if strings.Contains(got, "```mermaid") != (budget == 1024) {
				t.Fatalf("budget %d not applied independently", budget)
			}
			s := markdown.NewStream(opts)
			if streamed := s.Write(source.String()) + s.Flush(); streamed != got {
				t.Fatal("budget lost in stream")
			}
		})
	}
}
