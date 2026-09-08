package mermaid

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// longSpanSource needs (nodes-1)*(nodes-2)/2 virtual nodes despite having
// fewer than twice as many edges as real nodes.
func longSpanSource(nodes int) string {
	var source strings.Builder
	source.WriteString("flowchart TD\n")
	for i := 0; i < nodes-1; i++ {
		fmt.Fprintf(&source, "N%d --> N%d\n", i, i+1)
	}
	for i := 2; i < nodes; i++ {
		fmt.Fprintf(&source, "N0 --> N%d\n", i)
	}
	return source.String()
}

func TestRenderVirtualBudget(t *testing.T) {
	// Dense, short graphs can use the full virtual budget within the canvas
	// budget. One additional two-rank link needs exactly one more virtual node.
	for _, dir := range []string{"TD", "BT", "LR", "RL"} {
		t.Run(dir, func(t *testing.T) {
			source := strings.Replace(parallelSpanSource(18, 512), "TD", dir, 1)
			out, err := Render(source, Options{Width: maxCanvasDimension})
			if dir == "BT" {
				// Reversing this layout needs more canvas area, independently
				// of its legal virtual-node count.
				if !errors.Is(err, ErrLimitExceeded) || out != "" || !strings.Contains(err.Error(), "canvas cells") {
					t.Fatalf("BT: want canvas limit, got %d bytes, %v", len(out), err)
				}
			} else if err != nil || out == "" {
				t.Fatalf("8192 virtual nodes: got %d bytes, error %v", len(out), err)
			}
			if out, err := Render(source+"N0 --> N2\n", Options{Width: maxCanvasDimension}); !errors.Is(err, ErrLimitExceeded) || out != "" || !strings.Contains(err.Error(), "virtual nodes") {
				t.Fatalf("8193 virtual nodes: want empty output and virtual limit, got %d bytes, %v", len(out), err)
			}
		})
	}
}

func TestLongSpanIndependentBudgets(t *testing.T) {
	for _, tt := range []struct {
		nodes    int
		resource string
	}{
		{47, ""},               // 1035 virtual nodes: above the sibling's former 1024 cap.
		{101, "canvas cells"},  // 4950 virtual nodes fit, but the canvas does not.
		{130, "virtual nodes"}, // 8256 virtual nodes: rejected before expansion.
	} {
		t.Run(fmt.Sprint(tt.nodes), func(t *testing.T) {
			out, err := Render(longSpanSource(tt.nodes), Options{Width: maxCanvasDimension})
			if tt.resource == "" {
				if err != nil || out == "" {
					t.Fatalf("got %d bytes, %v", len(out), err)
				}
			} else if !errors.Is(err, ErrLimitExceeded) || out != "" || !strings.Contains(err.Error(), tt.resource) {
				t.Fatalf("want empty output and %s limit, got %d bytes, %v", tt.resource, len(out), err)
			}
		})
	}
}

func parallelSpanSource(nodes, links int) string {
	var source strings.Builder
	source.WriteString("flowchart TD\n")
	for i := 0; i < nodes-1; i++ {
		fmt.Fprintf(&source, "N%d --> N%d\n", i, i+1)
	}
	source.WriteString(strings.Repeat(fmt.Sprintf("N0 --> N%d\n", nodes-1), links))
	return source.String()
}

func BenchmarkRenderParallelVirtualNodes(b *testing.B) {
	for _, links := range []int{64, 96, 192, 320, 512} {
		b.Run(fmt.Sprintf("18nodes_%dvirtuals", links*16), func(b *testing.B) {
			source := parallelSpanSource(18, links)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := Render(source, Options{Width: maxCanvasDimension}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkRenderVirtualNodes(b *testing.B) {
	for _, nodes := range []int{47, 56, 64} {
		virtuals := (nodes - 1) * (nodes - 2) / 2
		b.Run(fmt.Sprintf("%dnodes_%dvirtuals", nodes, virtuals), func(b *testing.B) {
			source := longSpanSource(nodes)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := Render(source, Options{Width: maxCanvasDimension}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
