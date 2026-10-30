// Package markdown renders terminal-friendly Markdown and ASCII Mermaid diagrams.
// Rendering is opt-in with Options.Enabled; ANSI styling is separately opt-in.
package markdown

import (
	"github.com/ClarifiedLabs/mdcli/highlight"
	internal "github.com/ClarifiedLabs/mdcli/internal/markdown"
)

const (
	// DefaultWidth is the fallback diagram width, in terminal columns.
	DefaultWidth = internal.DefaultWidth
	// HorizontalRule is the fixed-width thematic break.
	HorizontalRule = internal.HorizontalRule
)

// Options controls Markdown rendering.
type Options struct {
	// Enabled leaves input byte-for-byte unchanged when false.
	Enabled bool
	// ANSI enables terminal styling. Callers decide whether output supports color.
	ANSI bool
	// ColorTheme selects syntax colors; zero is highlight.ThemeDark.
	ColorTheme highlight.Theme
	// Width wraps prose, lists and tables when positive. Flowcharts use DefaultWidth
	// when nonpositive; other diagrams and ordinary code keep their natural width.
	Width int
	// Prefix is prepended to nonempty output lines and counts toward diagram width.
	Prefix string
	// MermaidMaxVirtualNodes lowers the diagram layout budget. Nonpositive uses
	// 8192; positive values are capped at that hard maximum, per diagram.
	MermaidMaxVirtualNodes int
}

func (o Options) internal() internal.Options {
	return internal.Options{Enabled: o.Enabled, ANSI: o.ANSI, ColorTheme: o.ColorTheme,
		Width: o.Width, Prefix: o.Prefix, MermaidMaxVirtualNodes: o.MermaidMaxVirtualNodes}
}

// Render formats a complete block, preserving its final newline state.
func Render(text string, opts Options) string { return internal.Render(text, opts.internal()) }

// Stream incrementally renders Markdown. Tables and Mermaid fences are buffered
// for layout. A stream is stateful and must not be used concurrently.
type Stream struct{ stream *internal.Stream }

// NewStream creates a line-buffered renderer.
func NewStream(opts Options) *Stream { return &Stream{stream: internal.NewStream(opts.internal())} }

// Write consumes a delta and returns display-ready text (not an io.Writer count).
func (s *Stream) Write(text string) string { return s.stream.Write(text) }

// Flush finalizes pending lines and buffered blocks at EOF or a display boundary.
// Flushing an open Mermaid fence leaves subsequent source in plain-code mode.
func (s *Stream) Flush() string { return s.stream.Flush() }

// LineOpen reports whether the last emitted line lacks a newline.
func (s *Stream) LineOpen() bool { return s.stream.LineOpen() }

// AtLineBoundary reports whether external output can be inserted without a flush.
func (s *Stream) AtLineBoundary() bool { return s.stream.AtLineBoundary() }

// CloseLine records an external newline after an open display line.
func (s *Stream) CloseLine() { s.stream.CloseLine() }

// HasBufferedBlock reports a pending table or Mermaid layout. Avoid flushing it
// merely to repaint transient status output.
func (s *Stream) HasBufferedBlock() bool { return s.stream.HasBufferedBlock() }
