// Package viewer adapts the shared Markdown stream for the terminal CLI.
package viewer

import (
	"strings"

	"github.com/ClarifiedLabs/mdcli/internal/highlight"
	"github.com/ClarifiedLabs/mdcli/internal/markdown"
)

// Options controls document rendering.
type Options struct {
	// ANSI applies terminal styling when true.
	ANSI bool
	// Theme selects syntax colors; zero is dark.
	Theme highlight.Theme
	// Width defaults to 80 columns when nonpositive.
	Width int
}

// Render formats a document with diagrams and resource-limit explanations.
// Nonempty documents end in a newline, preserving the CLI's line-oriented output.
func Render(text string, opts Options) string {
	if text == "" {
		return ""
	}
	if opts.Width <= 0 {
		opts.Width = markdown.DefaultWidth
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return markdown.Render(text, markdown.Options{
		Enabled: true, ANSI: opts.ANSI, ColorTheme: opts.Theme,
		Width: opts.Width, MermaidLimitNotice: true,
	})
}
