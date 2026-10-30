package mermaid

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestSelectableVirtualBudget(t *testing.T) {
	for _, dir := range []string{"TD", "BT", "LR", "RL"} {
		for _, width := range []int{8, 80, maxCanvasDimension} {
			for _, budget := range []int{-1, 0, 1023, 1024, 8193} {
				t.Run(fmt.Sprintf("%s/%d/%d", dir, width, budget), func(t *testing.T) {
					source := strings.Replace(parallelSpanSource(18, 64), "TD", dir, 1) // exactly 1024 virtual nodes.
					out, err := Render(source, Options{Width: width, MaxVirtualNodes: budget})
					if budget == 1023 {
						if !errors.Is(err, ErrLimitExceeded) || out != "" || !strings.Contains(err.Error(), "maximum 1023") {
							t.Fatalf("budget not applied: %v", err)
						}
					} else if err != nil || out == "" {
						t.Fatalf("legal budget: %v", err)
					}
				})
			}
		}
	}
	// Caller cannot raise the fixed maximum.
	if _, err := Render(parallelSpanSource(18, 512)+"N0 --> N2\n", Options{Width: 4096, MaxVirtualNodes: 8193}); !errors.Is(err, ErrLimitExceeded) || !strings.Contains(err.Error(), "maximum 8192") {
		t.Fatalf("hard maximum bypassed: %v", err)
	}
}

func TestDisplayWidth(t *testing.T) {
	for _, tt := range []struct {
		text  string
		width int
	}{{"界界", 4}, {"e\u0301 ", 2}, {"hello", 5}} {
		if got := DisplayWidth(tt.text); got != tt.width {
			t.Fatalf("%q: %d != %d", tt.text, got, tt.width)
		}
	}
}
