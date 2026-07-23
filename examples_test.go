package mdcli_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ClarifiedLabs/mdcli/internal/viewer"
)

// fallbacksAllowed lists the examples that deliberately contain diagrams the
// renderer cannot draw, and how many such fences each one has. Every other
// example must render all of its diagrams as ASCII art.
var fallbacksAllowed = map[string]int{
	"08-fallbacks.md": 5,
}

// maxExampleWidth is the terminal width the examples are laid out for. Keeping
// diagrams inside it means they do not wrap on a normal terminal.
const maxExampleWidth = 100

// TestExamplesRender renders every shipped example and checks that its
// diagrams are actually drawn rather than falling back to a code fence, and
// that nothing overflows the width the examples advertise.
func TestExamplesRender(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("examples", "*.md"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no examples/*.md files found")
	}
	for _, path := range files {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			out := viewer.Render(string(src), viewer.Options{Width: maxExampleWidth})
			if strings.TrimSpace(out) == "" {
				t.Fatal("rendered to nothing")
			}

			// a surviving ```mermaid fence means a diagram was not drawn
			got := strings.Count(out, "```mermaid")
			if want := fallbacksAllowed[name]; got != want {
				t.Errorf("%d diagrams fell back to a code fence, want %d", got, want)
			}

			for i, line := range strings.Split(strings.TrimSuffix(out, "\n"), "\n") {
				if w := len([]rune(line)); w > maxExampleWidth {
					t.Errorf("line %d is %d columns, over the %d the examples target:\n%s",
						i+1, w, maxExampleWidth, line)
				}
			}

			// Styling must decorate the text, never change it: stripping the
			// escape codes from a colored render has to give the plain one, so
			// piping to a file loses formatting but no content.
			colored := viewer.Render(string(src), viewer.Options{
				ANSI:  true,
				Width: maxExampleWidth,
			})
			if !strings.Contains(colored, "\x1b[") {
				t.Error("colored render carries no escape codes; is styling reaching it?")
			}
			if got := stripANSI(colored); got != out {
				t.Errorf("colored output differs from plain once stripped:\n%s",
					firstDiff(got, out))
			}
		})
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && (s[i] < '@' || s[i] > '~') {
				i++
			}
			if i < len(s) {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// firstDiff reports the first line where two renderings disagree.
func firstDiff(got, want string) string {
	g := strings.Split(got, "\n")
	w := strings.Split(want, "\n")
	for i := 0; i < len(g) && i < len(w); i++ {
		if g[i] != w[i] {
			return "line " + strconv.Itoa(i+1) + "\n stripped: " + strconv.Quote(g[i]) +
				"\n plain:    " + strconv.Quote(w[i])
		}
	}
	return "line counts differ: stripped " + strconv.Itoa(len(g)) +
		", plain " + strconv.Itoa(len(w))
}
