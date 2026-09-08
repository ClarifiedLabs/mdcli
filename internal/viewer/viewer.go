// Package viewer renders a Markdown document for the terminal, combining the
// markdown renderer with ASCII Mermaid diagrams. Fenced code blocks tagged
// "mermaid" are rendered as ASCII art; every other block is passed through the
// markdown renderer unchanged. Only the standard library is used.
package viewer

import (
	"errors"
	"strings"
	"unicode"

	"github.com/ClarifiedLabs/mdcli/internal/highlight"
	"github.com/ClarifiedLabs/mdcli/internal/markdown"
	"github.com/ClarifiedLabs/mdcli/internal/mermaid"
)

const (
	fenceTick  = "```"
	fenceTilde = "~~~"
)

// Options controls document rendering.
type Options struct {
	// ANSI applies terminal styling (bold, italic, links, code) when true.
	ANSI bool
	// Theme selects the syntax highlighting palette when ANSI is true. Zero value is dark.
	Theme highlight.Theme
	// Width enables word wrapping for paragraphs and list bodies when positive.
	Width int
}

// Render formats a complete Markdown document. Mermaid code fences are rendered
// as ASCII diagrams; if a diagram cannot be rendered it falls back to a plain
// indented code fence. Resource-limit fallbacks include an explanation. Mermaid
// content is stripped of terminal controls regardless of the ANSI option.
func Render(text string, opts Options) string {
	if text == "" {
		return ""
	}
	stream := markdown.NewStream(markdown.Options{
		Enabled:    true,
		ANSI:       opts.ANSI,
		ColorTheme: opts.Theme,
		Width:      opts.Width,
	})

	lines := strings.Split(text, "\n")
	// A trailing newline yields a final empty element that is not a real line.
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var out strings.Builder
	inCode := false
	var codeMarker string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Inside a non-mermaid code fence: pass lines straight through until the
		// fence closes, without scanning for nested fences.
		if inCode {
			out.WriteString(stream.Write(line + "\n"))
			if strings.HasPrefix(strings.TrimSpace(line), codeMarker) {
				inCode = false
				codeMarker = ""
			}
			continue
		}

		marker, ok := fenceMarker(line)
		if !ok {
			out.WriteString(stream.Write(line + "\n"))
			continue
		}

		info := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), marker))
		if !strings.EqualFold(info, "mermaid") {
			// A regular code fence: let the markdown stream handle it, but track
			// it here so its contents are not scanned for mermaid fences.
			inCode = true
			codeMarker = marker
			out.WriteString(stream.Write(line + "\n"))
			continue
		}

		i = renderMermaidBlock(&out, stream, lines, i, marker)
	}

	out.WriteString(stream.Flush())
	return out.String()
}

// renderMermaidBlock renders the mermaid fence that opens at lines[start]. It
// returns the index of the last consumed line (the loop advances past it).
func renderMermaidBlock(out *strings.Builder, stream *markdown.Stream, lines []string, start int, marker string) int {
	var body []string
	j := start + 1
	closed := false
	for ; j < len(lines); j++ {
		if strings.HasPrefix(strings.TrimSpace(lines[j]), marker) {
			closed = true
			break
		}
		body = append(body, lines[j])
	}

	// Flush any buffered markdown (e.g. a table) so it precedes the diagram.
	out.WriteString(stream.Flush())

	source := strings.Join(body, "\n")
	// Keep the raw byte budget meaningful even when filtering would shrink the
	// input. Oversized source goes straight to Render's preflight rejection.
	if len(source) <= mermaid.MaxSourceBytes {
		source = diagramText(strings.ReplaceAll(source, "\t", "    "))
	}
	rendered, err := mermaid.Render(source)
	if err == nil {
		// Also filter after rendering, in case label decoding introduces controls.
		rendered = diagramText(rendered)
		out.WriteString(rendered)
		if !strings.HasSuffix(rendered, "\n") {
			out.WriteByte('\n')
		}
	} else {
		if errors.Is(err, mermaid.ErrLimitExceeded) {
			out.WriteString("[" + diagramText(err.Error()) + "; showing source]\n")
		}
		// Write plain code directly: removing a control can expose a fence
		// delimiter that must not re-enter Markdown parsing or highlighting.
		end := j
		if closed {
			end++
		}
		for _, line := range lines[start:end] {
			line = strings.TrimRight(strings.ReplaceAll(line, "\t", "    "), " \r")
			out.WriteString("  " + diagramText(line) + "\n")
		}
	}

	if closed {
		return j
	}
	return len(lines)
}

// diagramText treats Mermaid content as data, never terminal commands. Newlines
// are layout; all other control characters are removed. Normalize source tabs to
// spaces before calling this so they remain meaningful whitespace.
func diagramText(text string) string {
	return strings.Map(func(r rune) rune {
		if r != '\n' && unicode.IsControl(r) {
			return -1
		}
		return r
	}, text)
}

// fenceMarker reports whether line opens a fenced code block and returns the
// fence marker. The logic mirrors the markdown renderer's fence detection.
func fenceMarker(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, fenceTick) {
		return fenceTick, true
	}
	if strings.HasPrefix(trimmed, fenceTilde) {
		return fenceTilde, true
	}
	return "", false
}
