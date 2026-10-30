package mermaid

import (
	"errors"
	"strings"
	"testing"
)

func TestVirtualBudgetOnLaterAdaptiveCandidate(t *testing.T) {
	source := "flowchart TD\nA[" + strings.Repeat("long label ", 8) + "] --> B\nB --> C\nC --> D\nA --> C\nA --> D\n"
	lines, _ := preprocess(source)
	natural, err := parseFlowchart(lines)
	if err != nil {
		t.Fatal(err)
	}
	natural.maxVirtualNodes = 3
	naturalOut, err := natural.render()
	if err != nil {
		t.Fatal(err)
	}
	wrapped, err := parseFlowchart(lines)
	if err != nil {
		t.Fatal(err)
	}
	wrapped.maxVirtualNodes = 3
	for _, n := range wrapped.nodes {
		n.lines = wrapFlowText(strings.Join(n.lines, "\n"), 32)
	}
	want, err := wrapped.render()
	if err != nil {
		t.Fatal(err)
	}
	width := maxLineLen(strings.Split(want, "\n"))
	if width < minFlowchartWidth || maxLineLen(strings.Split(naturalOut, "\n")) <= width {
		t.Fatal("fixture must skip the natural candidate and fit a wrapped candidate")
	}
	got, err := Render(source, Options{Width: width, MaxVirtualNodes: 3})
	if err != nil || got != want {
		t.Fatalf("later candidate did not retain a legal budget: %v", err)
	}
	// An insufficient cap is a resource failure, not a reason to use a list.
	if out, err := Render(source, Options{Width: width, MaxVirtualNodes: 2}); out != "" || !errors.Is(err, ErrLimitExceeded) || !strings.Contains(err.Error(), "maximum 2") {
		t.Fatalf("adaptive fallback hid virtual budget: %v", err)
	}
}
