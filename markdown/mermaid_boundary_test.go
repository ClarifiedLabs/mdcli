package markdown_test

import (
	"strings"
	"testing"

	"github.com/ClarifiedLabs/mdcli/markdown"
)

func TestPublicMermaidExactCapSplitCloser(t *testing.T) {
	base := "flowchart TD\nA\n%%"
	body := base + strings.Repeat("x", 65536-len(base)-1) + "\n"
	for _, marker := range []string{"```", "~~~"} {
		opts := markdown.Options{Enabled: true}
		input := marker + "mermaid\n" + body
		if got := markdown.Render(input+marker, opts); strings.Contains(got, marker) || !strings.Contains(got, "| A |") || strings.HasSuffix(got, "\n") {
			t.Fatalf("%s: public Render rejected exact-cap EOF closer", marker)
		}
		s := markdown.NewStream(opts)
		if got := s.Write(input); got != "" {
			t.Fatal("exact-cap body was not buffered")
		}
		for _, c := range marker {
			if got := s.Write(string(c)); got != "" || !s.HasBufferedBlock() {
				t.Fatalf("%s: character-split closer charged as body", marker)
			}
		}
		got := s.Write("\n**after**\n") + s.Flush()
		if strings.Contains(got, marker) || !strings.Contains(got, "| A |") || !strings.HasSuffix(got, "after\n") {
			t.Fatalf("%s: public stream lost diagram or following Markdown", marker)
		}
	}
}
