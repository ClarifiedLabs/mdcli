package viewer

import (
	"github.com/ClarifiedLabs/mdcli/internal/markdown"
	"strings"
	"testing"
)

func TestAdapterUsesSharedStream(t *testing.T) {
	for _, input := range []string{"", "plain", "plain\n", "plain\n\n", "~~~mermaid\nflowchart TD\nA --> B\n~~~", "```text\n```mermaid\nflowchart TD\nA --> B\n```\n"} {
		normalized := input
		if normalized != "" && !strings.HasSuffix(normalized, "\n") {
			normalized += "\n"
		}
		want := markdown.Render(normalized, markdown.Options{Enabled: true, Width: 80, MermaidLimitNotice: true})
		if got := Render(input, Options{}); got != want {
			t.Fatalf("adapter diverged: %q != %q", got, want)
		}
	}
}
