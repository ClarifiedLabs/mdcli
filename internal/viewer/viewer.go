// Package viewer renders a Markdown document for the terminal, combining the
// markdown renderer with ASCII Mermaid diagrams. Fenced code blocks tagged
// "mermaid" are rendered as ASCII art; every other block is passed through the
// markdown renderer unchanged. Only the standard library is used.
package viewer

import (
	"strings"

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
// indented code fence.
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

	rendered, err := mermaid.Render(strings.Join(body, "\n"))
	if err == nil {
		out.WriteString(rendered)
		if !strings.HasSuffix(rendered, "\n") {
			out.WriteByte('\n')
		}
	} else {
		// Unrenderable diagram: fall back to a plain code fence.
		out.WriteString(stream.Write(lines[start] + "\n"))
		for _, b := range body {
			out.WriteString(stream.Write(b + "\n"))
		}
		if closed {
			out.WriteString(stream.Write(lines[j] + "\n"))
		}
	}

	if closed {
		return j
	}
	return len(lines)
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
