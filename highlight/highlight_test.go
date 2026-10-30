package highlight_test

import (
	"fmt"
	"github.com/ClarifiedLabs/mdcli/highlight"
	"regexp"
	"strings"
	"testing"
)

func ExampleNewWithTheme() {
	state := highlight.NewWithTheme("go", highlight.ThemeLight)
	colored := state.Line("func main() {}")
	fmt.Println(strings.Contains(colored, "\x1b["))
	// Output: true
}

func ExampleColorizeDiffWithTheme() {
	colored := highlight.ColorizeDiffWithTheme("main.go", "+func main() {}\n", highlight.ThemeDark)
	fmt.Println(strings.Contains(colored, "\x1b["))
	// Output: true
}

func TestPublicHighlightsPreserveBytes(t *testing.T) {
	strip := regexp.MustCompile(`\x1b\[[0-9;]*[mK]`)
	for _, theme := range []highlight.Theme{highlight.ThemeDark, highlight.ThemeLight} {
		var state *highlight.State = highlight.NewWithTheme("go", theme)
		line := "func main() {}"
		if got := strip.ReplaceAllString(state.Line(line), ""); got != line {
			t.Fatal(got)
		}
		var diff *highlight.DiffState = highlight.NewDiffWithTheme("go", theme)
		if got := strip.ReplaceAllString(diff.Line("+"+line), ""); got != "+"+line {
			t.Fatal(got)
		}
		for _, input := range []string{"--- a/main.go\n+++ b/main.go\n@@ -1 +1 @@\n-var x = 1\n+var x = 2", "+func main() {}\n"} {
			if got := strip.ReplaceAllString(highlight.ColorizeDiffWithTheme("main.go", input, theme), ""); got != input {
				t.Fatal(got)
			}
		}
	}
	var nilState *highlight.State
	var nilDiff *highlight.DiffState
	if nilState.Line("raw") != "raw" || nilDiff.Line("raw") != "raw" || highlight.New("not-a-language") != nil || highlight.New("go") == nil {
		t.Fatal("constructor/nil contract changed")
	}
}
