package markdown

import (
	"strings"
	"testing"
)

func TestHasBufferedBlock(t *testing.T) {
	stream := NewStream(Options{Enabled: true})
	if stream.HasBufferedBlock() {
		t.Fatal("new stream should not have buffered table")
	}
	stream.Write("| A | B |\n| --- | --- |\n")
	if !stream.HasBufferedBlock() {
		t.Fatal("buffered table header+separator should be reported")
	}
	stream.Write("| 1 | 2 |\n")
	if !stream.HasBufferedBlock() {
		t.Fatal("table with one data row should still be buffered")
	}
	if got := stream.Write("\n"); !strings.Contains(got, "| A") {
		t.Fatalf("non-table blank line should flush table, got %q", got)
	}
	if stream.HasBufferedBlock() {
		t.Fatal("table should not remain buffered after flush")
	}
	stream.Write("| X | Y |\n| --- | --- |\n")
	if !stream.HasBufferedBlock() {
		t.Fatal("second table should be buffered")
	}
	if got := stream.Flush(); !strings.Contains(got, "| X") {
		t.Fatalf("Flush should emit buffered table, got %q", got)
	}
	if stream.HasBufferedBlock() {
		t.Fatal("Flush should clear buffered table state")
	}
}

func TestMidTableFlushSplitsWithDifferentWidths(t *testing.T) {
	// The live renderer may call Flush at a status-line boundary (e.g.  to
	// repaint the wait counter).  If that happens in the middle of a buffered
	// markdown table, the table is split into two separately-formatted chunks
	// with different column widths, producing misaligned output.  The renderer
	// guards against this, but this test documents the underlying Stream
	// behavior so a future change is caught.
	// Use the real session table text where later rows are wider than early rows:
	full := "| agent | baseline → candidate (-no-env, request_bytes)  | Δ total      | vs system    |\n| --- | --- | --- | --- |\n| auto | 8282+13980=22275 → 7580+14313=~21906 see below | -369 (-1.7%) | -5.0% system |\n| independent | same | -369 | -5.0% |\n| explore (no edit/write) | 8524+10236=18773 → 7822+10355=18190 | -583 (-3.1%) | -8.3% system |\n| plan | 8949+13062=22024 → 8247+13112=~21372 | -652 | -7.8% |\n\n"
	whole := Render(full, Options{Enabled: true, Width: 240})
	// Simulate a buggy live flush after the third row (auto) – exactly how the
	// live renderer would have split the table before the fix.
	s := NewStream(Options{Enabled: true, Width: 240})
	var buggy strings.Builder
	for _, line := range strings.Split(strings.TrimSuffix(full, "\n"), "\n") {
		buggy.WriteString(s.Write(line + "\n"))
		if strings.Contains(line, "| auto") {
			buggy.WriteString(s.Flush())
		}
	}
	buggy.WriteString(s.Flush())
	if buggy.String() == whole {
		t.Fatal("mid-table Flush should produce different widths than whole-table Render (test assumes buggy behavior)")
	}
	// Whole-table render pads the short 'same' cell to the wide first column
	if !strings.Contains(whole, "| independent             | same") {
		t.Fatalf("whole-table Render should pad second half, got %q", whole)
	}
	// Buggy split: the second half was formatted with only the short rows' widths
	if strings.Contains(buggy.String(), "| independent             | same") {
		t.Fatal("buggy mid-table flush unexpectedly produced padded second half")
	}
	if !strings.Contains(buggy.String(), "| independent | same") {
		t.Fatalf("buggy output should contain unpadded second half, got %q", buggy.String())
	}
}
